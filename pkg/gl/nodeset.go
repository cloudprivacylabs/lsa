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

package gl

import (
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

// NodeSet is a set of nodes
type NodeSet struct {
	set   map[graph.Node]struct{}
	nodes []graph.Node
}

// NewNodeSet constructs a new nodeset from the given nodes
func NewNodeSet(node ...graph.Node) NodeSet {
	ret := NodeSet{set: make(map[graph.Node]struct{})}
	ret.Add(node...)
	return ret
}

// Add adds nodes to the set
func (n *NodeSet) Add(nodes ...graph.Node) {
	for _, k := range nodes {
		if !n.Has(k) {
			n.set[k] = struct{}{}
			n.nodes = append(n.nodes, k)
		}
	}
}

// Has returns true if node is in the set
func (n NodeSet) Has(node graph.Node) bool {
	_, ok := n.set[node]
	return ok
}

// Delete some nodes from the set
func (n *NodeSet) Delete(nodes ...graph.Node) {
	for _, k := range nodes {
		delete(n.set, k)
	}
	w := 0
	for i := 0; i < len(n.nodes); i++ {
		if _, ok := n.set[n.nodes[i]]; ok {
			n.nodes[w] = n.nodes[i]
			w++
		}
	}
	n.nodes = n.nodes[:w]
}

// Slice returns the nodes in a nodeset as a slice
func (n NodeSet) Slice() []graph.Node {
	return n.nodes
}

// Set returns the nodes as a map
func (n NodeSet) Map() map[graph.Node]struct{} {
	return n.set
}

func (n NodeSet) Len() int { return len(n.nodes) }

// EqualSet returns if the two nodesets are equal without taking into account the node ordering
func (n NodeSet) EqualSet(n2 NodeSet) bool {
	if n.Len() != n2.Len() {
		return false
	}
	for k := range n.Map() {
		if !n2.Has(k) {
			return false
		}
	}
	return true
}
