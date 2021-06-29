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

// LayerNode is the node type used for schema layer graphs
type LayerNode interface {
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

	// Clone returns a new layer node that is a copy of this one, but
	// the returned node is not connected to any graph, nor does it have
	// any edges.
	Clone() LayerNode

	// Connect this node to the target layer node with the given edge label
	Connect(LayerNode, string) LayerEdge

	// If this is an attribute node, returns true. If not, this is a semantic annotation nodex
	IsAttributeNode() bool

	// GetPropertyMap returns the name/value pairs of the node. The
	// values are either string or []string. When cloned, the new node
	// receives a deep copy of the map
	GetPropertyMap() map[string]*PropertyValue

	// Returns the compiled data map. This map is used to store
	// compilation information related to the node, and its contents are
	// unspecified. If the node is cloned with compiled data map, the
	// new node will get a shallow copy of the compiled data
	GetCompiledDataMap() map[interface{}]interface{}
}

// schemaNode is either an attribute node or an annotation attached to
// an attribute. The attribute nodes have types Attribute plus the
// specific type of the attribute. Other nodes will have their own
// types marking them as literal or IRI, or something
// else. Annotations cannot have Attribute or one of the attribute
// types
type schemaNode struct {
	digraph.NodeHeader

	// The types of the schema node
	types []string

	// Properties associated with the node. These are assumed to be JSON-types
	properties map[string]*PropertyValue
	// These can be set during compilation. They are shallow-cloned
	compiled map[interface{}]interface{}
}

func (a *schemaNode) GetCompiledDataMap() map[interface{}]interface{} { return a.compiled }

func (a *schemaNode) GetPropertyMap() map[string]*PropertyValue { return a.properties }

// NewLayerNode returns a new schema node with the given types
func NewLayerNode(ID string, types ...string) LayerNode {
	ret := schemaNode{
		properties: make(map[string]*PropertyValue),
		compiled:   make(map[interface{}]interface{}),
	}
	ret.AddTypes(types...)
	ret.SetLabel(ID)
	return &ret
}

// GetID returns the node ID
func (a *schemaNode) GetID() string {
	l := a.NodeHeader.Label()
	if l == nil {
		return ""
	}
	return l.(string)
}

// SetID sets the node ID
func (a *schemaNode) SetID(ID string) {
	a.SetLabel(ID)
}

// GetTypes returns the types of the node
func (a *schemaNode) GetTypes() []string {
	if a == nil {
		return nil
	}
	return a.types
}

// AddTypes adds new types to the schema node. The result is the
// set-union of the existing types and the given types
func (a *schemaNode) AddTypes(t ...string) {
	a.types = StringSetUnion(a.types, t)
}

// RemoveTypes removes the given set of types from the node.
func (a *schemaNode) RemoveTypes(t ...string) {
	a.types = StringSetSubtract(a.types, t)
}

// SetTypes sets the types of the node
func (a *schemaNode) SetTypes(t ...string) {
	a.types = make([]string, 0, len(t))
	a.AddTypes(t...)
}

// HasType returns true if the node has the given type
func (a *schemaNode) HasType(t string) bool {
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
func (a *schemaNode) Connect(target LayerNode, edgeLabel string) LayerEdge {
	edge := NewLayerEdge(edgeLabel)
	a.GetGraph().AddEdge(a, target, edge)
	return edge
}

// IsAttributeNode returns true if the node has Attribute type
func (a *schemaNode) IsAttributeNode() bool {
	return a != nil && a.HasType(AttributeTypes.Attribute)
}

// Clone returns a copy of the node data. The returned node has the
// same label, types, and properties. The Compiled map is directly
// assigned to the new node
func (a *schemaNode) Clone() LayerNode {
	ret := NewLayerNode(a.GetID(), a.GetTypes()...).(*schemaNode)
	ret.properties = CopyPropertyMap(a.properties)
	ret.compiled = a.compiled
	return ret
}

// GetParentAttribute returns the first immediate parent of the node that is
// an attribute and reached by an attribute edge.
func GetParentAttribute(a LayerNode) (LayerNode, LayerEdge) {
	for parents := a.AllIncomingEdges(); parents.HasNext(); {
		parent := parents.Next().(LayerEdge)
		if !parent.IsAttributeTreeEdge() {
			continue
		}
		nd, _ := parent.From().(LayerNode)
		if nd == nil {
			continue
		}
		if nd.HasType(AttributeTypes.Attribute) {
			return nd, parent
		}
	}
	return nil, nil
}
