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

	SetValue(interface{})
	GetValue() interface{}
	GetProperty(key string) (interface{}, bool)
	SetProperty(key string, value interface{})
	GetProperties() map[string]interface{}
}

// BasicDocumentNode is derived from BasicNode. It uses @value property to store the node value
type BasicDocumentNode struct {
	digraph.NodeHeader
	Value      interface{}
	Properties map[string]interface{}
}

// NewBasicDocumentNode returns an initialized basic document node with the given ID
func NewBasicDocumentNode(ID string) *BasicDocumentNode {
	ret := &BasicDocumentNode{Properties: make(map[string]interface{})}
	ret.SetLabel(ID)
	return ret
}

// SetValue sets the value
func (node *BasicDocumentNode) SetValue(value interface{}) {
	node.Value = value
}

// GetValue returns the value proeprty
func (node *BasicDocumentNode) GetValue() interface{} {
	return node.Value
}

func (node *BasicDocumentNode) GetProperty(key string) (interface{}, bool) {
	p, ok := node.Properties[key]
	return p, ok
}

func (node *BasicDocumentNode) SetProperty(key string, value interface{}) {
	node.Properties[key] = value
}

func (node *BasicDocumentNode) GetProperties() map[string]interface{} { return node.Properties }
