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

import ()

// GetSliceByTermsFunc is used in the Slice function to select nodes by
// the properties it contains.
func GetSliceByTermsFunc(includeTerms []string) func(*Layer, *SchemaNode) *SchemaNode {
	incl := make(map[string]struct{})
	for _, x := range includeTerms {
		incl[x] = struct{}{}
	}
	return func(layer *Layer, node *SchemaNode) *SchemaNode {
		properties := make(map[string]interface{})
		for k, v := range node.Properties {
			if _, ok := incl[k]; ok {
				properties[k] = v
			}
		}
		if len(properties) > 0 {
			newNode := node.Clone()
			newNode.Properties = properties
			return newNode
		}
		return nil
	}
}

// IncludeAllNodesInSlliceFunc includes all the nodes in the slice
var IncludeAllNodesInSliceFunc = func(layer *Layer, node *SchemaNode) *SchemaNode {
	return node.Clone()
}

func (layer *Layer) Slice(layerType string, nodeFilter func(*Layer, *SchemaNode) *SchemaNode) *Layer {
	ret := NewLayer("", layerType)
	rootNode := ret.GetRoot()
	for targets := layer.GetRoot().AllOutgoingEdges(); targets.HasNext(); {
		edge := targets.Next().(*SchemaEdge)
		if edge.IsAttributeTreeEdge() {
			newNode := slice(ret, layer.GetRoot(), nodeFilter, map[*SchemaNode]struct{}{})
			if newNode != nil {
				rootNode.Connect(newNode, edge.GetLabel())
			}
		}
	}
	return ret
}

func slice(targetLayer *Layer, sourceNode *SchemaNode, nodeFilter func(*Layer, *SchemaNode) *SchemaNode, ctx map[*SchemaNode]struct{}) *SchemaNode {
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
		edge := edges.Next().(*SchemaEdge)
		newTo := slice(targetLayer, edge.To().(*SchemaNode), nodeFilter, ctx)
		if newTo != nil {
			// If targetNode was filtered out, it has to be included now
			if targetNode == nil {
				targetNode = targetLayer.NewNode(sourceNode.GetID(), sourceNode.GetTypes()...)
			}
			// Add the edge
			targetLayer.AddEdge(targetNode, newTo, edge.Clone())
		}
	}
	return targetNode
}
