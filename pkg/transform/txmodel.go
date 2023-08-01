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

package transform

import (
	"fmt"

	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// txDocNode is used as an intermediate model for transformation procedures
type txDocNode struct {
	schemaNode *lpg.Node
	typeTerm   string
	id         string
	rawValue   string
	sourceNode *lpg.Node
	value      any
	valueTypes []string

	attributeIndex int
	attributeName  string

	children   []ls.ParsedDocNode
	properties map[string]any
}

func newTxDocNode(schemaNode *lpg.Node) *txDocNode {
	ret := &txDocNode{
		schemaNode: schemaNode,
		properties: make(map[string]any),
	}
	if v, ok := ls.GetPropertyValue(schemaNode, ls.ValueTypeTerm.Name); ok {
		ret.valueTypes = v.AsStringSlice()
	}
	return ret
}

func (node *txDocNode) findChildInstanceOf(schemaNode *lpg.Node) []*txDocNode {
	ret := make([]*txDocNode, 0)
	for _, x := range node.children {
		if x.(*txDocNode).schemaNode == schemaNode {
			ret = append(ret, x.(*txDocNode))
		}
	}
	return ret
}

func (node *txDocNode) String() string {
	var out string
	for _, x := range node.children {
		v := x.(*txDocNode).String()
		out += "  " + v + "\n"
	}
	return fmt.Sprintf("[%s %s]\n%s", ls.GetNodeID(node.schemaNode), node.rawValue, out)
}

func (node *txDocNode) GetSchemaNode() *lpg.Node        { return node.schemaNode }
func (node *txDocNode) GetID() string                   { return node.id }
func (node *txDocNode) GetAttributeIndex() int          { return node.attributeIndex }
func (node *txDocNode) GetAttributeName() string        { return node.attributeName }
func (node *txDocNode) GetChildren() []ls.ParsedDocNode { return node.children }
func (node *txDocNode) GetProperties() map[string]any   { return node.properties }
func (node *txDocNode) GetTypeTerm() string             { return node.typeTerm }
func (node *txDocNode) GetValueTypes() []string         { return node.valueTypes }
func (node *txDocNode) GetNativeValue() (any, bool) {
	if node.value == nil {
		return nil, false
	}
	return node.value, true
}

func (node *txDocNode) GetValue() string {
	return node.rawValue
}

func (node *txDocNode) GetSchemaNodeID() string {
	if node.schemaNode == nil {
		return ""
	}
	return ls.GetNodeID(node.schemaNode)
}
