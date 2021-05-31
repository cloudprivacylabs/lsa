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
	"strings"

	"github.com/bserdar/digraph"
	"github.com/cloudprivacylabs/lsa/pkg/layers"
	"github.com/piprate/json-gold/ld"
)

type inputNode struct {
	node      map[string]interface{}
	id        string
	processed bool
	types     map[string]struct{}
	graphNode *digraph.Node
}

// UnmarshalLayer unmarshals a schem ar overlay
func UnmarshalLayer(in interface{}) (*layers.Layer, error) {
	target := digraph.New()
	proc := ld.NewJsonLdProcessor()
	flattened, err := proc.Flatten(in, nil, nil)
	if err != nil {
		return nil, err
	}
	// In a flattened graph, the root object is the layer, with a link to attributes
	nodes, _ := flattened.([]interface{})
	if len(nodes) == 0 {
		return nil, layers.MakeErrInvalidInput("", "Cannot parse layer")
	}
	inputNodes := make(map[string]*inputNode)
	for _, node := range nodes {
		m, ok := node.(map[string]interface{})
		if !ok {
			continue
		}
		id := GetNodeID(m)
		if len(id) > 0 {
			inode := inputNode{node: m, id: id, types: make(map[string]struct{})}
			inputNodes[id] = &inode
			for _, t := range GetNodeTypes(inode.node) {
				inode.types[t] = struct{}{}
			}
			inode.graphNode = target.NewNode(id, layers.NewSchemaNode())
		}
	}

	// We have to find the object that has the type Schema or Overlay
	var layerNode *inputNode
	for _, node := range inputNodes {
		if _, ok := node.types[layers.SchemaTerm]; ok {
			if layerNode != nil {
				return nil, layers.MakeErrInvalidInput(node.id)
			}
			layerNode = node
		}
		if _, ok := node.types[layers.OverlayTerm]; ok {
			if layerNode != nil {
				return nil, layers.MakeErrInvalidInput(node.id)
			}
			layerNode = node
		}
	}
	if layerNode == nil {
		return nil, layers.ErrNotALayer
	}

	// Unmarshal all accessible nodes starting from the layer node
	if err := unmarshalAttributeNode(target, layerNode, inputNodes); err != nil {
		return nil, err
	}
	// Deal with annotations
	for _, node := range inputNodes {
		schNode, ok := node.graphNode.Payload.(*layers.SchemaNode)
		if !ok {
			continue
		}
		if !schNode.HasType(layers.AttributeTypes.Attribute) {
			continue
		}
		// This is an attribute node
		unmarshalAnnotations(target, node, inputNodes)
	}
	ret := &layers.Layer{Graph: target, RootNode: layerNode.graphNode}
	return ret, nil
}

func unmarshalAttributeNode(target *digraph.Graph, inode *inputNode, allNodes map[string]*inputNode) error {
	if inode.processed {
		return nil
	}
	inode.processed = true
	attribute := layers.NewSchemaNode(layers.AttributeTypes.Attribute)
	inode.graphNode.Payload = attribute
	// Process the nested attribute nodes
	for k, val := range inode.node {
		switch k {
		case "@id":
		case "@type":
			if arr, ok := val.([]interface{}); ok {
				for _, t := range arr {
					if str, ok := t.(string); ok {
						attribute.AddTypes(str)
					}
				}
			}
		case layers.TypeTerms.Attributes, layers.TypeTerms.AttributeList:
			attribute.AddTypes(layers.AttributeTypes.Object)
			// m must be an array of attributes
			attrArray, ok := val.([]interface{})
			if !ok {
				return layers.MakeErrInvalidInput(inode.id, "Array of attributes expected here")
			}
			attrArray = DescendToListElements(attrArray)
			for _, attr := range attrArray {
				// This must be a link
				attrNode := allNodes[GetNodeID(attr)]
				if attrNode == nil {
					return layers.MakeErrInvalidInput(inode.id, "Cannot follow link in attribute list")
				}
				if err := unmarshalAttributeNode(target, attrNode, allNodes); err != nil {
					return err
				}
				target.NewEdge(inode.graphNode, attrNode.graphNode, k, nil)
			}

		case layers.TypeTerms.Reference:
			attribute.AddTypes(layers.AttributeTypes.Reference)
			// There can be at most one reference
			oid := GetNodeID(val)
			if len(oid) == 0 {
				return layers.MakeErrInvalidInput(inode.id)
			}
			attribute.Properties[layers.TypeTerms.Reference] = layers.IRI(oid)

		case layers.TypeTerms.ArrayItems:
			attribute.AddTypes(layers.AttributeTypes.Array)
			// m must be an array of 1
			itemsArr, _ := val.([]interface{})
			switch len(itemsArr) {
			case 0:
				return layers.MakeErrInvalidInput(inode.id, "Invalid array items")
			case 1:
				itemsNode := allNodes[GetNodeID(itemsArr[0])]
				if itemsNode == nil {
					return layers.MakeErrInvalidInput(inode.id, "Cannot follow link to array items")
				}
				if err := unmarshalAttributeNode(target, itemsNode, allNodes); err != nil {
					return err
				}
				target.NewEdge(inode.graphNode, itemsNode.graphNode, k, nil)
			default:
				return layers.MakeErrInvalidInput(inode.id, "Multiple array items")
			}

		case layers.TypeTerms.AllOf, layers.TypeTerms.OneOf:
			if k == layers.TypeTerms.AllOf {
				attribute.AddTypes(layers.AttributeTypes.Composite)
			} else {
				attribute.AddTypes(layers.AttributeTypes.Polymorphic)
			}
			// m must be a list
			elements := GetListElements(val)
			if elements == nil {
				return layers.MakeErrInvalidInput(inode.id)
			}
			for _, element := range elements {
				nnode := allNodes[GetNodeID(element)]
				if nnode == nil {
					return layers.MakeErrInvalidInput(inode.id, "Cannot follow link")
				}
				if err := unmarshalAttributeNode(target, nnode, allNodes); err != nil {
					return err
				}
				target.NewEdge(inode.graphNode, nnode.graphNode, k, nil)
			}
		}
	}

	t := layers.GetAttributeTypes(attribute.GetTypes())
	switch len(t) {
	case 0:
		attribute.AddTypes(layers.AttributeTypes.Value)
	case 1:
	default:
		return layers.ErrMultipleTypes(inode.id)
	}
	return nil
}

