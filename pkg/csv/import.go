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
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/bserdar/digraph"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type ImportSpec struct {
	AttributeID TermSpec

	Terms []TermSpec
}

type TermSpec struct {
	// The term
	Term string `json:"term"`
	// The 0-based column containing term data
	TermCol int `json:"column"`
	// If nonempty, this template is used to build the term contents
	// with {{.term}} and {{.data}} in template context. {{.term}} gives
	// the Term, and {{.data}} gives the value of the term in the
	// current cell.
	TermTemplate string `json:"template"`
	// Is property an array
	Array bool `json:"array"`
	// Array separator character
	ArraySeparator string `json:"separator"`
}

type ErrColIndexOutOfBounds struct {
	For   string
	Index int
}

func (e ErrColIndexOutOfBounds) Error() string {
	s := fmt.Sprintf("Column index out of bounds: %d", e.Index)
	if e.For != "" {
		s += " " + e.For
	}
	return s
}

type ErrInvalidID struct {
	Row int
}

func (e ErrInvalidID) Error() string {
	return fmt.Sprintf("Invalid ID at row  %d", e.Row)
}

// Import a CSV schema. The CSV file is organized as columns, one
// column for base attribute names, and other columns for
// overlays. CSV does not support nested attributes. Returns an array
// of Layer objects
func Import(attributeID TermSpec, terms []TermSpec, startRow, nRows int, input [][]string) (*ls.Layer, error) {
	layer := ls.NewLayer()
	root := ls.NewNode("")
	layer.AddNode(root)
	digraph.Connect(layer.GetLayerInfoNode(), root, ls.NewEdge(ls.LayerRootTerm))

	var idTemplate *template.Template

	if len(attributeID.TermTemplate) > 0 {
		var err error
		idTemplate, err = template.New("").Parse(attributeID.TermTemplate)
		if err != nil {
			return nil, err
		}
	}
	templates := make([]*template.Template, len(terms))
	for i, t := range terms {
		if len(t.TermTemplate) > 0 {
			tmp, err := template.New("").Parse(t.TermTemplate)
			if err != nil {
				return nil, err
			}
			templates[i] = tmp
		}
	}

	nAttributeID := 0
	nTerms := make([]int, len(terms))
	for rowIndex, row := range input {
		fmt.Printf("Row: %d\n", rowIndex)
		if rowIndex >= startRow && (nRows == 0 || nAttributeID < nRows) {
			nAttributeID++
			if attributeID.TermCol < 0 || attributeID.TermCol >= len(row) {
				return nil, ErrColIndexOutOfBounds{For: "@id", Index: attributeID.TermCol}
			}
			id := strings.TrimSpace(row[attributeID.TermCol])
			fmt.Printf("id: %s\n", id)
			if len(id) == 0 {
				break
			}
			// If there is a template, run it
			if idTemplate != nil {
				var out bytes.Buffer
				if err := idTemplate.Execute(&out, map[string]interface{}{"term": "@id", "data": id}); err != nil {
					return nil, err
				}
				id = strings.TrimSpace(out.String())
			}
			if len(id) == 0 {
				return nil, ErrInvalidID{rowIndex}
			}
			attr := ls.NewNode(id, ls.AttributeTypes.Attribute, ls.AttributeTypes.Value)
			layer.AddNode(attr)
			digraph.Connect(root, attr, ls.NewEdge(ls.LayerTerms.AttributeList))
			for ti, term := range terms {
				nTerms[ti]++
				if term.TermCol < 0 || term.TermCol >= len(row) {
					return nil, ErrColIndexOutOfBounds{For: term.Term, Index: term.TermCol}
				}
				data := strings.TrimSpace(row[term.TermCol])
				if templates[ti] != nil {
					var out bytes.Buffer
					if err := templates[ti].Execute(&out, map[string]interface{}{"term": term.Term, "data": data}); err != nil {
						return nil, err
					}
				}
				if len(data) > 0 {
					if term.Array {
						elems := strings.Split(data, term.ArraySeparator)
						attr.GetProperties()[term.Term] = ls.StringSlicePropertyValue(elems)
					} else {
						attr.GetProperties()[term.Term] = ls.StringPropertyValue(data)
					}
				}
			}
		}
	}
	return layer, nil
}
