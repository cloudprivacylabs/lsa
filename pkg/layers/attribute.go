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
package layers

import (
	"github.com/bserdar/digraph"
)

// IRI is used for jsonld @id types
type IRI string

// IsAttributeTreeEdge returns true if the edge is an edge between two
// attribute nodes
func IsAttributeTreeEdge(edge digraph.Edge) bool {
	l := edge.Label()
	return l == TypeTerms.Attributes ||
		l == TypeTerms.AttributeList ||
		l == TypeTerms.ArrayItems ||
		l == TypeTerms.AllOf ||
		l == TypeTerms.OneOf
}

// IsAttributeNode returns true if the node has Attribute type
func IsAttributeNode(node digraph.Node) bool {
	payload, _ := node.(*SchemaNode)
	return payload != nil && payload.HasType(AttributeTypes.Attribute)
}

// GetAttributeTypes returns all recognized attribute types. This is
// mainly used for validation, to ensure there is only one attribute
// type
func GetAttributeTypes(types []string) []string {
	ret := make([]string, 0)
	for _, x := range types {
		if x == AttributeTypes.Value ||
			x == AttributeTypes.Object ||
			x == AttributeTypes.Array ||
			x == AttributeTypes.Reference ||
			x == AttributeTypes.Composite ||
			x == AttributeTypes.Polymorphic {
			ret = append(ret, x)
		}
	}
	return ret
}

// SchemaEdge is the payload for schema edges.
type SchemaEdge struct {
	digraph.EdgeHeader
	Properties map[string]interface{}
	Compiled   map[string]interface{}
}

// NewSchemaEdge returns a new initialized schema edge
func NewSchemaEdge(label interface{}) *SchemaEdge {
	ret := &SchemaEdge{Properties: make(map[string]interface{}),
		Compiled: make(map[string]interface{}),
	}
	ret.SetLabel(label)
	return ret
}

func (e *SchemaEdge) Clone() *SchemaEdge {
	ret := NewSchemaEdge(e.Label())
	ret.Properties = copyIntf(e.Properties).(map[string]interface{})
	return ret
}
