// Copyright 2021 Cloud Privacy Labs, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package transform

import (
	"errors"
	"fmt"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

type Reshaper struct {
	ls.Ingester
}

// ErrInvalidSchemaNodeType is returned if the schema node type cannot
// be projected (such as a reference, which cannot happen after
// compilation)
type ErrInvalidSchemaNodeType []string

func (e ErrInvalidSchemaNodeType) Error() string {
	return fmt.Sprintf("Invalid schema node type for reshaping: %v", []string(e))
}

var (
	ErrInvalidSource                = errors.New("Invalid source")
	ErrMultipleSourceNodesForObject = errors.New("Multiple source nodes specified for an object")
	ErrSourceMustBeString           = errors.New("source term value must be a string")
	ErrMultipleValues               = errors.New("Multiple values specified")
)

type ReshapeContext struct {
	parent *ReshapeContext
	// The expression language interpreter context
	symbols map[string]opencypher.Value
	// All schema nodes from the root to the current node
	schemaPath []graph.Node

	// The source graph
	sourceGraph graph.Graph
	targetGraph graph.Graph
}

func (p *ReshapeContext) CurrentSchemaNode() graph.Node {
	return p.schemaPath[len(p.schemaPath)-1]
}

func (p *ReshapeContext) setSymbols(ctx *opencypher.EvalContext) {
	if p.parent != nil {
		p.parent.setSymbols(ctx)
	}
	for k, v := range p.symbols {
		ctx.SetVar(k, v)
	}
}

func (p *ReshapeContext) getEvalContext() *opencypher.EvalContext {
	ctx := opencypher.NewEvalContext(p.sourceGraph)
	p.setSymbols(ctx)
	return ctx
}

func (p *ReshapeContext) nestedContext() *ReshapeContext {
	ret := *p
	ret.parent = p
	ret.symbols = make(map[string]opencypher.Value)
	return &ret
}

func (p *ReshapeContext) SetSymbolValue(name string, value opencypher.Value) {
	trc := p.parent
	for trc != nil {
		if _, ok := trc.symbols[name]; ok {
			trc.symbols[name] = value
			return
		}
		trc = trc.parent
	}
	p.symbols[name] = value
}

// Export the variables in the resultsets whose name match the given name
func (p *ReshapeContext) exportVar(name string, values []opencypher.Value) {
	for _, val := range values {
		resultSet, ok := val.Value.(opencypher.ResultSet)
		if !ok {
			continue
		}
		for _, row := range resultSet.Rows {
			rowValue, ok := row[name]
			if !ok {
				continue
			}

			p.SetSymbolValue(name, rowValue)
		}
	}
}

// Reshape builds a new graph in a shape that conforms to a target
// schema using the source graph
func (respaher *Reshaper) Reshape(sourceGraph, targetGraph graph.Graph) error {
	ctx := ReshapeContext{
		schemaPath:  []graph.Node{respaher.TargetSchema.GetSchemaRootNode()},
		sourceGraph: sourceGraph,
		targetGraph: targetGraph,
		symbols:     make(map[string]opencypher.Value),
	}
	return respaher.reshape(&ctx)
}

func (reshaper *Reshaper) reshape(context *ReshapeContext) (graph.Node, error) {
	schemaNode := context.CurrentSchemaNode()
	// If this is not a value node, create a sub-context
	if !schemaNode.GetLabels().Has(ls.AttributeTypeValue) {
		context = context.nestedContext()
	}
	// Evaluate expressions
	// Export values
	// Check conditionals
	expressionValues, err := reshaper.getExprs(context, schemaNode)
	if err != nil {
		return nil, err
	}
	for _, varname := range reshaper.getExportVars(schemaNode) {
		context.exportVar(varname, expressionValues)
	}
	// Check conditionals
	v, err := reshaper.checkConditionals(context, schemaNode)
	if !v || err != nil {
		return nil, err
	}

	switch {
	case schemaNode.GetLabels().Has(ls.AttributeTypeValue):
		return reshaper.value(context)
	case schemaNode.GetLabels().Has(ls.AttributeTypeObject):
		return reshaper.object(context)
	case schemaNode.GetLabels().Has(ls.AttributeTypeArray):
	case schemaNode.GetLabels().Has(ls.AttributeTypePolymorphic):
	}
	return nil, ErrInvalidSchemaNodeType(schemaNode.GetLabels().Slice())
}

func (reshaper *Reshaper) object(context *ReshapeContext) (graph.Node, error) {
	schemaNode := context.CurrentSchemaNode()
	attributes := graph.TargetNodes(ls.SortEdgesItr(schemaNode.GetEdgesWithLabel(graph.OutgoingEdge, ls.ObjectAttributesTerm)))
	attributes = append(attributes, graph.TargetNodes(ls.SortEdgesItr(schemaNode.GetEdgesWithLabel(graph.OutgoingEdge, ls.ObjectAttributeListTerm)))...)

	children := make([]graph.Node, 0)
	for _, a := range attributes {
		schemaAttribute := a.(graph.Node)
		context.schemaPath = append(context.schemaPath, schemaAttribute)
		newNode, err := reshaper.reshape(context)
		context.schemaPath = context.schemaPath[:len(context.schemaPath)-1]
		if err != nil {
			return nil, err
		}
		if newNode != nil {
			children = append(children, newNode)
		}
	}

	empty := len(children) == 0
	if empty {
		ifEmpty := ls.AsPropertyValue(schemaNode.GetProperty(ReshapeTerms.IfEmpty))
		if ifEmpty != nil && ifEmpty.IsString() && ifEmpty.AsString() == "true" {
			return nil, nil
		}
	}

	targetNode := context.targetGraph.NewNode([]string{ls.DocumentNodeTerm, ls.AttributeTypeObject}, nil)
	types := targetNode.GetLabels()
	types.Add(ls.FilterNonLayerTypes(schemaNode.GetLabels().Slice())...)
	targetNode.SetLabels(types)
	if reshaper.AddInstanceOfEdges {
		context.targetGraph.NewEdge(targetNode, schemaNode, ls.InstanceOfTerm, nil)
	}
	for _, c := range children {
		context.targetGraph.NewEdge(targetNode, c, ls.HasTerm, nil)
	}
	return targetNode, nil
}

func (reshaper *Reshaper) value(context *ReshapeContext) (graph.Node, error) {
	schemaNode := context.CurrentSchemaNode()

	source, err := getSource(context, schemaNode)
	if err != nil {
		return nil, err
	}
	empty := true
	var nodeValue interface{}
	getNodeValue := func(in interface{}) (interface{}, error) {
		if node, ok := in.(graph.Node); ok {
			return ls.GetNodeValue(node)
		}
		return in, nil
	}
	if source.Value != nil {
		if source.IsPrimitive() {
			nodeValue = source.Value
			empty = false
		} else if node, ok := source.Value.(graph.Node); ok {
			val, err := ls.GetNodeValue(node)
			if err != nil {
				return nil, err
			}
			nodeValue = val
			empty = false
		} else if rs, ok := source.Value.(opencypher.ResultSet); ok {
			switch len(rs.Rows) {
			case 0:
			case 1:
				if len(rs.Rows[0]) == 1 {
					for _, v := range rs.Rows[0] {
						nodeValue, err = getNodeValue(v.Value)
						if err != nil {
							return nil, err
						}
					}
					empty = false
				} else if len(rs.Rows[0]) > 1 {
					return nil, ErrMultipleValues
				}
			default:
				joinMethod := "join"
				prop := ls.AsPropertyValue(schemaNode.GetProperty(ReshapeTerms.JoinMethod))
				if prop != nil && prop.IsString() {
					joinMethod = prop.AsString()
				}
				joinDelimiter := " "
				prop = ls.AsPropertyValue(schemaNode.GetProperty(ReshapeTerms.JoinDelimiter))
				if prop != nil && prop.IsString() {
					joinDelimiter = prop.AsString()
				}

				values := make([]opencypher.Value, 0)
				for _, row := range rs.Rows {
					if len(row) > 1 {
						return nil, ErrMultipleValues
					}
					for _, v := range row {
						value, err := getNodeValue(v.Value)
						if err != nil {
							return nil, err
						}
						values = append(values, opencypher.ValueOf(value))
					}
				}
				result, err := JoinValues(values, joinMethod, joinDelimiter)
				if err != nil {
					return nil, err
				}
				nodeValue = result
				empty = false
			}
		}
	}

	if empty {
		ifEmpty := ls.AsPropertyValue(schemaNode.GetProperty(ReshapeTerms.IfEmpty))
		if ifEmpty != nil && ifEmpty.IsString() && ifEmpty.AsString() == "true" {
			empty = false
		}
	}
	if empty {
		return nil, nil
	}
	targetNode := context.targetGraph.NewNode([]string{ls.DocumentNodeTerm, ls.AttributeTypeValue}, nil)
	if reshaper.AddInstanceOfEdges {
		context.targetGraph.NewEdge(targetNode, schemaNode, ls.InstanceOfTerm, nil)
	}
	types := targetNode.GetLabels()
	types.Add(ls.FilterNonLayerTypes(schemaNode.GetLabels().Slice())...)
	targetNode.SetLabels(types)
	if nodeValue != nil {
		ls.SetNodeValue(targetNode, nodeValue)
	}
	return targetNode, nil
}

func getSource(context *ReshapeContext, schemaNode graph.Node) (opencypher.Value, error) {
	data, _ := schemaNode.GetProperty("$compiled_" + ReshapeTerms.ValueExpr)
	expr, ok := data.(opencypher.Evaluatable)
	if !ok {
		return opencypher.Value{}, nil
	}
	value, err := expr.Evaluate(context.getEvalContext())
	if err != nil {
		return opencypher.Value{}, err
	}
	return value, nil
}

func (reshaper *Reshaper) getExprs(context *ReshapeContext, schemaNode graph.Node) ([]opencypher.Value, error) {
	data, _ := schemaNode.GetProperty("$compiled_" + ReshapeTerms.Expressions)
	expr, ok := data.([]opencypher.Evaluatable)
	if !ok {
		return nil, nil
	}
	ret := make([]opencypher.Value, 0, len(expr))
	for _, x := range expr {
		evctx := context.getEvalContext()
		v, err := x.Evaluate(evctx)
		if err != nil {
			return nil, err
		}
		ret = append(ret, v)
	}
	return ret, nil
}

func (reshaper *Reshaper) getExportVars(schemaNode graph.Node) []string {
	data, _ := schemaNode.GetProperty("$compiled_" + ReshapeTerms.Export)
	slice, ok := data.([]string)
	if !ok {
		return nil
	}
	return slice
}

func (reshaper *Reshaper) checkConditionals(context *ReshapeContext, schemaNode graph.Node) (bool, error) {
	data, _ := schemaNode.GetProperty("$compiled_" + ReshapeTerms.If)
	slice, ok := data.([]opencypher.Evaluatable)
	if !ok {
		return true, nil
	}
	result := true
	for _, x := range slice {
		r, err := x.Evaluate(context.getEvalContext())
		if err != nil {
			return false, err
		}
		result, ok = r.AsBool()
		if !ok {
			return false, nil
		}
		if result == false {
			return false, nil
		}
	}
	return result, nil
}
