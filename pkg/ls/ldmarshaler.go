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

	"github.com/bserdar/digraph"
)

// LDMarshaler renders a graph in JSON-LD flattened format
type LDMarshaler struct {
	// If set, generates node identifiers from the given node. It should
	// be able to generate blank node IDs if the node is to be
	// represented as an RDF blank node, or if the node does not have an
	// ID.
	//
	// If not set, the default generator function uses the string node
	// id, or _b:<n> if the node does not have an id.
	NodeIDGeneratorFunc func(digraph.Node) string

	// If set, generates edge labels for the given edge. If it is not set,
	// the default is to use the edge label. If edge does not have a
	// label,
	EdgeLabelGeneratorFunc func(digraph.Edge) string
}

func (rd *LDMarshaler) Marshal(input *digraph.Graph) interface{} {
	type outputNode struct {
		id     string
		ldNode map[string]interface{}
	}
	blankNodeId := 0
	// Assign IDs to all nodes
	nodeIdMap := make(map[digraph.Node]outputNode)
	for nodes := input.AllNodes(); nodes.HasNext(); {
		node := nodes.Next()
		var idstr string
		if rd.NodeIDGeneratorFunc != nil {
			idstr = rd.NodeIDGeneratorFunc(node)
		} else {
			id := node.Label()
			if id == nil {
				idstr = fmt.Sprintf("_b:%d", blankNodeId)
				blankNodeId++
			} else if idstr = fmt.Sprint(id); len(idstr) == 0 {
				idstr = fmt.Sprintf("_b:%d", blankNodeId)
				blankNodeId++
			}
		}
		outNode := outputNode{ldNode: map[string]interface{}{"@id": idstr}, id: idstr}
		nodeIdMap[node] = outNode
	}

	type propertiesSupport interface {
		GetProperties() map[string]*PropertyValue
	}

	// Process the properties
	for gnode, onode := range nodeIdMap {
		switch n := gnode.(type) {
		case LayerNode:
			t := n.GetTypes()
			if len(t) > 0 {
				arr := make([]interface{}, 0, len(t))
				for _, x := range t {
					arr = append(arr, x)
				}
				onode.ldNode["@type"] = arr
			}
		case DocumentNode:
			if v := n.GetValue(); v != nil {
				onode.ldNode[AttributeValueTerm] = v
			}
			types := onode.ldNode["@type"]
			if types == nil {
				types = DocumentNodeTerm
			} else if s, ok := types.(string); ok {
				if s != DocumentNodeTerm {
					types = []interface{}{s, DocumentNodeTerm}
				}
			} else if arr, ok := types.([]interface{}); ok {
				hasType := false
				for _, c := range arr {
					if c == DocumentNodeTerm {
						hasType = true
						break
					}
				}
				if !hasType {
					types = append(arr, DocumentNodeTerm)
				}
			}
			onode.ldNode["@type"] = types
		}
		if prop, ok := gnode.(propertiesSupport); ok {
			for key, pvalue := range prop.GetProperties() {
				if pvalue.IsString() {
					onode.ldNode[key] = pvalue.AsString()
				} else if pvalue.IsStringSlice() {
					onode.ldNode[key] = pvalue.AsInterfaceSlice()
				}
			}
		}
	}

	// Process outgoing edges
	for gnode, onode := range nodeIdMap {
		for edges := gnode.AllOutgoingEdges(); edges.HasNext(); {
			edge := edges.Next()
			var labelStr string
			if rd.EdgeLabelGeneratorFunc != nil {
				labelStr = rd.EdgeLabelGeneratorFunc(edge)
			} else {
				id := edge.Label()
				if id == nil {
					labelStr = "http://www.w3.org/1999/02/22-rdf-syntax-ns#property"
				} else if labelStr = fmt.Sprint(id); len(labelStr) == 0 {
					labelStr = "http://www.w3.org/1999/02/22-rdf-syntax-ns#property"
				}
			}
			existing, ok := onode.ldNode[labelStr]
			if GetTermInfo(labelStr).IsList {
				if !ok {
					onode.ldNode[labelStr] = map[string]interface{}{"@list": []interface{}{map[string]interface{}{"@id": nodeIdMap[edge.To()].id}}}
				} else {
					x := existing.(map[string]interface{})["@list"].([]interface{})
					x = append(x, map[string]interface{}{"@id": nodeIdMap[edge.To()].id})
					existing.(map[string]interface{})["@list"] = x
				}
			} else {
				if !ok {
					onode.ldNode[labelStr] = map[string]interface{}{"@id": nodeIdMap[edge.To()].id}
				} else if arr, ok := existing.([]interface{}); ok {
					arr = append(arr, map[string]interface{}{"@id": nodeIdMap[edge.To()].id})
					onode.ldNode[labelStr] = arr
				} else {
					onode.ldNode[labelStr] = []interface{}{onode.ldNode[labelStr], map[string]interface{}{"@id": nodeIdMap[edge.To()].id}}
				}
			}
		}
	}
	graph := make([]interface{}, 0, len(nodeIdMap))
	for _, v := range nodeIdMap {
		graph = append(graph, v.ldNode)
	}
	return map[string]interface{}{"@graph": graph}
}

