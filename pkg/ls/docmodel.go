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

// DocumentNode is the graph node type used to store ingested document nodes
type DocumentNode interface {
	digraph.Node

	GetID() string
	SetValue(interface{})
	GetValue() interface{}
	GetProperty(key string) (*PropertyValue, bool)
	SetProperty(key string, value *PropertyValue)
	GetProperties() map[string]*PropertyValue
}

// BasicDocumentNode is derived from BasicNode. It uses @value property to store the node value
type BasicDocumentNode struct {
	digraph.NodeHeader
	Value      interface{}
	Properties map[string]*PropertyValue
}

// NewBasicDocumentNode returns an initialized basic document node with the given ID
func NewBasicDocumentNode(ID string) *BasicDocumentNode {
	ret := &BasicDocumentNode{Properties: make(map[string]*PropertyValue)}
	ret.SetLabel(ID)
	return ret
}

func (node *BasicDocumentNode) GetID() string { return node.Label().(string) }

// SetValue sets the value
func (node *BasicDocumentNode) SetValue(value interface{}) {
	node.Value = value
}

// GetValue returns the value proeprty
func (node *BasicDocumentNode) GetValue() interface{} {
	return node.Value
}

func (node *BasicDocumentNode) GetProperty(key string) (*PropertyValue, bool) {
	p, ok := node.Properties[key]
	return p, ok
}

func (node *BasicDocumentNode) SetProperty(key string, value *PropertyValue) {
	node.Properties[key] = value
}

func (node *BasicDocumentNode) GetProperties() map[string]*PropertyValue { return node.Properties }

// GetFilteredValue returns the field value processed by the schema
// value filters, and then the node value filters
func (node *BasicDocumentNode) GetFilteredValue() interface{} {
	schemaNode, _ := node.NextNode(InstanceOfTerm).(LayerNode)
	return GetFilteredValue(schemaNode, node)
}

// GetFilteredValue filters the value through the schema properties
// and then through the node properties before returning
func GetFilteredValue(schemaNode LayerNode, docNode DocumentNode) interface{} {
	value := docNode.GetValue()
	if schemaNode != nil {
		value = FilterValue(value, docNode, schemaNode.GetProperties())
	}
	return FilterValue(value, docNode, docNode.GetProperties())
}

// IsDocumentEdge returns true if the edge is a data edge term
func IsDocumentEdge(edge digraph.Edge) bool {
	switch edge.Label() {
	case DataEdgeTerms.ObjectAttributes, DataEdgeTerms.ArrayElements:
		return true
	}
	return false
}

func ForEachDocumentNode(root DocumentNode, f func(DocumentNode) bool) bool {
	return forEachDocumentNode(root, f, map[DocumentNode]struct{}{})
}

func forEachDocumentNode(root DocumentNode, f func(DocumentNode) bool, seen map[DocumentNode]struct{}) bool {
	if _, exists := seen[root]; exists {
		return true
	}
	seen[root] = struct{}{}

	if !f(root) {
		return false
	}

	for outgoing := root.AllOutgoingEdges(); outgoing.HasNext(); {
		edge := outgoing.Next()
		if !IsDocumentEdge(edge) {
			continue
		}
		next, ok := edge.To().(DocumentNode)
		if ok {
			if !forEachDocumentNode(next, f, seen) {
				return false
			}
		}
	}
	return true
}
