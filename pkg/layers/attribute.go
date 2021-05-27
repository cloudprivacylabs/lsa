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
package layers

import (
	"github.com/bserdar/digraph"
)

// IRI is used for jsonld @id types
type IRI string

// AttributeTypes defines the terms describing attribute types. Each
// attribute must have one of the attribute types plus the Attribute
// type, marking the object as an attribute.
var AttributeTypes = struct {
	Value       string
	Object      string
	Array       string
	Reference   string
	Composite   string
	Polymorphic string
	Attribute   string
}{
	Value:       LS + "Value",
	Object:      LS + "Object",
	Array:       LS + "Array",
	Reference:   LS + "Reference",
	Composite:   LS + "Composite",
	Polymorphic: LS + "Polymorphic",
	Attribute:   LS + "Attribute",
}

var AnnotationTerms = struct {
	Literal string
	IRI     string
}{
	Literal: LS + "Literal",
	IRI:     LS + "IRI",
}

// TypeTerms includes type specific terms recognized by the schema
// compiler. These are terms used to define elements of an attribute.
var TypeTerms = struct {
	// Unordered named attributes (json object)
	Attributes string
	// Ordered named attributes (json object, xml elements)
	AttributeList string
	// Reference to another schema. This will be resolved to another
	// schema during compilation
	Reference string
	// ArrayItems contains the definition for the items of the array
	ArrayItems string
	// All components of a composite attribute
	AllOf string
	// All options of a polymorphic attribute
	OneOf string
}{
	Attributes:    LS + "Object#attributes",
	AttributeList: LS + "Object#attributeList",
	Reference:     LS + "Reference#reference",
	ArrayItems:    LS + "Array#items",
	AllOf:         LS + "Composite#allOf",
	OneOf:         LS + "Polymorphic#oneOf",
}

// IsAttributeTreeEdge returns true if the edge is an edge between two
// attribute nodes
func IsAttributeTreeEdge(edge *digraph.Edge) bool {
	l := edge.Label()
	return l == TypeTerms.Attributes ||
		l == TypeTerms.AttributeList ||
		l == TypeTerms.ArrayItems ||
		l == TypeTerms.AllOf ||
		l == TypeTerms.OneOf
}

// IsAttributeNode returns true if the node has Attribute type
func IsAttributeNode(node *digraph.Node) bool {
	payload, _ := node.Payload.(*SchemaNode)
	return payload != nil && payload.HasType(AttributeTypes.Attribute)
}

// SchemaNode is the payload associated with all the nodes of a
// schema. The attribute nodes have types Attribute plus the specific
// type of the attribute. Other nodes will have their own types
// marking them as literal or IRI, or something else. Annotations
// cannot have Attribute or one of the attribute types
type SchemaNode struct {
	types    []string
	typesMap map[string]struct{}

	Properties map[string]interface{}
}

// NewSchemaNode returns a new schema node with the given types
func NewSchemaNode(types ...string) *SchemaNode {
	ret := SchemaNode{Properties: make(map[string]interface{})}
	ret.AddTypes(types...)
	return &ret
}

// GetAttributeTypes returns all recognized attribute types. This is
// mainly used for validation, to ensure there is only one attribute
// type
func GetAttributeTypes(types []string) []string {
	ret := make([]string, 0)
	for _, x := range types {
		if x == AttributeTypes.Value ||
			x == AttributeTypes.Object ||
			x == AttributeTypes.Array ||
			x == AttributeTypes.Reference ||
			x == AttributeTypes.Composite ||
			x == AttributeTypes.Polymorphic {
			ret = append(ret, x)
		}
	}
	return ret
}

func (a *SchemaNode) GetTypes() []string {
	if a == nil {
		return nil
	}
	return a.types
}

func (a *SchemaNode) AddTypes(t ...string) {
	if a.typesMap == nil {
		a.typesMap = make(map[string]struct{})
	}
	for _, x := range t {
		if _, exists := a.typesMap[x]; !exists {
			a.types = append(a.types, x)
			a.typesMap[x] = struct{}{}
		}
	}
}

func (a *SchemaNode) RemoveTypes(t ...string) {
	if a.typesMap == nil {
		return
	}
	for _, x := range t {
		delete(a.typesMap, x)
	}
	if len(a.typesMap) != len(a.types) {
		a.types = make([]string, 0, len(a.typesMap))
		for x := range a.typesMap {
			a.types = append(a.types, x)
		}
	}
}

func (a *SchemaNode) SetTypes(t ...string) {
	a.types = make([]string, 0, len(t))
	a.typesMap = make(map[string]struct{})
	a.AddTypes(t...)
}

func (a *SchemaNode) HasType(t string) bool {
	if a == nil || a.typesMap == nil {
		return false
	}
	_, exists := a.typesMap[t]
	return exists
}

// Clone returns a copy of the node data
func (a *SchemaNode) Clone() *SchemaNode {
	ret := NewSchemaNode(a.GetTypes()...)
	ret.Properties = copyIntf(a.Properties).(map[string]interface{})
	return ret
}

// GetParentAttribute returns the first immediate parent of the node that is
// an attribute and reached by an attribute edge.
func GetParentAttribute(node *digraph.Node) *digraph.Node {
	for parents := node.AllIncomingEdges(); parents.HasNext(); {
		parent := parents.Next()
		if !IsAttributeTreeEdge(parent) {
			continue
		}
		nd, _ := parent.From().Payload.(*SchemaNode)
		if nd == nil {
			continue
		}
		if nd.HasType(AttributeTypes.Attribute) {
			return parent.From()
		}
	}
	return nil
}

// SchemaEdge is the payload for schema edges.
type SchemaEdge struct {
	Properties map[string]interface{}
}

// NewSchemaEdge returns a new initialized schema edge
func NewSchemaEdge() *SchemaEdge {
	return &SchemaEdge{Properties: make(map[string]interface{})}
}

func (e *SchemaEdge) Clone() *SchemaEdge {
	ret := NewSchemaEdge()
	ret.Properties = copyIntf(e.Properties).(map[string]interface{})
	return ret
}
