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

type reshapeContext struct {
	*ls.Context
	parent        *reshapeContext
	schemaPath    []graph.Node
	generatedPath []graph.Node
	mapContext    []graph.Node
	builder       ls.GraphBuilder
	symbols       map[string]opencypher.Value
	sourceGraph   graph.Graph
}

func (ctx *reshapeContext) getSchemaNode() graph.Node { return ctx.schemaPath[len(ctx.schemaPath)-1] }
func (ctx *reshapeContext) getParentGraphNode() graph.Node {
	if len(ctx.generatedPath) == 0 {
		return nil
	}
	return ctx.generatedPath[len(ctx.generatedPath)-1]
}

func (ctx *reshapeContext) pushGeneratedNode(node graph.Node) {
	ctx.generatedPath = append(ctx.generatedPath, node)
}

func (ctx *reshapeContext) popGeneratedNode() {
	ctx.generatedPath = ctx.generatedPath[:len(ctx.generatedPath)-1]
}

func (ctx *reshapeContext) pushSchemaNode(node graph.Node) {
	ctx.schemaPath = append(ctx.schemaPath, node)
}

func (ctx *reshapeContext) popSchemaNode() {
	ctx.schemaPath = ctx.schemaPath[:len(ctx.schemaPath)-1]
}

func (ctx *reshapeContext) pushMapContext(node graph.Node) {
	ctx.mapContext = append(ctx.mapContext, node)
}

func (ctx *reshapeContext) popMapContext() {
	ctx.mapContext = ctx.mapContext[:len(ctx.mapContext)-1]
}

