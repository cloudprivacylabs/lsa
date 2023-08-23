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

package jsonld

import (
	"fmt"
	"strings"

	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/piprate/json-gold/ld"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type unmarshalInfo struct {
	id        string
	typ       string
	ldNode    map[string]any
	graphNode *lpg.Node
}

type ldId string

// compact will follow the links in the input ldNode and create a compacted view of the node
func compact(ldNode any, attributeNodes map[string]unmarshalInfo, ldNodes map[string]any, loop map[string]struct{}) (any, error) {
	arr, ok := ldNode.([]any)
	if ok {
		// Dealing with an array
		if len(arr) == 0 {
			return nil, nil
		}
		if len(arr) == 1 {
			return compact(arr[0], attributeNodes, ldNodes, loop)
		}
		ret := make([]any, 0, len(arr))
		for _, x := range arr {
			val, err := compact(x, attributeNodes, ldNodes, loop)
			if err != nil {
				return nil, err
			}
			ret = append(ret, val)
		}
		return ret, nil
	}

	m, ok := ldNode.(map[string]any)
	if ok {
		if id, ok := m["@id"]; ok {
			if idstr, ok := id.(string); ok {
				if _, has := loop[idstr]; has {
					return nil, ls.MakeErrInvalidInput(idstr, "Loop in graph")
				}
				loop[idstr] = struct{}{}
				defer func() {
					delete(loop, idstr)
				}()

				// Cannot reference an attribute from here
				if _, has := attributeNodes[idstr]; has {
					return nil, ls.MakeErrInvalidInput(idstr, "Illegal reference to an attribute node")
				}
				// Has node with this id?
				referredNode := ldNodes[idstr]
				if referredNode != nil {
					return compact(referredNode, attributeNodes, ldNodes, loop)
				}
			}
		}
		if val, ok := m["@value"]; ok {
			return compact(val, attributeNodes, ldNodes, loop)
		}
		if lst, ok := m["@list"]; ok {
			return compact(lst, attributeNodes, ldNodes, loop)
		}
		if lst, ok := m["@set"]; ok {
			return compact(lst, attributeNodes, ldNodes, loop)
		}
		ret := make(map[string]any)
		for key, value := range m {
			val, err := compact(value, attributeNodes, ldNodes, loop)
			if err != nil {
				return nil, err
			}
			ret[key] = val
		}
		return ret, nil
	}
	return ldNode, nil
}

func collectProperties(attr unmarshalInfo, attributeNodes map[string]unmarshalInfo, ldNodes map[string]any) error {
	for key, value := range attr.ldNode {
		if key == "@id" || key == "@type" {
			continue
		}
		if key == ls.LayerRootTerm.Name ||
			key == ls.ObjectAttributeListTerm.Name ||
			key == ls.ObjectAttributesTerm.Name ||
			key == ls.AllOfTerm.Name ||
			key == ls.OneOfTerm.Name ||
			key == ls.ArrayItemsTerm.Name {
			continue
		}
		// Remaining properties are to be recorded in the graph node,
		// based on the type of the property
		pvalue, err := compact(value, attributeNodes, ldNodes, map[string]struct{}{})
		if err != nil {
			return err
		}
		attr.graphNode.SetProperty(key, ls.NewPropertyValue(key, pvalue))
	}
	return nil
}

