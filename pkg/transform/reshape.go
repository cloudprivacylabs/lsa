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
	"fmt"
	"strings"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher"
	"github.com/cloudprivacylabs/opencypher/graph"
)

const TRANSFORM = ls.LS + "transform/"

type Reshaper struct {
	TargetSchema *ls.Layer
	Builder      ls.GraphBuilder
	Script       *TransformScript
}

type path[X any] struct {
	items []X
}

func (p *path[X]) push(x X) *path[X] {
	p.items = append(p.items, x)
	return p
}

func (p *path[X]) pop() {
	p.items = p.items[:len(p.items)-1]
}

func (p *path[X]) last() X {
	return p.items[len(p.items)-1]
}

func (p *path[X]) empty() bool { return len(p.items) == 0 }

type reshapeContext struct {
	*ls.Context
	parent      *reshapeContext
	symbols     map[string]opencypher.Value
	sourceGraph graph.Graph

	schemaPath    path[graph.Node]
	generatedPath path[*txDocNode]
	mapContext    path[graph.Node]
}

func (ctx *reshapeContext) sub() *reshapeContext {
	ret := *ctx
	ret.parent = ctx
	ret.symbols = make(map[string]opencypher.Value)
	return &ret
}

// Export the named variables in the resultset
func (ctx *reshapeContext) exportVars(row map[string]opencypher.Value) {
	for varName, val := range row {
		if opencypher.IsNamedResult(varName) {
			ctx.setSymbolValue(varName, val)
		}
	}
}

// Export the named result columns
func (ctx *reshapeContext) exportResults(rs opencypher.ResultSet) error {
	if len(rs.Rows) == 0 {
		return nil
	}
	if len(rs.Rows) > 1 {
		return fmt.Errorf("Multiple results in the resultset")
	}
	for _, row := range rs.Rows {
		for varName, val := range row {
			if opencypher.IsNamedResult(varName) {
				ctx.setSymbolValue(varName, val)
			}
		}
	}
	return nil
}

func (ctx *reshapeContext) setSymbolValue(name string, value opencypher.Value) {
	// If the symbol is already defined, set that. Otherwise, define the symbol at this level
	trc := ctx.parent
	for trc != nil {
		if _, ok := trc.symbols[name]; ok {
			trc.symbols[name] = value
			return
		}
		trc = trc.parent
	}
	ctx.symbols[name] = value
}

func (ctx *reshapeContext) fillEvalContext(ectx *opencypher.EvalContext) {
	if ctx.parent != nil {
		ctx.parent.fillEvalContext(ectx)
	}
	for k, v := range ctx.symbols {
		ectx.SetVar(k, v)
	}
}

func (ctx *reshapeContext) getEvalContext() *opencypher.EvalContext {
	ectx := opencypher.NewEvalContext(ctx.sourceGraph)
	ctx.fillEvalContext(ectx)
	return ectx
}

func (reshaper Reshaper) fillNodes(roots []*txDocNode) error {
	for i := range roots {
		if err := reshaper.fillNode(roots[i]); err != nil {
			return err
		}
	}
	return nil
}

func (reshaper Reshaper) fillNode(node *txDocNode) error {
	if node.schemaNode != nil {
		if node.schemaNode.GetLabels().Has(ls.AttributeTypeArray) {
			node.typeTerm = ls.AttributeTypeArray
		} else if node.schemaNode.GetLabels().Has(ls.AttributeTypeObject) {
			node.typeTerm = ls.AttributeTypeObject
		} else {
			node.typeTerm = ls.AttributeTypeValue
		}
	}
	if node.sourceNode != nil {
		val, err := ls.GetNodeValue(node.sourceNode)
		if err != nil {
			return err
		}
		node.value = val
		node.sourceNode.ForEachProperty(func(key string, value interface{}) bool {
			if strings.HasPrefix(key, ls.LS) {
				return true
			}
			p, ok := value.(*ls.PropertyValue)
			if !ok {
				return true
			}
			node.properties[key] = p
			return true
		})
	}
	nodes := make([]*txDocNode, 0, len(node.children))
	for i := range node.children {
		nodes = append(nodes, node.children[i].(*txDocNode))
	}
	return reshaper.fillNodes(nodes)
}

