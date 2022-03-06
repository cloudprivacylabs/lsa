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
)

// GetSliceByTermsFunc is used in the Slice function to select nodes
// by the properties it contains. If includeAttributeNodes is true,
// attributes nodes are included unconditionally even though the node
// does not contain any of the terms.
func GetSliceByTermsFunc(includeTerms []string, includeAttributeNodes bool) func(*Layer, graph.Node) graph.Node {
	incl := make(map[string]struct{})
	for _, x := range includeTerms {
		incl[x] = struct{}{}
	}
	return func(layer *Layer, nd graph.Node) graph.Node {
		includeNode := false
		if includeAttributeNodes && IsAttributeNode(nd) {
			includeNode = true
		}
		properties := make(map[string]interface{})
		hasProperties := false
		nd.ForEachProperty(func(k string, v interface{}) bool {
			if prop, ok := v.(*PropertyValue); ok {
				if _, ok := incl[k]; ok {
					properties[k] = prop
					hasProperties = true
				}
			} else {
				properties[k] = v
			}
			return true
		})
		if hasProperties || includeNode {
			newNode := layer.Graph.NewNode(nd.GetLabels().Slice(), properties)
			return newNode
		}
		return nil
	}
}

// IncludeAllNodesInSliceFunc includes all the nodes in the slice
var IncludeAllNodesInSliceFunc = func(layer *Layer, nd graph.Node) graph.Node {
	return CloneNode(nd, layer.Graph)
}

func (layer *Layer) Slice(layerType string, nodeFilter func(*Layer, graph.Node) graph.Node) *Layer {
	ret := NewLayer()
	ret.SetLayerType(layerType)

	sourceRoot := layer.GetSchemaRootNode()
	if sourceRoot == nil {
		return ret
	}
	rootNode := nodeFilter(ret, sourceRoot)
	if rootNode == nil {
		rootNode = CloneNode(sourceRoot, ret.Graph)
	}
	for targets := sourceRoot.GetEdges(graph.OutgoingEdge); targets.Next(); {
		edge := targets.Edge()
		if IsAttributeTreeEdge(edge) {
			newNode := slice(ret, edge.GetTo(), nodeFilter, map[graph.Node]struct{}{})
			if newNode != nil {
				ret.Graph.NewEdge(rootNode, newNode, edge.GetLabel(), nil)
			}
		}
	}
	ret.Graph.NewEdge(ret.GetLayerRootNode(), rootNode, LayerRootTerm, nil)
	return ret
}

func slice(targetLayer *Layer, sourceNode graph.Node, nodeFilter func(*Layer, graph.Node) graph.Node, ctx map[graph.Node]struct{}) graph.Node {
	// Avoid loops
	if _, seen := ctx[sourceNode]; seen {
		return nil
	}
	ctx[sourceNode] = struct{}{}
	defer func() {
		delete(ctx, sourceNode)
	}()

	// Try to filter first. This may return nil
	targetNode := nodeFilter(targetLayer, sourceNode)

	for edges := sourceNode.GetEdges(graph.OutgoingEdge); edges.Next(); {
		edge := edges.Edge()
		newTo := slice(targetLayer, edge.GetTo(), nodeFilter, ctx)
		if newTo != nil {
			// If targetNode was filtered out, it has to be included now
			if targetNode == nil {
				targetNode = targetLayer.Graph.NewNode(sourceNode.GetLabels().Slice(), nil)
				if len(GetAttributeID(sourceNode)) > 0 {
					SetAttributeID(targetNode, GetAttributeID(sourceNode))
				}
			}
			targetLayer.Graph.NewEdge(targetNode, newTo, edge.GetLabel(), CloneProperties(edge))
		}
	}
	return targetNode
}
