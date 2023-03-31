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

package csv

import (
	"fmt"
	"strings"

	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type rootNode struct {
	schemaNode *lpg.Node
	id         string
	children   []ls.ParsedDocNode
	properties map[string]interface{}
}

func (i rootNode) GetSchemaNode() *lpg.Node              { return i.schemaNode }
func (i rootNode) GetTypeTerm() string                   { return ls.AttributeTypeObject }
func (i rootNode) GetValue() string                      { return "" }
func (i rootNode) GetValueTypes() []string               { return nil }
func (i rootNode) GetChildren() []ls.ParsedDocNode       { return i.children }
func (i rootNode) GetID() string                         { return i.id }
func (i rootNode) GetProperties() map[string]interface{} { return i.properties }
func (i rootNode) GetAttributeIndex() int                { return 0 }
func (i rootNode) GetAttributeName() string              { return "" }

type cellNode struct {
	schemaNode *lpg.Node
	value      string
	name       string
	index      int
	id         string
	properties map[string]interface{}
}

func (i cellNode) GetSchemaNode() *lpg.Node              { return i.schemaNode }
func (i cellNode) GetTypeTerm() string                   { return ls.AttributeTypeValue }
func (i cellNode) GetValue() string                      { return i.value }
func (i cellNode) GetValueTypes() []string               { return nil }
func (i cellNode) GetChildren() []ls.ParsedDocNode       { return nil }
func (i cellNode) GetID() string                         { return i.id }
func (i cellNode) GetProperties() map[string]interface{} { return i.properties }
func (i cellNode) GetAttributeIndex() int                { return i.index }
func (i cellNode) GetAttributeName() string              { return i.name }

type Parser struct {
	OnlySchemaAttributes bool
	IngestNullValues     bool
	SchemaNode           *lpg.Node
	ColumnNames          []string
}

type parserContext struct {
	context    *ls.Context
	schemaNode *lpg.Node
	baseID     string
}

func (ing Parser) ParseDoc(context *ls.Context, baseID string, row []string) (ls.ParsedDocNode, error) {
	ctx := parserContext{
		context:    context,
		schemaNode: ing.SchemaNode,
		baseID:     baseID,
	}
	return ing.parseRow(ctx, row)
}

func (ing Parser) parseRow(ctx parserContext, row []string) (ls.ParsedDocNode, error) {
	if ctx.schemaNode == nil && ing.OnlySchemaAttributes {
		return nil, nil
	}
	// Get all attributes
	var allAttributes []*lpg.Node
	attributes := make(map[string][]*lpg.Node)
	if ctx.schemaNode != nil {
		for edges := ctx.schemaNode.GetEdges(lpg.OutgoingEdge); edges.Next(); {
			edge := edges.Edge()
			node := edge.GetTo()
			if !ls.IsAttributeNode(node) {
				continue
			}
			allAttributes = append(allAttributes, node)
			name := ls.AsPropertyValue(node.GetProperty(ls.AttributeNameTerm)).AsString()
			if len(name) > 0 {
				attributes[name] = append(attributes[name], node)
			}
		}
	}

	children := make([]ls.ParsedDocNode, 0, len(row))
	var id []string
	if len(ctx.baseID) > 0 {
		id = []string{ctx.baseID, ""}
	} else {
		id = []string{""}
	}
	for columnIndex, columnData := range row {
		columnData = strings.TrimSpace(columnData)
		if len(columnData) == 0 && !ing.IngestNullValues {
			continue
		}
		var newChild *cellNode
		var columnName string
		if columnIndex < len(ing.ColumnNames) {
			columnName = ing.ColumnNames[columnIndex]
		}
		var schemaNode *lpg.Node
		// if column header exists, assign schemaNode to corresponding value in attributes map
		if len(columnName) > 0 {
			schemaNodes := attributes[columnName]
			if len(schemaNodes) > 1 {
				return nil, ls.ErrInvalidSchema(fmt.Sprintf("Multiple elements with key '%s'", columnName))
			}
			if len(schemaNodes) == 1 {
				schemaNode = schemaNodes[0]
			}
			id[len(id)-1] = columnName
			newChild = &cellNode{
				schemaNode: schemaNode,
				value:      columnData,
				name:       columnName,
				id:         strings.Join(id, "."),
				index:      columnIndex,
				properties: make(map[string]any),
			}
		} else if ctx.schemaNode != nil || !ing.OnlySchemaAttributes {
			var schemaNode *lpg.Node
			for _, attr := range allAttributes {
				p := ls.AsPropertyValue(attr.GetProperty(ls.AttributeIndexTerm))
				if p == nil {
					continue
				}
				if p.IsInt() && p.AsInt() == columnIndex {
					schemaNode = attr
					break
				}
			}
			id[len(id)-1] = fmt.Sprint(columnIndex)
			newChild = &cellNode{
				schemaNode: schemaNode,
				value:      columnData,
				id:         strings.Join(id, "."),
				index:      columnIndex,
				properties: make(map[string]any),
			}
		}
		if newChild != nil {
			newChild.properties = make(map[string]interface{})
			newChild.properties[ls.AttributeIndexTerm] = ls.IntPropertyValue(ls.AttributeIndexTerm, columnIndex)
			if len(columnName) > 0 {
				newChild.properties[ls.AttributeNameTerm] = ls.StringPropertyValue(ls.AttributeNameTerm, columnName)
			}
			children = append(children, newChild)
		}
	}
	if len(children) > 0 {
		node := &rootNode{
			schemaNode: ctx.schemaNode,
			id:         ctx.baseID,
			children:   make([]ls.ParsedDocNode, 0, len(children)),
			properties: make(map[string]any),
		}
		for _, x := range children {
			node.children = append(node.children, x)
		}
		return node, nil
	}
	return nil, nil
}
