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
	"sort"

	"github.com/bserdar/digraph"
)

type Edge interface {
	digraph.Edge

	GetLabelStr() string

	// Clone returns a new edge that is a copy of this one but
	// unconnected to any nodes
	Clone() Edge

	GetProperties() map[string]*PropertyValue

	GetCompiledDataMap() map[interface{}]interface{}
}

// edge is a labeled graph edge between two nodes
type edge struct {
	digraph.EdgeHeader
	properties map[string]*PropertyValue
	compiled   map[interface{}]interface{}
}

func (edge *edge) GetCompiledDataMap() map[interface{}]interface{} { return edge.compiled }

func (edge *edge) GetProperties() map[string]*PropertyValue { return edge.properties }

// NewEdge returns a new initialized  edge
func NewEdge(label string) Edge {
	ret := &edge{
		EdgeHeader: digraph.NewEdgeHeader(label),
		properties: make(map[string]*PropertyValue),
		compiled:   make(map[interface{}]interface{}),
	}
	return ret
}

// GetLabelStr returns the edge label
func (edge *edge) GetLabelStr() string {
	if edge == nil {
		return ""
	}
	l := edge.GetLabel()
	if l == nil {
		return ""
	}
	return l.(string)
}

// IsAttributeTreeEdge returns true if the edge is an edge between two
// attribute nodes
func IsAttributeTreeEdge(edge Edge) bool {
	if edge == nil {
		return false
	}
	l := edge.GetLabelStr()
	return l == LayerTerms.Attributes ||
		l == LayerTerms.AttributeList ||
		l == LayerTerms.ArrayItems ||
		l == LayerTerms.AllOf ||
		l == LayerTerms.OneOf
}

// Clone returns a copy of the schema edge
func (e *edge) Clone() Edge {
	return CloneWithLabel(e, e.GetLabelStr())
}

// CloneWithLabel returns a copy of the schema edge with a new label
func CloneWithLabel(e Edge, label string) Edge {
	ret := NewEdge(label).(*edge)
	p := ret.GetProperties()
	for k, v := range e.GetProperties() {
		p[k] = v.Clone()
	}
	return ret
}

// SortEdges sorts edges by their target node index
func SortEdges(edges []Edge) {
	sort.Slice(edges, func(i, j int) bool { return edges[i].GetTo().(Node).GetIndex() < edges[j].GetTo().(Node).GetIndex() })
}

// SortEdgesItr sorts the edges by index
func SortEdgesItr(edges digraph.Edges) digraph.Edges {
	e := make([]Edge, 0)
	for edges.HasNext() {
		e = append(e, edges.Next().(Edge))
	}
	SortEdges(e)
	arr := make([]digraph.Edge, 0, len(e))
	for _, x := range e {
		arr = append(arr, x)
	}
	return digraph.Edges{&digraph.EdgeArrayIterator{arr}}
}

// An EdgeSet is a set of edges
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
