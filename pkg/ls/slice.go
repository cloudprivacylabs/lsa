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
	"github.com/cloudprivacylabs/lpg/v2"
)

// GetSliceByTermsFunc is used in the Slice function to select nodes
// by the properties it contains. If includeAttributeNodes is true,
// attributes nodes are included unconditionally even though the node
// does not contain any of the terms.
func GetSliceByTermsFunc(includeTerms []string, includeAttributeNodes bool) func(*Layer, *lpg.Node) *lpg.Node {
	incl := make(map[string]struct{})
	for _, x := range includeTerms {
		incl[x] = struct{}{}
	}
	return func(layer *Layer, nd *lpg.Node) *lpg.Node {
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
var IncludeAllNodesInSliceFunc = func(layer *Layer, nd *lpg.Node) *lpg.Node {
	return CloneNode(nd, layer.Graph)
}

func (layer *Layer) Slice(layerType string, nodeFilter func(*Layer, *lpg.Node) *lpg.Node) *Layer {
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
	nodeMap := make(map[*lpg.Node]*lpg.Node)
	for targets := sourceRoot.GetEdges(lpg.OutgoingEdge); targets.Next(); {
		edge := targets.Edge()
		if IsAttributeTreeEdge(edge) {
			newNode := slice(ret, edge.GetTo(), nodeFilter, nodeMap)
			if newNode != nil {
				ret.Graph.NewEdge(rootNode, newNode, edge.GetLabel(), nil)
			}
		}
	}
	ret.Graph.NewEdge(ret.GetLayerRootNode(), rootNode, LayerRootTerm, nil)
	return ret
}

func slice(targetLayer *Layer, sourceNode *lpg.Node, nodeFilter func(*Layer, *lpg.Node) *lpg.Node, nodeMap map[*lpg.Node]*lpg.Node) *lpg.Node {
	// If the sourceNode was seen before, link to it
	if tgt, ok := nodeMap[sourceNode]; ok {
		return tgt
	}
	// Try to filter first. This may return nil
	targetNode := nodeFilter(targetLayer, sourceNode)
	if targetNode != nil {
		nodeMap[sourceNode] = targetNode
	}

	for edges := sourceNode.GetEdges(lpg.OutgoingEdge); edges.Next(); {
		edge := edges.Edge()
		newTo := slice(targetLayer, edge.GetTo(), nodeFilter, nodeMap)
		if newTo != nil {
			// If targetNode was filtered out, it has to be included now
			if targetNode == nil {
				targetNode = targetLayer.Graph.NewNode(sourceNode.GetLabels().Slice(), nil)
				nodeMap[sourceNode] = targetNode
				if len(GetAttributeID(sourceNode)) > 0 {
					SetAttributeID(targetNode, GetAttributeID(sourceNode))
				}
			}
			targetLayer.Graph.NewEdge(targetNode, newTo, edge.GetLabel(), CloneProperties(edge))
		}
	}
	return targetNode
}