func UnmarshalLayer(in any, interner ls.Interner) (*ls.Layer, error) {
	if interner == nil {
		interner = ls.NewInterner()
	}
	ldProcessor := ld.NewJsonLdProcessor()
	flattened, err := ldProcessor.Flatten(in, nil, nil)
	if err != nil {
		return nil, err
	}
	if m, ok := flattened.(map[string]any); ok {
		flattened = m["@graph"]
	}
	nodes, _ := flattened.([]any)
	if len(nodes) == 0 {
		return nil, ls.MakeErrInvalidInput("", "Layer graph has no nodes")
	}
	layerGraph := ls.NewLayerGraph()
	// Create nodes
	attributeNodes := make(map[string]unmarshalInfo)
	var rootNode *lpg.Node
	var rootNodeLD map[string]any

	stringSlice := func(in any) []string {
		if str, ok := in.(string); ok {
			return []string{str}
		}
		arr, ok := in.([]any)
		if !ok {
			return []string{}
		}
		ret := make([]string, 0, len(arr))
		for _, x := range arr {
			if s, ok := x.(string); ok {
				ret = append(ret, s)
			}
		}
		return ret
	}

	singleObject := func(in any) map[string]any {
		if arr, ok := in.([]any); ok {
			if len(arr) == 1 {
				in = arr[0]
			}
		}
		if m, ok := in.(map[string]any); ok {
			return m
		}
		return nil
	}

	ldNodes := make(map[string]any)
	// This loop will collect all top-level nodes: schema root and all attributes
	for _, item := range nodes {
		node, _ := item.(map[string]any)
		if node == nil {
			return nil, ls.MakeErrInvalidInput("", "Invalid JSON-LD graph")
		}
		id, ok := node["@id"].(string)
		if !ok {
			return nil, ls.MakeErrInvalidInput("", fmt.Sprintf("Node ID is not a string: %v", node["@id"]))
		}
		ldNodes[id] = node
		typeSet := lpg.NewStringSet(stringSlice(node["@type"])...)
		isOverlay := typeSet.Has(ls.OverlayTerm.Name)
		isSchema := typeSet.Has(ls.SchemaTerm.Name)
		if isOverlay && isSchema {
			return nil, ls.MakeErrInvalidInput(id, "Only one of schema or overlay is allowed")
		}
		attributeTypes := ls.FilterAttributeTypes(typeSet.Slice())
		if len(attributeTypes) > 1 {
			return nil, ls.MakeErrInvalidInput(id, fmt.Sprintf("Multiple attribute types: %v", attributeTypes))
		}
		if len(attributeTypes) == 1 {
			// An attribute node
			if strings.HasPrefix(id, "_:") {
				return nil, ls.MakeErrInvalidInput(id, "Attribute cannot be blank node - missing @id")
			}
			typeSet.Add(ls.AttributeNodeTerm.Name)
			props := make(map[string]any)
			props[ls.NodeIDTerm.Name] = ls.NewPropertyValue(ls.NodeIDTerm.Name, id)
			attrNode := layerGraph.NewNode(typeSet.Slice(), props)
			attributeNodes[id] = unmarshalInfo{
				id:        id,
				typ:       attributeTypes[0],
				ldNode:    node,
				graphNode: attrNode,
			}
		} else if isSchema || isOverlay {
			// Must be the root node
			if rootNode != nil {
				return nil, ls.MakeErrInvalidInput(id, "Multiple root nodes")
			}
			if strings.HasPrefix(id, "_:") {
				return nil, ls.MakeErrInvalidInput(id, "Schema root cannot be blank node - missing @id")
			}
			rootNodeLD = node
			props := make(map[string]any)
			props[ls.NodeIDTerm.Name] = ls.NewPropertyValue(ls.NodeIDTerm.Name, id)
			rootNode = layerGraph.NewNode(typeSet.Slice(), props)
		}
	}
	if rootNode == nil {
		return nil, ls.MakeErrInvalidInput("", "Not a schema or an overlay")
	}

	// Link graphNode to its children using term edges
	link := func(root unmarshalInfo, elements []any, term string) error {
		for _, el := range elements {
			element, ok := el.(map[string]any)
			if !ok {
				return ls.MakeErrInvalidInput(root.id, "Unrecognized child element")
			}
			childId, _ := element["@id"].(string)
			if len(childId) == 0 {
				return ls.MakeErrInvalidInput(root.id, "Child element without @id")
			}
			childInfo, ok := attributeNodes[childId]
			if !ok {
				return ls.MakeErrInvalidInput(root.id, fmt.Sprintf("Cannot find child with id %s", childId))
			}
			layerGraph.NewEdge(root.graphNode, childInfo.graphNode, term, nil)
		}
		return nil
	}

	// Link attributes
	for _, attributeNode := range attributeNodes {
		switch attributeNode.typ {
		case ls.AttributeTypeObject.Name:
			if lst, ok := attributeNode.ldNode[ls.ObjectAttributeListTerm.Name].(map[string]any); ok {
				if listEl, ok := lst["@list"]; ok {
					if elements, ok := listEl.([]any); ok {
						if err := link(attributeNode, elements, ls.ObjectAttributeListTerm.Name); err != nil {
							return nil, err
						}
					}
				}
			}
			if element, ok := attributeNode.ldNode[ls.ObjectAttributesTerm.Name].(map[string]any); ok {
				if err := link(attributeNode, []any{element}, ls.ObjectAttributesTerm.Name); err != nil {
					return nil, err
				}
			}
			if elements, ok := attributeNode.ldNode[ls.ObjectAttributesTerm.Name].([]any); ok {
				if err := link(attributeNode, elements, ls.ObjectAttributesTerm.Name); err != nil {
					return nil, err
				}
			}
		case ls.AttributeTypeArray.Name:
			arrayElements := singleObject(attributeNode.ldNode[ls.ArrayItemsTerm.Name])
			if arrayElements == nil {
				return nil, ls.MakeErrInvalidInput(attributeNode.id, "Array declaration does not have array elements")
			}
			if err := link(attributeNode, []any{arrayElements}, ls.ArrayItemsTerm.Name); err != nil {
				return nil, err
			}

		case ls.AttributeTypeComposite.Name:
			if lst, ok := attributeNode.ldNode[ls.AllOfTerm.Name].(map[string]any); ok {
				if listEl, ok := lst["@list"]; ok {
					if elements, ok := listEl.([]any); ok {
						if err := link(attributeNode, elements, ls.AllOfTerm.Name); err != nil {
							return nil, err
						}
					}
				}
			}
		case ls.AttributeTypePolymorphic.Name:
			if lst, ok := attributeNode.ldNode[ls.OneOfTerm.Name].(map[string]any); ok {
				if listEl, ok := lst["@list"]; ok {
					if elements, ok := listEl.([]any); ok {
						if err := link(attributeNode, elements, ls.OneOfTerm.Name); err != nil {
							return nil, err
						}
					}
				}
			}
		}
	}

	// Collect properties for attributes
	for _, attributeNode := range attributeNodes {
		if err := collectProperties(attributeNode, attributeNodes, ldNodes); err != nil {
			return nil, err
		}
	}

	// Link root node to the layer
	layerRoot := singleObject(rootNodeLD[ls.LayerRootTerm.Name])
	if len(layerRoot) == 0 {
		// This is only valid if there are no attributes
		if len(attributeNodes) == 0 {
			return ls.NewLayerFromRootNode(rootNode), nil
		}
		return nil, ls.MakeErrInvalidInput("", "Schema has no layer")
	}
	layerRootId, _ := layerRoot["@id"].(string)
	if len(layerRootId) == 0 {
		return nil, ls.MakeErrInvalidInput("", "Schema has no layer")
	}
	layerNode, ok := attributeNodes[layerRootId]
	if !ok {
		return nil, ls.MakeErrInvalidInput("", "Schema has no layer")
	}
	layerGraph.NewEdge(rootNode, layerNode.graphNode, ls.LayerRootTerm.Name, nil)
	return ls.NewLayerFromRootNode(rootNode), nil
}

