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

type LayerEdge interface {
	digraph.Edge

	GetLabel() string
	IsAttributeTreeEdge() bool

	// Clone returns a new edge that is a copy of this one but
	// unconnected to any nodes
	Clone() LayerEdge

	GetPropertyMap() map[string]*PropertyValue

	GetCompiledDataMap() map[interface{}]interface{}
}

// schemaEdge is a labeled graph edge between two schema nodes
type schemaEdge struct {
	digraph.EdgeHeader
	properties map[string]*PropertyValue
	compiled   map[interface{}]interface{}
}

func (edge *schemaEdge) GetCompiledDataMap() map[interface{}]interface{} { return edge.compiled }

func (edge *schemaEdge) GetPropertyMap() map[string]*PropertyValue { return edge.properties }

// NewLayerEdge returns a new initialized schema edge
func NewLayerEdge(label string) LayerEdge {
	ret := &schemaEdge{
		properties: make(map[string]*PropertyValue),
		compiled:   make(map[interface{}]interface{}),
	}
	ret.SetLabel(label)
	return ret
}

// GetLabel returns the edge label
func (edge *schemaEdge) GetLabel() string {
	if edge == nil {
		return ""
	}
	l := edge.Label()
	if l == nil {
		return ""
	}
	return l.(string)
}

// IsAttributeTreeEdge returns true if the edge is an edge between two
// attribute nodes
func (edge *schemaEdge) IsAttributeTreeEdge() bool {
	if edge == nil {
		return false
	}
	l := edge.Label()
	return l == LayerTerms.Attributes ||
		l == LayerTerms.AttributeList ||
		l == LayerTerms.ArrayItems ||
		l == LayerTerms.AllOf ||
		l == LayerTerms.OneOf
}

// Clone returns a copy of the schema edge
func (e *schemaEdge) Clone() LayerEdge {
	ret := NewLayerEdge(e.GetLabel()).(*schemaEdge)
	ret.properties = CopyPropertyMap(e.properties)
	return ret
}
