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

// GetSliceByTermsFunc is used in the Slice function to select nodes
// by the properties it contains. If includeAttributeNodes is true,
// attributes nodes are included unconditionally even though the node
// does not contain any of the terms.
func GetSliceByTermsFunc(includeTerms []string, includeAttributeNodes bool) func(*Layer, LayerNode) LayerNode {
	incl := make(map[string]struct{})
	for _, x := range includeTerms {
		incl[x] = struct{}{}
	}
	return func(layer *Layer, node LayerNode) LayerNode {
		includeNode := false
		if includeAttributeNodes && node.IsAttributeNode() {
			includeNode = true
		}
		properties := make(map[string]*PropertyValue)
		for k, v := range node.GetPropertyMap() {
			if _, ok := incl[k]; ok {
				properties[k] = v
			}
		}
		if len(properties) > 0 || includeNode {
			newNode := node.Clone().(*schemaNode)
			newNode.properties = properties
			return newNode
		}
		return nil
	}
}

// IncludeAllNodesInSliceFunc includes all the nodes in the slice
var IncludeAllNodesInSliceFunc = func(layer *Layer, node LayerNode) LayerNode {
	return node.Clone()
}

func (layer *Layer) Slice(layerType string, nodeFilter func(*Layer, LayerNode) LayerNode) *Layer {
	ret := NewLayer()
	ret.SetLayerType(layerType)
	rootNode := NewLayerNode("")
	ret.AddNode(rootNode)
	sourceRoot := layer.GetObjectInfoNode()
	if sourceRoot != nil {
		rootNode.SetID(sourceRoot.GetID())
		rootNode.SetTypes(sourceRoot.GetTypes()...)
	}
	hasNodes := false
	if sourceRoot != nil {
		for targets := sourceRoot.AllOutgoingEdges(); targets.HasNext(); {
			edge := targets.Next().(LayerEdge)
			if edge.IsAttributeTreeEdge() {
				newNode := slice(ret, edge.To().(LayerNode), nodeFilter, map[LayerNode]struct{}{})
				if newNode != nil {
					rootNode.Connect(newNode, edge.GetLabel())
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

func slice(targetLayer *Layer, sourceNode LayerNode, nodeFilter func(*Layer, LayerNode) LayerNode, ctx map[LayerNode]struct{}) LayerNode {
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

	for edges := sourceNode.AllOutgoingEdges(); edges.HasNext(); {
		edge := edges.Next().(LayerEdge)
		newTo := slice(targetLayer, edge.To().(LayerNode), nodeFilter, ctx)
		if newTo != nil {
			// If targetNode was filtered out, it has to be included now
			if targetNode == nil {
				targetNode = targetLayer.NewNode(sourceNode.GetID(), sourceNode.GetTypes()...)
			}
			// Add the edge
			if targetNode.GetGraph() == nil {
				targetLayer.AddNode(targetNode)
			}
			targetLayer.AddEdge(targetNode, newTo, edge.Clone())
		}
	}
	if targetNode != nil && targetNode.GetGraph() == nil {
		targetLayer.AddNode(targetNode)
	}
	return targetNode
}
