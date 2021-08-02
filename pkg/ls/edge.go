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

type Edge interface {
	digraph.Edge

	GetLabelStr() string
	IsAttributeTreeEdge() bool

	// Indexes are used to index ordered array edges.
	GetIndex() int
	SetIndex(int)

	// Clone returns a new edge that is a copy of this one but
	// unconnected to any nodes
	Clone() Edge
	CloneWithLabel(string) Edge

	GetProperties() map[string]*PropertyValue

	GetCompiledDataMap() map[interface{}]interface{}
}

// edge is a labeled graph edge between two nodes
type edge struct {
	digraph.EdgeHeader
	index      int
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

func (edge *edge) GetIndex() int  { return edge.index }
func (edge *edge) SetIndex(i int) { edge.index = i }

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
func (edge *edge) IsAttributeTreeEdge() bool {
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
	return e.CloneWithLabel(e.GetLabelStr())
}

// CloneWithLabel returns a copy of the schema edge with a new label
func (e *edge) CloneWithLabel(label string) Edge {
	ret := NewEdge(label).(*edge)
	ret.properties = CopyPropertyMap(e.properties)
	return ret
}
