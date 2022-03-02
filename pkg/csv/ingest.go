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

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

const CSV = ls.LS + "csv/"

// Ingester is a wrapper for the ls/Ingester struct
type Ingester struct {
	ls.Ingester
	ColumnNames []string
}

type propertyValueItem struct {
	of    string
	name  string
	value string
}

// Ingest a row of CSV data. The `data` slice is a row of data. The ID
// is the assigned identifier for the resulting object. ID is empty,
// IDs are assigned based on schema-dictated identifiers.
func (ingester Ingester) Ingest(context *ls.Context, targetGraph graph.Graph, data []string, ID string) (graph.Node, error) {
	ingester.PreserveNodePaths = true
	path, schemaRoot := ingester.Start(context, ID)
	// Retrieve map of schema attribute nodes from schemaRoot
	attributes, err := ingester.GetObjectAttributeNodes(context, schemaRoot)
	if err != nil {
		return nil, err
	}
	// initialize slice for storing children nodes
	children := make([]graph.Node, 0, len(data))
	childrenSchemaNodes := make(map[graph.Node]graph.Node)
	propertyValueQueue := make([]propertyValueItem, 0, len(data))
	// Iterate through each column of the CSV row
	for columnIndex, columnData := range data {
		if !ingester.IngestEmptyValues && len(columnData) == 0 {
			continue
		}
		var columnName string
		if columnIndex < len(ingester.ColumnNames) {
			columnName = ingester.ColumnNames[columnIndex]
		}
		var schemaNode graph.Node
		var newPath ls.NodePath
		// if column header exists, assign schemaNode to corresponding value in attributes map
		if len(columnName) > 0 {
			schemaNodes := attributes[columnName]
			if len(schemaNodes) > 1 {
				return nil, ls.ErrInvalidSchema(fmt.Sprintf("Multiple elements with key '%s'", columnName))
			}
			if len(schemaNodes) == 1 {
				schemaNode = schemaNodes[0]
			}
			newPath = path.AppendString(columnName)
		} else if ingester.Schema != nil || !ingester.OnlySchemaAttributes {
			schemaNode, _ = ingester.Schema.FindFirstAttribute(func(n graph.Node) bool {
				p := ls.AsPropertyValue(n.GetProperty(ls.AttributeIndexTerm))
				if p == nil {
					return false
				}
				return p.IsInt() && p.AsInt() == columnIndex
			})
			newPath = path.AppendInt(columnIndex)
		}

		propertyOf, propertyName := ls.GetAsProperty(schemaNode)
		if len(propertyName) != 0 || len(propertyOf) != 0 {
			if len(propertyName) == 0 {
				propertyName = columnName
			}
			propertyValueQueue = append(propertyValueQueue, propertyValueItem{
				of:    propertyOf,
				name:  propertyName,
				value: columnData,
			})
		} else if schemaNode != nil || !ingester.OnlySchemaAttributes {
			// create a new value node only if there is a matching schema node
			newNode, err := ingester.Value(context, targetGraph, newPath, schemaNode, columnData)
			if err != nil {
				return nil, err
			}
			if len(columnName) > 0 {
				newNode.SetProperty(ls.AttributeNameTerm, ls.StringPropertyValue(columnName))
			}
			newNode.SetProperty(ls.AttributeIndexTerm, ls.StringPropertyValue(fmt.Sprint(columnIndex)))
			children = append(children, newNode)
			// keep a reference to the schemaNode
			childrenSchemaNodes[newNode] = schemaNode
		}
	}

	// create a new object node
	retNode, err := ingester.Object(context, targetGraph, path, schemaRoot, children)
	if err != nil {
		return nil, err
	}
	for _, pv := range propertyValueQueue {
		var parentNode graph.Node
		if len(pv.of) > 0 {
			for child, sch := range childrenSchemaNodes {
				if sch != nil && ls.GetNodeID(sch) == pv.of {
					if parentNode != nil {
						return nil, ls.ErrMultipleParentNodes{pv.of}
					}
					parentNode = child
				}
			}
			if parentNode == nil {
				return nil, ls.ErrNoParentNode{pv.of}
			}
		} else {
			parentNode = retNode
		}
		parentNode.SetProperty(pv.name, ls.StringPropertyValue(pv.value))
	}
	ingester.Finish(context, retNode, nil)
	return retNode, nil
}
