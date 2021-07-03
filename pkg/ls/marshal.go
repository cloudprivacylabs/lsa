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
	"errors"
	"fmt"
	"strings"

	"github.com/piprate/json-gold/ld"
)

type inputNode struct {
	node      map[string]interface{}
	id        string
	types     []string
	processed bool
	graphNode LayerNode
}

// UnmarshalLayer unmarshals a schem ar overlay
func UnmarshalLayer(in interface{}) (*Layer, error) {
	proc := ld.NewJsonLdProcessor()
	flattened, err := proc.Flatten(in, nil, nil)
	if err != nil {
		return nil, err
	}
	// In a flattened graph, the root object is the layer, with a link to attributes
	nodes, _ := flattened.([]interface{})
	if len(nodes) == 0 {
		return nil, MakeErrInvalidInput("", "Cannot parse layer")
	}

	inputNodes := make(map[string]*inputNode)
	for _, node := range nodes {
		m, ok := node.(map[string]interface{})
		if !ok {
			continue
		}
		inode := inputNode{node: m}
		inode.types = GetNodeTypes(m)
		inode.id = GetNodeID(m)
		inputNodes[inode.id] = &inode
	}

	// Find the root node: there must be one node with overlay or schema type
	var rootNode *inputNode
	for _, v := range inputNodes {
		for _, t := range v.types {
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
	// The root node must connect to the layer node
	layerRoot := inputNodes[GetNodeID(rootNode.node[LayerRootTerm])]
	targetType := GetNodeID(rootNode.node[TargetType])
	target := NewLayer()
	rootNode.graphNode = target.GetLayerInfoNode()
	rootNode.graphNode.SetTypes(rootNode.types...)
	rootNode.graphNode.SetID(rootNode.id)
	if layerRoot != nil {
		if ld.IsURL(layerRoot.id) {
			layerRoot.graphNode = target.NewNode(layerRoot.id)
		} else {
			layerRoot.graphNode = target.NewNode("")
		}
		target.GetLayerInfoNode().Connect(layerRoot.graphNode, LayerRootTerm)
	}

	for _, node := range inputNodes {
		if node.graphNode == nil {
			node.graphNode = target.NewNode(node.id)
		}
	}

	unmarshalAnnotations(target, rootNode, inputNodes)

	if len(target.GetLayerType()) == 0 {
		return nil, ErrNotALayer
	}

	if layerRoot != nil {
		// Unmarshal all accessible nodes starting from the layer node
		if err := unmarshalAttributeNode(target, layerRoot, inputNodes); err != nil {
			return nil, err
		}
	}
	// Deal with annotations
	for _, node := range inputNodes {
		if node.graphNode != nil {
			if !node.graphNode.HasType(AttributeTypes.Attribute) {
				continue
			}
		}
		// This is an attribute node
		unmarshalAnnotations(target, node, inputNodes)
	}
	if len(targetType) > 0 {
		target.SetTargetType(targetType)
	}
	return target, nil
}

func unmarshalAttributeNode(target *Layer, inode *inputNode, allNodes map[string]*inputNode) error {
	if inode.processed {
		return nil
	}
	inode.processed = true
	attribute := inode.graphNode
	attribute.AddTypes(AttributeTypes.Attribute)
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
		case LayerTerms.Attributes, LayerTerms.AttributeList:
			attribute.AddTypes(AttributeTypes.Object)
			// m must be an array of attributes. It can be under a @list
			attrArray, ok := val.([]interface{})
			if !ok {
				return MakeErrInvalidInput(inode.id, "Array of attributes expected here")
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
			for _, attr := range attrArray {
				// This must be a link
				follow := GetNodeID(attr)
				attrNode := allNodes[follow]
				if attrNode == nil {
					return MakeErrInvalidInput(inode.id, "Cannot follow link in attribute list:"+follow)
				}
				if err := unmarshalAttributeNode(target, attrNode, allNodes); err != nil {
					return err
				}
				target.AddEdge(inode.graphNode, attrNode.graphNode, NewLayerEdge(k))
			}

		case LayerTerms.Reference:
			attribute.AddTypes(AttributeTypes.Reference)
			// There can be at most one reference
			oid := GetNodeID(val)
			if len(oid) == 0 {
				return MakeErrInvalidInput(inode.id)
			}
			attribute.GetPropertyMap()[LayerTerms.Reference] = StringPropertyValue(oid)

		case LayerTerms.ArrayItems:
			attribute.AddTypes(AttributeTypes.Array)
			// m must be an array of 1
			itemsArr, _ := val.([]interface{})
			switch len(itemsArr) {
			case 0:
				return MakeErrInvalidInput(inode.id, "Invalid array items")
			case 1:
				itemsNode := allNodes[GetNodeID(itemsArr[0])]
				if itemsNode == nil {
					return MakeErrInvalidInput(inode.id, "Cannot follow link to array items")
				}
				if err := unmarshalAttributeNode(target, itemsNode, allNodes); err != nil {
					return err
				}
				target.AddEdge(inode.graphNode, itemsNode.graphNode, NewLayerEdge(k))
			default:
				return MakeErrInvalidInput(inode.id, "Multiple array items")
			}

		case LayerTerms.AllOf, LayerTerms.OneOf:
			if k == LayerTerms.AllOf {
				attribute.AddTypes(AttributeTypes.Composite)
			} else {
				attribute.AddTypes(AttributeTypes.Polymorphic)
			}
			// m must be a list
			elements := GetListElements(val)
			if elements == nil {
				return MakeErrInvalidInput(inode.id, "@list expected")
			}
			for _, element := range elements {
				nnode := allNodes[GetNodeID(element)]
				if nnode == nil {
					return MakeErrInvalidInput(inode.id, "Cannot follow link")
				}
				if err := unmarshalAttributeNode(target, nnode, allNodes); err != nil {
					return err
				}
				target.AddEdge(inode.graphNode, nnode.graphNode, NewLayerEdge(k))
			}
		}
	}

	t := FilterAttributeTypes(attribute.GetTypes())
	switch len(t) {
	case 0:
		attribute.AddTypes(AttributeTypes.Value)
	case 1:
	default:
		return ErrMultipleTypes(inode.id)
	}
	return nil
}

func unmarshalAnnotations(target *Layer, node *inputNode, allNodes map[string]*inputNode) {
	for key, value := range node.node {
		switch key {
		case "@id", "@type",
			LayerTerms.Attributes,
			LayerTerms.AttributeList,
			LayerTerms.Reference,
			LayerTerms.ArrayItems,
			LayerTerms.AllOf,
			LayerTerms.OneOf,
			LayerRootTerm:
		default:
			if strings.HasPrefix(key, "@") {
				break
			}
			// value must be an array
			arr, ok := value.([]interface{})
			if !ok {
				break
			}
			setValue := func(v string) {
				value := node.graphNode.GetPropertyMap()[key]
				if value == nil {
					node.graphNode.GetPropertyMap()[key] = StringPropertyValue(v)
				} else if value.IsStringSlice() {
					node.graphNode.GetPropertyMap()[key] = StringSlicePropertyValue(append(value.AsStringSlice(), v))
				} else {
					node.graphNode.GetPropertyMap()[key] = StringSlicePropertyValue([]string{value.AsString(), v})
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
						setValue(fmt.Sprint(v))
					} else if v := m["@id"]; v != nil {
						if id, ok := v.(string); ok {
							// Is this a link?
							referencedNode := allNodes[id]
							if referencedNode == nil {
								setValue(id)
							} else {
								target.AddEdge(node.graphNode, referencedNode.graphNode, NewLayerEdge(key))
							}
						}
					}
				}
			}
		}
	}
}