// // Marshals the layer into an expanded jsonld document
// func MarshalLayer(layer *Layer) (any, error) {
// 	schRoot := layer.GetSchemaRootNode()
// 	var layerOut any
// 	nodeMap := make(map[*lpg.Node]string)
// 	if schRoot != nil {
// 		var err error
// 		layerOut, err = marshalNode(layer, schRoot, nodeMap)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	attrOverlays := make([]any, 0)
// 	for edges := layer.GetLayerRootNode().GetEdgesWithLabel(lpg.OutgoingEdge, AttributeOverlaysTerm.Name); edges.Next(); {
// 		attr := edges.Edge().GetTo()
// 		attrOut, err := marshalNode(layer, attr, nodeMap)
// 		if err != nil {
// 			return nil, err
// 		}
// 		attrOverlays = append(attrOverlays, attrOut)
// 	}
// 	v := map[string]any{}
// 	if len(attrOverlays) > 0 {
// 		v[AttributeOverlaysTerm.Name] = []any{
// 			map[string]any{"@list": attrOverlays}}
// 	}
// 	if layerOut != nil {
// 		v[LayerRootTerm.Name] = []any{layerOut}
// 	}
// 	if id := layer.GetID(); len(id) > 0 {
// 		v["@id"] = id
// 	}
// 	if t := layer.GetLayerType(); len(t) > 0 {
// 		v["@type"] = []string{t}
// 	}
// 	layer.GetLayerRootNode().ForEachProperty(func(k string, value any) bool {
// 		if _, p := value.(PropertyValue); !p {
// 			return true
// 		}
// 		val, err := GetTermMarshaler(k).MarshalLd(layer, layer.GetLayerRootNode(), k)
// 		if err != nil {
// 			return false
// 		}
// 		if val != nil {
// 			v[k] = val
// 		}
// 		return true
// 	})
// 	return []any{v}, nil
// }

