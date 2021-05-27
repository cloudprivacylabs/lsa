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

const LS = "https://layeredschemas.org/"

const SchemaTerm = LS + "Schema"
const OverlayTerm = LS + "Overlay"

const TargetType = LS + "targetType"

type Layer struct {
	*digraph.Graph
	RootNode *digraph.Node
}

func NewLayer() *Layer {
	return &Layer{Graph: digraph.New()}
}

// Clone returns a copy of the layer
func (l *Layer) Clone() *Layer {
	ret := &Layer{Graph: digraph.New()}
	nodeMap := make(map[*digraph.Node]*digraph.Node)
	for nodes := l.AllNodes(); nodes.HasNext(); {
		oldNode := nodes.Next()
		newNode := ret.NewNode(oldNode.Label())
		newNode.Payload = oldNode.Payload.(*SchemaNode).Clone()
		nodeMap[oldNode] = newNode
	}
	ret.RootNode = nodeMap[l.RootNode]
	for nodes := l.AllNodes(); nodes.HasNext(); {
		node := nodes.Next()
		for edges := node.AllOutgoingEdges(); edges.HasNext(); {
			edge := edges.Next()
			var p interface{}
			if edge.Payload != nil {
				p = edge.Payload.(*SchemaEdge).Clone()
			}
			ret.NewEdge(nodeMap[edge.From()], nodeMap[edge.To()], edge.Label(), p)
		}
	}
	return ret
}

// GetID returns the ID of the layer, which is the ID of the root node
func (l *Layer) GetID() string {
	return l.RootNode.Label().(string)
}

// GetLayerType returns the layer type, SchemaTerm or OverlayTerm.
func (l *Layer) GetLayerType() string {
	schNode := l.RootNode.Payload.(*SchemaNode)
	if schNode.HasType(SchemaTerm) {
		return SchemaTerm
	}
	if schNode.HasType(OverlayTerm) {
		return OverlayTerm
	}
	return ""
}

func (l *Layer) NewNode(label interface{}, types ...string) *digraph.Node {
	return l.Graph.NewNode(label, NewSchemaNode(types...))
}

// GetTargetTypes returns the value of the targetType field
func (l *Layer) GetTargetTypes() []string {
	schNode := l.RootNode.Payload.(*SchemaNode)
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
func (l *Layer) ForEachAttribute(f func(*digraph.Node) bool) {
	var forEachAttribute func(*digraph.Node, func(*digraph.Node) bool) bool
	forEachAttribute = func(root *digraph.Node, f func(*digraph.Node) bool) bool {
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
			next := edge.To()
			np, _ := next.Payload.(*SchemaNode)
			if np.HasType(AttributeTypes.Attribute) {
				if !forEachAttribute(next, f) {
					return false
				}
			}
		}
		return true
	}

	forEachAttribute(l.RootNode, f)
}
