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

package ls

import (
	"fmt"

	"github.com/cloudprivacylabs/opencypher/graph"
)

// ValuesetLookupRequest specifies an optional list of lookup tables
// and key-value pairs to lookup. If the lookup tables list is empty,
// the valueset lookup should check all compatible tables. The
// key-values may contain a single value with empty key for simple
// dictionay lookups
type ValuesetLookupRequest struct {
	TableIDs  []string
	KeyValues map[string]string
}

// ValuesetLookupResponse returns the key-value pairs that should be
// inserted into the graph
type ValuesetLookupResponse struct {
	KeyValues map[string]string
}

var (
	// ValuesetContextTerm specifies the context under which valueset related values are to be looked up.
	//
	// For instance, consider a row of data:
	//
	//   root ----> item1
	//        ----> item2
	//
	// If item1 and item2 are values that can be used as an input to
	// valueset lookup, then the context should be set to root
	//
	// A node must contain either the ValuesetContextTerm or the ValuesetTablesTerm to be used in lookup
	ValuesetContextTerm = NewTerm(LS, "vs/context", false, false, OverrideComposition, nil)

	// ValuesetTablesTerm specifies the list of table IDs to
	// lookup. This is optional.  A node must contain either the
	// ValuesetContextTerm or the ValuesetTablesTerm to be used in
	// lookup
	ValuesetTablesTerm = NewTerm(LS, "vs/valuesets", false, false, OverrideComposition, nil)

	// ValuesetRequestKeysTerm specifies the keys that will be used in
	// the valueset request. These keys are interpreted by the valueset
	// lookup.
	ValuesetRequestKeysTerm = NewTerm(LS, "vs/requestKeys", false, false, OverrideComposition, nil)

	// ValuesetRequestValuesTerm contains entries matching
	// ValuesetRequestKeysTerm. It specifies the schema node IDs of the
	// nodes containing values to lookup
	ValuesetRequestValuesTerm = NewTerm(LS, "vs/requestValues", false, false, OverrideComposition, nil)

	// ValuesetResultKeys term contains the keys that will be returned
	// from the valueset lookup. Values of these keys will be inserted under the context
	ValuesetResultKeysTerm = NewTerm(LS, "vs/resultKeys", false, false, OverrideComposition, nil)

	// ValuesetResultValuesTerm specifies the schema node IDs for the
	// nodes that will receive the matching key values. If there is only one, resultKeys is optional
	ValuesetResultValuesTerm = NewTerm(LS, "vs/resultValues", false, false, OverrideComposition, nil)
)

// ValuesetInfo describes value set information for a schema node.
//
// A schema node containing valueset information requests the use of a
// dictionary or valueset to manipulate the graph. The results of
// valueset lookup can be placed under new nodes of the schema node.
//
// The valueset or dictionary lookup is envisioned as an external
// service. This service is called with the list of valueset ids, and
// a set of key:value pairs as the request. The response is another
// set of key:value pairs. Valueset processing adds the response to
// the graph.
//
// The valueset context node specifies a node that is either the
// current node or one of its ancestors. All the nodes for the
// valueset lookup are found under this context, and valueset lookup
// results are added under this context node.
//
// The valueset info may specify one or mode valuesets to lookup
// (TableIDs). If none specified, all compatible valuesets should be
// looked up.
//
type ValuesetInfo struct {
	// If the valueset lookup requires a single value, the attribute id
	// of the source node. Otherwise, the root node containing all the
	// required values. Results of the valueset lookup will also be
	// inserted under this node. If this is empty, then the current node
	// is the context node
	ContextID string

	// Optional lookup table IDs. If omitted, all compatible tables are
	// looked up
	TableIDs []string

	// Ordered list of valueset keys. The request to the valueset
	// function will use these as the request keys
	RequestKeys []string

	// Ordered list of attribute ids containing valueset request
	// values. The elements of this match the keys array
	RequestValues []string

	// The keys of the valueset result
	ResultKeys []string

	// The attribute ids of the nodes under this node to receive values
	ResultValues []string

	// The schemanode containing the valueset info
	SchemaNode graph.Node
}

