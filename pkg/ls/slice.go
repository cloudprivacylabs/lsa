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

// GetSliceByTermsFunc is used in the Slice function to select nodes
// by the properties it contains. If includeAttributeNodes is true,
// attributes nodes are included unconditionally even though the node
// does not contain any of the terms.
func GetSliceByTermsFunc(includeTerms []string, includeAttributeNodes bool) func(*Layer, Node) Node {
	incl := make(map[string]struct{})
	for _, x := range includeTerms {
		incl[x] = struct{}{}
	}
	return func(layer *Layer, nd Node) Node {
		includeNode := false
		if includeAttributeNodes && IsAttributeNode(nd) {
			includeNode = true
		}
		properties := make(map[string]*PropertyValue)
		for k, v := range nd.GetProperties() {
			if _, ok := incl[k]; ok {
				properties[k] = v
			}
		}
		if len(properties) > 0 || includeNode {
			newNode := nd.Clone().(*node)
			newNode.properties = properties
			return newNode
		}
		return nil
	}
}

// IncludeAllNodesInSliceFunc includes all the nodes in the slice
var IncludeAllNodesInSliceFunc = func(layer *Layer, nd Node) Node {
	return nd.Clone().(*node)
}

func (layer *Layer) Slice(layerType string, nodeFilter func(*Layer, Node) Node) *Layer {
	ret := NewLayer()
	ret.SetLayerType(layerType)
	rootNode := NewNode("")
	ret.AddNode(rootNode)
	sourceRoot := layer.GetSchemaRootNode()
	if sourceRoot != nil {
		rootNode.SetID(sourceRoot.GetID())
		rootNode.SetTypes(sourceRoot.GetTypes()...)
	}
	hasNodes := false
	if sourceRoot != nil {
		for targets := sourceRoot.GetAllOutgoingEdges(); targets.HasNext(); {
			edge := targets.Next().(Edge)
			if IsAttributeTreeEdge(edge) {
				newNode := slice(ret, edge.GetTo().(Node), nodeFilter, map[Node]struct{}{})
				if newNode != nil {
					rootNode.Connect(newNode, edge.GetLabelStr())
					hasNodes = true
				}
			}
		}
	}
	if hasNodes {
		ret.GetLayerInfoNode().Connect(rootNode, LayerRootTerm)
	}
	return ret
}

func slice(targetLayer *Layer, sourceNode Node, nodeFilter func(*Layer, Node) Node, ctx map[Node]struct{}) Node {
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

	for edges := sourceNode.GetAllOutgoingEdges(); edges.HasNext(); {
		edge := edges.Next().(Edge)
		newTo := slice(targetLayer, edge.GetTo().(Node), nodeFilter, ctx)
		if newTo != nil {
			// If targetNode was filtered out, it has to be included now
			if targetNode == nil {
				targetNode = targetLayer.NewNode(sourceNode.GetID(), sourceNode.GetTypes()...)
			}
			digraph.Connect(targetNode, newTo, edge.Clone())
		}
	}
	return targetNode
}
