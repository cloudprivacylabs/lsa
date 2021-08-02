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
	"github.com/bserdar/digraph"
)

// Node is the node type used for schema layer graphs
type Node interface {
	digraph.Node

	// Return the types of the node
	GetTypes() []string

	// Returns true if the node has the given type
	HasType(string) bool

	// Add the given types to the node type
	AddTypes(...string)

	// Remove the given types from the node
	RemoveTypes(...string)

	// Sets the types of the node to the given types
	SetTypes(...string)

	// Return node ID
	GetID() string

	// Set node ID
	SetID(string)

	// Clone returns a new node that is a copy of this one, but the
	// returned node is not connected
	Clone() Node

	// Connect this node to the target layer node with the given edge label
	Connect(Node, string) Edge

	// If this is an attribute node, returns true.
	IsAttributeNode() bool

	// If this is a document node, returns true.
	IsDocumentNode() bool

	// Value of the document node, nil if the node is not a document node
	GetValue() interface{}

	SetValue(interface{})

	// GetProperties returns the name/value pairs of the node. The
	// values are either string or []string. When cloned, the new node
	// receives a deep copy of the map
	GetProperties() map[string]*PropertyValue

	// Returns the compiled data map. This map is used to store
	// compilation information related to the node, and its contents are
	// unspecified. If the node is cloned with compiled data map, the
	// new node will get a shallow copy of the compiled data
	GetCompiledDataMap() map[interface{}]interface{}
}

// node is either an attribute node, document node, or an annotation
// node.  The attribute nodes have types Attribute plus the specific
// type of the attribute. Other nodes will have their own types
// marking them as literal or IRI, or something else. Annotations
// cannot have Attribute or one of the attribute types
type node struct {
	digraph.NodeHeader

	// The types of the schema node
	types []string

	// value for document nodes
	value interface{}

	// Properties associated with the node. These are assumed to be JSON-types
	properties map[string]*PropertyValue
	// These can be set during compilation. They are shallow-cloned
	compiled map[interface{}]interface{}
}

func (a *node) GetCompiledDataMap() map[interface{}]interface{} { return a.compiled }

func (a *node) GetProperties() map[string]*PropertyValue { return a.properties }

func (a *node) GetValue() interface{} { return a.value }

func (a *node) SetValue(value interface{}) { a.value = value }

func (a *node) IsDocumentNode() bool {
	return a.HasType(DocumentNodeTerm)
}

// NewNode returns a new node with the given types
func NewNode(ID string, types ...string) Node {
	ret := node{
		properties: make(map[string]*PropertyValue),
		compiled:   make(map[interface{}]interface{}),
	}
	ret.AddTypes(types...)
	ret.SetLabel(ID)
	return &ret
}

// GetID returns the node ID
func (a *node) GetID() string {
	l := a.GetLabel()
	if l == nil {
		return ""
	}
	return l.(string)
}

// SetID sets the node ID
func (a *node) SetID(ID string) {
	a.SetLabel(ID)
}

// GetTypes returns the types of the node
func (a *node) GetTypes() []string {
	if a == nil {
		return nil
	}
	return a.types
}

// AddTypes adds new types to the schema node. The result is the
// set-union of the existing types and the given types
func (a *node) AddTypes(t ...string) {
	for i := range t {
		t[i] = knownTerm(t[i])
	}
	a.types = StringSetUnion(a.types, t)
}

// RemoveTypes removes the given set of types from the node.
func (a *node) RemoveTypes(t ...string) {
	a.types = StringSetSubtract(a.types, t)
}

// SetTypes sets the types of the node
func (a *node) SetTypes(t ...string) {
	a.types = make([]string, 0, len(t))
	a.AddTypes(t...)
}

// HasType returns true if the node has the given type
func (a *node) HasType(t string) bool {
	if a == nil {
		return false
	}
	for _, x := range a.types {
		if t == x {
			return true
		}
	}
	return false
}

// Connect this node with the target node using an edge with the given label
func (a *node) Connect(target Node, edgeLabel string) Edge {
	edge := NewEdge(edgeLabel)
	digraph.Connect(a, target, edge)
	return edge
}

// IsAttributeNode returns true if the node has Attribute type
func (a *node) IsAttributeNode() bool {
	return a != nil && a.HasType(AttributeTypes.Attribute)
}

// Clone returns a copy of the node data. The returned node has the
// same label, types, and properties. The Compiled map is directly
// assigned to the new node
func (a *node) Clone() Node {
	ret := NewNode(a.GetID(), a.GetTypes()...).(*node)
	ret.properties = CopyPropertyMap(a.properties)
	ret.compiled = a.compiled
	return ret
}

// GetAttributeEdgeBetweenNodes returns the attribute edges between
// two nodes. If there are no direct edges, return nil
func GetLayerEdgeBetweenNodes(source, target Node) Edge {
	for edges := source.GetAllOutgoingEdges(); edges.HasNext(); {
		edge := edges.Next().(Edge)
		if edge.IsAttributeTreeEdge() && edge.GetTo() == target {
			return edge
		}
	}
	return nil
}

// GetFilteredValue returns the field value processed by the schema
// value filters, and then the node value filters
func (node *node) GetFilteredValue() interface{} {
	schemaNode, _ := node.Next(InstanceOfTerm).(Node)
	return GetFilteredValue(schemaNode, node)
}

// GetFilteredValue filters the value through the schema properties
// and then through the node properties before returning
func GetFilteredValue(schemaNode, docNode Node) interface{} {
	value := docNode.GetValue()
	if schemaNode != nil {
		value = FilterValue(value, docNode, schemaNode.GetProperties())
	}
	return FilterValue(value, docNode, docNode.GetProperties())
}

// IsDocumentEdge returns true if the edge is a data edge term
func IsDocumentEdge(edge digraph.Edge) bool {
	switch edge.GetLabel() {
	case DataEdgeTerms.ObjectAttributes, DataEdgeTerms.ArrayElements:
		return true
	}
	return false
}
