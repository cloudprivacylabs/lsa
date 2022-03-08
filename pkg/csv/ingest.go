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

// Ingest a row of CSV data. The `data` slice is a row of data. The ID
// is the assigned identifier for the resulting object. ID is empty,
// IDs are assigned based on schema-dictated identifiers.
func (ingester Ingester) Ingest(context *ls.Context, data []string, ID string) (graph.Node, error) {
	ictx := ingester.Start(context, ID)
	// Retrieve map of schema attribute nodes from schemaRoot
	attributes, err := ls.GetObjectAttributeNodes(ictx.GetSchemaNode())
	if err != nil {
		return nil, err
	}
	// create a new object node
	_, retNode, err := ingester.Object(ictx)
	if err != nil {
		return nil, err
	}
	ctx := ictx.NewLevel(retNode)
	// Iterate through each column of the CSV row
	for columnIndex, columnData := range data {
		var columnName string
		if columnIndex < len(ingester.ColumnNames) {
			columnName = ingester.ColumnNames[columnIndex]
		}
		var schemaNode graph.Node
		// if column header exists, assign schemaNode to corresponding value in attributes map
		var node graph.Node
		if len(columnName) > 0 {
			schemaNodes := attributes[columnName]
			if len(schemaNodes) > 1 {
				return nil, ls.ErrInvalidSchema(fmt.Sprintf("Multiple elements with key '%s'", columnName))
			}
			if len(schemaNodes) == 1 {
				schemaNode = schemaNodes[0]
			}

			_, node, err = ingester.Value(ctx.New(columnName, schemaNode), columnData)
			if err != nil {
				return nil, err
			}
		} else if ingester.Schema != nil || !ingester.OnlySchemaAttributes {
			schemaNode, _ = ingester.Schema.FindFirstAttribute(func(n graph.Node) bool {
				p := ls.AsPropertyValue(n.GetProperty(ls.AttributeIndexTerm))
				if p == nil {
					return false
				}
				return p.IsInt() && p.AsInt() == columnIndex
			})
			if schemaNode != nil {
				_, node, err = ingester.Value(ctx.New(columnIndex, schemaNode), columnData)
				if err != nil {
					return nil, err
				}
			}
		}
		if node != nil {
			if len(columnName) > 0 {
				node.SetProperty(ls.AttributeNameTerm, ls.StringPropertyValue(columnName))
			}
			node.SetProperty(ls.AttributeIndexTerm, ls.StringPropertyValue(fmt.Sprint(columnIndex)))
		}
	}

	ingester.Finish(ictx, retNode)
	return retNode, nil
}
