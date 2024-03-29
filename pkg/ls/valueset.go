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

	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/opencypher"
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
	//
	// If context is empty, entity root is assumed
	ValuesetContextTerm = RegisterStringTerm(NewTerm(LS, "vs/context").SetComposition(OverrideComposition).SetTags(SchemaElementTag))
	// ValuesetContextExprTerm is an opencyper expression that gives the context node using "this" as the current node
	ValuesetContextExprTerm = RegisterStringTerm(NewTerm(LS, "vs/contextExpr").SetComposition(OverrideComposition).SetTags(SchemaElementTag))

	// ValuesetTablesTerm specifies the list of table IDs to
	// lookup. This is optional.  A node must contain either the
	// ValuesetContextTerm or the ValuesetTablesTerm to be used in
	// lookup
	ValuesetTablesTerm = RegisterStringSliceTerm(NewTerm(LS, "vs/valuesets").SetComposition(OverrideComposition).SetTags(SchemaElementTag))

	// ValuesetRequestKeysTerm specifies the keys that will be used in
	// the valueset request. These keys are interpreted by the valueset
	// lookup.
	ValuesetRequestKeysTerm = RegisterStringSliceTerm(NewTerm(LS, "vs/requestKeys").SetComposition(OverrideComposition).SetTags(SchemaElementTag))

	// ValuesetRequestValuesTerm contains entries matching
	// ValuesetRequestKeysTerm. It specifies the schema node IDs of the
	// nodes containing values to lookup
	ValuesetRequestValuesTerm = RegisterStringSliceTerm(NewTerm(LS, "vs/requestValues").SetComposition(OverrideComposition).SetTags(SchemaElementTag))

	// ValuesetRequestTerm specifies one or more openCypher expressions
	// that builds up a valuest lookup request. The named results of
	// those expressions are added to the request key/value pairs
	ValuesetRequestTerm = RegisterStringSliceTerm(NewTerm(LS, "vs/request").SetComposition(OverrideComposition).SetTags(SchemaElementTag).SetMetadata(CompileOCSemantics{}))

	// ValuesetResultKeys term contains the keys that will be returned
	// from the valueset lookup. Values of these keys will be inserted under the context
	ValuesetResultKeysTerm = RegisterStringSliceTerm(NewTerm(LS, "vs/resultKeys").SetComposition(OverrideComposition).SetTags(SchemaElementTag))

	// ValuesetResultValuesTerm specifies the schema node IDs for the
	// nodes that will receive the matching key values. If there is only
	// one, resultKeys is optional The result value nodes must be a
	// direct descendant of one of the nodes from the document node up
	// to the context node.
	ValuesetResultValuesTerm = RegisterStringSliceTerm(NewTerm(LS, "vs/resultValues").SetComposition(OverrideComposition).SetTags(SchemaElementTag))

	// ValuesetResultContext determines the node under which the results
	// will be added. This is needed if the results will be added under
	// a different entity attached to the valueset context.
	ValuesetResultContextTerm = RegisterStringTerm(NewTerm(LS, "vs/resultContext").SetComposition(OverrideComposition).SetTags(SchemaElementTag))
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
type ValuesetInfo struct {
	// If the valueset lookup requires a single value, the attribute id
	// of the source node. Otherwise, the root node containing all the
	// required values. If this is empty, then the entity root
	// node is the context node
	ContextID string

	ContextExpr opencypher.Evaluatable

	// Optional lookup table IDs. If omitted, all compatible tables are
	// looked up
	TableIDs []string

	// Ordered list of valueset keys. The request to the valueset
	// function will use these as the request keys
	RequestKeys []string

	// Ordered list of attribute ids containing valueset request
	// values. The elements of this match the keys array
	RequestValues []string

	// This is initialized if RequestValues is nonempty. It is
	// initialized to point to the schema paths corresponding to the
	// request value entries
	requestSchemaPaths [][]*lpg.Node

	// Request expressions
	RequestExprs []opencypher.Evaluatable

	// The keys of the valueset result
	ResultKeys []string

	// The attribute ids of the nodes under this node to receive values
	ResultValues []string

	// If empty, will be set to the ContextID. The results will be added under this node.
	ResultContext string

	// The schemanode containing the valueset info
	SchemaNode *lpg.Node
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

func init() {
	RegisterNewDocGraphHook(func(g *lpg.Graph) {
		g.AddNodePropertyIndex(ValuesetContextTerm.Name, lpg.HashIndex)
		g.AddNodePropertyIndex(ValuesetContextExprTerm.Name, lpg.HashIndex)
	})
	RegisterNewLayerGraphHook(func(g *lpg.Graph) {
		g.AddNodePropertyIndex(ValuesetContextTerm.Name, lpg.HashIndex)
		g.AddNodePropertyIndex(ValuesetContextExprTerm.Name, lpg.HashIndex)
	})
}

// ValueSetInfoFromNode parses the valueset information from a
// node. Returns nil if the node does not have valueset info
func ValuesetInfoFromNode(layer *Layer, node *lpg.Node) (*ValuesetInfo, error) {
	ctxp := ValuesetContextTerm.PropertyValue(node)
	ctexpr := ValuesetContextExprTerm.PropertyValue(node)
	tablep := ValuesetTablesTerm.PropertyValue(node)
	if len(ctexpr) == 0 && len(ctxp) == 0 && len(tablep) == 0 {
		return nil, nil
	}
	ret := &ValuesetInfo{
		ContextID:     ctxp,
		TableIDs:      tablep,
		SchemaNode:    node,
		ResultValues:  ValuesetResultValuesTerm.PropertyValue(node),
		ResultKeys:    ValuesetResultKeysTerm.PropertyValue(node),
		RequestValues: ValuesetRequestValuesTerm.PropertyValue(node),
		RequestKeys:   ValuesetRequestKeysTerm.PropertyValue(node),
		ResultContext: ValuesetResultContextTerm.PropertyValue(node),
	}
	if len(ctexpr) > 0 {
		var err error
		ret.ContextExpr, err = opencypher.Parse(ctexpr)
		if err != nil {
			return nil, err
		}
	}
	ret.RequestExprs = CompileOCSemantics{}.Compiled(node, ValuesetRequestTerm.Name)
	entityRoot := GetLayerEntityRoot(node)
	if len(ret.ContextID) == 0 && ret.ContextExpr == nil {
		if entityRoot != nil {
			ret.ContextID = GetNodeID(entityRoot)
		}
	}
	if len(ret.RequestValues) > 0 {
		contextNode := layer.GetAttributeByID(ret.ContextID)
		var contextPath []*lpg.Node
		IterateDescendantsp(entityRoot, func(n *lpg.Node, p []*lpg.Node) bool {
			if GetNodeID(n) == ret.ContextID {
				contextPath = p
				return false
			}
			return true
		}, OnlyAttributeNodes, false, false)
		if contextPath == nil {
			return nil, fmt.Errorf("Not found: %s", ret.ContextID)
		}
		ret.requestSchemaPaths = make([][]*lpg.Node, len(ret.RequestValues))
		// These must be under context node
		for i, v := range ret.RequestValues {
			node := layer.GetAttributeByID(v)
			if node == nil {
				return nil, fmt.Errorf("Attribute %s not found", v)
			}
			var found []*lpg.Node
			IterateDescendantsp(contextNode, func(n *lpg.Node, p []*lpg.Node) bool {
				if n == node {
					found = make([]*lpg.Node, len(p))
					copy(found, p)
					return false
				}
				return true
			}, OnlyAttributeNodes, false, false)
			if len(found) < 1 {
				return nil, fmt.Errorf("Not found: %s", v)
			}
			ret.requestSchemaPaths[i] = make([]*lpg.Node, len(contextPath)+len(found)-1)
			copy(ret.requestSchemaPaths[i], contextPath)
			copy(ret.requestSchemaPaths[i][len(contextPath):], found[1:])
		}
	}
	return ret, nil
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
func (vsi *ValuesetInfo) GetRequest(ctx *Context, contextDocumentNode, vsiDocumentNode *lpg.Node) (map[string]string, error) {
	ret := make(map[string]string)
	if len(vsi.RequestExprs) > 0 {
		if vsiDocumentNode == nil {
			return nil, ErrInvalidValuesetSpec{Msg: fmt.Sprintf("An opencypher expression is given for %s, but there is no document node", GetNodeID(vsi.SchemaNode))}
		}
		evalctx := NewEvalContext(vsiDocumentNode.GetGraph())
		evalctx.SetVar("this", opencypher.ValueOf(vsiDocumentNode))
		for index, expr := range vsi.RequestExprs {
			result, err := expr.Evaluate(evalctx)
			if err != nil {
				return nil, err
			}
			if result.Get() == nil {
				continue
			}
			rs, ok := result.Get().(opencypher.ResultSet)
			if !ok {
				continue
			}
			if len(rs.Rows) == 0 {
				continue
			}
			if len(rs.Rows) > 1 {
				return nil, ErrInvalidValuesetSpec{Msg: fmt.Sprintf("Multiple results for expression %d", index)}
			}
			for k, v := range rs.Rows[0] {
				if opencypher.IsNamedVar(k) {
					value := v.Get()
					if value == nil {
						continue
					}
					if node, ok := value.(*lpg.Node); ok {
						ret[k], _ = GetRawNodeValue(node)
					} else {
						ret[k] = fmt.Sprint(value)
					}
				}
			}
		}
	}
	if len(vsi.RequestExprs) == 0 && len(vsi.RequestValues) == 0 {
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
	// There are some request value fields under this node. Collect them.
	for index, reqv := range vsi.RequestValues {
		if reqv == SchemaNodeIDTerm.PropertyValue(contextDocumentNode) {
			value, _ := GetRawNodeValue(contextDocumentNode)
			if len(vsi.RequestKeys) == 0 {
				ret[""] = value
			} else {
				ret[vsi.RequestKeys[index]] = value
			}
		} else {
			// Locate a child node
			// match (n)-[]->({SchemaNodeIDTerm:reqv})
			ctx.GetLogger().Debug(map[string]interface{}{"valueset.GetRequest": "locate child node", "requested": reqv, "start": contextDocumentNode})
			// If ingestAs for the schema node is "property", get the parent node
			attrRef, found := GetAttributeReferenceBySchemaPath(vsi.requestSchemaPaths[index], contextDocumentNode)
			if found {
				if !attrRef.IsProperty() {
					if len(vsi.RequestKeys) == 0 {
						ret[""], _ = GetRawNodeValue(attrRef.Node)
					} else {
						ret[vsi.RequestKeys[index]], _ = GetRawNodeValue(attrRef.Node)
					}
				} else {
					pv, _ := attrRef.AsPropertyValue()
					str, _ := pv.Value().(string)
					if len(vsi.RequestKeys) == 0 {
						ret[""] = str
					} else {
						ret[vsi.RequestKeys[index]] = str
					}
				}
			}
		}
	}
	return ret, nil
}

// getContextNodes returns the contexts node for the given document
func (vsi *ValuesetInfo) getContextNodes(g *lpg.Graph, contextID string) ([]*lpg.Node, error) {
	pattern := lpg.Pattern{
		{
			Properties: map[string]any{SchemaNodeIDTerm.Name: SchemaNodeIDTerm.MustPropertyValue(contextID)},
		},
	}
	return pattern.FindNodes(g, nil)
}

// GetContextNode returns the context node for the given document
// node. The context node must be the node itself, or an ancestor of
// the node
func (vsi *ValuesetInfo) GetContextNode(docNode *lpg.Node) (*lpg.Node, error) {
	pattern := lpg.Pattern{
		{
			Properties: map[string]any{SchemaNodeIDTerm.Name: SchemaNodeIDTerm.MustPropertyValue(vsi.ContextID)},
		},
		{
			Min: 0,
			Max: -1,
		},
		{
			Name: "start",
		},
	}
	ps := lpg.PatternSymbol{}
	ps.AddNode(docNode)
	nodes, err := pattern.FindNodes(docNode.GetGraph(), map[string]*lpg.PatternSymbol{"start": &ps})
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
func (vsi *ValuesetInfo) GetDocNodes(g *lpg.Graph) []*lpg.Node {
	pattern := lpg.Pattern{
		{
			Properties: map[string]any{SchemaNodeIDTerm.Name: SchemaNodeIDTerm.MustPropertyValue(GetNodeID(vsi.SchemaNode))},
		}}
	nodes, err := pattern.FindNodes(g, nil)
	if err != nil {
		panic(err)
	}
	return nodes
}

// contextDocumentNode is the document node that is the context
// root. resultSchemaNodeID is the node id of the result node. This
// will search children of contextDocumentNode to find the parent node
// of resultSchemaNodeID instance and if exists, resultSchemaNodeID
// itself
func (vsi *ValuesetInfo) findResultNodes(contextDocumentNode, contextSchemaNode, resultSchemaNode *lpg.Node) (resultParent *lpg.Node, resultNodes []*lpg.Node, err error) {
	path := GetAttributePath(contextSchemaNode, resultSchemaNode)
	if path == nil {
		return nil, nil, nil
	}

	idPath := make([]string, 0, len(path))
	for _, x := range path {
		idPath = append(idPath, GetNodeID(x))
	}

	resultParent = contextDocumentNode
	depth := 0
	IterateDescendants(contextDocumentNode, func(node *lpg.Node) bool {
		schemaNodeID := SchemaNodeIDTerm.PropertyValue(node)
		if len(schemaNodeID) == 0 {
			return true
		}
		for i := depth; i < len(idPath); i++ {
			if idPath[i] == schemaNodeID {
				depth = i + 1
				if i == len(idPath)-1 {
					// Found the node
					resultNodes = []*lpg.Node{path[i]}
					return false
				}
				// Found an ancestor
				resultParent = node
				break
			}
		}
		return true
	}, FollowEdgesInEntity, false)

	return
}

func (vsi *ValuesetInfo) createResultNodes(ctx *Context, builder GraphBuilder, layer *Layer, contextDocumentNode, contextSchemaNode *lpg.Node, resultSchemaNodeID string, resultValue string) error {
	// There is value. If there is a node, update it. Otherwise, insert it
	resultSchemaNode := layer.GetAttributeByID(resultSchemaNodeID)
	if resultSchemaNode == nil {
		return ErrValueset{SchemaNodeID: vsi.ContextID, Msg: fmt.Sprintf("Target schema node %s does not exist in layer", resultSchemaNodeID)}
	}
	resultParent, resultNodes, err := vsi.findResultNodes(contextDocumentNode, contextSchemaNode, resultSchemaNode)
	if err != nil {
		return err
	}

	switch len(resultNodes) {
	case 0: // insert it
		ctx.GetLogger().Debug(map[string]interface{}{"valueset.createResultNodes": "inserting", "schId": resultSchemaNodeID})
		parent := resultParent
		if parent == nil {
			parent = contextDocumentNode
		}
		switch GetIngestAs(resultSchemaNode) {
		case "node":
			_, err := EnsurePath(contextDocumentNode, nil, contextSchemaNode, resultSchemaNode, func(parentDocNode, childSchemaNode *lpg.Node) (*lpg.Node, error) {
				if GetNodeID(childSchemaNode) == resultSchemaNodeID {
					_, n, err := builder.RawValueAsNode(childSchemaNode, parentDocNode, resultValue)
					if err != nil {
						return nil, ErrValueset{SchemaNodeID: vsi.ContextID, Msg: fmt.Sprintf("Cannot create new node: %s", err.Error())}
					}
					ctx.GetLogger().Debug(map[string]interface{}{"valueset.createResultNodes": "insert", "schId": resultSchemaNode, "newNode": n})
					return n, nil
				}
				newNode := InstantiateSchemaNode(builder.targetGraph, childSchemaNode, true, map[*lpg.Node]*lpg.Node{})
				builder.GetGraph().NewEdge(parentDocNode, newNode, GetOutputEdgeLabel(parentDocNode), nil)
				return newNode, nil
			})
			if err != nil {
				return ErrValueset{SchemaNodeID: vsi.ContextID, Msg: fmt.Sprintf("Cannot create path: %s", err.Error())}
			}

		case "edge":
			_, err := builder.RawValueAsEdge(resultSchemaNode, parent, resultValue)
			if err != nil {
				return ErrValueset{SchemaNodeID: vsi.ContextID, Msg: fmt.Sprintf("Cannot create new node: %s", err.Error())}
			}
		case "property":
			err := builder.RawValueAsProperty(resultSchemaNode, []*lpg.Node{parent}, resultValue)
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

func (vsi *ValuesetInfo) ApplyValuesetResponse(ctx *Context, builder GraphBuilder, layer *Layer, contextDocumentNode, contextSchemaNode, resultContextSchemaNode *lpg.Node, result ValuesetLookupResponse) error {
	if len(result.KeyValues) == 0 {
		return nil
	}
	if len(vsi.ResultKeys) == 0 && len(vsi.ResultValues) == 0 {
		ctx.GetLogger().Debug(map[string]interface{}{"mth": "valueset.applyThisNode", "contextSchemaNode": contextSchemaNode})
		// Target is this document node
		if len(result.KeyValues) != 1 {
			return ErrValueset{SchemaNodeID: vsi.ContextID, Msg: "Multiple results from valueset lookup, but no ResultKeys specified in the schema"}
		}
		if !contextSchemaNode.HasLabel(AttributeTypeValue.Name) {
			return ErrValueset{SchemaNodeID: GetNodeID(contextSchemaNode), Msg: "Trying to set the value of a non-value node using valueset"}
		}
		for _, v := range result.KeyValues {
			SetRawNodeValue(contextDocumentNode, v)
		}
		return nil
	}
	// We have to make sure the context document node that will receive the result exists in the document
	resultContextDocumentNode := contextDocumentNode
	if resultContextSchemaNode != contextSchemaNode {
		entityRootSchemaNode := GetLayerEntityRoot(contextSchemaNode)
		entityRootDocNode := GetEntityRootNode(contextDocumentNode)
		var err error
		//		resultContextDocumentNode, err = EnsurePath(contextDocumentNode, nil, contextSchemaNode, resultContextSchemaNode, func(parentDocNode, childSchemaNode *lpg.Node) (*lpg.Node, error) {
		resultContextDocumentNode, err = EnsurePath(entityRootDocNode, nil, entityRootSchemaNode, resultContextSchemaNode, func(parentDocNode, childSchemaNode *lpg.Node) (*lpg.Node, error) {
			newNode := InstantiateSchemaNode(builder.targetGraph, childSchemaNode, true, map[*lpg.Node]*lpg.Node{})
			builder.GetGraph().NewEdge(parentDocNode, newNode, GetOutputEdgeLabel(parentDocNode), nil)
			return newNode, nil
		})
		if err != nil {
			return err
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
		if err := vsi.createResultNodes(ctx, builder, layer, resultContextDocumentNode, resultContextSchemaNode, resultNodeID, resultValue); err != nil {
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
			resultNodes := FindChildInstanceOf(resultContextDocumentNode, resultNodeID)
			for _, n := range resultNodes {
				n.DetachAndRemove()
			}
			return nil
		}
		if err := vsi.createResultNodes(ctx, builder, layer, resultContextDocumentNode, resultContextSchemaNode, resultNodeID, resultValue); err != nil {
			return err
		}
	}
	return nil
}

// ProcessByContextNode processes the value set of the given context  node and the schema node containing the vsi
func (prc *ValuesetProcessor) ProcessByContextNode(ctx *Context, builder GraphBuilder, contextDocNode, contextSchemaNode, resultContextSchemaNode, vsiDocNode *lpg.Node, vsi *ValuesetInfo) error {
	kv, err := vsi.GetRequest(ctx, contextDocNode, vsiDocNode)
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
			if err := vsi.ApplyValuesetResponse(ctx, builder, prc.layer, contextDocNode, contextSchemaNode, resultContextSchemaNode, result); err != nil {
				return err
			}
		}
	}
	return nil
}

// Process processes the value set of the given context document node and the schema node containing the vsi
func (prc *ValuesetProcessor) Process(ctx *Context, builder GraphBuilder, vsiDocNode, contextSchemaNode, resultContextSchemaNode *lpg.Node, vsi *ValuesetInfo) error {
	contextDocNode, err := vsi.GetContextNode(vsiDocNode)
	if err != nil {
		return err
	}
	ctx.GetLogger().Debug(map[string]interface{}{"mth": "valueset.process", "vsi": vsi, "contextNode": contextSchemaNode, "docNode": vsiDocNode, "contextDocNode": contextDocNode})
	return prc.ProcessByContextNode(ctx, builder, contextDocNode, contextSchemaNode, resultContextSchemaNode, vsiDocNode, vsi)
}

type ValuesetProcessor struct {
	layer      *Layer
	lookupFunc func(*Context, ValuesetLookupRequest) (ValuesetLookupResponse, error)
	vsis       []ValuesetInfo
	tables     []string
}

func NewValuesetProcessor(layer *Layer, lookupFunc func(*Context, ValuesetLookupRequest) (ValuesetLookupResponse, error), tables []string) (ValuesetProcessor, error) {
	ret := ValuesetProcessor{
		layer:      layer,
		lookupFunc: lookupFunc,
		tables:     tables,
	}
	if err := ret.init(); err != nil {
		return ret, err
	}
	return ret, nil
}

func (prc *ValuesetProcessor) init() error {
	if prc.vsis != nil {
		return nil
	}
	var err error
	seen := make(map[*lpg.Node]struct{})
	scan := func(nodes lpg.NodeIterator) {
		for nodes.Next() {
			node := nodes.Node()
			if _, exists := seen[node]; exists {
				continue
			}
			seen[node] = struct{}{}
			vsi, e := ValuesetInfoFromNode(prc.layer, node)
			if e != nil {
				err = e
				return
			}
			if vsi == nil {
				continue
			}
			if len(prc.tables) > 0 {
				hasTable := false
				for _, ptable := range prc.tables {
					for _, t := range vsi.TableIDs {
						if t == ptable {
							hasTable = true
							break
						}
					}
					if hasTable {
						break
					}
				}
				if !hasTable {
					continue
				}
			}
			prc.vsis = append(prc.vsis, *vsi)
		}
	}
	scan(prc.layer.Graph.GetNodesWithProperty(ValuesetContextTerm.Name))
	scan(prc.layer.Graph.GetNodesWithProperty(ValuesetContextExprTerm.Name))
	scan(prc.layer.Graph.GetNodesWithProperty(ValuesetTablesTerm.Name))
	if err != nil {
		return err
	}
	return nil
}

func (prc *ValuesetProcessor) ProcessGraphValueset(ctx *Context, builder GraphBuilder, vsi *ValuesetInfo) error {

	vsiDocNodes := vsi.GetDocNodes(builder.GetGraph())
	ctx.GetLogger().Debug(map[string]interface{}{"mth": "processGraphValueset", "stage": "looking up context nodes", "vsi": vsi})
	contextSchemaNode := prc.layer.GetAttributeByID(vsi.ContextID)
	if contextSchemaNode == nil {
		ctx.GetLogger().Debug(map[string]interface{}{"mth": "processGraphValueset", "stage": "No context node", "id": vsi.ContextID})
		return nil
	}
	resultContextSchemaNode := contextSchemaNode
	if len(vsi.ResultContext) > 0 {
		resultContextSchemaNode = prc.layer.GetAttributeByID(vsi.ResultContext)
		if resultContextSchemaNode == nil {
			ctx.GetLogger().Debug(map[string]interface{}{"mth": "processGraphValueset", "stage": "No result context node", "id": vsi.ResultContext})
			return nil
		}
	}
	ctx.GetLogger().Debug(map[string]interface{}{"mth": "processGraphValueset", "stage": "found context node", "vsi": vsi})
	if len(vsiDocNodes) == 0 {
		contextNodes, err := vsi.getContextNodes(builder.GetGraph(), vsi.ContextID)
		if err != nil {
			return err
		}
		ctx.GetLogger().Debug(map[string]interface{}{"mth": "processGraphValueset", "numContextNodes": len(contextNodes)})
		for _, contextNode := range contextNodes {
			if err := prc.ProcessByContextNode(ctx, builder, contextNode, contextSchemaNode, resultContextSchemaNode, nil, vsi); err != nil {
				return err
			}
		}
		return nil
	}
	for _, vsiDocNode := range vsiDocNodes {
		if err := prc.Process(ctx, builder, vsiDocNode, contextSchemaNode, resultContextSchemaNode, vsi); err != nil {
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
