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

// GetNodeTypes returns the node @type. The argument must be a map
func GetNodeTypes(node interface{}) []string {
	m, ok := node.(map[string]interface{})
	if !ok {
		return nil
	}
	arr, ok := m["@type"].([]interface{})
	if ok {
		ret := make([]string, 0, len(arr))
		for _, x := range arr {
			s, _ := x.(string)
			if len(s) > 0 {
				ret = append(ret, s)
			}
		}
		return ret
	}
	return nil
}

// GetKeyValue returns the value of the key in the node. The node must
// be a map
func GetKeyValue(key string, node interface{}) (interface{}, bool) {
	var m map[string]interface{}
	arr, ok := node.([]interface{})
	if ok {
		if len(arr) == 1 {
			m, _ = arr[0].(map[string]interface{})
		}
	} else {
		m, _ = node.(map[string]interface{})
	}
	if m == nil {
		return "", false
	}
	v, ok := m[key]
	return v, ok
}

// GetStringValue returns a string value from the node with the
// key. The node must be a map
func GetStringValue(key string, node interface{}) string {
	v, _ := GetKeyValue(key, node)
	if v == nil {
		return ""
	}
	return v.(string)
}

// GetNodeID returns the node @id. The argument must be a map
func GetNodeID(node interface{}) string {
	return GetStringValue("@id", node)
}

// GetLDListElements returns the elements of a @list node. The input can
// be a [{"@list":elements}] or {@list:elements}. If the input cannot
// be interpreted as a list, returns nil
func GetLDListElements(node interface{}) []interface{} {
	var m map[string]interface{}
	if arr, ok := node.([]interface{}); ok {
		if len(arr) == 1 {
			m, _ = arr[0].(map[string]interface{})
		}
	}
	if m == nil {
		m, _ = node.(map[string]interface{})
	}
	if len(m) == 0 {
		return []interface{}{}
	}
	lst, ok := m["@list"]
	if !ok {
		return nil
	}
	elements, ok := lst.([]interface{})
	if !ok {
		return nil
	}
	return elements
}

// If in is a @list, returns its elements
func DescendToListElements(in []interface{}) []interface{} {
	if len(in) == 1 {
		if m, ok := in[0].(map[string]interface{}); ok {
			if l, ok := m["@list"]; ok {
				if a, ok := l.([]interface{}); ok {
					return a
				}
			}
		}
	}
	return in
}

// LDMarshaler renders a graph in JSON-LD flattened format
type LDMarshaler struct {
	// If set, generates node identifiers from the given node. It should
	// be able to generate blank node IDs if the node is to be
	// represented as an RDF blank node, or if the node does not have an
	// ID.
	//
	// If not set, the default generator function uses the string node
	// id, or _b:<n> if the node does not have an id.
	NodeIDGeneratorFunc func(Node) string

	// If set, generates edge labels for the given edge. If it is not set,
	// the default is to use the edge label. If edge does not have a
	// label,
	EdgeLabelGeneratorFunc func(Edge) string
}

func (rd *LDMarshaler) Marshal(input *digraph.Graph) interface{} {
	type outputNode struct {
		id     string
		ldNode map[string]interface{}
	}
	blankNodeId := 0
	// Assign IDs to all nodes
	nodeIdMap := make(map[Node]outputNode)
	for nodes := input.AllNodes(); nodes.HasNext(); {
		node := nodes.Next().(Node)
		var idstr string
		if rd.NodeIDGeneratorFunc != nil {
			idstr = rd.NodeIDGeneratorFunc(node)
		} else {
			id := node.GetLabel()
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
		if gnode.IsAttributeNode() {
			t := gnode.GetTypes()
			if len(t) > 0 {
				arr := make([]interface{}, 0, len(t))
				for _, x := range t {
					arr = append(arr, x)
				}
				onode.ldNode["@type"] = arr
			}
		} else if gnode.IsDocumentNode() {
			if v := gnode.GetValue(); v != nil {
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
		for edges := gnode.GetAllOutgoingEdges(); edges.HasNext(); {
			edge := edges.Next().(Edge)
			var labelStr string
			if rd.EdgeLabelGeneratorFunc != nil {
				labelStr = rd.EdgeLabelGeneratorFunc(edge)
			} else {
				id := edge.GetLabel()
				if id == nil {
					labelStr = "http://www.w3.org/1999/02/22-rdf-syntax-ns#property"
				} else if labelStr = fmt.Sprint(id); len(labelStr) == 0 {
					labelStr = "http://www.w3.org/1999/02/22-rdf-syntax-ns#property"
				}
			}
			existing, ok := onode.ldNode[labelStr]
			if GetTermInfo(labelStr).IsList {
				if !ok {
					onode.ldNode[labelStr] = map[string]interface{}{"@list": []interface{}{map[string]interface{}{"@id": nodeIdMap[edge.GetTo().(Node)].id}}}
				} else {
					x := existing.(map[string]interface{})["@list"].([]interface{})
					x = append(x, map[string]interface{}{"@id": nodeIdMap[edge.GetTo().(Node)].id})
					existing.(map[string]interface{})["@list"] = x
				}
			} else {
				if !ok {
					onode.ldNode[labelStr] = map[string]interface{}{"@id": nodeIdMap[edge.GetTo().(Node)].id}
				} else if arr, ok := existing.([]interface{}); ok {
					arr = append(arr, map[string]interface{}{"@id": nodeIdMap[edge.GetTo().(Node)].id})
					onode.ldNode[labelStr] = arr
				} else {
					onode.ldNode[labelStr] = []interface{}{onode.ldNode[labelStr], map[string]interface{}{"@id": nodeIdMap[edge.GetTo().(Node)].id}}
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
		inode.graphNode = NewNode(inode.id, inode.types...)
		target.AddNode(inode.graphNode)
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
					val, vals, _, err := getValuesOrIDs(v)
					if err != nil {
						return nil, err
					}
					if len(vals) == 1 {
						inode.graphNode.SetValue(vals[0])
					} else {
						inode.graphNode.SetValue(val)
					}
				default:
					value, values, ids, err := getValuesOrIDs(v)
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
								digraph.Connect(inode.graphNode, tgt.graphNode, NewEdge(k))
							}
						}
					}
				}
			}

		default:
			inode.graphNode = NewNode(inode.id, inode.types...)
			target.AddNode(inode.graphNode)
			for k, v := range inode.node {
				value, values, ids, err := getValuesOrIDs(v)
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
							digraph.Connect(inode.graphNode, tgt.graphNode, NewEdge(k))
						}
					}
				}
			}
		}
	}
	return target, nil
}