type ErrInvalidValuesetSpec struct {
	Msg string
}

func (e ErrInvalidValuesetSpec) Error() string {
	return fmt.Sprintf("Value set error: %s", e.Msg)
}

type ErrValueset struct {
	Msg          string
	SchemaNodeID string
}

func (e ErrValueset) Error() string {
	return fmt.Sprintf("Value set processing error %s: %s", e.SchemaNodeID, e.Msg)
}

// ValueSetInfoFromNode parses the valuese information from a
// node. Returns nil if the node does not have valueset info
func ValuesetInfoFromNode(node graph.Node) *ValuesetInfo {
	ctxp := AsPropertyValue(node.GetProperty(ValuesetContextTerm))
	tablep := AsPropertyValue(node.GetProperty(ValuesetTablesTerm))
	if ctxp == nil && tablep == nil {
		return nil
	}
	ret := &ValuesetInfo{
		ContextID:     ctxp.AsString(),
		TableIDs:      tablep.MustStringSlice(),
		RequestKeys:   AsPropertyValue(node.GetProperty(ValuesetRequestKeysTerm)).MustStringSlice(),
		RequestValues: AsPropertyValue(node.GetProperty(ValuesetRequestValuesTerm)).MustStringSlice(),
		ResultKeys:    AsPropertyValue(node.GetProperty(ValuesetResultKeysTerm)).MustStringSlice(),
		ResultValues:  AsPropertyValue(node.GetProperty(ValuesetResultValuesTerm)).MustStringSlice(),
		SchemaNode:    node,
	}
	if len(ret.ContextID) == 0 {
		ret.ContextID = GetNodeID(node)
	}
	return ret
}

// GetRequest builds a valueset service request in the form of
// key:value pairs. All the request values are expected to be under the
// contextDocumentNode.
//
// If ValuesetInfo contains no RequestValues, then the
// vsiDocumentNode is used as the request value. If there are
// RequestKeys specified, there must be only one, and that is used as
// the key. Otherwise, the request is prepared with empty key.
//
// vsiDocumentNode can be nil
//
// If ValuesetInfo contains RequestValues, the request values
func (vsi *ValuesetInfo) GetRequest(contextDocumentNode, vsiDocumentNode graph.Node) (map[string]string, error) {
	if len(vsi.RequestValues) == 0 {
		if vsiDocumentNode == nil {
			return nil, ErrInvalidValuesetSpec{Msg: fmt.Sprintf("Document node is nil for the value set in context node %v", contextDocumentNode)}
		}
		value, _ := GetRawNodeValue(vsiDocumentNode)
		// Document node is the source node
		// There can be at most one key
		if len(vsi.RequestKeys) > 1 {
			return nil, ErrInvalidValuesetSpec{Msg: "Multiple request keys"}
		}
		if len(vsi.RequestKeys) == 1 {
			return map[string]string{vsi.RequestKeys[0]: value}, nil
		}
		return map[string]string{"": value}, nil
	}

	if !((len(vsi.RequestValues) == 1 && len(vsi.RequestKeys) == 0) ||
		(len(vsi.RequestValues) == len(vsi.RequestKeys))) {
		return nil, ErrInvalidValuesetSpec{Msg: "Inconsistent request keys and values"}
	}
	// Here, either there is one value and no key, or there are keys
	// and values of same length
	req := make(map[string]string)
	// There are some request value fields under this node. Collect them.
	for index, reqv := range vsi.RequestValues {
		if reqv == AsPropertyValue(contextDocumentNode.GetProperty(SchemaNodeIDTerm)).AsString() {
			value, _ := GetRawNodeValue(contextDocumentNode)
			if len(vsi.RequestKeys) == 0 {
				req[""] = value
			} else {
				req[vsi.RequestKeys[index]] = value
			}
		} else {
			// Locate a child node
			// match (n)-[]->({SchemaNodeIDTerm:reqv})
			pattern := graph.Pattern{
				{
					Name: "n",
				},
				{
					Min: 1,
					Max: -1,
				},
				{
					Properties: map[string]interface{}{SchemaNodeIDTerm: StringPropertyValue(reqv)},
				}}
			p := graph.PatternSymbol{}
			p.Add(contextDocumentNode)
			acc, err := pattern.FindPaths(contextDocumentNode.GetGraph(), map[string]*graph.PatternSymbol{"n": &p})
			if err != nil {
				return nil, err
			}
			nodes := acc.GetTailNodes()
			if len(nodes) > 1 {
				return nil, ErrInvalidValuesetSpec{Msg: fmt.Sprintf("Multiple nodes instance of %s", reqv)}
			}
			if len(nodes) == 1 {
				if len(vsi.RequestKeys) == 0 {
					req[""], _ = GetRawNodeValue(nodes[0])
				} else {
					req[vsi.RequestKeys[index]], _ = GetRawNodeValue(nodes[0])
				}
			}
		}
	}
	return req, nil
}

