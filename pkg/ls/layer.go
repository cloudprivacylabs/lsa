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

// IRI is used as property values
type IRI string

// A Layer is either a schema or an overlay. It keeps the definition
// of a layer as a directed labeled property graph.
type Layer struct {
	*digraph.Graph
	root *SchemaNode
}

// SchemaNode is either an attribute node or an annotation attached to
// an attribute. The attribute nodes have types Attribute plus the
// specific type of the attribute. Other nodes will have their own
// types marking them as literal or IRI, or something
// else. Annotations cannot have Attribute or one of the attribute
// types
type SchemaNode struct {
	digraph.NodeHeader

	// The types of the schema node
	types []string

	// Properties associated with the node. These are assumed to be JSON-types and IRI
	Properties map[string]interface{}
	// These can be set during compilation. They are shallow-cloned
	Compiled map[string]interface{}
}

// NewSchemaNode returns a new schema node with the given types
func NewSchemaNode(ID string, types ...string) *SchemaNode {
	ret := SchemaNode{Properties: make(map[string]interface{}),
		Compiled: make(map[string]interface{}),
	}
	ret.AddTypes(types...)
	ret.SetLabel(ID)
	return &ret
}

// GetID returns the node ID
func (a *SchemaNode) GetID() string {
	l := a.NodeHeader.Label()
	if l == nil {
		return ""
	}
	return l.(string)
}

// SetID sets the node ID
func (a *SchemaNode) SetID(ID string) {
	a.SetLabel(ID)
}

// GetTypes returns the types of the node
func (a *SchemaNode) GetTypes() []string {
	if a == nil {
		return nil
	}
	return a.types
}

// AddTypes adds new types to the schema node. The result is the
// set-union of the existing types and the given types
func (a *SchemaNode) AddTypes(t ...string) {
	a.types = StringSetUnion(a.types, t)
}

// RemoveTypes removes the given set of types from the node.
func (a *SchemaNode) RemoveTypes(t ...string) {
	a.types = StringSetSubtract(a.types, t)
}

// SetTypes sets the types of the node
func (a *SchemaNode) SetTypes(t ...string) {
	a.types = make([]string, 0, len(t))
	a.AddTypes(t...)
}

