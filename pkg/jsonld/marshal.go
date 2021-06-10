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
	"encoding/json"
	"errors"
	"strings"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/piprate/json-gold/ld"
)

type inputNode struct {
	node      map[string]interface{}
	id        string
	processed bool
	graphNode *ls.SchemaNode
}

// UnmarshalLayer unmarshals a schem ar overlay
func UnmarshalLayer(in interface{}) (*ls.Layer, error) {
	proc := ld.NewJsonLdProcessor()
	flattened, err := proc.Flatten(in, nil, nil)
	if err != nil {
		return nil, err
	}
	// In a flattened graph, the root object is the layer, with a link to attributes
	nodes, _ := flattened.([]interface{})
	if len(nodes) == 0 {
		return nil, ls.MakeErrInvalidInput("", "Cannot parse layer")
	}

	target := ls.NewLayer("")
	inputNodes := make(map[string]*inputNode)
	var layerNode *inputNode
	for _, node := range nodes {
		m, ok := node.(map[string]interface{})
		if !ok {
			continue
		}
		types := GetNodeTypes(m)
		id := GetNodeID(m)
		rootNode := false
		for _, t := range types {
			if t == ls.SchemaTerm || t == ls.OverlayTerm {
				if len(target.GetLayerType()) > 0 {
					return nil, ls.MakeErrInvalidInput("", "Not a valid layer")
				}
				target.SetLayerType(t)
				target.SetID(id)
				rootNode = true
			}
		}
		if len(id) > 0 {
			inode := inputNode{node: m, id: id}
			inputNodes[id] = &inode
			if rootNode {
				inode.graphNode = target.GetRoot()
				layerNode = &inode
			} else {
				inode.graphNode = target.NewNode(id, types...)
			}
		}
	}
	if len(target.GetLayerType()) == 0 {
		return nil, ls.ErrNotALayer
	}

	// Unmarshal all accessible nodes starting from the layer node
	if err := unmarshalAttributeNode(target, layerNode, inputNodes); err != nil {
		return nil, err
	}
	// Deal with annotations
	for _, node := range inputNodes {
		if !node.graphNode.HasType(ls.AttributeTypes.Attribute) {
			continue
		}
		// This is an attribute node
		unmarshalAnnotations(target, node, inputNodes)
	}
	return target, nil
}

func unmarshalAttributeNode(target *ls.Layer, inode *inputNode, allNodes map[string]*inputNode) error {
	if inode.processed {
		return nil
	}
	inode.processed = true
	attribute := inode.graphNode
	attribute.AddTypes(ls.AttributeTypes.Attribute)
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
		case ls.LayerTerms.Attributes, ls.LayerTerms.AttributeList:
			attribute.AddTypes(ls.AttributeTypes.Object)
			// m must be an array of attributes
			attrArray, ok := val.([]interface{})
			if !ok {
				return ls.MakeErrInvalidInput(inode.id, "Array of attributes expected here")
			}
			for _, attr := range attrArray {
				// This must be a link
				attrNode := allNodes[GetNodeID(attr)]
				if attrNode == nil {
					return ls.MakeErrInvalidInput(inode.id, "Cannot follow link in attribute list")
				}
				if err := unmarshalAttributeNode(target, attrNode, allNodes); err != nil {
					return err
				}
				target.AddEdge(inode.graphNode, attrNode.graphNode, ls.NewSchemaEdge(k))
			}

		case ls.LayerTerms.Reference:
			attribute.AddTypes(ls.AttributeTypes.Reference)
			// There can be at most one reference
			oid := GetNodeID(val)
			if len(oid) == 0 {
				return ls.MakeErrInvalidInput(inode.id)
			}
			attribute.Properties[ls.LayerTerms.Reference] = oid

		case ls.LayerTerms.ArrayItems:
			attribute.AddTypes(ls.AttributeTypes.Array)
			// m must be an array of 1
			itemsArr, _ := val.([]interface{})
			switch len(itemsArr) {
			case 0:
				return ls.MakeErrInvalidInput(inode.id, "Invalid array items")
			case 1:
				itemsNode := allNodes[GetNodeID(itemsArr[0])]
				if itemsNode == nil {
					return ls.MakeErrInvalidInput(inode.id, "Cannot follow link to array items")
				}
				if err := unmarshalAttributeNode(target, itemsNode, allNodes); err != nil {
					return err
				}
				target.AddEdge(inode.graphNode, itemsNode.graphNode, ls.NewSchemaEdge(k))
			default:
				return ls.MakeErrInvalidInput(inode.id, "Multiple array items")
			}

		case ls.LayerTerms.AllOf, ls.LayerTerms.OneOf:
			if k == ls.LayerTerms.AllOf {
				attribute.AddTypes(ls.AttributeTypes.Composite)
			} else {
				attribute.AddTypes(ls.AttributeTypes.Polymorphic)
			}
			// m must be a list
			elements := GetListElements(val)
			if elements == nil {
				return ls.MakeErrInvalidInput(inode.id)
			}
			for _, element := range elements {
				nnode := allNodes[GetNodeID(element)]
				if nnode == nil {
					return ls.MakeErrInvalidInput(inode.id, "Cannot follow link")
				}
				if err := unmarshalAttributeNode(target, nnode, allNodes); err != nil {
					return err
				}
				target.AddEdge(inode.graphNode, nnode.graphNode, ls.NewSchemaEdge(k))
			}
		}
	}

	t := ls.FilterAttributeTypes(attribute.GetTypes())
	switch len(t) {
	case 0:
		attribute.AddTypes(ls.AttributeTypes.Value)
	case 1:
	default:
		return ls.ErrMultipleTypes(inode.id)
	}
	return nil
}