// GetContextNodes returns the contexts node for the given document
func (vsi *ValuesetInfo) GetContextNodes(g graph.Graph) ([]graph.Node, error) {
	pattern := graph.Pattern{
		{
			Properties: map[string]interface{}{SchemaNodeIDTerm: StringPropertyValue(vsi.ContextID)},
		},
	}
	return pattern.FindNodes(g, nil)
}

// GetContextNode returns the context node for the given document
// node. The context node must be the node itself, or an ancestor of
// the node
func (vsi *ValuesetInfo) GetContextNode(docNode graph.Node) (graph.Node, error) {
	pattern := graph.Pattern{
		{
			Properties: map[string]interface{}{SchemaNodeIDTerm: StringPropertyValue(vsi.ContextID)},
		},
		{
			Min: 0,
			Max: -1,
		},
		{
			Name: "start",
		},
	}
	ps := graph.PatternSymbol{}
	ps.AddNode(docNode)
	nodes, err := pattern.FindNodes(docNode.GetGraph(), map[string]*graph.PatternSymbol{"start": &ps})
	if err != nil {
		return nil, err
	}
	if len(nodes) > 1 {
		return nil, ErrValueset{SchemaNodeID: GetNodeID(vsi.SchemaNode), Msg: "Multiple context nodes"}
	}
	if len(nodes) == 0 {
		return nil, nil
	}
	return nodes[0], nil
}

// GetDocNodes returns the document nodes that are instance of the vsi schema node
func (vsi *ValuesetInfo) GetDocNodes(g graph.Graph) []graph.Node {
	pattern := graph.Pattern{
		{
			Properties: map[string]interface{}{SchemaNodeIDTerm: StringPropertyValue(GetNodeID(vsi.SchemaNode))},
		}}
	nodes, err := pattern.FindNodes(g, nil)
	if err != nil {
		panic(err)
	}
	return nodes
}

