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

// A Walk specifies an edge-node-edge-node... predicate sequence. The
// predicates alter, it starts with an edge predicate, and then
// follows a node predicate, edge predicate, etc. In the end, the walk
// specifies all the nodes selected by the paths defined by the
// edge-node predicates.
type Walk struct {
	steps []stepPredicate
}

// AnyEdgePredicate accepts all edges
var AnyEdgePredicate = func(Edge) bool { return true }

// AnyNodePredicate accepts all nodes
var AnyNodePredicate = func(Node) bool { return true }

type stepPredicate struct {
	nodePredicate func(Node) bool
	edgePredicate func(Edge) bool
}

// NewWalk creates a new walk
func NewWalk() *Walk {
	return &Walk{}
}

// Step adds a new edge and node to the walk
func (w *Walk) Step(edge func(Edge) bool, node func(Node) bool) *Walk {
	w.steps = append(w.steps, stepPredicate{edgePredicate: edge, nodePredicate: node})
	return w
}

type walkState struct {
	walk *Walk
	at   int
}

func (w *walkState) next(set NodeSet) NodeSet {
	// Follow sorted edges
	ret := NewNodeSet()
	edges := make([]Edge, 0)
	for _, node := range set.Slice() {
		for outgoing := node.GetAllOutgoingEdges(); outgoing.HasNext(); {
			edge := outgoing.Next().(Edge)
			if w.walk.steps[w.at].edgePredicate(edge) {
				edges = append(edges, edge)
			}
		}
		SortEdges(edges)
		for _, e := range edges {
			node := e.GetTo().(Node)
			if !ret.Has(node) && w.walk.steps[w.at].nodePredicate(node) {
				ret.Add(node)
			}
		}
	}
	w.at++
	return ret
}

// Walk the walk starting at the given nodes, and return all the nodes arrived
func (w *Walk) Walk(start []Node) []Node {
	set := NewNodeSet(start...)
	state := walkState{walk: w, at: 0}
	for state.at = range w.steps {
		set = state.next(set)
	}
	return set.Slice()
}
