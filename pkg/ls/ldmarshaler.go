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

// LDMarshaler renders a graph in JSON-LD flattened format
type LDMarshaler struct {
	// If set, generates node identifiers from the given node. It should
	// be able to generate blank node IDs if the node is to be
	// represented as an RDF blank node, or if the node does not have an
	// ID.
	//
	// If not set, the default generator function uses the string node
	// id, or _b:<n> if the node does not have an id.
	NodeIDGeneratorFunc func(graph.Node) string

	// If set, generates edge labels for the given edge. If it is not set,
	// the default is to use the edge label. If edge does not have a
	// label,
	EdgeLabelGeneratorFunc func(graph.Edge) string
}

func (rd *LDMarshaler) Marshal(input graph.Graph) interface{} {
	type outputNode struct {
		id     string
		ldNode map[string]interface{}
	}
	blankNodeId := 0
	// Assign IDs to all nodes
	nodeIdMap := make(map[graph.Node]outputNode)
	for nodes := input.GetNodes(); nodes.Next(); {
		node := nodes.Node()
		var idstr string
		if rd.NodeIDGeneratorFunc != nil {
			idstr = rd.NodeIDGeneratorFunc(node)
		} else {
			id := GetAttributeID(node)
			if len(id) == 0 {
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
		t := gnode.GetLabels()
		if t.Len() > 0 {
			if t.Len() == 1 {
				onode.ldNode["@type"] = t.Slice()[0]
			} else {
				arr := make([]interface{}, 0, t.Len())
				for _, x := range t.Slice() {
					arr = append(arr, x)
				}
				onode.ldNode["@type"] = arr
			}
		}
		if IsDocumentNode(gnode) {
			if v, ok := GetRawNodeValue(gnode); ok {
				onode.ldNode[NodeValueTerm] = v
			}
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
		for edges := gnode.GetEdges(graph.OutgoingEdge); edges.Next(); {
			edge := edges.Edge()
			var labelStr string
			if rd.EdgeLabelGeneratorFunc != nil {
				labelStr = rd.EdgeLabelGeneratorFunc(edge)
			} else {
				labelStr = edge.GetLabel()
				if len(labelStr) == 0 {
					labelStr = "http://www.w3.org/1999/02/22-rdf-syntax-ns#property"
				}
			}
			existing, ok := onode.ldNode[labelStr]
			if GetTermInfo(labelStr).IsList {
				if !ok {
					onode.ldNode[labelStr] = map[string]interface{}{"@list": []interface{}{map[string]interface{}{"@id": nodeIdMap[edge.GetTo()].id}}}
				} else {
					x := existing.(map[string]interface{})["@list"].([]interface{})
					x = append(x, map[string]interface{}{"@id": nodeIdMap[edge.GetTo()].id})
					existing.(map[string]interface{})["@list"] = x
				}
			} else {
				if !ok {
					onode.ldNode[labelStr] = map[string]interface{}{"@id": nodeIdMap[edge.GetTo()].id}
				} else if arr, ok := existing.([]interface{}); ok {
					arr = append(arr, map[string]interface{}{"@id": nodeIdMap[edge.GetTo()].id})
					onode.ldNode[labelStr] = arr
				} else {
					onode.ldNode[labelStr] = []interface{}{onode.ldNode[labelStr], map[string]interface{}{"@id": nodeIdMap[edge.GetTo()].id}}
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

// getValuesOrIDs returns the @values, or @ids contained in the interface
// This can be a single value, an array, or a @list
func getValuesOrIDs(in interface{}) (value string, values, ids []string, err error) {
	if in == nil {
		return
	}
	if arr, ok := in.([]interface{}); ok {
		if len(arr) == 1 {
			return getValuesOrIDs(arr[0])
		}
		for _, el := range arr {
			val, vals, i, e := getValuesOrIDs(el)
			if e != nil {
				return "", nil, nil, e
			}
			if vals == nil && i == nil {
				values = append(values, val)
			} else {
				if vals != nil {
					values = append(values, vals...)
				}
				if i != nil {
					ids = append(ids, i...)
				}
			}
		}
		if len(values) > 0 && len(ids) > 0 {
			return "", nil, nil, ErrInvalidJsonLdGraph
		}
		return
	}

	if m, ok := in.(map[string]interface{}); ok {
		if lst, ok := m["@list"]; ok {
			return getValuesOrIDs(lst)
		}
		if id, ok := m["@id"]; ok {
			return "", nil, []string{fmt.Sprint(id)}, nil
		}
		if v, ok := m["@value"]; ok {
			return fmt.Sprint(v), nil, nil, nil
		}
		return "", nil, nil, ErrInvalidJsonLdGraph
	}
	value = fmt.Sprint(in)
	return
}

// UnmarshalJSONLDGraph Unmarshals a graph in JSON-LD format
func UnmarshalJSONLDGraph(input interface{}, target graph.Graph, interner Interner) error {
	if interner == nil {
		interner = NewInterner()
	}
	inputNodes, err := getNodesFromGraph(input, interner)
	if err != nil {
		return err
	}
	hasType := func(t string, types []string) bool {
		for _, x := range types {
			if t == x {
				return true
			}
		}
		return false
	}

	// Generate a graph node for each input node to populate the graph
	for _, inode := range inputNodes {
		inode.GraphNode = target.NewNode(inode.Types, nil)
		if len(inode.ID) > 0 {
			SetNodeID(inode.GraphNode, inode.ID)
		}
	}

	// Deal with properties and edges
	type propertySupport interface {
		GetProperties() map[string]*PropertyValue
	}
	for _, inode := range inputNodes {
		switch {
		case hasType(DocumentNodeTerm, inode.Types):
			// A document node
			for k, v := range inode.Node {
				if v == nil {
					continue
				}
				switch k {
				case "@id", "@type":
				case NodeValueTerm:
					val, vals, _, err := getValuesOrIDs(v)
					if err != nil {
						return err
					}
					if len(vals) == 1 {
						SetRawNodeValue(inode.GraphNode, vals[0])
					} else {
						SetRawNodeValue(inode.GraphNode, val)
					}
				default:
					value, values, ids, err := getValuesOrIDs(v)
					if err != nil {
						return err
					}
					if values == nil && ids == nil {
						inode.GraphNode.SetProperty(k, StringPropertyValue(value))
					} else if values != nil {
						inode.GraphNode.SetProperty(k, StringSlicePropertyValue(values))
					} else if ids != nil {
						for _, id := range ids {
							tgt := inputNodes[id]
							if tgt != nil {
								target.NewEdge(inode.GraphNode, tgt.GraphNode, k, nil)
							}
						}
					}
				}
			}

		default:
			for k, v := range inode.Node {
				value, values, ids, err := getValuesOrIDs(v)
				if err != nil {
					return err
				}
				if values == nil && ids == nil {
					inode.GraphNode.SetProperty(k, StringPropertyValue(value))
				} else if values != nil {
					inode.GraphNode.SetProperty(k, StringSlicePropertyValue(values))
				} else if ids != nil {
					for _, id := range ids {
						tgt := inputNodes[id]
						if tgt != nil {
							target.NewEdge(inode.GraphNode, tgt.GraphNode, k, nil)
						}
					}
				}
			}
		}
	}
	return nil
}