// Unmarshal a graph
func UnmarshalGraph(input interface{}) (*digraph.Graph, error) {
	inputNodes, err := getNodesFromGraph(input)
	if err != nil {
		return nil, err
	}
	hasType := func(t string, types []string) bool {
		for _, x := range types {
			if t == x {
				return true
			}
		}
		return false
	}
	target := digraph.New()

	// Generate a graph node for each input node to populate the graph
	for _, inode := range inputNodes {
		switch {
		case hasType(DocumentNodeTerm, inode.types):
			// A document node
			inode.docNode = NewBasicDocumentNode(inode.id)
			target.AddNode(inode.docNode)

		default:
			inode.graphNode = NewLayerNode(inode.id, inode.types...)
			target.AddNode(inode.graphNode)
		}
	}

	// Deal with properties and edges
	type propertySupport interface {
		GetProperties() map[string]*PropertyValue
	}
	for _, inode := range inputNodes {
		switch {
		case hasType(DocumentNodeTerm, inode.types):
			// A document node
			for k, v := range inode.node {
				if v == nil {
					continue
				}
				switch k {
				case AttributeValueTerm:
					val, vals, _, err := GetValuesOrIDs(v)
					if err != nil {
						return nil, err
					}
					if len(vals) == 1 {
						inode.docNode.SetValue(vals[0])
					} else {
						inode.docNode.SetValue(val)
					}
				default:
					value, values, ids, err := GetValuesOrIDs(v)
					if err != nil {
						return nil, err
					}
					if values == nil && ids == nil {
						inode.docNode.GetProperties()[k] = StringPropertyValue(value)
					} else if values != nil {
						inode.docNode.GetProperties()[k] = StringSlicePropertyValue(values)
					} else if ids != nil {
						for _, id := range ids {
							tgt := inputNodes[id]
							if tgt != nil {
								var t digraph.Node
								if tgt.graphNode != nil {
									t = tgt.graphNode
								} else {
									t = tgt.docNode
								}
								target.AddEdge(inode.docNode, t, digraph.NewBasicEdge(k, nil))
							}
						}
					}
				}
			}

		default:
			inode.graphNode = NewLayerNode(inode.id, inode.types...)
			target.AddNode(inode.graphNode)
			for k, v := range inode.node {
				value, values, ids, err := GetValuesOrIDs(v)
				if err != nil {
					return nil, err
				}
				if values == nil && ids == nil {
					inode.graphNode.GetProperties()[k] = StringPropertyValue(value)
				} else if values != nil {
					inode.graphNode.GetProperties()[k] = StringSlicePropertyValue(values)
				} else if ids != nil {
					for _, id := range ids {
						tgt := inputNodes[id]
						if tgt != nil {
							var t digraph.Node
							if tgt.graphNode != nil {
								t = tgt.graphNode
							} else {
								t = tgt.docNode
							}
							target.AddEdge(inode.graphNode, t, digraph.NewBasicEdge(k, nil))
						}
					}
				}
			}
		}
	}
	return target, nil
}