func (reshaper Reshaper) Reshape(ctx *ls.Context, sourceGraph graph.Graph) error {
	c := reshapeContext{
		Context:     ctx,
		symbols:     make(map[string]opencypher.Value),
		sourceGraph: sourceGraph,
	}
	c.schemaPath.push(reshaper.TargetSchema.GetSchemaRootNode())
	roots, err := reshaper.reshapeNode(&c)
	if err != nil {
		return err
	}
	if err := reshaper.fillNodes(roots); err != nil {
		return err
	}
	for _, root := range roots {
		_, err := ls.Ingest(reshaper.Builder, root)
		if err != nil {
			return err
		}
	}
	return nil
}

// Returns true if operation produced any output
func (reshaper Reshaper) reshapeNode(ctx *reshapeContext) ([]*txDocNode, error) {
	// last element of schemaPath is the schema node to reshape
	ctx = ctx.sub()
	schemaNode := ctx.schemaPath.last()
	schemaNodeID := ls.GetNodeID(schemaNode)
	ctx.GetLogger().Debug(map[string]interface{}{"mth": "reshapeNode", "building": schemaNodeID})

	// Evaluate expressions and export values
	for _, expr := range EvaluateTermSemantics.GetEvaluatables(reshaper.Script.GetProperties(schemaNode)) {
		evalContext := ctx.getEvalContext()
		v, err := expr.Evaluate(evalContext)
		if err != nil {
			return nil, wrapReshapeError(err, schemaNodeID)
		}
		ctx.GetLogger().Info(map[string]interface{}{"reshape": schemaNodeID,
			"evaluateTermExpr": EvaluateTermSemantics.Get(reshaper.Script.GetProperties(schemaNode)),
			"result":           v})
		if rs, ok := v.Get().(opencypher.ResultSet); ok {
			if len(rs.Rows) > 0 {
				if err := ctx.exportResults(rs); err != nil {
					return nil, wrapReshapeError(err, schemaNodeID)
				}
			}
		}
	}
	// If the node is marked as a map context, evaluate that expr
	if ev := MapContextSemantics.GetEvaluatable(reshaper.Script.GetProperties(schemaNode)); ev != nil {
		evalContext := ctx.getEvalContext()
		mapContext, err := ev.Evaluate(evalContext)
		if err != nil {
			return nil, wrapReshapeError(err, schemaNodeID)
		}
		ctx.GetLogger().Debug(map[string]interface{}{"reshape": schemaNodeID, "mapContext": mapContext})
		node, err := getAtMostOneNode(mapContext)
		if err != nil {
			return nil, wrapReshapeError(err, schemaNodeID)
		}
		if node != nil {
			ctx.mapContext.push(node)
			defer ctx.mapContext.pop()
		}
	}

	// This is the case where the schema node has `mapProperty:
	// propName`. In this case, nodes under the map context that has the
	// property `propName: schemaNodeId` will be selected as the source
	// nodes.
	if mapProperty := ls.AsPropertyValue(reshaper.Script.GetProperties(schemaNode).GetProperty(MapPropertyTerm)).AsString(); len(mapProperty) > 0 {
		ctx.GetLogger().Debug(map[string]interface{}{"reshape": schemaNodeID, "valueFrom": MapPropertyTerm})
		// Find the nodes under the map context whose mapProperty property points to schemaNodeID
		nodeValues := reshaper.findNodesUnderMapContext(ctx, mapProperty, []string{schemaNodeID})
		return reshaper.generateOutput(ctx, nodeValues)
	}

	// This is the case where the script contains a direct mapping for
	// the node. There is a mapping with source and target in the
	// script, and the source nodes are the nodes whose schemaNodeId are
	// the source values.
	if nodeMappings := reshaper.Script.GetMappingsByTarget(schemaNodeID); len(nodeMappings) != 0 {
		ctx.GetLogger().Debug(map[string]interface{}{"reshape": schemaNodeID, "valueFrom": "map by target"})
		// Find the nodes under the map context whose source node ID is given in
		nodeValues := reshaper.findNodesUnderMapContext(ctx, ls.SchemaNodeIDTerm, getMappingSources(nodeMappings))
		return reshaper.generateOutput(ctx, nodeValues)
	}

	processValueExpr := func(evalContext *opencypher.EvalContext, term string) ([]opencypher.Value, error) {
		evaluatables := ValueExprTermSemantics.GetEvaluatables(term, reshaper.Script.GetProperties(schemaNode))
		if len(evaluatables) == 0 {
			return nil, nil
		}
		ret := make([]opencypher.Value, 0, len(evaluatables))
		for _, evaluatable := range evaluatables {
			sv, err := evaluatable.Evaluate(evalContext)
			if err != nil {
				return nil, err
			}
			if isEmptyValue(sv) {
				continue
			}
			ret = append(ret, sv)
			if term == ValueExprTerm || term == ValueExprFirstTerm {
				break
			}
		}
		return ret, nil
	}

	// Evaluate value expressions
	{
		evalContext := ctx.getEvalContext()
		v1, err := processValueExpr(evalContext, ValueExprFirstTerm)
		if err != nil {
			return nil, wrapReshapeError(err, schemaNodeID)
		}
		v2, err := processValueExpr(evalContext, ValueExprAllTerm)
		if err != nil {
			return nil, wrapReshapeError(err, schemaNodeID)
		}
		v3, err := processValueExpr(evalContext, ValueExprTerm)
		if err != nil {
			return nil, wrapReshapeError(err, schemaNodeID)
		}
		results := make([]opencypher.Value, 0, len(v1)+len(v2)+len(v3))
		results = append(results, v1...)
		results = append(results, v2...)
		results = append(results, v3...)
		if len(results) > 0 {
			return reshaper.generateOutput(ctx, results)
		}
	}

	// If not a value node, try reshaping subtree
	if !schemaNode.HasLabel(ls.AttributeTypeValue) {
		v, err := reshaper.handleNode(ctx, nil)
		if err != nil {
			return nil, err
		}
		return []*txDocNode{v}, nil
	}
	return nil, nil
}

