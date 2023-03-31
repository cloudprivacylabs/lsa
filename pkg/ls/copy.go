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

// CopyGraph source to target using the optional node/edge selectors. Return a node map from the in to target nodes
func CopyGraph(target, source *lpg.Graph, nodeSelector func(*lpg.Node) bool, edgeSelector func(*lpg.Edge) bool) map[*lpg.Node]*lpg.Node {
	if nodeSelector == nil {
		nodeSelector = func(*lpg.Node) bool { return true }
	}
	if edgeSelector == nil {
		edgeSelector = func(*lpg.Edge) bool { return true }
	}
	nodeMap := make(map[*lpg.Node]*lpg.Node)
	for n := source.GetNodes(); n.Next(); {
		node := n.Node()
		if nodeSelector(node) {
			newNode := CloneNode(node, target)
			nodeMap[node] = newNode
		}
	}
	for edges := source.GetEdges(); edges.Next(); {
		edge := edges.Edge()
		from := nodeMap[edge.GetFrom()]
		to := nodeMap[edge.GetTo()]
		if from == nil || to == nil {
			continue
		}
		if edgeSelector(edge) {
			CloneEdge(from, to, edge, target)
		}
	}
	return nodeMap
}

// CopyGraphWithCloneFunc source to target using the optional node/edge cloners. Return a node map from the in to target nodes
func CopyGraphWithCloneFunc(target, source *lpg.Graph, nodeCloner func(*lpg.Node, *lpg.Graph) *lpg.Node, edgeCloner func(*lpg.Edge, map[*lpg.Node]*lpg.Node, *lpg.Graph) *lpg.Edge) map[*lpg.Node]*lpg.Node {
	if nodeCloner == nil {
		nodeCloner = CloneNode
	}
	if edgeCloner == nil {
		edgeCloner = func(source *lpg.Edge, nodeMap map[*lpg.Node]*lpg.Node, g *lpg.Graph) *lpg.Edge {
			from := nodeMap[source.GetFrom()]
			to := nodeMap[source.GetTo()]
			if from == nil || to == nil {
				return nil
			}
			return CloneEdge(from, to, source, g)
		}
	}
	nodeMap := make(map[*lpg.Node]*lpg.Node)
	for n := source.GetNodes(); n.Next(); {
		node := n.Node()
		newNode := nodeCloner(node, target)
		if newNode != nil {
			nodeMap[node] = newNode
		}
	}
	for edges := source.GetEdges(); edges.Next(); {
		edge := edges.Edge()
		edgeCloner(edge, nodeMap, target)
	}
	return nodeMap
}