func (vsi *ValuesetInfo) createResultNodes(ctx *Context, builder GraphBuilder, layer *Layer, contextDocumentNode graph.Node, resultSchemaNodeID string, resultValue string) error {
	// There is value. If there is a node, update it. Otherwise, insert it
	resultSchemaNode := layer.GetAttributeByID(resultSchemaNodeID)
	if resultSchemaNode == nil {
		return ErrValueset{SchemaNodeID: vsi.ContextID, Msg: fmt.Sprintf("Target schema node %s does not exist in layer", resultSchemaNodeID)}
	}
	resultNodes := FindChildInstanceOf(contextDocumentNode, resultSchemaNodeID)
	switch len(resultNodes) {
	case 0: // insert it
		ctx.GetLogger().Debug(map[string]interface{}{"valueset.createResultNodes": "inserting"})
		switch GetIngestAs(resultSchemaNode) {
		case "node":
			_, _, err := builder.ValueAsNode(resultSchemaNode, contextDocumentNode, resultValue)
			if err != nil {
				return ErrValueset{SchemaNodeID: vsi.ContextID, Msg: fmt.Sprintf("Cannot create new node: %s", err.Error())}
			}
			ctx.GetLogger().Debug(map[string]interface{}{"valueset.createResultNodes": "insert", "schma": resultSchemaNode, "parent": contextDocumentNode})
		case "edge":
			_, err := builder.ValueAsEdge(resultSchemaNode, contextDocumentNode, resultValue)
			if err != nil {
				return ErrValueset{SchemaNodeID: vsi.ContextID, Msg: fmt.Sprintf("Cannot create new node: %s", err.Error())}
			}
		case "property":
			err := builder.ValueAsProperty(resultSchemaNode, []graph.Node{contextDocumentNode}, resultValue)
			if err != nil {
				return ErrValueset{SchemaNodeID: vsi.ContextID, Msg: fmt.Sprintf("Cannot create new node: %s", err.Error())}
			}
		}
	case 1: // update it
		switch GetIngestAs(resultSchemaNode) {
		case "node", "edge":
			SetRawNodeValue(resultNodes[0], resultValue)
		default:
			return ErrValueset{SchemaNodeID: vsi.ContextID, Msg: "Cannot update value in property, inconsistent graph"}
		}
	}
	return nil
}

func (vsi *ValuesetInfo) ApplyValuesetResponse(ctx *Context, builder GraphBuilder, layer *Layer, contextDocumentNode, contextSchemaNode graph.Node, result ValuesetLookupResponse) error {
	if len(result.KeyValues) == 0 {
		return nil
	}
	if len(vsi.ResultKeys) == 0 && len(vsi.ResultValues) == 0 {
		ctx.GetLogger().Debug(map[string]interface{}{"mth": "valueset.applyThisNode", "contextSchemaNode": contextSchemaNode})
		// Target is this document node
		if len(result.KeyValues) != 1 {
			return ErrValueset{SchemaNodeID: vsi.ContextID, Msg: "Multiple results from valueset lookup, but no ResultKeys specified in the schema"}
		}
		if !contextSchemaNode.HasLabel(AttributeTypeValue) {
			return ErrValueset{SchemaNodeID: GetNodeID(contextSchemaNode), Msg: "Trying to set the value of a non-value node using valueset"}
		}
		for _, v := range result.KeyValues {
			SetRawNodeValue(contextDocumentNode, v)
			return nil
		}
	}
	// If only one resultValues is specified and there is one result
	if len(vsi.ResultKeys) == 0 && len(vsi.ResultValues) == 1 && len(result.KeyValues) == 1 {
		ctx.GetLogger().Debug(map[string]interface{}{"mth": "valueset.applyOneNode"})
		resultNodeID := vsi.ResultValues[0]
		var resultValue string
		for _, v := range result.KeyValues {
			resultValue = v
		}
		if err := vsi.createResultNodes(ctx, builder, layer, contextDocumentNode, resultNodeID, resultValue); err != nil {
			return err
		}
		return nil
	}
	ctx.GetLogger().Debug(map[string]interface{}{"mth": "valueset.keyValues"})
	for resultKeyIndex, resultKey := range vsi.ResultKeys {
		if len(vsi.ResultValues) < resultKeyIndex {
			continue
		}
		// Assign resultValue to resultNodeID
		resultNodeID := vsi.ResultValues[resultKeyIndex]
		resultValue, ok := result.KeyValues[resultKey]
		if !ok {
			// No value. If there is a resultNodeID, remove it
			resultNodes := FindChildInstanceOf(contextDocumentNode, resultNodeID)
			for _, n := range resultNodes {
				n.DetachAndRemove()
			}
			return nil
		}
		if err := vsi.createResultNodes(ctx, builder, layer, contextDocumentNode, resultNodeID, resultValue); err != nil {
			return err
		}
	}
	return nil
}

