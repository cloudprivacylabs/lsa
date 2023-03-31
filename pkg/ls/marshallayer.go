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
	"sort"
	"strings"

	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/piprate/json-gold/ld"
)

type LDNode struct {
	Node      map[string]interface{}
	ID        string
	Types     []string
	GraphNode *lpg.Node
	processed bool
}

func getNodesFromGraph(in interface{}, interner Interner) (map[string]*LDNode, error) {
	proc := ld.NewJsonLdProcessor()
	flattened, err := proc.Flatten(in, nil, nil)
	if err != nil {
		return nil, err
	}
	if m, ok := flattened.(map[string]interface{}); ok {
		flattened = m["@graph"]
	}
	// In a flattened graph, the root object is the layer, with a link to attributes
	nodes, _ := flattened.([]interface{})
	if len(nodes) == 0 {
		return nil, MakeErrInvalidInput("", "Cannot parse layer")
	}

	inputNodes := make(map[string]*LDNode)
	for _, node := range nodes {
		m, ok := node.(map[string]interface{})
		if !ok {
			continue
		}
		inode := LDNode{Node: m}
		inode.Types = InternSlice(interner, LDGetNodeTypes(m))
		inode.ID = LDGetNodeID(m)
		inputNodes[inode.ID] = &inode
	}
	return inputNodes, nil
}

// UnmarshalLayer unmarshals a schema ar overlay
func UnmarshalLayer(in interface{}, interner Interner) (*Layer, error) {
	if interner == nil {
		interner = NewInterner()
	}
	inputNodes, err := getNodesFromGraph(in, interner)
	if err != nil {
		return nil, err
	}
	// Find the root node: there must be one node with overlay or schema type
	var rootNode *LDNode
	for _, v := range inputNodes {
		for _, t := range v.Types {
			if t == SchemaTerm || t == OverlayTerm {
				if rootNode != nil {
					return nil, MakeErrInvalidInput("Multiple root nodes")
				}
				rootNode = v
			}
		}
	}
	if rootNode == nil {
		return nil, MakeErrInvalidInput("No schema or overlay type node")
	}
	targetType := LDGetNodeValue(rootNode.Node[ValueTypeTerm])
	target := NewLayer()
	rootNode.GraphNode = target.GetLayerRootNode()
	rootNode.GraphNode.SetLabels(lpg.NewStringSet(rootNode.Types...))
	target.SetID(rootNode.ID)
	if len(target.GetID()) == 0 || target.GetID() == "./" || strings.HasPrefix(target.GetID(), "_") {
		return nil, MakeErrInvalidInput("No layer @id")
	}
	// The root node must connect to the layer node
	layerRoot := inputNodes[LDGetNodeID(rootNode.Node[LayerRootTerm])]
	if layerRoot != nil {
		layerRoot.GraphNode = target.Graph.NewNode([]string{AttributeNodeTerm}, nil)
		if ld.IsURL(layerRoot.ID) {
			SetAttributeID(layerRoot.GraphNode, layerRoot.ID)
		}
		if strings.HasPrefix(layerRoot.ID, "_") {
			return nil, MakeErrInvalidInput("layer root cannot be blank node. Enter a unique @id")
		}
		target.Graph.NewEdge(target.GetLayerRootNode(), layerRoot.GraphNode, LayerRootTerm, nil)
	}

	for _, node := range inputNodes {
		if node.GraphNode == nil {
			node.GraphNode = target.Graph.NewNode(nil, nil)
		}
	}

	if len(target.GetLayerType()) == 0 {
		return nil, ErrNotALayer
	}

	if layerRoot != nil {
		// Unmarshal all accessible nodes starting from the layer node
		if err := unmarshalAttributeNode(target, layerRoot, inputNodes, interner); err != nil {
			return nil, err
		}
	}
	// Deal with annotations
	for _, node := range inputNodes {
		if node.GraphNode == nil {
			continue
		}
		if !node.GraphNode.GetLabels().Has(AttributeNodeTerm) && node != rootNode {
			continue
		}
		// This is an attribute node
		if err := unmarshalAnnotations(target, node, inputNodes, interner); err != nil {
			return nil, err
		}
	}
	// If this is an overlay, deal with attributeOverlays
	if target.GetLayerType() == OverlayTerm {
		for _, attr := range LDGetListElements(rootNode.Node[AttributeOverlaysTerm]) {
			attrNode, ok := attr.(map[string]interface{})
			if !ok {
				continue
			}
			id := LDGetNodeID(attrNode)
			if len(id) == 0 {
				continue
			}
			inputNode, ok := inputNodes[id]
			if !ok {
				return nil, MakeErrInvalidInput(id, "Cannot follow link")
			}
			if err := unmarshalAttributeNode(target, inputNode, inputNodes, interner); err != nil {
				return nil, err
			}
			if err := unmarshalAnnotations(target, inputNode, inputNodes, interner); err != nil {
				return nil, err
			}
		}
	}

	if len(targetType) > 0 {
		target.SetValueType(targetType)
	}
	return target, nil
}

