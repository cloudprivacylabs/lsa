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
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
	"golang.org/x/text/encoding"
)

// A Layer is either a schema or an overlay. It keeps the definition
// of a layer as a directed labeled property graph.
//
// The root node of the layer keeps layer identifying information. The
// root node is connected to the schema node which contains the actual
// object defined by the layer.
type Layer struct {
	graph.Graph
	layerInfo graph.Node
}

// NewLayer returns a new empty layer
func NewLayer() *Layer {
	ret := &Layer{Graph: graph.NewOCGraph()}
	ret.layerInfo = ret.NewNode(nil, nil)
	return ret
}

// Clone returns a copy of the layer
func (l *Layer) Clone() *Layer {
	targetGraph := graph.NewOCGraph()
	nodeMap := graph.CopyGraph(l.Graph, targetGraph, func(key string, value interface{}) interface{} {
		if p, ok := value.(*PropertyValue); ok {
			return p.Clone()
		}
		return value
	})
	ret := &Layer{
		Graph:     targetGraph,
		layerInfo: nodeMap[l.layerInfo],
	}
	return ret
}

// CloneInto clones the layer into the targetgraph
func (l *Layer) CloneInto(targetGraph graph.Graph) (*Layer, map[graph.Node]graph.Node) {
	ret := &Layer{Graph: targetGraph}
	nodeMap := graph.CopyGraph(l.Graph, targetGraph, func(key string, value interface{}) interface{} {
		if p, ok := value.(*PropertyValue); ok {
			return p.Clone()
		}
		return value
	})
	ret.layerInfo = nodeMap[l.layerInfo]
	return ret, nodeMap
}

// GetLayerRootNode returns the root node of the schema
func (l *Layer) GetLayerRootNode() graph.Node { return l.layerInfo }

// GetSchemaRootNode returns the root node of the object defined by the schema
func (l *Layer) GetSchemaRootNode() graph.Node {
	x := graph.TargetNodes(l.layerInfo.GetEdgesWithLabel(graph.OutgoingEdge, LayerRootTerm))
	if len(x) != 1 {
		return nil
	}
	return x[0]
}

// GetID returns the ID of the layer
func (l *Layer) GetID() string {
	v, _ := l.layerInfo.GetProperty(LayerIDTerm)
	s, _ := v.(string)
	return s
}

// SetID sets the ID of the layer
func (l *Layer) SetID(ID string) {
	l.layerInfo.SetProperty(LayerIDTerm, ID)
}

// GetLayerType returns the layer type, SchemaTerm or OverlayTerm.
func (l *Layer) GetLayerType() string {
	labels := l.layerInfo.GetLabels()
	if labels.Has(SchemaTerm) {
		return SchemaTerm
	}
	if labels.Has(OverlayTerm) {
		return OverlayTerm
	}
	return ""
}

// SetLayerType sets if the layer is a schema or an overlay
func (l *Layer) SetLayerType(t string) {
	if t != SchemaTerm && t != OverlayTerm {
		panic("Invalid layer type:" + t)
	}
	labels := l.layerInfo.GetLabels()
	labels.Remove(SchemaTerm, OverlayTerm)
	labels.Add(t)
	l.layerInfo.SetLabels(labels)
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
		enc = AsPropertyValue(oi.GetProperty(CharacterEncodingTerm)).AsString()
		if len(enc) == 0 {
			return encoding.Nop, nil
		}
	}
	return UnknownEncodingIndex.Encoding(enc)
}

// GetTargetType returns the value of the targetType field from the
// layer information node
func (l *Layer) GetTargetType() string {
	return AsPropertyValue(l.layerInfo.GetProperty(TargetType)).AsString()
}

// SetTargetType sets the targe types of the layer
func (l *Layer) SetTargetType(t string) {
	if oldT := l.GetTargetType(); len(oldT) > 0 {
		if oin := l.GetSchemaRootNode(); oin != nil {
			labels := oin.GetLabels()
			labels.Remove(oldT)
			oin.SetLabels(labels)
		}
	}
	if len(t) > 0 {
		l.layerInfo.SetProperty(TargetType, StringPropertyValue(t))
		if oin := l.GetSchemaRootNode(); oin != nil {
			labels := oin.GetLabels()
			labels.Add(t)
			oin.SetLabels(labels)
		}
	}
}

