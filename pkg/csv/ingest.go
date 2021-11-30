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
)

const CSV = ls.LS + "csv/"

type Ingester struct {
	ls.Ingester
	ColumnNames []string
}

func (ingester Ingester) Ingest(data []string, ID string) (ls.Node, error) {
	path, schemaRoot := ingester.Start(ID)
	attributes, err := ingester.GetObjectAttributeNodes(schemaRoot)
	if err != nil {
		return nil, err
	}
	children := make([]ls.Node, 0, len(data))
	for columnIndex, columnData := range data {
		var columnName string
		if columnIndex < len(ingester.ColumnNames) {
			columnName = ingester.ColumnNames[columnIndex]
		}
		var schemaNode ls.Node
		var newPath []interface{}
		if len(columnName) > 0 {
			schemaNode = attributes[columnName]
			newPath = append(path, columnName)
		} else if ingester.Schema != nil {
			schemaNode, _ = ingester.Schema.FindFirstAttribute(func(n ls.Node) bool {
				p := n.GetProperties()[ls.AttributeIndexTerm]
				if p == nil {
					return false
				}
				return p.IsInt() && p.AsInt() == columnIndex
			})
			newPath = append(path, columnIndex)
		}

		newNode, err := ingester.Value(newPath, schemaNode, columnData)
		if err != nil {
			return nil, err
		}
		if len(columnName) > 0 {
			newNode.GetProperties()[ls.AttributeNameTerm] = ls.StringPropertyValue(columnName)
		}
		newNode.GetProperties()[ls.AttributeIndexTerm] = ls.StringPropertyValue(fmt.Sprint(columnIndex))
		children = append(children, newNode)
	}
	return ingester.Object(path, schemaRoot, children)
}