// ProcessByContextNode processes the value set of the given context  node and the schema node containing the vsi
func (prc *ValuesetProcessor) ProcessByContextNode(ctx *Context, builder GraphBuilder, contextDocNode, contextSchemaNode, vsiDocNode graph.Node, vsi *ValuesetInfo) error {
	kv, err := vsi.GetRequest(contextDocNode, vsiDocNode)
	if err != nil {
		return err
	}
	ctx.GetLogger().Debug(map[string]interface{}{"mth": "valueset.process", "request": kv})
	if len(kv) != 0 {
		// Perform the lookup
		result, err := prc.lookupFunc(ctx, ValuesetLookupRequest{
			TableIDs:  vsi.TableIDs,
			KeyValues: kv,
		})
		if err != nil {
			return err
		}
		ctx.GetLogger().Debug(map[string]interface{}{"mth": "valueset.process", "result": result, "contextDocNode": contextDocNode})
		// If there is nonzero result, put it back into the doc
		if len(result.KeyValues) > 0 {
			if err := vsi.ApplyValuesetResponse(ctx, builder, prc.layer, contextDocNode, contextSchemaNode, result); err != nil {
				return err
			}
		}
	}
	return nil
}

// Process processes the value set of the given context document node and the schema node containing the vsi
func (prc *ValuesetProcessor) Process(ctx *Context, builder GraphBuilder, vsiDocNode, contextSchemaNode graph.Node, vsi *ValuesetInfo) error {
	contextDocNode, err := vsi.GetContextNode(vsiDocNode)
	if err != nil {
		return err
	}
	ctx.GetLogger().Debug(map[string]interface{}{"mth": "valueset.process", "contextNode": contextSchemaNode, "docNode": vsiDocNode})
	return prc.ProcessByContextNode(ctx, builder, contextDocNode, contextSchemaNode, vsiDocNode, vsi)
}

type ValuesetProcessor struct {
	layer      *Layer
	lookupFunc func(*Context, ValuesetLookupRequest) (ValuesetLookupResponse, error)
	vsis       []ValuesetInfo
}

func NewValuesetProcessor(layer *Layer, lookupFunc func(*Context, ValuesetLookupRequest) (ValuesetLookupResponse, error)) ValuesetProcessor {
	ret := ValuesetProcessor{
		layer:      layer,
		lookupFunc: lookupFunc,
	}
	ret.init()
	return ret
}

func (prc *ValuesetProcessor) init() {
	if prc.vsis != nil {
		return
	}
	for nodes := prc.layer.Graph.GetNodes(); nodes.Next(); {
		node := nodes.Node()
		vsi := ValuesetInfoFromNode(node)
		if vsi == nil {
			continue
		}
		prc.vsis = append(prc.vsis, *vsi)
	}
}

func (prc *ValuesetProcessor) ProcessGraphValueset(ctx *Context, builder GraphBuilder, vsi *ValuesetInfo) error {
	vsiDocNodes := vsi.GetDocNodes(builder.GetGraph())
	contextSchemaNode := prc.layer.GetAttributeByID(vsi.ContextID)
	if contextSchemaNode == nil {
		return nil
	}
	ctx.GetLogger().Debug(map[string]interface{}{"mth": "processGraphValueset", "stage": "found context node", "vsi": vsi})
	if len(vsiDocNodes) == 0 {
		contextNodes, err := vsi.GetContextNodes(builder.GetGraph())
		if err != nil {
			return err
		}
		for _, contextNode := range contextNodes {
			if err := prc.ProcessByContextNode(ctx, builder, contextNode, contextSchemaNode, nil, vsi); err != nil {
				return err
			}
		}
		return nil
	}
	for _, vsiDocNode := range vsiDocNodes {
		if err := prc.Process(ctx, builder, vsiDocNode, contextSchemaNode, vsi); err != nil {
			return err
		}
	}
	return nil
}

func (prc *ValuesetProcessor) ProcessGraph(ctx *Context, builder GraphBuilder) error {
	ctx.GetLogger().Debug(map[string]interface{}{"mth": "processGraph", "nVSI": len(prc.vsis)})
	for i := range prc.vsis {
		if err := prc.ProcessGraphValueset(ctx, builder, &prc.vsis[i]); err != nil {
			return err
		}
	}
	return nil
}
