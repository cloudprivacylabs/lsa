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
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

// ApplyLayer applies the given layer onto the graph.
//
// The annotations of the given layer will be composed with all
// matching nodes of the graph.
func ApplyLayer(ctx *ls.Context, g graph.Graph, layer *ls.Layer) error {
	var applyErr error
	// Process each node of the layer
	layer.ForEachAttribute(func(layerNode graph.Node, layerPath []graph.Node) bool {
		layerNodeID := ls.GetAttributeID(layerNode)
		if len(layerNodeID) == 0 {
			return true
		}
		// Find document graph nodes for this layer node
		pattern := graph.Pattern{
			{
				Labels: graph.NewStringSet(ls.DocumentNodeTerm),
				Properties: map[string]interface{}{
					ls.SchemaNodeIDTerm: ls.StringPropertyValue(layerNodeID),
				},
			}}
		nodes, err := pattern.FindNodes(g, nil)
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
		// Find schema graph nodes for this layer node
		// This is required if schema nodes were not embedded
		pattern = graph.Pattern{
			{
				Labels: graph.NewStringSet(ls.AttributeNodeTerm),
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
	})
	return applyErr
}