func unmarshalAnnotations(target *digraph.Graph, node *inputNode, allNodes map[string]*inputNode) {
	for key, value := range node.node {
		switch key {
		case "@id", "@type",
			layers.TypeTerms.Attributes,
			layers.TypeTerms.AttributeList,
			layers.TypeTerms.Reference,
			layers.TypeTerms.ArrayItems,
			layers.TypeTerms.AllOf,
			layers.TypeTerms.OneOf:
		default:
			if strings.HasPrefix(key, "@") {
				break
			}
			// value must be an array
			arr, ok := value.([]interface{})
			if !ok {
				break
			}
			payload := node.graphNode.Payload.(*layers.SchemaNode)
			setValue := func(v interface{}) {
				value := payload.Properties[key]
				if value == nil {
					payload.Properties[key] = v
				} else if a, ok := value.([]interface{}); ok {
					a = append(a, v)
					payload.Properties[key] = a
				} else {
					payload.Properties[key] = []interface{}{value, v}
				}
			}
			// If list, descend to its elements
			arr = DescendToListElements(arr)
			for _, element := range arr {
				m, ok := element.(map[string]interface{})
				if !ok {
					continue
				}
				// This is a value or an @id
				if len(m) == 1 {
					if v := m["@value"]; v != nil {
						setValue(v)
					} else if v := m["@id"]; v != nil {
						if id, ok := v.(string); ok {
							// Is this a link?
							referencedNode := allNodes[id]
							if referencedNode == nil {
								setValue(layers.IRI(id))
							} else {
								target.NewEdge(node.graphNode, referencedNode.graphNode, key, nil)
							}
						}
					}
				}
			}
		}
	}
}

// Marshals the layer into a flattened jsonld document
func MarshalLayer(layer *layers.Layer) interface{} {
	return []interface{}{marshalNode(layer.RootNode)}
}

func marshalNode(node *digraph.Node) interface{} {
	m := make(map[string]interface{})
	if node.Label() != nil {
		m["@id"] = node.Label().(string)
	}
	payload, _ := node.Payload.(*layers.SchemaNode)
	if payload != nil {
		t := payload.GetTypes()
		if len(t) > 0 {
			m["@type"] = t
		}

		for k, v := range payload.Properties {
			switch val := v.(type) {
			case layers.IRI:
				m[k] = []interface{}{map[string]interface{}{"@id": string(val)}}
			case []interface{}:
				arr := make([]interface{}, 0)
				for _, elem := range val {
					switch t := elem.(type) {
					case layers.IRI:
						arr = append(arr, map[string]interface{}{"@id": string(t)})
					default:
						arr = append(arr, map[string]interface{}{"@value": t})
					}
				}
				m[k] = arr
			default:
				m[k] = []interface{}{map[string]interface{}{"@value": val}}
			}
		}

		edges := node.AllOutgoingEdges()
		for edges.HasNext() {
			edge := edges.Next()
			switch edge.Label() {
			case layers.TypeTerms.AttributeList, layers.TypeTerms.Attributes, layers.TypeTerms.ArrayItems:
				existing := m[edge.Label().(string)]
				if existing == nil {
					existing = []interface{}{}
				}
				if arr, ok := existing.([]interface{}); ok {
					arr = append(arr, marshalNode(edge.To()))
					m[edge.Label().(string)] = arr
				}

			case layers.TypeTerms.AllOf, layers.TypeTerms.OneOf:
				existing := m[edge.Label().(string)]
				if existing == nil {
					existing = []interface{}{}
				}
				if arr, ok := existing.([]interface{}); ok {
					if len(arr) == 0 {
						arr = []interface{}{map[string]interface{}{"@list": []interface{}{}}}
					}
					if len(arr) == 1 {
						m, ok := arr[0].(map[string]interface{})
						if ok {
							elements := m["@list"]
							if elements == nil {
								elements = []interface{}{}
								m["@list"] = elements
							}
							elements = append(elements.([]interface{}), marshalNode(edge.To()))
							m["@list"] = elements
						}
					}
				}
			}
		}
	}
	return m
}
