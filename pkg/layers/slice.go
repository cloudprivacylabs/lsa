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

func (layer *Layer) Slice(layerType string, nodeFilter func(*Layer, *digraph.Node) *digraph.Node) *Layer {
	ret := &Layer{Graph: digraph.New()}
	ret.RootNode = slice(ret, layer.RootNode, nodeFilter, map[*digraph.Node]struct{}{})
	if ret.RootNode == nil {
		ret.RootNode = ret.NewNode(layer.RootNode.Label())
	}
	ret.RootNode.Payload.(*SchemaNode).RemoveTypes(SchemaTerm, OverlayTerm)
	ret.RootNode.Payload.(*SchemaNode).AddTypes(layerType)
	return ret
}

func slice(targetLayer *Layer, source *digraph.Node, nodeFilter func(*Layer, *digraph.Node) *digraph.Node, ctx map[*digraph.Node]struct{}) *digraph.Node {
	// Avoid loops
	if _, seen := ctx[source]; seen {
		return nil
	}
	ctx[source] = struct{}{}
	defer func() {
		delete(ctx, source)
	}()

	// Try to filter first. This may return nil
	targetNode := nodeFilter(targetLayer, source)

	for edges := source.AllOutgoingEdges(); edges.HasNext(); {
		edge := edges.Next()
		newTo := slice(targetLayer, edge.To(), nodeFilter, ctx)
		if newTo != nil {
			// If targetNode was filtered out, it has to be included now
			if targetNode == nil {
				targetNode = targetLayer.NewNode(source.Label(), source.Payload.(*SchemaNode).GetTypes()...)
			}
			// Add the edge
			targetLayer.NewEdge(targetNode, newTo, edge.Label(), edge.Payload)
		}
	}
	return targetNode
}

// Recursively copy in
func copyIntf(in interface{}) interface{} {
	if arr, ok := in.([]interface{}); ok {
		ret := make([]interface{}, 0, len(arr))
		for _, x := range arr {
			ret = append(ret, copyIntf(x))
		}
		return ret
	}
	if m, ok := in.(map[string]interface{}); ok {
		ret := make(map[string]interface{}, len(m))
		for k, v := range m {
			ret[k] = copyIntf(v)
		}
		return ret
	}
	return in
}
