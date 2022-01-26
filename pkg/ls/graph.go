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

// GraphIndex stores a fast-accessible representation of a graph.
type GraphIndex struct {
	// Key: nodeID
	nodesByID map[string]UnorderedNodeSet
	// Key: node.types
	nodesByType map[string]UnorderedNodeSet
	// Key: target node, label
	edgesByTarget map[Node]map[string]EdgeSet
}

// UnorderedNodeSet keeps nodes
type UnorderedNodeSet map[Node]struct{}

func (n UnorderedNodeSet) Add(node Node)    { n[node] = struct{}{} }
func (n UnorderedNodeSet) Remove(node Node) { delete(n, node) }
func (n UnorderedNodeSet) Len() int         { return len(n) }

// ForEach calls f for each member of n until f returns false. Returns
// true if all elements are processed
func (n UnorderedNodeSet) ForEach(f func(Node) bool) bool {
	if len(n) == 0 {
		return true
	}
	for k := range n {
		if !f(k) {
			return false
		}
	}
	return true
}

// EdgeSet keeps edges
type EdgeSet map[Edge]struct{}

// NewEdgeSet creates a new edge set containing the given edges
func NewEdgeSet(edge ...Edge) EdgeSet {
	ret := make(EdgeSet, len(edge))
	for _, k := range edge {
		ret[k] = struct{}{}
	}
	return ret
}

// Slice returns edges in the set as a slice
func (set EdgeSet) Slice() []Edge {
	ret := make([]Edge, 0, len(set))
	for k := range set {
		ret = append(ret, k)
	}
	return ret
}

func (e EdgeSet) Add(edge Edge)    { e[edge] = struct{}{} }
func (e EdgeSet) Remove(edge Edge) { delete(e, edge) }
func (e EdgeSet) Len() int         { return len(e) }

// ForEach calls f for each member of e until f returns false. Returns
// true if all elements are processed
func (e EdgeSet) ForEach(f func(Edge) bool) bool {
	if len(e) == 0 {
		return true
	}
	for k := range e {
		if !f(k) {
			return false
		}
	}
	return true
}

// NewGraphIndex creates a new empty graph index
func NewGraphIndex() *GraphIndex {
	return &GraphIndex{
		nodesByID:     make(map[string]UnorderedNodeSet),
		nodesByType:   make(map[string]UnorderedNodeSet),
		edgesByTarget: make(map[Node]map[string]EdgeSet),
	}
}

// AddTree adds all accessible nodes starting from root into the index
func (index *GraphIndex) AddTree(root Node) {
	IterateDescendants(root, func(node Node, _ []Node) bool {
		index.AddNode(node)
		return true
	}, func(edge Edge, _ []Node) EdgeFuncResult {
		index.AddEdge(edge)
		return FollowEdgeResult
	}, false)
}

func addStringNodeMap(target map[string]UnorderedNodeSet, key string, value Node) {
	set, ok := target[key]
	if !ok {
		set = make(UnorderedNodeSet)
		target[key] = set
	}
	set.Add(value)
}

func addNodeStringEdgeMap(target map[Node]map[string]EdgeSet, key1 Node, key2 string, value Edge) {
	stringSet, ok := target[key1]
	if !ok {
		stringSet = make(map[string]EdgeSet)
		target[key1] = stringSet
	}
	edgeSet, ok := stringSet[key2]
	if !ok {
		edgeSet = make(EdgeSet)
		stringSet[key2] = edgeSet
	}
	edgeSet.Add(value)
}

// AddNode adds a single node, without the edges, into the index
func (index *GraphIndex) AddNode(node Node) {
	addStringNodeMap(index.nodesByID, node.GetID(), node)
	for _, s := range node.GetTypes().Slice() {
		addStringNodeMap(index.nodesByType, s, node)
	}
}

// AddEdge adds a single edge to the index, without the nodes
func (index *GraphIndex) AddEdge(edge Edge) {
	addNodeStringEdgeMap(index.edgesByTarget, edge.GetTo().(Node), edge.GetLabelStr(), edge)
}

// ScanNodes calls f for each node until f returns false. Returns
// false if f returned false, true if not
func (index *GraphIndex) ScanNodes(f func(Node) bool) bool {
	for _, v := range index.nodesByID {
		if !v.ForEach(f) {
			return false
		}
	}
	return true
}

// ScanNodesByAnyType calls f for nodes that have any one of the given
// types until f returns false. Returns false if f returns false, true
// if not
func (index *GraphIndex) ScanNodesByAnyType(f func(Node) bool, types ...string) bool {
	for typeIndex, currentType := range types {
		currentSet := index.nodesByType[currentType]
		if !currentSet.ForEach(func(node Node) bool {
			// Is this node seen before? Node is seen if it also has one of
			// the types in the types array that we scanned before this
			if !node.GetTypes().HasAny(types[:typeIndex]...) {
				if !f(node) {
					return false
				}
			}
			return true
		}) {
			return false
		}
	}
	return true
}

// ScanSourceNodes calls f for each source node of target, until f
// returns false. Returns false if f returns false, true otherwise
func (index *GraphIndex) ScanSourceNodes(f func(Edge) bool, target Node) bool {
	set := index.edgesByTarget[target]
	if len(set) == 0 {
		return true
	}
	for _, v := range set {
		if !v.ForEach(f) {
			return false
		}
	}
	return true
}

// ScanSourceNodesByLabel calls f for each source node of target, until f
// returns false. Returns false if f returns false, true otherwise
func (index *GraphIndex) ScanSourceNodesByLabel(f func(Edge) bool, target Node, label string) bool {
	set := index.edgesByTarget[target]
	if len(set) == 0 {
		return true
	}
	edges := set[label]
	if len(edges) == 0 {
		return true
	}
	return edges.ForEach(f)
}
