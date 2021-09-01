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

// Copy in to target using the node selectors. Return a node map from the in to target nodes
func Copy(in, target *digraph.Graph, nodeSelector func(Node) bool, edgeSelector func(Edge) bool) map[Node]Node {
	nodeMap := make(map[Node]Node)
	for n := in.GetAllNodes(); n.HasNext(); {
		node := n.Next().(Node)
		if nodeSelector == nil || nodeSelector(node) {
			newNode := node.Clone()
			nodeMap[node] = newNode
			target.AddNode(newNode)
		}
	}
	for src, dest := range nodeMap {
		edges := src.Out()
		for edges.HasNext() {
			edge := edges.Next().(Edge)
			newTo := nodeMap[edge.GetTo().(Node)]
			if newTo != nil {
				if edgeSelector == nil || edgeSelector(edge) {
					digraph.Connect(dest, newTo, edge.Clone())
				}
			}
		}
	}
	return nodeMap
}