// Marshals the layer into an expanded jsonld document
func MarshalLayer(layer *Layer) interface{} {
	return []interface{}{marshalNode(layer.GetLayerInfoNode())}
}

func marshalNode(node LayerNode) interface{} {
	m := make(map[string]interface{})
	if node.Label() != nil {
		s := node.Label().(string)
		if len(s) > 0 {
			m["@id"] = s
		}
	}
	t := node.GetTypes()
	if len(t) > 0 {
		m["@type"] = t
	}

	for k, v := range node.GetPropertyMap() {
		if k == LayerTerms.Reference {
			m[k] = []interface{}{map[string]interface{}{"@id": v.AsString()}}
		} else {
			if v.IsString() {
				m[k] = []interface{}{map[string]interface{}{"@value": v.AsString()}}
			} else if v.IsStringSlice() {
				arr := make([]interface{}, 0)
				for _, elem := range v.AsStringSlice() {
					arr = append(arr, map[string]interface{}{"@value": elem})
				}
				m[k] = arr
			}
		}
	}

	edges := node.AllOutgoingEdges()
	for edges.HasNext() {
		edge := edges.Next().(LayerEdge)
		toNode := marshalNode(edge.To().(LayerNode))
		existing := m[edge.Label().(string)]
		switch edge.Label() {
		case LayerTerms.AttributeList, LayerTerms.AllOf, LayerTerms.OneOf:
			if existing == nil {
				m[edge.Label().(string)] = []interface{}{map[string]interface{}{"@list": []interface{}{toNode}}}
			} else {
				listMap := existing.([]interface{})[0].(map[string]interface{})
				listMap["@list"] = append(listMap["@list"].([]interface{}), toNode)
			}

		case LayerTerms.Attributes:
			if existing == nil {
				m[LayerTerms.Attributes] = []interface{}{toNode}
			} else {
				m[LayerTerms.Attributes] = append(m[LayerTerms.Attributes].([]interface{}), toNode)
			}

		case LayerTerms.ArrayItems:
			m[LayerTerms.ArrayItems] = []interface{}{toNode}

		default:
			if existing == nil {
				m[edge.Label().(string)] = []interface{}{toNode}
			} else {
				m[edge.Label().(string)] = append(m[edge.Label().(string)].([]interface{}), toNode)
			}
		}
	}
	return m
}

