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
	layerInfo LayerNode
}

// NewLayer returns a new empty layer
func NewLayer() *Layer {
	ret := &Layer{Graph: digraph.New()}
	ret.layerInfo = NewLayerNode("")
	ret.AddNode(ret.layerInfo)
	return ret
}

// Clone returns a copy of the layer
func (l *Layer) Clone() *Layer {
	ret := &Layer{Graph: digraph.New()}
	nodeMap := digraph.Copy(ret.Graph, l.Graph, func(node digraph.Node) digraph.Node {
		return node.(LayerNode).Clone()
	},
		func(edge digraph.Edge) digraph.Edge {
			return edge.(LayerEdge).Clone()
		})
	if x := nodeMap[l.layerInfo]; x != nil {
		ret.layerInfo = x.(LayerNode)
	}
	return ret
}

// GetLayerInfoNode returns the root node of the schema
func (l *Layer) GetLayerInfoNode() LayerNode { return l.layerInfo }

// GetObjectInfoNode returns the root node of the object defined by the schema
func (l *Layer) GetObjectInfoNode() LayerNode {
	x := l.layerInfo.NextNode(LayerRootTerm)
	if x == nil {
		return nil
	}
	return x.(LayerNode)
}

// GetID returns the ID of the layer
func (l *Layer) GetID() string {
	return l.layerInfo.Label().(string)
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
	oi := l.GetObjectInfoNode()
	var enc string
	if oi != nil {
		enc, _ = oi.GetPropertyMap()[CharacterEncodingTerm].(string)
		if len(enc) == 0 {
			return encoding.Nop, nil
		}
	}
	return UnknownEncodingIndex.Encoding(enc)
}

// NewNode creates a new node for the layer with the given ID and
// types, and adds the node to the layer
func (l *Layer) NewNode(ID string, types ...string) LayerNode {
	ret := NewLayerNode(ID, types...)
	l.AddNode(ret)
	return ret
}

// GetTargetTypes returns the value of the targetType field from the
// layer information node
func (l *Layer) GetTargetTypes() []string {
	v := l.layerInfo.GetPropertyMap()[TargetType]
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

// SetTargetTypes sets the target types of the layer
func (l *Layer) SetTargetTypes(t ...string) {
	arr := make([]interface{}, 0, len(t))
	for _, x := range t {
		arr = append(arr, x)
	}
	l.layerInfo.GetPropertyMap()[TargetType] = arr
}

// ForEachAttribute calls f with each attribute node, depth first. If
// f returns false, iteration stops
func (l *Layer) ForEachAttribute(f func(LayerNode) bool) bool {
	oi := l.GetObjectInfoNode()
	if oi != nil {
		return ForEachAttributeNode(oi, f)
	}
	return true
}

// ForEachAttributeNode calls f with each attribute node, depth
// first. If f returns false, iteration stops. This function avoids loops
func ForEachAttributeNode(root LayerNode, f func(LayerNode) bool) bool {
	return forEachAttributeNode(root, f, map[LayerNode]struct{}{})
}

func forEachAttributeNode(root LayerNode, f func(LayerNode) bool, loop map[LayerNode]struct{}) bool {
	if _, exists := loop[root]; exists {
		return true
	}
	loop[root] = struct{}{}
	defer delete(loop, root)

	if root.IsAttributeNode() {
		if !f(root) {
			return false
		}
	}
	for outgoing := root.AllOutgoingEdges(); outgoing.HasNext(); {
		edge := outgoing.Next().(LayerEdge)
		if !edge.IsAttributeTreeEdge() {
			continue
		}
		next := edge.To().(LayerNode)
		if next.HasType(AttributeTypes.Attribute) {
			if !forEachAttributeNode(next, f, loop) {
				return false
			}
		}
	}
	return true
}
