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

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/terms"
)

type ImportSpec struct {
	AttributeIDColumn int         `json:"attributeIdColumn"`
	ObjectType        string      `json:"objectType"`
	Layers            []LayerSpec `json:"layers"`
}

type LayerSpec struct {
	Output  string          `json:"output"`
	Columns []ColumnProfile `json:"columns"`
}

// Import a CSV input and slice into a base schema and its
// overlays. The CSV file is organized as columns, one column for base
// attribute names, and other columns for overlays. CSV does not
// support nested attributes. Returns an array of *SchemaBase and
// *SchemaLayer objects
func Import(spec ImportSpec, input [][]string) (layers []*ls.Layer, err error) {
	getColValue := func(row []string, col int) (value string, ok bool) {
		if col >= 0 && col < len(row) {
			return row[col], true
		}
		return "", false
	}
	for _, layer := range spec.Layers {
		newLayer := ls.NewLayer()
		newLayer.TargetType = []string{spec.ObjectType}
		layers = append(layers, newLayer)

		for _, row := range input {
			attribute := ls.NewAttribute(newLayer.Root)
			for _, col := range layer.Columns {
				if col.Index == spec.AttributeIDColumn {
					continue
				}
				colValue, ok := getColValue(row, col.Index)
				if !ok {
					continue
				}
				t := col.Type
				if len(t) == 0 {
					term, ok := ls.Terms[col.Name]
					if ok {
						switch term.GetContainerType() {
						case terms.SetTermType:
							switch term.GetValueType() {
							case terms.ValueTermType:
								t = "@set"
							case terms.IDTermType:
								t = "@idlist"
							case terms.ObjectTermType:
								return nil, fmt.Errorf("Unsupported term: %s", col.Name)
							}

						case terms.ListTermType:
							switch term.GetValueType() {
							case terms.ValueTermType:
								t = "@list"
							case terms.IDTermType:
								t = "@idlist"
							case terms.ObjectTermType:
								return nil, fmt.Errorf("Unsupported term: %s", col.Name)
							}
						case terms.MonadicTermType:
							switch term.GetValueType() {
							case terms.ValueTermType:
								t = "@value"
							case terms.IDTermType:
								t = "@id"
							case terms.ObjectTermType:
								return nil, fmt.Errorf("Unsupported term: %s", col.Name)
							}
						}
					}
				}
				if col.Name == ls.LayerTerms.Attributes.GetTerm() ||
					col.Name == ls.LayerTerms.AllOf.GetTerm() ||
					col.Name == ls.LayerTerms.OneOf.GetTerm() ||
					col.Name == ls.LayerTerms.ArrayItems.GetTerm() {
					return nil, fmt.Errorf("%s cannot be used in CSV", col.Name)
				}
				switch t {
				case "@value":
					attribute.Values[col.Name] = []map[string]interface{}{{"@value": colValue}}
				case "@id":
					if col.Name == ls.LayerTerms.Reference.GetTerm() {
						attribute.Type = &ls.ReferenceType{colValue}
					} else {
						attribute.Values[col.Name] = []map[string]interface{}{{"@id": colValue}}
					}
				case "@list", "@set":
					lst := make([]interface{}, 0)
					for _, x := range strings.Split(colValue, ",") {
						lst = append(lst, map[string]interface{}{"@value": x})
					}
					attribute.Values[col.Name] = []interface{}{map[string]interface{}{t: lst}}
				case "@idlist":
					lst := make([]interface{}, 0)
					for _, x := range strings.Split(colValue, ",") {
						lst = append(lst, map[string]interface{}{"@id": x})
					}
					attribute.Values[col.Name] = []interface{}{map[string]interface{}{"@list": lst}}
				}
			}
			id, ok := getColValue(row, spec.AttributeIDColumn)
			if ok {
				attribute.ID = id
				if err := newLayer.Root.GetAttributes().Add(attribute, newLayer); err != nil {
					return nil, err
				}
			}
		}
	}
	return
}