func (ctx *reshapeContext) getMapContext() graph.Node {
	if len(ctx.mapContext) == 0 {
		return nil
	}
	return ctx.mapContext[len(ctx.mapContext)-1]
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

func (ctx *reshapeContext) nestedContext() *reshapeContext {
	ret := *ctx
	ret.parent = ctx
	ret.symbols = make(map[string]opencypher.Value)
	return &ret
}

// Export the variables in the resultset
func (ctx *reshapeContext) exportVars(row map[string]opencypher.Value) {
	for varName, val := range row {
		ctx.setSymbolValue(varName, val)
	}
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

func (reshaper Reshaper) Reshape(ctx *ls.Context, sourceGraph graph.Graph) error {
	c := reshapeContext{
		Context:     ctx,
		symbols:     make(map[string]opencypher.Value),
		schemaPath:  []graph.Node{reshaper.TargetSchema.GetSchemaRootNode()},
		sourceGraph: sourceGraph,
		builder:     reshaper.Builder,
	}
	_, err := reshaper.reshapeNode(&c)
	return err
}

func getResultSetColumn(rs opencypher.ResultSet, key string) ([]opencypher.Value, bool) {
	ret := make([]opencypher.Value, 0)
	foundKey := false
	for _, row := range rs.Rows {
		val, ok := row[key]
		if ok {
			foundKey = true
			ret = append(ret, val)
		}
	}
	return ret, foundKey
}

// Returns true if operation produced any output
func (reshaper Reshaper) reshapeNode(ctx *reshapeContext) (bool, error) {
	// last element of schemaPath is the schema node to reshape
	schemaNode := ctx.getSchemaNode()
	schemaNodeID := ls.GetNodeID(schemaNode)
	ctx.GetLogger().Debug(map[string]interface{}{"reshape": schemaNodeID})

	// Evaluate expressions and export values
	for _, expr := range EvaluateTermSemantics.GetEvaluatables(reshaper.Script.GetProperties(schemaNode)) {
		evalContext := ctx.getEvalContext()
		v, err := expr.Evaluate(evalContext)
		if err != nil {
			return false, wrapReshapeError(err, schemaNodeID)
		}
		ctx.GetLogger().Info(map[string]interface{}{"reshape": schemaNodeID,
			"evaluateTermExpr": EvaluateTermSemantics.Get(reshaper.Script.GetProperties(schemaNode)),
			"result":           v})
		if rs, ok := v.Value.(opencypher.ResultSet); ok {
			if len(rs.Rows) > 1 {
				return false, wrapReshapeError(fmt.Errorf("Multiple values for resultset"), schemaNodeID)
			}
			if len(rs.Rows) == 1 {
				ctx.exportVars(rs.Rows[0])
			}
		} else {
			return false, wrapReshapeError(fmt.Errorf("evaluate result is not a resultset"), schemaNodeID)
		}
	}

	// The node can get its value in one of the following ways:
	//
	//    mapProperty: If set, get the mapped value, which is []Node, and process each
	//    valueExpr: If has values, get the results of the expression, process each
	//
	// otherwise, if this is a non-value node, process children and
	// create the node if there are nonempty children
	// Is it a map node?
	if mapProperty := ls.AsPropertyValue(reshaper.Script.GetProperties(schemaNode).GetProperty(MapPropertyTerm)).AsString(); len(mapProperty) > 0 {
		ctx.GetLogger().Debug(map[string]interface{}{"reshape": schemaNodeID, "state": "Running map", "mapProperty": mapProperty})
		// Find the nodes under the map context whose mapProperty property points to this
		nodes, err := reshaper.getMapNodes(ctx, mapProperty, schemaNodeID)
		if err != nil {
			return false, wrapReshapeError(err, schemaNodeID)
		}
		ctx.GetLogger().Debug(map[string]interface{}{"reshape": schemaNodeID, "nMapNodes": len(nodes)})
		v, err := reshaper.handle(ctx, opencypher.Value{Value: nodes})
		if err != nil {
			return false, wrapReshapeError(err, schemaNodeID)
		}
		return v, nil
	}
	// Script may contain a direct mapping
	if nodeMapping := reshaper.Script.GetMappingByTarget(schemaNodeID); nodeMapping != nil {
		nodes, err := reshaper.getNodeMappingNodes(ctx, nodeMapping)
		if err != nil {
			return false, wrapReshapeError(err, schemaNodeID)
		}
		ctx.GetLogger().Debug(map[string]interface{}{"reshape": schemaNodeID, "nMapNodes": len(nodes)})
		v, err := reshaper.handle(ctx, opencypher.Value{Value: nodes})
		if err != nil {
			return false, wrapReshapeError(err, schemaNodeID)
		}
		return v, nil
	}

	if evaluatables := ValueExprTermSemantics.GetEvaluatables(reshaper.Script.GetProperties(schemaNode)); len(evaluatables) != 0 {
		// Determine the source nodes for ingestion
		evalContext := ctx.getEvalContext()
		for _, evaluatable := range evaluatables {
			sv, err := evaluatable.Evaluate(evalContext)
			if err != nil {
				return false, wrapReshapeError(err, schemaNodeID)
			}
			ctx.GetLogger().Info(map[string]interface{}{"reshape": schemaNodeID,
				"valueExpr": ValueExprTermSemantics.Get(reshaper.Script.GetProperties(schemaNode)),
				"result":    sv})
			if isEmptyValue(sv) {
				continue
			}
			ctx.GetLogger().Debug(map[string]interface{}{"reshape": schemaNodeID, "stage": "Evaluated value expression"})
			v, err := reshaper.handle(ctx, sv)
			if err != nil {
				return false, wrapReshapeError(err, schemaNodeID)
			}
			return v, nil
		}
		return false, nil
	}

	// If not a value node, try reshaping subtree
	if !schemaNode.HasLabel(ls.AttributeTypeValue) {
		v, err := reshaper.handleNonValueNode(ctx)
		if err != nil {
			return false, wrapReshapeError(err, schemaNodeID)
		}
		return v, nil
	}
	return false, nil
}

func isEmptyValue(v opencypher.Value) bool {
	if v.Value == nil {
		return true
	}
	rs, ok := v.Value.(opencypher.ResultSet)
	if !ok {
		return false
	}
	if len(rs.Rows) == 0 {
		return true
	}
	return false
}

func (reshaper Reshaper) evaluateMapContext(ctx *reshapeContext) (bool, error) {
	schemaNode := ctx.getSchemaNode()
	ev := MapContextSemantics.GetEvaluatable(reshaper.Script.GetProperties(schemaNode))
	if ev == nil {
		return false, nil
	}
	evalContext := ctx.getEvalContext()
	mapContext, err := ev.Evaluate(evalContext)
	if err != nil {
		return false, nil
	}
	ctx.GetLogger().Debug(map[string]interface{}{"reshape": ls.GetNodeID(schemaNode), "mapContext": mapContext})
	// Map context can be a node, or a node slice with one element, or an rs with a node
	switch c := mapContext.Value.(type) {
	case graph.Node:
		ctx.pushMapContext(c)
		return true, nil

	case []graph.Node:
		switch len(c) {
		case 1:
			ctx.pushMapContext(c[0])
			return true, nil
		case 0:
			ctx.pushMapContext(nil)
			return true, nil
		default:
			return false, fmt.Errorf("Map context has multiple nodes")
		}
	case opencypher.ResultSet:
		switch len(c.Rows) {
		case 0:
			ctx.pushMapContext(nil)
			return true, nil
		case 1:
			if len(c.Rows[0]) != 1 {
				return false, fmt.Errorf("Map context resultset has to have one column")
			}
			for _, v := range c.Rows[0] {
				n, ok := v.Value.(graph.Node)
				if !ok {
					return false, fmt.Errorf("Map context expression must return a single node")
				}
				ctx.pushMapContext(n)
				return true, nil
			}
		default:
			return false, fmt.Errorf("Map context has multiple items")
		}
	}
	return false, nil
}

func (reshaper Reshaper) getNodeMappingNodes(ctx *reshapeContext, mapping *NodeMapping) ([]graph.Node, error) {
	mapContext := ctx.getMapContext()
	ctx.GetLogger().Debug(map[string]interface{}{"reshaper.getNodeMappingNodes.context": mapContext})
	var found []graph.Node
	checkNodeProperty := func(node graph.Node) bool {
		schNode := ls.AsPropertyValue(node.GetProperty(ls.SchemaNodeIDTerm)).AsString()
		if schNode == mapping.SourceNodeID {
			found = append(found, node)
			return true
		}
		return false
	}
	if mapContext != nil {
		ls.IterateDescendants(mapContext, func(node graph.Node) bool {
			checkNodeProperty(node)
			return true
		}, ls.OnlyDocumentNodes, true)
	} else {
		for nodes := ctx.sourceGraph.GetNodes(); nodes.Next(); {
			node := nodes.Node()
			if !node.HasLabel(ls.DocumentNodeTerm) {
				continue
			}
			checkNodeProperty(node)
		}
	}
	return found, nil
}

func (reshaper Reshaper) getMapNodes(ctx *reshapeContext, mapProperty, schemaNodeID string) ([]graph.Node, error) {
	mapContext := ctx.getMapContext()
	ctx.GetLogger().Debug(map[string]interface{}{"reshaper.getMapNodes.context": mapContext})
	var found []graph.Node
	checkNodeProperty := func(node graph.Node) bool {
		pv := ls.AsPropertyValue(node.GetProperty(mapProperty))
		if pv == nil {
			return false
		}
		if pv.IsString() {
			if pv.AsString() == schemaNodeID {
				found = append(found, node)
				return true
			}
		}
		if pv.IsStringSlice() {
			for _, x := range pv.AsStringSlice() {
				if x == schemaNodeID {
					found = append(found, node)
					return true
				}
			}
		}
		return false
	}
	if mapContext != nil {
		ls.IterateDescendants(mapContext, func(node graph.Node) bool {
			checkNodeProperty(node)
			return true
		}, ls.OnlyDocumentNodes, true)
	} else {
		for nodes := ctx.sourceGraph.GetNodes(); nodes.Next(); {
			node := nodes.Node()
			if !node.HasLabel(ls.DocumentNodeTerm) {
				continue
			}
			checkNodeProperty(node)
		}
	}
	return found, nil
}

// Returns true if operation produced any output
func (reshaper Reshaper) handle(ctx *reshapeContext, value opencypher.Value) (bool, error) {
	schemaNodeID := ls.GetNodeID(ctx.getSchemaNode())
	if value.Value == nil {
		ctx.GetLogger().Debug(map[string]interface{}{"reshape": schemaNodeID, "stage": "source is nil"})
		return false, nil
	}
	if value.IsPrimitive() {
		ctx.GetLogger().Debug(map[string]interface{}{"reshape": schemaNodeID, "stage": "source is primitive"})
		// We have a single value. Nothing to export
		return reshaper.handleValue(ctx, value.Value)
	}
	if node, ok := value.Value.(graph.Node); ok {
		ctx.GetLogger().Debug(map[string]interface{}{"reshape": schemaNodeID, "stage": "source is a node"})
		return reshaper.handleNode(ctx, node)
	}
	if nodes, ok := value.Value.([]graph.Node); ok {
		ctx.GetLogger().Debug(map[string]interface{}{"reshape": schemaNodeID, "stage": "source is a []node"})
		return reshaper.handleNodeSlice(ctx, nodes)
	}
	if rs, ok := value.Value.(opencypher.ResultSet); ok {
		ctx.GetLogger().Debug(map[string]interface{}{"reshape": schemaNodeID, "stage": "source is a result set"})
		if len(rs.Rows) == 0 {
			ctx.GetLogger().Debug(map[string]interface{}{"reshape": schemaNodeID, "stage": "no results"})
			// No results
			return false, nil
		}
		return reshaper.handleResultSet(ctx, rs)

	}
	return false, ErrReshape{
		Wrapped:      fmt.Errorf("Unhandled result type: %T", value.Value),
		SchemaNodeID: schemaNodeID,
	}
}

// Returns true if operation produces output
func (reshaper Reshaper) handleValue(ctx *reshapeContext, value interface{}) (bool, error) {
	ctx.GetLogger().Debug(map[string]interface{}{"reshape.handleValue": value})
	schemaNode := ctx.getSchemaNode()
	// We have a value. The schema node must be a value
	if !schemaNode.HasLabel(ls.AttributeTypeValue) {
		return false, fmt.Errorf("Schema node is not a value, but a value is given")
	}
	parent := ctx.getParentGraphNode()
	if parent != nil && !parent.HasLabel(ls.AttributeTypeArray) {
		// If the parent is not an array, it might already have content
		// for this schema node. In that case, do we append to the
		// previous one, or combine them
		siblings := ls.FindChildInstanceOf(parent, ls.GetNodeID(schemaNode))
		if len(siblings) > 0 {
			ctx.GetLogger().Debug(map[string]interface{}{"reshape.handleValue": value, "siblings": len(siblings)})
			// There are already instances of this node
			// Can there be multiple?
			joinWithp, ok := reshaper.Script.GetProperties(schemaNode).GetProperty(JoinWithTerm)
			if ok {
				// join them
				lastSibling := siblings[len(siblings)-1]
				siblingValue, err := ls.GetNodeValue(lastSibling)
				if err != nil {
					return false, err
				}
				newValue := JoinValues([]string{fmt.Sprint(siblingValue), fmt.Sprint(value)}, joinWithp.(*ls.PropertyValue).AsString())
				if err := ls.SetNodeValue(lastSibling, newValue); err != nil {
					return false, err
				}
				return true, nil
			}
			// Multiple siblings, no join. Is multiple allowed?
			if ls.AsPropertyValue(reshaper.Script.GetProperties(schemaNode).GetProperty(MultipleTerm)).AsString() != "true" {
				return false, fmt.Errorf("Multiple values in a single value context")
			}
		}
	}
	// If we are here, we'll create a node for this value
	switch ls.GetIngestAs(schemaNode) {
	case "node":
		_, newNode, err := ctx.builder.ValueAsNode(schemaNode, parent, "")
		if err != nil {
			return false, err
		}
		return true, ls.SetNodeValue(newNode, value)

	case "edge":
		edge, err := ctx.builder.ValueAsEdge(schemaNode, parent, "")
		if err != nil {
			return false, err
		}
		return true, ls.SetNodeValue(edge.GetTo(), value)
	case "property":
		return true, ctx.builder.ValueAsProperty(schemaNode, ctx.generatedPath, fmt.Sprint(value))
	}
	return false, nil
}

func (reshaper Reshaper) handleNode(ctx *reshapeContext, node graph.Node) (bool, error) {
	ctx.GetLogger().Debug(map[string]interface{}{"reshape.handleNode": node})
	val, err := ls.GetNodeValue(node)
	if err != nil {
		return false, err
	}
	return reshaper.handleValue(ctx, val)
}

func (reshaper Reshaper) handleRow(ctx *reshapeContext, row map[string]opencypher.Value) (bool, error) {
	ctx.GetLogger().Debug(map[string]interface{}{"reshape.handleRow": row})
	if len(row) == 0 {
		return false, nil
	}
	mc, err := reshaper.evaluateMapContext(ctx)
	if err != nil {
		return false, nil
	}
	if mc {
		defer ctx.popMapContext()
	}
	schemaNode := ctx.getSchemaNode()
	parent := ctx.getParentGraphNode()
	multi := ls.AsPropertyValue(reshaper.Script.GetProperties(schemaNode).GetProperty(MultipleTerm)).AsString() == "true"
	switch {
	case schemaNode.HasLabel(ls.AttributeTypeValue):
		// If there is a single column, use that
		if len(row) == 1 {
			for _, v := range row {
				return reshaper.handle(ctx, v)
			}
		}
		return false, fmt.Errorf("Multi-column result set assigned to value")

	case schemaNode.HasLabel(ls.AttributeTypeObject) || schemaNode.HasLabel(ls.AttributeTypeArray):
		var siblings []graph.Node
		if parent != nil {
			siblings = ls.FindChildInstanceOf(parent, ls.GetNodeID(schemaNode))
		}
		if len(siblings) > 0 && !multi {
			return false, fmt.Errorf("Multiple values in resultset for %s", ls.GetNodeID(schemaNode))
		}
		return reshaper.handleNonValueNode(ctx)
	}
	return false, nil
}

func (reshaper Reshaper) handleNonValueNode(ctx *reshapeContext) (bool, error) {
	// Dealing with a non-value node. If the operation generates output,
	// keep it. Otherwise, remove it
	schemaNode := ctx.getSchemaNode()
	newNode := ctx.builder.NewNode(schemaNode)
	parentNode := ctx.getParentGraphNode()
	ctx.pushGeneratedNode(newNode)
	defer ctx.popGeneratedNode()
	hasOutput := false
	switch {
	case schemaNode.HasLabel(ls.AttributeTypeObject):
		nodes := ls.GetObjectAttributeNodes(schemaNode)
		ls.SortNodes(nodes)
		for _, node := range nodes {
			ctx.GetLogger().Debug(map[string]interface{}{"reshaper.handleNonValueNode": node})
			ctx.pushSchemaNode(node)
			mc, err := reshaper.evaluateMapContext(ctx)
			if err != nil {
				return false, nil
			}
			r, err := reshaper.reshapeNode(ctx)
			if mc {
				ctx.popMapContext()
			}
			ctx.popSchemaNode()
			if err != nil {
				return false, err
			}
			if r {
				hasOutput = true
			}
		}

	case schemaNode.HasLabel(ls.AttributeTypeArray):
		elemNode := ls.GetArrayElementNode(schemaNode)
		ctx.pushSchemaNode(elemNode)
		mc, err := reshaper.evaluateMapContext(ctx)
		if err != nil {
			return false, nil
		}
		r, err := reshaper.reshapeNode(ctx)
		if mc {
			ctx.popMapContext()
		}
		ctx.popSchemaNode()
		if err != nil {
			return false, err
		}
		hasOutput = r
	}
	if !hasOutput {
		newNode.DetachAndRemove()
	} else if parentNode != nil {
		parentNode.GetGraph().NewEdge(parentNode, newNode, ls.HasTerm, nil)
	}
	return hasOutput, nil
}

func (reshaper Reshaper) handleResultSet(ctx *reshapeContext, rs opencypher.ResultSet) (bool, error) {
	ctx.GetLogger().Debug(map[string]interface{}{"reshape.handleResultSet": len(rs.Rows)})
	schemaNode := ctx.getSchemaNode()
	hasResults := false
	for index, row := range rs.Rows {
		ctx.GetLogger().Debug(map[string]interface{}{
			"reshape": ls.GetNodeID(schemaNode),
			"stage":   "processing resultset for each value",
			"row":     index,
		})
		// Export values
		ctx.exportVars(row)
		mc, err := reshaper.evaluateMapContext(ctx)
		if err != nil {
			return false, nil
		}
		r, err := reshaper.handleRow(ctx, row)
		if mc {
			ctx.popMapContext()
		}
		if err != nil {
			return hasResults, err
		}
		if r {
			hasResults = true
		}
	}
	return hasResults, nil
}

func (reshaper Reshaper) handleNodeSlice(ctx *reshapeContext, nodes []graph.Node) (bool, error) {
	ctx.GetLogger().Debug(map[string]interface{}{"reshape.handleNodeSlice": len(nodes)})
	schemaNode := ctx.getSchemaNode()
	hasResults := false
	for index, node := range nodes {
		ctx.GetLogger().Debug(map[string]interface{}{
			"reshape": ls.GetNodeID(schemaNode),
			"stage":   "processing nodes for each value",
			"row":     index,
		})
		mc, err := reshaper.evaluateMapContext(ctx)
		if err != nil {
			return false, nil
		}
		if !mc {
			ctx.pushMapContext(node)
		}
		parent := ctx.getParentGraphNode()
		multi := ls.AsPropertyValue(reshaper.Script.GetProperties(schemaNode).GetProperty(MultipleTerm)).AsString() == "true"
		switch {
		case schemaNode.HasLabel(ls.AttributeTypeValue):
			r, err := reshaper.handle(ctx, opencypher.Value{Value: node})
			if err != nil {
				ctx.popMapContext()
				return false, err
			}
			if r {
				hasResults = true
			}

		case schemaNode.HasLabel(ls.AttributeTypeObject) || schemaNode.HasLabel(ls.AttributeTypeArray):
			var siblings []graph.Node
			if parent != nil {
				siblings = ls.FindChildInstanceOf(parent, ls.GetNodeID(schemaNode))
			}
			ctx.GetLogger().Debug(map[string]interface{}{"reshape": ls.GetNodeID(schemaNode),
				"type":     schemaNode.GetLabels(),
				"parent":   parent,
				"siblings": siblings})
			if len(siblings) > 0 && !multi {
				ctx.popMapContext()
				return false, fmt.Errorf("Multiple values in mapping")
			}
			r, err := reshaper.handleNonValueNode(ctx)
			if err != nil {
				ctx.popMapContext()
				return false, err
			}
			if r {
				hasResults = true
			}
		}
		ctx.popMapContext()
	}
	return hasResults, nil
}
