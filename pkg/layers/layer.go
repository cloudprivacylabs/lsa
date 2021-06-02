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

// A Layer is either a schema or an overlay. It keeps the definition
// of a layer as a directed labeled property graph.
type Layer struct {
	g    *digraph.Graph
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
func NewSchemaNode(label interface{}, types ...string) *SchemaNode {
	ret := SchemaNode{Properties: make(map[string]interface{}),
		Compiled: make(map[string]interface{}),
	}
	ret.AddTypes(types...)
	ret.SetLabel(label)
	return &ret
}

func (a *SchemaNode) GetTypes() []string {
	if a == nil {
		return nil
	}
	return a.types
}

func (a *SchemaNode) AddTypes(t ...string) {
	for _, x := range t {
		if !a.HasType(x) {
			a.types = append(a.types, x)
		}
	}
}

func (a *SchemaNode) RemoveTypes(t ...string) {
	w := 0
	for i, x := range a.types {
		found := false
		for _, q := range t {
			if q == x {
				found = true
				break
			}
		}
		if !found {
			a.types[w] = x
			w++
		}
	}
	a.types = a.types[:w]
}

func (a *SchemaNode) SetTypes(t ...string) {
	a.types = make([]string, 0, len(t))
	a.AddTypes(t...)
}

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

// Clone returns a copy of the node data
func (a *SchemaNode) Clone() *SchemaNode {
	ret := NewSchemaNode(a.Label(), a.GetTypes()...)
	ret.Properties = copyIntf(a.Properties).(map[string]interface{})
	ret.Compiled = a.Compiled
	return ret
}

// GetParentAttribute returns the first immediate parent of the node that is
// an attribute and reached by an attribute edge.
func (a *SchemaNode) GetParentAttribute() *SchemaNode {
	for parents := a.AllIncomingEdges(); parents.HasNext(); {
		parent := parents.Next()
		if !IsAttributeTreeEdge(parent) {
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

func NewLayer() *Layer {
	ret := &Layer{Graph: digraph.New()}
	ret.RootNode = ret.NewNode(nil)
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
	ret.RootNode = nodeMap[l.RootNode]
	for nodes := l.AllNodes(); nodes.HasNext(); {
		node := nodes.Next().(*SchemaNode)
		for edges := node.AllOutgoingEdges(); edges.HasNext(); {
			edge := edges.Next().(*SchemaEdge)
			ret.AddEdge(nodeMap[edge.From().(*SchemaNode)], nodeMap[edge.To().(*SchemaNode)], edge.Clone())
		}
	}
	return ret
}

// GetID returns the ID of the layer, which is the ID of the root node
func (l *Layer) GetID() string {
	return l.RootNode.Label().(string)
}

// SetID sets the ID of the layer, which is the ID of the root node
func (l *Layer) SetID(ID string) {
	l.RootNode.SetLabel(ID)
}

// GetLayerType returns the layer type, SchemaTerm or OverlayTerm.
func (l *Layer) GetLayerType() string {
	schNode := l.RootNode
	if schNode.HasType(SchemaTerm) {
		return SchemaTerm
	}
	if schNode.HasType(OverlayTerm) {
		return OverlayTerm
	}
	return ""
}

func (l *Layer) NewNode(label interface{}, types ...string) *SchemaNode {
	ret := NewSchemaNode(label, types...)
	l.AddNode(ret)
	return ret
}

// GetTargetTypes returns the value of the targetType field
func (l *Layer) GetTargetTypes() []string {
	schNode := l.RootNode
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
		if IsAttributeNode(root) {
			if !f(root) {
				return false
			}
		}
		for outgoing := root.AllOutgoingEdges(); outgoing.HasNext(); {
			edge := outgoing.Next()
			if !IsAttributeTreeEdge(edge) {
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

	forEachAttribute(l.RootNode, f)
}
