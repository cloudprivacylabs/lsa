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
	ErrInvalidSource      = errors.New("Invalid source")
	ErrInvalidSourceValue = errors.New("Invalid source value")
	ErrSourceMustBeString = errors.New("source term value must be a string")
	ErrMultipleValues     = errors.New("Multiple values/result columns found")
)

type ReshapeContext struct {
	*ls.Context
	parent *ReshapeContext
	// The expression language interpreter context
	symbols map[string]opencypher.Value

	// The source graph
	sourceGraph graph.Graph
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
func (reshaper *Reshaper) Reshape(lsContext *ls.Context, sourceGraph graph.Graph) error {
	ictx := reshaper.Start(lsContext, "")
	ctx := ReshapeContext{
		Context:     lsContext,
		sourceGraph: sourceGraph,
		symbols:     make(map[string]opencypher.Value),
	}

	_, err := reshaper.reshape(&ctx, ictx)
	return err
}

func (reshaper *Reshaper) reshape(context *ReshapeContext, ictx ls.IngestionContext) ([]graph.Node, error) {
	schemaNode := ictx.GetSchemaNode()
	// Nested context to keep symbols available to the nodes connected to this one
	context = context.nestedContext()
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
	// Get source values
	source, err := getSource(context, schemaNode)
	if err != nil {
		return nil, err
	}
	nativeValues, err := getNativeSourceValues(source)
	if err != nil {
		return nil, err
	}

	var ret []graph.Node
	switch {
	case schemaNode.GetLabels().Has(ls.AttributeTypeValue):
		ret, err = reshaper.value(context, ictx, nativeValues)
	case schemaNode.GetLabels().Has(ls.AttributeTypeObject):
		ret, err = reshaper.object(context, ictx, nativeValues)
	case schemaNode.GetLabels().Has(ls.AttributeTypeArray):
	case schemaNode.GetLabels().Has(ls.AttributeTypePolymorphic):
	default:
		return nil, ErrInvalidSchemaNodeType(schemaNode.GetLabels().Slice())
	}
	for _, x := range ret {
		x.RemoveProperty(ReshapeTerms.If)
		x.RemoveProperty(ReshapeTerms.Export)
		x.RemoveProperty(ReshapeTerms.Expressions)
		x.RemoveProperty(ReshapeTerms.ValueExpr)
		x.RemoveProperty(ReshapeTerms.IfEmpty)
		x.RemoveProperty(ReshapeTerms.JoinMethod)
		x.RemoveProperty(ReshapeTerms.JoinDelimiter)
	}
	return ret, nil
}

func (reshaper *Reshaper) object(context *ReshapeContext, ictx ls.IngestionContext, values []namedValue) ([]graph.Node, error) {
	schemaNode := ictx.GetSchemaNode()
	attributes := graph.TargetNodes(ls.SortEdgesItr(schemaNode.GetEdgesWithLabel(graph.OutgoingEdge, ls.ObjectAttributesTerm)))
	attributes = append(attributes, graph.TargetNodes(ls.SortEdgesItr(schemaNode.GetEdgesWithLabel(graph.OutgoingEdge, ls.ObjectAttributeListTerm)))...)

	ingestObject := func() (graph.Node, error) {
		_, objectNode, err := reshaper.Object(ictx)
		if err != nil {
			return nil, err
		}
		newCtx := ictx.NewLevel(objectNode)
		empty := true
		for _, schemaAttribute := range attributes {
			name := ls.AsPropertyValue(schemaAttribute.GetProperty(ls.AttributeNameTerm)).AsString()
			_, err := reshaper.reshape(context, newCtx.New(name, schemaAttribute))
			if err != nil {
				return nil, err
			}
			empty = false
		}

		if empty {
			ifEmpty := ls.AsPropertyValue(schemaNode.GetProperty(ReshapeTerms.IfEmpty))
			if ifEmpty != nil && ifEmpty.IsString() && ifEmpty.AsString() == "true" {
				objectNode.DetachAndRemove()
				return nil, nil
			}
		}

		return objectNode, err
	}

	if values == nil {
		v, err := ingestObject()
		if err != nil {
			return nil, err
		}
		return []graph.Node{v}, nil
	}

	ret := make([]graph.Node, 0)
	for _, val := range values {
		// Define the symbol
		context.symbols[val.name] = opencypher.ValueOf(val)
		v, err := ingestObject()
		if err != nil {
			return nil, err
		}
		if v != nil {
			ret = append(ret, v)
		}
	}

	return ret, nil
}

func (reshaper *Reshaper) value(context *ReshapeContext, ictx ls.IngestionContext, values []namedValue) ([]graph.Node, error) {
	schemaNode := ictx.GetSchemaNode()
	empty := true
	var nodeValue string
	if len(values) > 1 {
		empty = false
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
		vals := make([]interface{}, 0, len(values))
		for _, x := range values {
			vals = append(vals, x.value)
		}
		result, err := JoinValues(vals, joinMethod, joinDelimiter)
		if err != nil {
			return nil, err
		}
		nodeValue = result
	} else if len(values) == 1 {
		nodeValue = fmt.Sprint(values[0].value)
		empty = false
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
	_, node, err := reshaper.Value(ictx, nodeValue)
	if err != nil {
		return nil, err
	}
	return []graph.Node{node}, nil
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

type namedValue struct {
	name  string
	value interface{}
}

func getNativeSourceValues(in opencypher.Value) ([]namedValue, error) {
	// If we have source values, collect the native values
	if in.Value == nil {
		return nil, nil
	}

	if in.IsPrimitive() {
		return []namedValue{{value: in.Value}}, nil
	}

	if node, ok := in.Value.(graph.Node); ok {
		value, err := ls.GetNodeValue(node)
		return []namedValue{{value: value}}, err
	}

	if rs, ok := in.Value.(opencypher.ResultSet); ok {
		// A resultset. There must be one column only
		switch len(rs.Rows) {
		case 0: // No result
			return nil, nil
		case 1:
			if len(rs.Rows[0]) > 1 {
				return nil, ErrMultipleValues
			}
			for k, v := range rs.Rows[0] {
				val, err := getNativeSourceValues(v)
				if err != nil {
					return nil, err
				}
				return []namedValue{{name: k, value: val[0].value}}, nil
			}
		default:
			values := make([]namedValue, 0)
			for _, row := range rs.Rows {
				if len(row) > 1 {
					return nil, ErrMultipleValues
				}
				for k, v := range row {
					val, err := getNativeSourceValues(v)
					if err != nil {
						return nil, err
					}
					values = append(values, namedValue{name: k, value: val[0].value})
				}
			}
			return values, nil
		}
	}
	return nil, ErrInvalidSourceValue
}
