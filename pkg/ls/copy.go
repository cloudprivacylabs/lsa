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

// CopyGraph source to target using the optional node/edge selectors. Return a node map from the in to target nodes
func CopyGraph(target, source graph.Graph, nodeSelector func(graph.Node) bool, edgeSelector func(graph.Edge) bool) map[graph.Node]graph.Node {
	if nodeSelector == nil {
		nodeSelector = func(graph.Node) bool { return true }
	}
	if edgeSelector == nil {
		edgeSelector = func(graph.Edge) bool { return true }
	}
	nodeMap := make(map[graph.Node]graph.Node)
	for n := source.GetNodes(); n.Next(); {
		node := n.Node()
		if nodeSelector(node) {
			newNode := CloneNode(node, target)
			nodeMap[node] = newNode
		}
	}
	for edges := source.GetEdges(); edges.Next(); {
		edge := edges.Edge()
		if edgeSelector(edge) {
			CloneEdge(nodeMap[edge.GetFrom()], nodeMap[edge.GetTo()], edge, target)
		}
	}
	return nodeMap
}
