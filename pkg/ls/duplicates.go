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

// ConsolidateDuplicateEntities finds all nodes that have identical
// nonempty entityId with the given label, and merges those that share labels
func ConsolidateDuplicateEntities(g *lpg.Graph, label string) {
	type rootNode struct {
		node *lpg.Node
		id   []string
	}

	allRoots := make([]rootNode, 0)
	for nodes := g.GetNodesWithAllLabels(lpg.NewStringSet(label)); nodes.Next(); {
		node := nodes.Node()
		id := AsPropertyValue(node.GetProperty(EntityIDTerm)).MustStringSlice()
		if len(id) == 0 {
			continue
		}
		allRoots = append(allRoots, rootNode{
			node: node,
			id:   id,
		})
	}

	type nodeGroup struct {
		head     rootNode
		children []rootNode
	}

	groups := make([]nodeGroup, 0, len(allRoots))

	sameId := func(id1, id2 []string) bool {
		if len(id1) != len(id2) {
			return false
		}
		for i := range id1 {
			if id1[i] != id2[i] {
				return false
			}
		}
		return true
	}

	for _, root := range allRoots {
		found := false
		for i, grp := range groups {
			if !sameId(grp.head.id, root.id) {
				continue
			}
			found = true
			groups[i].children = append(groups[i].children, root)
		}
		if !found {
			groups = append(groups, nodeGroup{head: root})
		}
	}
	for _, x := range groups {
		for _, child := range x.children {
			rm := make([]*lpg.Edge, 0)
			for edges := child.node.GetEdges(lpg.IncomingEdge); edges.Next(); {
				edge := edges.Edge()
				rm = append(rm, edge)
				g.NewEdge(edge.GetFrom(), x.head.node, edge.GetLabel(), CloneProperties(edge))
			}
			for _, edge := range rm {
				edge.Remove()
			}
			child.node.DetachAndRemove()
		}
	}
}
