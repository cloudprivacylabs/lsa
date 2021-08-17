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
	"golang.org/x/text/encoding"
)

// A Layer is either a schema or an overlay. It keeps the definition
// of a layer as a directed labeled property graph.
//
// The root node of the layer keeps layer identifying information. The
// root node is connected to the schema node which contains the actual
// object defined by the layer.
type Layer struct {
	*digraph.Graph
	layerInfo Node
}

// NewLayer returns a new empty layer
func NewLayer() *Layer {
	ret := &Layer{Graph: digraph.New()}
	ret.layerInfo = NewNode("")
	ret.AddNode(ret.layerInfo)
	return ret
}

// Clone returns a copy of the layer
func (l *Layer) Clone() *Layer {
	ret := &Layer{Graph: digraph.New()}
	nodeMap := digraph.Copy(ret.Graph, l.Graph, func(node digraph.Node) digraph.Node {
		return node.(Node).Clone()
	},
		func(edge digraph.Edge) digraph.Edge {
			return edge.(Edge).Clone()
		})
	if x := nodeMap[l.layerInfo]; x != nil {
		ret.layerInfo = x.(Node)
	}
	return ret
}

// GetLayerInfoNode returns the root node of the schema
func (l *Layer) GetLayerInfoNode() Node { return l.layerInfo }

// GetSchemaRootNode returns the root node of the object defined by the schema
func (l *Layer) GetSchemaRootNode() Node {
	x := l.layerInfo.Next(LayerRootTerm)
	if x == nil {
		return nil
	}
	return x.(Node)
}

// GetID returns the ID of the layer
func (l *Layer) GetID() string {
	return l.layerInfo.GetLabel().(string)
}

// SetID sets the ID of the layer
func (l *Layer) SetID(ID string) {
	l.layerInfo.SetLabel(ID)
}

// GetLayerType returns the layer type, SchemaTerm or OverlayTerm.
func (l *Layer) GetLayerType() string {
	if l.layerInfo.HasType(SchemaTerm) {
		return SchemaTerm
	}
	if l.layerInfo.HasType(OverlayTerm) {
		return OverlayTerm
	}
	return ""
}

// SetLayerType sets if the layer is a schema or an overlay
func (l *Layer) SetLayerType(t string) {
	if t != SchemaTerm && t != OverlayTerm {
		panic("Invalid layer type:" + t)
	}
	l.layerInfo.RemoveTypes(SchemaTerm, OverlayTerm)
	l.layerInfo.AddTypes(t)
}

// GetEncoding returns the encoding that should be used to
// ingest/export data using this layer. The encoding information is
// taken from the schema root node characterEncoding annotation. If missing,
// the default encoding is used, which does not perform any character
// translation
func (l *Layer) GetEncoding() (encoding.Encoding, error) {
	oi := l.GetSchemaRootNode()
	var enc string
	if oi != nil {
		enc = oi.GetProperties()[CharacterEncodingTerm].AsString()
		if len(enc) == 0 {
			return encoding.Nop, nil
		}
	}
	return UnknownEncodingIndex.Encoding(enc)
}

// NewNode creates a new node for the layer with the given ID and
// types, and adds the node to the layer
func (l *Layer) NewNode(ID string, types ...string) Node {
	ret := NewNode(ID, types...)
	l.AddNode(ret)
	return ret
}

// GetTargetType returns the value of the targetType field from the
// layer information node
func (l *Layer) GetTargetType() string {
	v := l.layerInfo.GetProperties()[TargetType]
	if v == nil {
		return ""
	}
	return v.AsString()
}

// SetTargetType sets the targe types of the layer
func (l *Layer) SetTargetType(t string) {
	if oldT := l.GetTargetType(); len(oldT) > 0 {
		if oin := l.GetSchemaRootNode(); oin != nil {
			oin.RemoveTypes(oldT)
		}
	}
	l.layerInfo.GetProperties()[TargetType] = StringPropertyValue(t)
	if oin := l.GetSchemaRootNode(); oin != nil {
		oin.AddTypes(t)
	}
}

// ForEachAttribute calls f with each attribute node, depth first. If
// f returns false, iteration stops
func (l *Layer) ForEachAttribute(f func(Node, []Node) bool) bool {
	oi := l.GetSchemaRootNode()
	if oi != nil {
		return ForEachAttributeNode(oi, f)
	}
	return true
}

// ForEachAttributeOrdered calls f with each attribute node, depth
// first and in order. If f returns false, iteration stops
func (l *Layer) ForEachAttributeOrdered(f func(Node, []Node) bool) bool {
	oi := l.GetSchemaRootNode()
	if oi != nil {
		return ForEachAttributeNodeOrdered(oi, f)
	}
	return true
}

// RenameBlankNodes will call namerFunc for each blank node, so they
// can be renamed and won't cause name clashes
func (l *Layer) RenameBlankNodes(namer func(Node)) {
	for nodes := l.GetAllNodes(); nodes.HasNext(); {
		node := nodes.Next().(Node)
		id := node.GetID()
		if len(id) == 0 || id[0] == '_' {
			namer(node)
		}
	}
}

// GetPath returns the path to the given attribute node
func (l *Layer) GetAttributePath(node Node) []Node {
	var ret []Node
	ForEachAttributeNode(l.GetSchemaRootNode(), func(n Node, path []Node) bool {
		if n == node {
			ret = path
			return false
		}
		return true
	})
	return ret
}

// FindAttributeByID returns the attribute and the path to it
func (l *Layer) FindAttributeByID(id string) (Node, []Node) {
	return l.FindFirstAttribute(func(n Node) bool { return n.GetID() == id })
}

// FindFirstAttribute returns the first attribute for which the predicate holds
func (l *Layer) FindFirstAttribute(predicate func(Node) bool) (Node, []Node) {
	var node Node
	var path []Node
	ForEachAttributeNode(l.GetSchemaRootNode(), func(n Node, p []Node) bool {
		if predicate(n) {
			node = n
			path = p
			return false
		}
		return true
	})
	return node, path
}

// ForEachAttributeNode calls f with each attribute node, depth
// first. Path contains all the nodes from root to the current
// node. If f returns false, iteration stops. This function visits
// each node only once
func ForEachAttributeNode(root Node, f func(node Node, path []Node) bool) bool {
	return forEachAttributeNode(root, make([]Node, 0, 32), f, map[Node]struct{}{}, false)
}

func forEachAttributeNode(root Node, path []Node, f func(Node, []Node) bool, loop map[Node]struct{}, ordered bool) bool {
	if _, exists := loop[root]; exists {
		return true
	}
	loop[root] = struct{}{}

	path = append(path, root)
	if root.IsAttributeNode() {
		if !f(root, path) {
			return false
		}
	}

	outgoing := root.GetAllOutgoingEdges()
	if ordered {
		outgoing = SortEdgesItr(outgoing)
	}

	for outgoing.HasNext() {
		edge := outgoing.Next().(Edge)
		if !edge.IsAttributeTreeEdge() {
			continue
		}
		next := edge.GetTo().(Node)
		if next.HasType(AttributeTypes.Attribute) {
			if !forEachAttributeNode(next, path, f, loop, ordered) {
				return false
			}
		}
	}
	return true
}

// ForEachAttributeNodeOrdered calls f with each attribute node, depth
// first, preserving order. Path contains all the nodes from root to the current
// node. If f returns false, iteration stops. This function visits
// each node only once
func ForEachAttributeNodeOrdered(root Node, f func(node Node, path []Node) bool) bool {
	return forEachAttributeNode(root, make([]Node, 0, 32), f, map[Node]struct{}{}, true)
}
