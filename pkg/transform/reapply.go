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

package transform

import (
	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// ApplyLayer applies the given layer onto the graph.
//
// The annotations of the given layer will be composed with all
// matching nodes of the graph. If reinterpretValues is set, the
// operation will get the node value, compose, and set it back, so
// this can be used for type conversions.
func ApplyLayer(ctx *ls.Context, g *lpg.Graph, layer *ls.Layer, reinterpretValues bool) error {
	var applyErr error

	processNode := func(layerNode *lpg.Node) bool {
		layerNodeID := ls.GetAttributeID(layerNode)
		if len(layerNodeID) == 0 {
			return true
		}
		// Find document graph nodes for this layer node
		pattern := lpg.Pattern{
			{
				Labels: lpg.NewStringSet(ls.DocumentNodeTerm),
				Properties: map[string]interface{}{
					ls.SchemaNodeIDTerm: ls.StringPropertyValue(ls.SchemaNodeIDTerm, layerNodeID),
				},
			}}
		nodes, err := pattern.FindNodes(g, nil)
		if err != nil {
			applyErr = err
			return false
		}
		for _, node := range nodes {
			var value interface{}
			if reinterpretValues && node.HasLabel(ls.AttributeTypeValue) {
				value, err = ls.GetNodeValue(node)
				if err != nil {
					applyErr = err
					return false
				}
			}
			if err := ls.ComposeProperties(ctx, node, layerNode); err != nil {
				applyErr = err
				return false
			}
			if reinterpretValues && node.HasLabel(ls.AttributeTypeValue) {
				if err := ls.SetNodeValue(node, value); err != nil {
					applyErr = err
					return false
				}
			}
		}
		// Find schema graph nodes for this layer node
		// This is required if schema nodes were not embedded
		pattern = lpg.Pattern{
			{
				Labels: lpg.NewStringSet(ls.AttributeNodeTerm),
				Properties: map[string]interface{}{
					ls.NodeIDTerm: layerNodeID,
				},
			}}
		nodes, err = pattern.FindNodes(g, nil)
		if err != nil {
			applyErr = err
			return false
		}
		for _, node := range nodes {
			if err := ls.ComposeProperties(ctx, node, layerNode); err != nil {
				applyErr = err
				return false
			}
		}
		return true
	}

	for _, layerNode := range layer.GetOverlayAttributes() {
		processNode(layerNode)
	}
	// Process each node of the layer
	layer.ForEachAttribute(func(layerNode *lpg.Node, _ []*lpg.Node) bool {
		return processNode(layerNode)
	})
	return applyErr
}
