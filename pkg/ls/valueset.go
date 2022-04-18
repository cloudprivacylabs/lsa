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

type ValuesetLookupRequest struct {
	TableIDs  []string
	KeyValues map[string]string
}

type ValuesetLookupResponse struct {
	KeyValues map[string]string
}

var (
	ValuesetSourceIDTerm      = NewTerm(LS, "vs/source", false, false, OverrideComposition, nil)
	ValuesetTablesTerm        = NewTerm(LS, "vs/valuesets", false, false, OverrideComposition, nil)
	ValuesetRequestKeysTerm   = NewTerm(LS, "vs/requestKeys", false, false, OverrideComposition, nil)
	ValuesetRequestValuesTerm = NewTerm(LS, "vs/requestValues", false, false, OverrideComposition, nil)
	ValuesetResultKeysTerm    = NewTerm(LS, "vs/resultKeys", false, false, OverrideComposition, nil)
	ValuesetResultValuesTerm  = NewTerm(LS, "vs/resultValues", false, false, OverrideComposition, nil)
)

type ValuesetInfo struct {
	// If the valueset lookup requires a single value, the attribute id
	// of the source node. Otherwise, the root node containing all the
	// required values.
	SourceID string

	// Optional lookup table IDs. If ommitted, all compatible tables are
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
}

type ErrInvalidValuesetSpec struct {
	Msg string
}

func (e ErrInvalidValuesetSpec) Error() string {
	return fmt.Sprintf("Value set error: %s", e.Msg)
}

type ErrValueset struct {
	Msg string
}

func (e ErrValueset) Error() string {
	return fmt.Sprintf("Value set processing error: %s", e.Msg)
}

// ValueSetInfoFromNode parses the valuese information from a
// node. Returns nil if the node does not have valueset info
func ValuesetInfoFromNode(node graph.Node) *ValuesetInfo {
	v, ok := node.GetProperty(ValuesetSourceIDTerm)
	if !ok {
		return nil
	}
	pv, _ := v.(*PropertyValue)
	if pv == nil {
		return nil
	}
	ret := &ValuesetInfo{
		SourceID:      pv.AsString(),
		TableIDs:      AsPropertyValue(node.GetProperty(ValuesetTablesTerm)).MustStringSlice(),
		RequestKeys:   AsPropertyValue(node.GetProperty(ValuesetRequestKeysTerm)).MustStringSlice(),
		RequestValues: AsPropertyValue(node.GetProperty(ValuesetRequestValuesTerm)).MustStringSlice(),
		ResultKeys:    AsPropertyValue(node.GetProperty(ValuesetResultKeysTerm)).MustStringSlice(),
		ResultValues:  AsPropertyValue(node.GetProperty(ValuesetResultValuesTerm)).MustStringSlice(),
	}
	if len(ret.SourceID) == 0 {
		return nil
	}
	return ret
}

func (vsi *ValuesetInfo) GetRequest(sourceDocumentNode graph.Node) (map[string]string, error) {
	value, _ := GetRawNodeValue(sourceDocumentNode)
	if len(vsi.RequestValues) == 0 {
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
		if reqv == AsPropertyValue(sourceDocumentNode.GetProperty(SchemaNodeIDTerm)).AsString() {
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
			p.Add(sourceDocumentNode)
			acc, err := pattern.FindPaths(sourceDocumentNode.GetGraph(), map[string]*graph.PatternSymbol{"n": &p})
			if err != nil {
				return nil, err
			}
			nodes := acc.GetTailNodes()
			if len(nodes) > 1 {
				return nil, ErrInvalidValuesetSpec{Msg: fmt.Sprintf("Multiple nodes instance of %s", reqv)}
			}
			if len(nodes) == 1 {
				req[vsi.RequestKeys[index]], _ = GetRawNodeValue(nodes[0])
			}
		}
	}
	return req, nil
}