func unmarshalAnnotations(target *ls.Layer, node *inputNode, allNodes map[string]*inputNode) {
	for key, value := range node.node {
		switch key {
		case "@id", "@type",
			ls.LayerTerms.Attributes,
			ls.LayerTerms.AttributeList,
			ls.LayerTerms.Reference,
			ls.LayerTerms.ArrayItems,
			ls.LayerTerms.AllOf,
			ls.LayerTerms.OneOf:
		default:
			if strings.HasPrefix(key, "@") {
				break
			}
			// value must be an array
			arr, ok := value.([]interface{})
			if !ok {
				break
			}
			setValue := func(v interface{}) {
				value := node.graphNode.Properties[key]
				if value == nil {
					node.graphNode.Properties[key] = v
				} else if a, ok := value.([]interface{}); ok {
					a = append(a, v)
					node.graphNode.Properties[key] = a
				} else {
					node.graphNode.Properties[key] = []interface{}{value, v}
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
								setValue(id)
							} else {
								target.AddEdge(node.graphNode, referencedNode.graphNode, ls.NewSchemaEdge(key))
							}
						}
					}
				}
			}
		}
	}
}

// Marshals the layer into a flattened jsonld document
func MarshalLayer(layer *ls.Layer) interface{} {
	return []interface{}{marshalNode(layer.GetRoot())}
}

func marshalNode(node *ls.SchemaNode) interface{} {
	m := make(map[string]interface{})
	if node.Label() != nil {
		m["@id"] = node.Label().(string)
	}
	t := node.GetTypes()
	if len(t) > 0 {
		m["@type"] = t
	}

	for k, v := range node.Properties {
		if k == ls.LayerTerms.Reference {
			m[k] = []interface{}{map[string]interface{}{"@id": v}}
		} else {
			switch val := v.(type) {
			case []interface{}:
				arr := make([]interface{}, 0)
				for _, elem := range val {
					arr = append(arr, map[string]interface{}{"@value": elem})
				}
				m[k] = arr
			default:
				m[k] = []interface{}{map[string]interface{}{"@value": val}}
			}
		}
	}

	edges := node.AllOutgoingEdges()
	for edges.HasNext() {
		edge := edges.Next().(*ls.SchemaEdge)
		switch edge.Label() {
		case ls.LayerTerms.AttributeList, ls.LayerTerms.Attributes, ls.LayerTerms.ArrayItems:
			existing := m[edge.Label().(string)]
			if existing == nil {
				existing = []interface{}{}
			}
			if arr, ok := existing.([]interface{}); ok {
				arr = append(arr, marshalNode(edge.To().(*ls.SchemaNode)))
				m[edge.Label().(string)] = arr
			}

		case ls.LayerTerms.AllOf, ls.LayerTerms.OneOf:
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
						elements = append(elements.([]interface{}), marshalNode(edge.To().(*ls.SchemaNode)))
						m["@list"] = elements
					}
				}
			}
		}
	}
	return m
}

type SchemaManifest struct {
	ID         string   `json:"@id"`
	Type       string   `json:"@type"`
	TargetType string   `json:"https://lschema.org/targetType"`
	Bundle     string   `json:"https://lschema.org/SchemaManifest#bundle,omitempty"`
	Schema     string   `json:"https://lschema.org/SchemaManifest#schema"`
	Overlays   []string `json:"https://lschema.org/SchemaManifest#overlays,omitempty"`
}

var ErrNotASchemaManifest = errors.New("Not a schema manifest")

// Unmarshals the given jsonld document into a schema manifest
func UnmarshalSchemaManifest(in interface{}) (*ls.SchemaManifest, error) {
	var m SchemaManifest
	proc := ld.NewJsonLdProcessor()
	compacted, err := proc.Compact(in, map[string]interface{}{}, nil)
	if err != nil {
		return nil, err
	}
	d, _ := json.Marshal(compacted)
	err = json.Unmarshal(d, &m)
	if err != nil {
		return nil, err
	}
	if m.Type != ls.SchemaManifestTerm {
		return nil, ErrNotASchemaManifest
	}
	return (*ls.SchemaManifest)(&m), nil
}

// MarshalSchemaNanifest returns a compact jsonld document for the manifest
func MarshalSchemaManifest(manifest *ls.SchemaManifest) interface{} {
	m := (*SchemaManifest)(manifest)
	m.Type = ls.SchemaManifestTerm
	d, _ := json.Marshal(m)
	var v interface{}
	json.Unmarshal(d, &v)
	return v
}