// func marshalNode(layer *Layer, node *lpg.Node, nodeMap map[*lpg.Node]string) (any, error) {
// 	if nodeId, ok := nodeMap[node]; ok {
// 		return []any{map[string]any{"@id": nodeId}}, nil
// 	}
// 	nodeMap[node] = GetNodeID(node)
// 	m := make(map[string]any)
// 	s := GetAttributeID(node)
// 	if len(s) > 0 {
// 		m["@id"] = s
// 	}
// 	t := node.GetLabels()
// 	if t.Len() > 0 {
// 		m["@type"] = t.SortedSlice()
// 	}

// 	var err error
// 	node.ForEachProperty(func(k string, value any) bool {
// 		if _, p := value.(PropertyValue); !p {
// 			return true
// 		}
// 		val, err := GetTermMarshaler(k).MarshalLd(layer, node, k)
// 		if err != nil {
// 			return false
// 		}
// 		if val != nil {
// 			m[k] = val
// 		}
// 		return true
// 	})
// 	if err != nil {
// 		return nil, err
// 	}

// 	edges := lpg.EdgeSlice(node.GetEdges(lpg.OutgoingEdge))
// 	sort.Slice(edges, func(i, j int) bool {
// 		return GetNodeIndex(edges[i].GetTo()) < GetNodeIndex(edges[j].GetTo())
// 	})
// 	for _, edge := range edges {
// 		toNode, err := marshalNode(layer, edge.GetTo(), nodeMap)
// 		if err != nil {
// 			return nil, err
// 		}
// 		existing := m[edge.GetLabel()]
// 		switch edge.GetLabel() {
// 		case ObjectAttributeListTerm.Name, AllOfTerm.Name, OneOfTerm.Name:
// 			if existing == nil {
// 				m[edge.GetLabel()] = []any{map[string]any{"@list": []any{toNode}}}
// 			} else {
// 				listMap := existing.([]any)[0].(map[string]any)
// 				listMap["@list"] = append(listMap["@list"].([]any), toNode)
// 			}

// 		case ObjectAttributesTerm.Name:
// 			if existing == nil {
// 				m[ObjectAttributesTerm.Name] = []any{toNode}
// 			} else {
// 				m[ObjectAttributesTerm.Name] = append(m[ObjectAttributesTerm.Name].([]any), toNode)
// 			}

// 		case ArrayItemsTerm.Name:
// 			m[ArrayItemsTerm.Name] = []any{toNode}

// 		default:
// 			if existing == nil {
// 				m[edge.GetLabel()] = []any{toNode}
// 			} else {
// 				m[edge.GetLabel()] = append(m[edge.GetLabel()].([]any), toNode)
// 			}
// 		}
// 	}
// 	return m, nil
// }