func (ingester *Ingester) ProcessValueset(ctx IngestionContext, rootDocumentNode, schemaNode graph.Node) error {
	if !schemaNode.GetLabels().Has(AttributeNodeTerm) {
		return nil
	}
	vsi := ValuesetInfoFromNode(schemaNode)
	if vsi == nil {
		return nil
	}
	// Node has valueset info
	ctx.GetLogger().Debug(map[string]interface{}{"mth": "valueset.process", "schemaNodeId": GetNodeID(schemaNode)})

	// Find all instances of the source node under the given root document node
	pattern := graph.Pattern{
		{
			Name: "root",
		},
		{
			Min: 1,
			Max: -1,
		},
		{
			Properties: map[string]interface{}{SchemaNodeIDTerm: StringPropertyValue(vsi.SourceID)},
		}}
	p := graph.PatternSymbol{}
	p.Add(rootDocumentNode)
	acc, err := pattern.FindPaths(rootDocumentNode.GetGraph(), map[string]*graph.PatternSymbol{"root": &p})
	if err != nil {
		return err
	}
	nodes := acc.GetTailNodes()
	ctx.GetLogger().Debug(map[string]interface{}{"mth": "valueset.process", "numInstances": len(nodes)})

	rootDocumentNodeAttrID := AsPropertyValue(rootDocumentNode.GetProperty(SchemaNodeIDTerm)).AsString()
	rootDocumentNodeAttrs := FindNodeByID(schemaNode.GetGraph(), rootDocumentNodeAttrID)
	if len(rootDocumentNodeAttrs) != 1 {
		return ErrValueset{Msg: fmt.Sprintf("Cannot find the schema for the root node: %s", rootDocumentNodeAttrID)}
	}
	rootDocumentNodeAttr := rootDocumentNodeAttrs[0]

	// Process each source node. The source node may itself be the
	// value source, or all values are under it
	for _, sourceNode := range nodes {
		ctx.GetLogger().Debug(map[string]interface{}{"mth": "valueset.process",
			"sourceNode": sourceNode})
		kv, err := vsi.GetRequest(sourceNode)
		if err != nil {
			return err
		}
		ctx.GetLogger().Debug(map[string]interface{}{"mth": "valueset.process", "request": kv})
		if len(kv) != 0 {
			// Perform the lookup
			result, err := ingester.ValuesetFunc(ValuesetLookupRequest{
				TableIDs:  vsi.TableIDs,
				KeyValues: kv,
			})
			if err != nil {
				return err
			}
			ctx.GetLogger().Debug(map[string]interface{}{"mth": "valueset.process", "result": result})
			// If there is nonzero result, put it back into the doc
			if len(result.KeyValues) > 0 {
				// The target schema node is `schemaNode`
				// The root of the source document node is `rootDocumentNode`
				// Find the schema path from schemaNode to the schema node of rootDocumentNode
				path := GetAttributePath(rootDocumentNodeAttr, schemaNode)
				if len(path) == 0 {
					return ErrValueset{Msg: fmt.Sprintf("Cannot find schema path to %s", GetNodeID(schemaNode))}
				}
				// Tandem walk: root node is already known, and do not create the target node itself
				ictx := newIngestionContext(ctx.Context, "", rootDocumentNodeAttr)
				ictx = ictx.NewLevel(rootDocumentNode)
				currentParent := rootDocumentNode
				for i := 1; i < len(path)-1; i++ {
					// Find the node under the current root that is
					// instance of the schema path element
					children := FindChildInstanceOf(currentParent, GetNodeID(path[i]))
					if len(children) > 1 {
						return ErrValueset{Msg: fmt.Sprintf("Multiple children while setting valueset results: %s", GetNodeID(path[i]))}
					}
					if len(children) == 0 {
						// TODO: do a property ingester node creation here

					}
					currentParent = children[0]
					ictx = ictx.New("", path[i])
					ictx = ictx.NewLevel(currentParent)
				}
				// Create the new node
				ictx = ictx.New("", schemaNode)
				// Do we already have the target node?
				children := FindChildInstanceOf(currentParent, GetNodeID(schemaNode))
				if len(children) > 1 {
					return ErrValueset{Msg: fmt.Sprintf("Multiple nodes already exist for valueset target %s", GetNodeID(schemaNode))}
				}
				switch {
				case schemaNode.GetLabels().Has(AttributeTypeValue):
					if len(result.KeyValues) != 1 {
						return ErrValueset{Msg: "Multiple values cannot be set to value node: %s"}
					}
					switch len(children) {
					case 0:
						// Create the new node
						for _, v := range result.KeyValues {
							ingester.Value(ictx, v)
						}
					case 1:
						// Overwrite the existing node
						for _, v := range result.KeyValues {
							SetRawNodeValue(children[0], v)
						}
					}
				case schemaNode.GetLabels().Has(AttributeTypeObject):
					// Create the object node
					_, _, node, nictx, err := ingester.Instantiate(ictx)
					if err != nil {
						return ErrValueset{Msg: fmt.Sprintf("In %s: %s", GetNodeID(schemaNode), err)}
					}
					nictx = nictx.NewLevel(node)
					// Create children
					// TODO: For now, this only creates single-level flat structures
					if len(vsi.ResultValues) == 0 {
						return ErrValueset{Msg: fmt.Sprintf("In %s: target is an object node, but there are no ids to receive result values", GetNodeID(schemaNode))}
					}
					for i, key := range vsi.ResultKeys {
						value, ok := result.KeyValues[key]
						if !ok {
							continue
						}
						// Get the schema node for this one
						valueSchemaNode := FindNodeByID(schemaNode.GetGraph(), vsi.ResultValues[i])
						if len(valueSchemaNode) != 1 {
							return ErrValueset{Msg: fmt.Sprintf("Expecting one %s node in schema, got %d", vsi.ResultValues[i], len(valueSchemaNode))}
						}

						// Insert a value node
						_, _, node, _, err := ingester.Instantiate(nictx.New("", valueSchemaNode[0]))
						if err != nil {
							return ErrValueset{Msg: fmt.Sprintf("While setting %s: %s", GetNodeID(valueSchemaNode[0]), err)}
						}
						SetRawNodeValue(node, value)
					}

				default:
					return ErrValueset{Msg: fmt.Sprintf("Unsupported schema node type: %s", GetNodeID(schemaNode))}
				}
			}
		}
	}

	return nil
}
