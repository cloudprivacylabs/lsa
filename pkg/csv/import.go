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
	"github.com/cloudprivacylabs/opencypher/graph"
	"github.com/cloudprivacylabs/lsa/pkg/validators"
)

type TermSpec struct {
	// The term
	Term string `json:"term"`
	// If nonempty, this template is used to build the term contents
	// with {{.term}}, and {{.row}} in template context. {{.term}} gives
	// the Term, and {{.row}} gives the cells of the current row
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
func Import(attributeID string, terms []TermSpec, startRow, nRows int, idRows []int, entityID string, required string, input [][]string) (*ls.Layer, error) {
	layer := ls.NewLayer()
	root := layer.Graph.NewNode([]string{ls.AttributeTypeObject, ls.AttributeNodeTerm}, nil)
	layer.Graph.NewEdge(layer.GetLayerRootNode(), root, ls.LayerRootTerm, nil)

	idTemplate, err := template.New("").Parse(attributeID)
	if err != nil {
		return nil, err
	}

	var entityIDTemplate *template.Template
	if len(entityID) > 0 {
		entityIDTemplate, err = template.New("").Parse(entityID)
		if err != nil {
			return nil, err
		}
	}

	var requiredTemplate *template.Template
	if len(required) > 0 {
		requiredTemplate, err = template.New("").Parse(required)
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
	entityIDFields := make([]string, 0)
	requiredFields := make([]string, 0)
	for rowIndex, row := range input {
		runtmp := func(t *template.Template, term string) (string, error) {
			if t == nil {
				return "", nil
			}
			var out bytes.Buffer
			if err := t.Execute(&out, map[string]interface{}{"term": term, "row": row}); err != nil {
				return "", err
			}
			return strings.TrimSpace(out.String()), nil
		}
		if rowIndex >= startRow && (nRows == 0 || nAttributeID < nRows) {
			nAttributeID++
			id, err := runtmp(idTemplate, "@id")
			if err != nil {
				return nil, err
			}
			if len(id) == 0 {
				return nil, ErrInvalidID{rowIndex}
			}
			for i, x := range idRows {
				if rowIndex == x {
					for len(entityIDFields) <= i {
						entityIDFields = append(entityIDFields, "")
					}
					entityIDFields[i] = id
					break
				}
			}

			if entityIDTemplate != nil {
				s, err := runtmp(entityIDTemplate, "")
				if err != nil {
					return nil, err
				}
				if s == "true" {
					entityIDFields = append(entityIDFields, id)
				}
			}

			if requiredTemplate != nil {
				s, err := runtmp(requiredTemplate, "")
				if err != nil {
					return nil, err
				}
				if s == "true" {
					requiredFields = append(requiredFields, id)
				}
			}

			var attr graph.Node
			attr = layer.Graph.NewNode([]string{ls.AttributeNodeTerm, ls.AttributeTypeValue}, nil)
			ls.SetNodeID(attr, id)
			layer.Graph.NewEdge(root, attr, ls.ObjectAttributeListTerm, nil)
			ls.SetNodeIndex(attr, index)
			index++
			for ti, term := range terms {
				nTerms[ti]++
				data, err := runtmp(templates[ti], term.Term)
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
	if len(entityIDFields) > 0 {
		var v *ls.PropertyValue
		if len(entityIDFields) == 1 {
			v = ls.StringPropertyValue(entityIDFields[0])
		} else {
			v = ls.StringSlicePropertyValue(entityIDFields)
		}
		root.SetProperty(ls.EntityIDFieldsTerm, v)
	}
	if len(requiredFields) > 0 {
		root.SetProperty(validators.RequiredTerm, ls.StringSlicePropertyValue(requiredFields))
	}
	return layer, nil
}