func unmarshalAttributeNode(target *Layer, inode *LDNode, allNodes map[string]*LDNode, interner Interner) error {
	if inode.processed {
		return nil
	}
	inode.processed = true
	attribute := inode.GraphNode
	types := attribute.GetLabels()
	types.Add(AttributeNodeTerm)
	if len(inode.ID) == 0 {
		return MakeErrInvalidInput("", fmt.Sprintf("Parsing %s: Attribute node without an ID: %v", target.GetID(), inode.Node))
	}
	if strings.HasPrefix(inode.ID, "_") {
		return MakeErrInvalidInput("", fmt.Sprintf("Parsing %s: Attribute node does not have an ID: %v", target.GetID(), inode.Node))
	}

	SetAttributeID(attribute, inode.ID)

	// Process the nested attribute nodes
	if arr, ok := inode.Node["@type"].([]interface{}); ok {
		for _, t := range arr {
			if str, ok := t.(string); ok {
				types.Add(interner.Intern(str))
			}
		}
	}
	attribute.SetLabels(types)

	switch {
	case types.Has(AttributeTypeObject):
		// m must be an array of attributes. It can be under a @list
		k := ObjectAttributesTerm
		val, ok := inode.Node[k]
		if !ok {
			k = ObjectAttributeListTerm
			val = inode.Node[k]
		}
		attrArray, ok := val.([]interface{})
		if !ok {
			break
		}
		if len(attrArray) == 1 {
			if m, ok := attrArray[0].(map[string]interface{}); ok {
				if l, ok := m["@list"]; ok {
					if a, ok := l.([]interface{}); ok {
						attrArray = a
					}
				}
			}
		}
		for index, attr := range attrArray {
			// This must be a link
			follow := LDGetNodeID(attr)
			attrNode := allNodes[follow]
			if attrNode == nil {
				return MakeErrInvalidInput(inode.ID, fmt.Sprintf("Parsing %s: Cannot follow link in attribute list: %v", target.GetID(), follow))
			}
			if err := unmarshalAttributeNode(target, attrNode, allNodes, interner); err != nil {
				return err
			}
			SetNodeIndex(attrNode.GraphNode, index)
			target.Graph.NewEdge(inode.GraphNode, attrNode.GraphNode, k, nil)
		}

	case types.Has(AttributeTypeReference):
		// There can be at most one reference
		oid := LDGetNodeValue(inode.Node[ReferenceTerm])
		if len(oid) == 0 {
			return MakeErrInvalidInput(inode.ID, fmt.Sprintf("Parsing %s: No references in reference node", target.GetID()))
		}
		attribute.SetProperty(ReferenceTerm, StringPropertyValue(ReferenceTerm, oid))

	case types.Has(AttributeTypeArray):
		// m must be an array of 1
		itemsArr, _ := inode.Node[ArrayItemsTerm].([]interface{})
		switch len(itemsArr) {
		case 0:
			// Allowed in an overlay
			if target.GetLayerType() == OverlayTerm {
				break
			}
			return MakeErrInvalidInput(inode.ID, fmt.Sprintf("Parsing %s: Invalid array items", target.GetID()))
		case 1:
			itemsNode := allNodes[LDGetNodeID(itemsArr[0])]
			if itemsNode == nil {
				return MakeErrInvalidInput(inode.ID, fmt.Sprintf("Parsing %s: Cannot follow link to array items", target.GetID()))
			}
			if err := unmarshalAttributeNode(target, itemsNode, allNodes, interner); err != nil {
				return err
			}
			target.Graph.NewEdge(inode.GraphNode, itemsNode.GraphNode, ArrayItemsTerm, nil)
		default:
			return MakeErrInvalidInput(inode.ID, fmt.Sprintf("Parsing %s: Multiple array items", target.GetID()))
		}

	case types.Has(AttributeTypeComposite) || types.Has(AttributeTypePolymorphic):
		var t string
		if types.Has(AttributeTypeComposite) {
			t = AllOfTerm
		} else {
			t = OneOfTerm
		}
		// m must be a list
		elements := LDGetListElements(inode.Node[t])
		if elements == nil {
			return MakeErrInvalidInput(inode.ID, fmt.Sprintf("Parsing %s: @list expected", target.GetID()))
		}
		for index, element := range elements {
			nnode := allNodes[LDGetNodeID(element)]
			if nnode == nil {
				return MakeErrInvalidInput(inode.ID, fmt.Sprintf("Parsing %s: Cannot follow link", target.GetID()))
			}
			if err := unmarshalAttributeNode(target, nnode, allNodes, interner); err != nil {
				return err
			}
			SetNodeIndex(nnode.GraphNode, index)
			target.Graph.NewEdge(inode.GraphNode, nnode.GraphNode, t, nil)
		}
	}
	types = attribute.GetLabels()
	t := FilterAttributeTypes(types.Slice())
	if len(t) > 1 {
		return ErrMultipleTypes(fmt.Sprintf("%s: %s", inode.ID, t))
	}
	return nil
}