func (reshaper Reshaper) handlePrimitiveValue(ctx *reshapeContext, input interface{}) (*txDocNode, error) {
	if input == nil {
		return nil, nil
	}
	ret := newTxDocNode(ctx.schemaPath.last())
	ret.value = input

	return ret, nil
}

func (reshaper Reshaper) handleNode(ctx *reshapeContext, input graph.Node) (*txDocNode, error) {
	ret := newTxDocNode(ctx.schemaPath.last())
	ret.sourceNode = input
	ctx.generatedPath.push(ret)
	defer ctx.generatedPath.pop()
	// Descend into the schema
	switch {
	case ret.schemaNode.GetLabels().Has(ls.AttributeTypeObject):
		children := ls.GetObjectAttributeNodes(ret.schemaNode)
		ls.SortNodes(children)
		for _, child := range children {
			ctx.schemaPath.push(child)
			r, err := reshaper.reshapeNode(ctx)
			ctx.schemaPath.pop()
			if err != nil {
				return nil, err
			}
			for _, x := range r {
				ret.children = append(ret.children, x)
			}
		}
		if len(ret.children) > 0 {
			return ret, nil
		}

	case ret.schemaNode.GetLabels().Has(ls.AttributeTypeArray):
		elemNode := ls.GetArrayElementNode(ret.schemaNode)
		ctx.schemaPath.push(elemNode)
		defer ctx.schemaPath.pop()
		r, err := reshaper.reshapeNode(ctx)
		if err != nil {
			return nil, err
		}
		for _, x := range r {
			ret.children = append(ret.children, x)
		}
		if len(ret.children) > 0 {
			return ret, nil
		}

	case ret.schemaNode.GetLabels().Has(ls.AttributeTypeValue):
		if input == nil {
			return nil, nil
		}
	}
	return ret, nil
}