// GetEntityIDNodes returns the entity ID nodes for the layer.
func (l *Layer) GetEntityIDNodes() []graph.Node {
	root := l.GetSchemaRootNode()
	if root == nil {
		return nil
	}
	return GetEntityIDNodes(root)
}

// ForEachAttribute calls f with each attribute node, depth first. If
// f returns false, iteration stops
func (l *Layer) ForEachAttribute(f func(graph.Node, []graph.Node) bool) bool {
	oi := l.GetSchemaRootNode()
	if oi != nil {
		return ForEachAttributeNode(oi, f)
	}
	return true
}

// ForEachAttributeOrdered calls f with each attribute node, depth
// first and in order. If f returns false, iteration stops
func (l *Layer) ForEachAttributeOrdered(f func(graph.Node, []graph.Node) bool) bool {
	oi := l.GetSchemaRootNode()
	if oi != nil {
		return ForEachAttributeNodeOrdered(oi, f)
	}
	return true
}

// RenameBlankNodes will call namerFunc for each blank node, so they
// can be renamed and won't cause name clashes
func (l *Layer) RenameBlankNodes(namer func(graph.Node)) {
	for nodes := l.GetNodes(); nodes.Next(); {
		node := nodes.Node()
		id := GetAttributeID(node)
		if len(id) == 0 || id[0] == '_' {
			namer(node)
		}
	}
}

// GetPath returns the path to the given attribute node
func (l *Layer) GetAttributePath(node graph.Node) []graph.Node {
	var ret []graph.Node
	ForEachAttributeNode(l.GetSchemaRootNode(), func(n graph.Node, path []graph.Node) bool {
		if n == node {
			ret = path
			return false
		}
		return true
	})
	return ret
}

// FindAttributeByID returns the attribute and the path to it
func (l *Layer) FindAttributeByID(id string) (graph.Node, []graph.Node) {
	return l.FindFirstAttribute(func(n graph.Node) bool { return GetAttributeID(n) == id })
}

// FindFirstAttribute returns the first attribute for which the predicate holds
func (l *Layer) FindFirstAttribute(predicate func(graph.Node) bool) (graph.Node, []graph.Node) {
	var node graph.Node
	var path []graph.Node
	ForEachAttributeNode(l.GetSchemaRootNode(), func(n graph.Node, p []graph.Node) bool {
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
func ForEachAttributeNode(root graph.Node, f func(node graph.Node, path []graph.Node) bool) bool {
	return forEachAttributeNode(root, make([]graph.Node, 0, 32), f, map[graph.Node]struct{}{}, false)
}

func forEachAttributeNode(root graph.Node, path []graph.Node, f func(graph.Node, []graph.Node) bool, loop map[graph.Node]struct{}, ordered bool) bool {
	if _, exists := loop[root]; exists {
		return true
	}
	loop[root] = struct{}{}

	path = append(path, root)
	if IsAttributeNode(root) {
		if !f(root, path) {
			return false
		}
	}

	outgoing := root.GetEdges(graph.OutgoingEdge)
	if ordered {
		outgoing = SortEdgesItr(outgoing)
	}

	for outgoing.Next() {
		edge := outgoing.Edge()
		if !IsAttributeTreeEdge(edge) {
			continue
		}
		next := edge.GetTo()
		if next.GetLabels().Has(AttributeNodeTerm) {
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
func ForEachAttributeNodeOrdered(root graph.Node, f func(node graph.Node, path []graph.Node) bool) bool {
	return forEachAttributeNode(root, make([]graph.Node, 0, 32), f, map[graph.Node]struct{}{}, true)
}

// GetAttributeID returns the attributeID
func GetAttributeID(node graph.Node) string {
	v, _ := node.GetProperty(NodeIDTerm)
	s, _ := v.(string)
	return s
}

// SetAttrributeID sets the attribute ID
func SetAttributeID(node graph.Node, ID string) {
	node.SetProperty(NodeIDTerm, ID)
}