// HasType returns true if the node has the given type
func (a *SchemaNode) HasType(t string) bool {
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
func (a *SchemaNode) Connect(target *SchemaNode, edgeLabel string) *SchemaEdge {
	edge := NewSchemaEdge(edgeLabel)
	a.GetGraph().AddEdge(a, target, edge)
	return edge
}

// IsAttributeNode returns true if the node has Attribute type
func (a *SchemaNode) IsAttributeNode() bool {
	return a != nil && a.HasType(AttributeTypes.Attribute)
}

// Clone returns a copy of the node data. The returned node has the
// same label, types, and properties. The Compiled map is directly
// assigned to the new node
func (a *SchemaNode) Clone() *SchemaNode {
	ret := NewSchemaNode(a.GetID(), a.GetTypes()...)
	ret.Properties = copyIntf(a.Properties).(map[string]interface{})
	ret.Compiled = a.Compiled
	return ret
}

// GetParentAttribute returns the first immediate parent of the node that is
// an attribute and reached by an attribute edge.
func (a *SchemaNode) GetParentAttribute() *SchemaNode {
	for parents := a.AllIncomingEdges(); parents.HasNext(); {
		parent := parents.Next().(*SchemaEdge)
		if !parent.IsAttributeTreeEdge() {
			continue
		}
		nd, _ := parent.From().(*SchemaNode)
		if nd == nil {
			continue
		}
		if nd.HasType(AttributeTypes.Attribute) {
			return nd
		}
	}
	return nil
}

// SchemaEdge is a labeled graph edge between two schema nodes
type SchemaEdge struct {
	digraph.EdgeHeader
	Properties map[string]interface{}
	Compiled   map[string]interface{}
}

// NewSchemaEdge returns a new initialized schema edge
func NewSchemaEdge(label string) *SchemaEdge {
	ret := &SchemaEdge{Properties: make(map[string]interface{}),
		Compiled: make(map[string]interface{}),
	}
	ret.SetLabel(label)
	return ret
}

// GetLabel returns the edge label
func (edge *SchemaEdge) GetLabel() string {
	if edge == nil {
		return ""
	}
	l := edge.Label()
	if l == nil {
		return ""
	}
	return l.(string)
}

// IsAttributeTreeEdge returns true if the edge is an edge between two
// attribute nodes
func (edge *SchemaEdge) IsAttributeTreeEdge() bool {
	if edge == nil {
		return false
	}
	l := edge.Label()
	return l == TypeTerms.Attributes ||
		l == TypeTerms.AttributeList ||
		l == TypeTerms.ArrayItems ||
		l == TypeTerms.AllOf ||
		l == TypeTerms.OneOf
}

// Clone returns a copy of the schema edge
func (e *SchemaEdge) Clone() *SchemaEdge {
	ret := NewSchemaEdge(e.GetLabel())
	ret.Properties = copyIntf(e.Properties).(map[string]interface{})
	return ret
}

// NewLayer returns a new empty layer with the given ID and type
func NewLayer(ID string, typ ...string) *Layer {
	ret := &Layer{Graph: digraph.New()}
	ret.root = ret.NewNode(ID, typ...)
	ret.root.AddTypes(AttributeTypes.Attribute)
	return ret
}

// Clone returns a copy of the layer
func (l *Layer) Clone() *Layer {
	ret := &Layer{Graph: digraph.New()}
	nodeMap := make(map[*SchemaNode]*SchemaNode)
	for nodes := l.AllNodes(); nodes.HasNext(); {
		oldNode := nodes.Next().(*SchemaNode)
		newNode := oldNode.Clone()
		nodeMap[oldNode] = newNode
	}
	ret.root = nodeMap[l.root]
	for nodes := l.AllNodes(); nodes.HasNext(); {
		node := nodes.Next().(*SchemaNode)
		for edges := node.AllOutgoingEdges(); edges.HasNext(); {
			edge := edges.Next().(*SchemaEdge)
			ret.AddEdge(nodeMap[edge.From().(*SchemaNode)], nodeMap[edge.To().(*SchemaNode)], edge.Clone())
		}
	}
	return ret
}

// GetRoot returns the root node of the schema
func (l *Layer) GetRoot() *SchemaNode { return l.root }

// GetID returns the ID of the layer, which is the ID of the root node
func (l *Layer) GetID() string {
	return l.root.Label().(string)
}

// SetID sets the ID of the layer, which is the ID of the root node
func (l *Layer) SetID(ID string) {
	l.root.SetLabel(ID)
}

// GetLayerType returns the layer type, SchemaTerm or OverlayTerm.
func (l *Layer) GetLayerType() string {
	if l.root.HasType(SchemaTerm) {
		return SchemaTerm
	}
	if l.root.HasType(OverlayTerm) {
		return OverlayTerm
	}
	return ""
}

// SetLayerType sets if the layer is a schema or an overlay
func (l *Layer) SetLayerType(t string) {
	if t != SchemaTerm && t != OverlayTerm {
		panic("Invalid layer type:" + t)
	}
	l.root.RemoveTypes(SchemaTerm, OverlayTerm)
	l.root.AddTypes(t)
}

// NewNode creates a new node for the layer with the given ID and types
func (l *Layer) NewNode(ID string, types ...string) *SchemaNode {
	ret := NewSchemaNode(ID, types...)
	l.AddNode(ret)
	return ret
}

// GetTargetTypes returns the value of the targetType field
func (l *Layer) GetTargetTypes() []string {
	schNode := l.root
	v := schNode.Properties[TargetType]
	if arr, ok := v.([]interface{}); ok {
		ret := make([]string, len(arr))
		for _, x := range arr {
			ret = append(ret, x.(string))
		}
		return ret
	}
	if str, ok := v.(string); ok {
		return []string{str}
	}
	return nil
}

// ForEachAttribute calls f with each attribute node, depth first. If
// f returns false, iteration stops
func (l *Layer) ForEachAttribute(f func(*SchemaNode) bool) {
	var forEachAttribute func(*SchemaNode, func(*SchemaNode) bool) bool
	forEachAttribute = func(root *SchemaNode, f func(*SchemaNode) bool) bool {
		if root.IsAttributeNode() {
			if !f(root) {
				return false
			}
		}
		for outgoing := root.AllOutgoingEdges(); outgoing.HasNext(); {
			edge := outgoing.Next().(*SchemaEdge)
			if !edge.IsAttributeTreeEdge() {
				continue
			}
			next := edge.To().(*SchemaNode)
			if next.HasType(AttributeTypes.Attribute) {
				if !forEachAttribute(next, f) {
					return false
				}
			}
		}
		return true
	}

	forEachAttribute(l.root, f)
}