var ErrNotASchemaManifest = errors.New("Not a schema manifest")

// Unmarshals the given jsonld document into a schema manifest
func UnmarshalSchemaManifest(in interface{}) (*SchemaManifest, error) {
	proc := ld.NewJsonLdProcessor()
	compacted, err := proc.Compact(in, map[string]interface{}{}, nil)
	if err != nil {
		return nil, err
	}
	ret := SchemaManifest{}
	for k, v := range compacted {
		switch k {
		case "@id":
			ret.ID = v.(string)
		case "@type":
			ret.Type = v.(string)
		case TargetType:
			ret.TargetType = GetNodeID(v)
		case BundleTerm:
			ret.Bundle = GetNodeID(v)
		case SchemaBaseTerm:
			ret.Schema = GetNodeID(v)
		case OverlaysTerm:
			for _, x := range GetListElements(v) {
				ret.Overlays = append(ret.Overlays, GetNodeID(x))
			}
		}
	}
	if ret.Type != SchemaManifestTerm {
		return nil, ErrNotASchemaManifest
	}
	return &ret, nil
}

// MarshalSchemaNanifest returns a compact jsonld document for the manifest
func MarshalSchemaManifest(manifest *SchemaManifest) interface{} {
	m := make(map[string]interface{})
	m["@id"] = manifest.ID
	m["@type"] = SchemaManifestTerm
	m[TargetType] = map[string]interface{}{"@id": manifest.TargetType}
	if len(manifest.Bundle) > 0 {
		m[BundleTerm] = map[string]interface{}{"@id": manifest.Bundle}
	}
	m[SchemaBaseTerm] = map[string]interface{}{"@id": manifest.Schema}
	if len(manifest.Overlays) > 0 {
		arr := make([]interface{}, 0, len(manifest.Overlays))
		for _, x := range manifest.Overlays {
			arr = append(arr, map[string]interface{}{"@id": x})
		}
		m[OverlaysTerm] = map[string]interface{}{"@list": arr}
	}
	return m
}
