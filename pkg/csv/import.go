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

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
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
	// with {{.term}}, {{.data}}, and {{.row}} in template context. {{.term}} gives
	// the Term, and {{.data}} gives the value of the term in the
	// current cell.
	TermTemplate string `json:"template"`
	// Is property an array
	Array bool `json:"array"`
	// Array separator character
	ArraySeparator string `json:"separator"`
}

type AttributeSpec struct {
	// The 0-based column containing term data
	TermCol int `json:"column"`
	// If nonempty, this template is used to build the term contents
	// with {{.term}}, {{.data}}, and {{.row}} in template context. {{.term}} gives
	// the Term, and {{.data}} gives the value of the term in the
	// current cell.
	TermTemplate string `json:"template"`

	// If evaluates to a nonempty string, the attribute is an array whose elements are of this type
	ArrayTypeTemplate string `json:"arrayTypeTemplate"`
	ArrayIDTemplate   string `json:"arrayIdTemplate"`
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
func Import(attributeID AttributeSpec, terms []TermSpec, startRow, nRows int, input [][]string) (*ls.Layer, error) {
	layer := ls.NewLayer()
	root := layer.NewNode([]string{ls.AttributeTypeObject, ls.AttributeNodeTerm}, nil)
	layer.NewEdge(layer.GetLayerRootNode(), root, ls.LayerRootTerm, nil)

	var idTemplate, arrayIdTemplate, arrayTypeTemplate *template.Template

	if len(attributeID.TermTemplate) > 0 {
		var err error
		idTemplate, err = template.New("").Parse(attributeID.TermTemplate)
		if err != nil {
			return nil, err
		}
	}
	if len(attributeID.ArrayTypeTemplate) > 0 {
		var err error
		arrayTypeTemplate, err = template.New("").Parse(attributeID.ArrayTypeTemplate)
		if err != nil {
			return nil, err
		}
	}
	if len(attributeID.ArrayIDTemplate) > 0 {
		var err error
		arrayIdTemplate, err = template.New("").Parse(attributeID.ArrayIDTemplate)
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
	index := 0
	for rowIndex, row := range input {
		runtmp := func(t *template.Template, term, data string) (string, error) {
			if t == nil {
				return data, nil
			}
			var out bytes.Buffer
			if err := t.Execute(&out, map[string]interface{}{"term": term, "data": data, "row": row}); err != nil {
				return "", err
			}
			return strings.TrimSpace(out.String()), nil
		}
		if rowIndex >= startRow && (nRows == 0 || nAttributeID < nRows) {
			var err error
			nAttributeID++
			if attributeID.TermCol < 0 || attributeID.TermCol >= len(row) {
				return nil, ErrColIndexOutOfBounds{For: "@id", Index: attributeID.TermCol}
			}
			id := strings.TrimSpace(row[attributeID.TermCol])
			if len(id) == 0 {
				break
			}
			id, err = runtmp(idTemplate, "@id", id)
			if err != nil {
				return nil, err
			}
			if len(id) == 0 {
				return nil, ErrInvalidID{rowIndex}
			}
			var attr graph.Node
			if arrayIdTemplate != nil {
				attr = layer.NewNode([]string{ls.AttributeNodeTerm, ls.AttributeTypeArray}, nil)
				ls.SetNodeID(attr, id)
				arrId, err := runtmp(arrayIdTemplate, "@id", id)
				if err != nil {
					return nil, err
				}
				if len(arrId) == 0 {
					attr = layer.NewNode([]string{ls.AttributeNodeTerm, ls.AttributeTypeValue}, nil)
					ls.SetNodeID(attr, id)
				} else {
					typ, err := runtmp(arrayTypeTemplate, "@id", id)
					if err != nil {
						return nil, err
					}
					if len(typ) == 0 {
						typ = ls.AttributeTypeValue
					}
					elems := layer.NewNode([]string{ls.AttributeNodeTerm, typ}, nil)
					ls.SetNodeID(elems, arrId)
					layer.NewEdge(attr, elems, ls.ArrayItemsTerm, nil)
				}
			} else {
				attr = layer.NewNode([]string{ls.AttributeNodeTerm, ls.AttributeTypeValue}, nil)
				ls.SetNodeID(attr, id)
			}
			layer.NewEdge(root, attr, ls.ObjectAttributeListTerm, nil)
			ls.SetNodeIndex(attr, index)
			index++
			for ti, term := range terms {
				nTerms[ti]++
				if term.TermCol < 0 || term.TermCol >= len(row) {
					return nil, ErrColIndexOutOfBounds{For: term.Term, Index: term.TermCol}
				}
				data := strings.TrimSpace(row[term.TermCol])
				data, err = runtmp(templates[ti], term.Term, data)
				if err != nil {
					return nil, err
				}
				if len(data) > 0 {
					if term.Array {
						elems := strings.Split(data, term.ArraySeparator)
						attr.SetProperty(term.Term, ls.StringSlicePropertyValue(elems))
					} else {
						attr.SetProperty(term.Term, ls.StringPropertyValue(data))
					}
				}
			}
		}
	}
	return layer, nil
}