func (reshaper Reshaper) generateOutput(ctx *reshapeContext, input interface{}) ([]*txDocNode, error) {
	if input == nil {
		return nil, nil
	}
	// Handle primitives separately
	if val, ok := input.(opencypher.Value); ok {
		if opencypher.IsValuePrimitive(val) {
			// We have a single value. Nothing to export
			v, err := reshaper.handlePrimitiveValue(ctx, val.Get())
			if err != nil {
				return nil, err
			}
			if v == nil {
				return nil, nil
			}
			return []*txDocNode{v}, nil
		}
	}
	schemaNode := ctx.schemaPath.last()
	switch values := input.(type) {
	case graph.Node:
		v, err := reshaper.handleNode(ctx, values)
		if err != nil {
			return nil, err
		}
		return []*txDocNode{v}, nil
	case []graph.Node:
		ret := make([]*txDocNode, 0, len(values))
		for _, x := range values {
			v, err := reshaper.generateOutput(ctx, x)
			if err != nil {
				return nil, err
			}
			ret = append(ret, v...)
		}
		if len(ret) == 0 {
			return nil, nil
		}
		return ret, nil
	case opencypher.Value:
		return reshaper.generateOutput(ctx, values.Get())
	case []opencypher.Value:
		ret := make([]*txDocNode, 0, len(values))
		for _, x := range values {
			v, err := reshaper.generateOutput(ctx, x)
			if err != nil {
				return nil, err
			}
			ret = append(ret, v...)
		}
		if len(ret) == 0 {
			return nil, nil
		}
		return ret, nil
	case opencypher.ResultSet:
		ret := make([]*txDocNode, 0, len(values.Rows))
		for _, row := range values.Rows {
			if len(row) == 0 {
				continue
			}
			if !schemaNode.HasLabel(ls.AttributeTypeValue) {
				for k, v := range row {
					if opencypher.IsNamedResult(k) {
						ctx.setSymbolValue(k, v)
					}
				}
				result, err := reshaper.handleNode(ctx, nil)
				if err != nil {
					return nil, err
				}
				ret = append(ret, result)
			} else {
				if len(row) > 1 {
					return nil, fmt.Errorf("Resultset has multiple columns where one expected")
				}
				for k, v := range row {
					if opencypher.IsNamedResult(k) {
						ctx.setSymbolValue(k, v)
					}
					result, err := reshaper.generateOutput(ctx, v)
					if err != nil {
						return nil, err
					}
					ret = append(ret, result...)
				}
			}
		}
		if len(ret) == 0 {
			return nil, nil
		}
		return ret, nil
	}
	panic(fmt.Sprintf("Unhandled input type: %T", input))
}

func (reshaper Reshaper) findNodesUnderMapContext(ctx *reshapeContext, propertyKey string, values []string) []graph.Node {
	ret := make([]graph.Node, 0)
	has := func(s string) bool {
		if len(s) == 0 {
			return false
		}
		for _, x := range values {
			if x == s {
				return true
			}
		}
		return false
	}
	if ctx.mapContext.empty() {
		for nodes := ctx.sourceGraph.GetNodesWithProperty(propertyKey); nodes.Next(); {
			node := nodes.Node()
			s := ls.AsPropertyValue(node.GetProperty(propertyKey)).AsString()
			if has(s) {
				ret = append(ret, node)
			}
		}
		return ret
	}

	mc := ctx.mapContext.last()
	ls.IterateDescendants(mc, func(node graph.Node) bool {
		s := ls.AsPropertyValue(node.GetProperty(propertyKey)).AsString()
		if has(s) {
			ret = append(ret, node)
		}
		return true
	}, ls.OnlyDocumentNodes, false)
	return ret
}

func isEmptyValue(v opencypher.Value) bool {
	if v.Get() == nil {
		return true
	}
	rs, ok := v.Get().(opencypher.ResultSet)
	if !ok {
		return false
	}
	if len(rs.Rows) == 0 {
		return true
	}
	if len(rs.Rows) == 1 {
		if len(rs.Rows[0]) == 1 {
			for _, v := range rs.Rows[0] {
				if v.Get() == nil {
					return true
				}
			}
		}
	}
	return false
}

func getAtMostOneNode(value opencypher.Value) (graph.Node, error) {
	if value == nil {
		return nil, nil
	}
	val := value.Get()
	if val == nil {
		return nil, nil
	}
	node, ok := val.(graph.Node)
	if ok {
		return node, nil
	}
	nodes, ok := val.([]graph.Node)
	if ok {
		if len(nodes) == 0 {
			return nil, nil
		}
		if len(nodes) > 1 {
			return nil, fmt.Errorf("Multiple nodes in result where one required")
		}
		return nodes[0], nil
	}
	rs, ok := val.(opencypher.ResultSet)
	if !ok {
		return nil, fmt.Errorf("Unhandled result type: %T", rs)
	}
	if len(rs.Rows) == 0 {
		return nil, nil
	}
	if len(rs.Rows) > 1 {
		return nil, fmt.Errorf("Result set has multiple rows where one node required")
	}
	row := rs.Rows[0]
	if len(row) == 0 {
		return nil, nil
	}
	if len(row) > 1 {
		return nil, fmt.Errorf("Resultset has multiple columns, only one result is required")
	}
	for _, v := range row {
		return getAtMostOneNode(v)
	}
	return nil, nil
}