func unmarshalAnnotations(target *Layer, node *LDNode, allNodes map[string]*LDNode, interner Interner) error {
	for key, value := range node.Node {
		key = interner.Intern(key)
		if key[0] == '@' ||
			key == ObjectAttributesTerm ||
			key == ObjectAttributeListTerm ||
			key == ReferenceTerm ||
			key == ArrayItemsTerm ||
			key == AllOfTerm ||
			key == OneOfTerm ||
			key == LayerRootTerm {
			continue
		}

		// Get the unmarshaler for the term
		if err := GetTermMarshaler(key).UnmarshalLd(target, key, value, node, allNodes, interner); err != nil {
			return err
		}
	}
	return nil
}

// Marshals the layer into an expanded jsonld document
func MarshalLayer(layer *Layer) (interface{}, error) {
	schRoot := layer.GetSchemaRootNode()
	var layerOut interface{}
	nodeMap := make(map[*lpg.Node]string)
	if schRoot != nil {
		var err error
		layerOut, err = marshalNode(layer, schRoot, nodeMap)
		if err != nil {
			return nil, err
		}
	}
	attrOverlays := make([]interface{}, 0)
	for edges := layer.GetLayerRootNode().GetEdgesWithLabel(lpg.OutgoingEdge, AttributeOverlaysTerm); edges.Next(); {
		attr := edges.Edge().GetTo()
		attrOut, err := marshalNode(layer, attr, nodeMap)
		if err != nil {
			return nil, err
		}
		attrOverlays = append(attrOverlays, attrOut)
	}
	v := map[string]interface{}{}
	if len(attrOverlays) > 0 {
		v[AttributeOverlaysTerm] = []interface{}{
			map[string]interface{}{"@list": attrOverlays}}
	}
	if layerOut != nil {
		v[LayerRootTerm] = []interface{}{layerOut}
	}
	if id := layer.GetID(); len(id) > 0 {
		v["@id"] = id
	}
	if t := layer.GetLayerType(); len(t) > 0 {
		v["@type"] = []string{t}
	}
	layer.GetLayerRootNode().ForEachProperty(func(k string, value interface{}) bool {
		if _, p := value.(*PropertyValue); !p {
			return true
		}
		val, err := GetTermMarshaler(k).MarshalLd(layer, layer.GetLayerRootNode(), k)
		if err != nil {
			return false
		}
		if val != nil {
			v[k] = val
		}
		return true
	})
	return []interface{}{v}, nil
}

func marshalNode(layer *Layer, node *lpg.Node, nodeMap map[*lpg.Node]string) (interface{}, error) {
	if nodeId, ok := nodeMap[node]; ok {
		return []interface{}{map[string]interface{}{"@id": nodeId}}, nil
	}
	nodeMap[node] = GetNodeID(node)
	m := make(map[string]interface{})
	s := GetAttributeID(node)
	if len(s) > 0 {
		m["@id"] = s
	}
	t := node.GetLabels()
	if t.Len() > 0 {
		m["@type"] = t.SortedSlice()
	}

	var err error
	node.ForEachProperty(func(k string, value interface{}) bool {
		if _, p := value.(*PropertyValue); !p {
			return true
		}
		val, err := GetTermMarshaler(k).MarshalLd(layer, node, k)
		if err != nil {
			return false
		}
		if val != nil {
			m[k] = val
		}
		return true
	})
	if err != nil {
		return nil, err
	}

	edges := lpg.EdgeSlice(node.GetEdges(lpg.OutgoingEdge))
	sort.Slice(edges, func(i, j int) bool {
		return GetNodeIndex(edges[i].GetTo()) < GetNodeIndex(edges[j].GetTo())
	})
	for _, edge := range edges {
		toNode, err := marshalNode(layer, edge.GetTo(), nodeMap)
		if err != nil {
			return nil, err
		}
		existing := m[edge.GetLabel()]
		switch edge.GetLabel() {
		case ObjectAttributeListTerm, AllOfTerm, OneOfTerm:
			if existing == nil {
				m[edge.GetLabel()] = []interface{}{map[string]interface{}{"@list": []interface{}{toNode}}}
			} else {
				listMap := existing.([]interface{})[0].(map[string]interface{})
				listMap["@list"] = append(listMap["@list"].([]interface{}), toNode)
			}

		case ObjectAttributesTerm:
			if existing == nil {
				m[ObjectAttributesTerm] = []interface{}{toNode}
			} else {
				m[ObjectAttributesTerm] = append(m[ObjectAttributesTerm].([]interface{}), toNode)
			}

		case ArrayItemsTerm:
			m[ArrayItemsTerm] = []interface{}{toNode}

		default:
			if existing == nil {
				m[edge.GetLabel()] = []interface{}{toNode}
			} else {
				m[edge.GetLabel()] = append(m[edge.GetLabel()].([]interface{}), toNode)
			}
		}
	}
	return m, nil
}
