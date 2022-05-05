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
	"github.com/cloudprivacylabs/lsa/pkg/validators"
	"github.com/cloudprivacylabs/opencypher/graph"
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

			var attr graph.Node
			attr = layer.Graph.NewNode([]string{ls.AttributeNodeTerm, ls.AttributeTypeValue}, nil)
			ls.SetNodeID(attr, id)
			layer.Graph.NewEdge(root, attr, ls.ObjectAttributeListTerm, nil)
			ls.SetNodeIndex(attr, index)
			if requiredTemplate != nil {
				s, err := runtmp(requiredTemplate, "")
				if err != nil {
					return nil, err
				}
				if s == "true" {
					attr.SetProperty(validators.RequiredTerm, ls.StringPropertyValue("true"))
				}
			}

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
	return layer, nil
}

// ImportSchema imports a schema from a CSV file. The CSV file is organized as follows:
//
//   @id,   @type,    entityId,   <term>,     <term>
//  layerId, Schema,attrId ,...
//  layerId, Overlay,        , true,        true   --> true means include this attribute in overlay
//  attrId, Object,          , termValue, termValue
//  attrId, Value, true/false, termValue, termValue
//   ...
//
// The terms are expanded using the JSON-LD context given.
func ImportSchema(ctx *ls.Context, rows [][]string, context map[string]interface{}) ([]*ls.Layer, error) {
	// Locate the header row
	headerRowIndex := -1
	for index, row := range rows {
		if len(row) > 1 {
			if row[0] == "@id" && row[1] == "@type" {
				headerRowIndex = index
				break
			}
		}
	}
	if headerRowIndex == -1 {
		return nil, fmt.Errorf("Cannot locate the header row. The header row must have @id and @type in the first two columns.")
	}

	header := make([]string, 0, len(rows[headerRowIndex]))
	for _, x := range rows[headerRowIndex] {
		header = append(header, strings.TrimSpace(x))
	}

	layerInfo := make([]map[string][]string, 0, len(rows))
	attrRows := make([]map[string][]string, 0, len(rows))

	rowToMap := func(row []string, rowIndex int) (map[string][]string, error) {
		ret := make(map[string][]string, len(row))
		for i := range row {
			value := strings.TrimSpace(row[i])
			if len(value) == 0 {
				continue
			}
			if len(header) <= i {
				return nil, fmt.Errorf("At row %d: More columns than headers", rowIndex)
			}
			ret[header[i]] = append(ret[header[i]], value)
		}
		if len(ret["@id"]) == 0 || len(ret["@type"]) == 0 {
			return nil, nil
		}
		return ret, nil
	}
	// Parse spreadsheet
	for index := headerRowIndex + 1; index < len(rows); index++ {
		row := rows[index]
		if len(row) < 2 {
			continue
		}
		mrow, err := rowToMap(row, index)
		if err != nil {
			return nil, err
		}
		if mrow == nil {
			continue
		}
		typ := mrow["@type"]
		if len(typ) != 1 {
			continue
		}
		if len(mrow["@id"]) != 1 {
			continue
		}
		// This must be an overlay or schema
		if typ[0] == "Schema" || typ[0] == ls.SchemaTerm || typ[0] == "Overlay" || typ[0] == ls.OverlayTerm {
			layerInfo = append(layerInfo, mrow)
		} else {
			// Attr row
			if len(layerInfo) == 0 {
				return nil, fmt.Errorf("The schema must have at least one layer definition row")
			}
			attrRows = append(attrRows, mrow)
		}
	}

	ret := make([]*ls.Layer, 0, len(layerInfo))
	copyMap := func(target map[string]interface{}, source, filter map[string][]string) {
		for k, v := range source {
			if !strings.HasPrefix(k, "@") && filter != nil {
				// Check source if this is to be set
				bvalue := filter[k]
				if len(bvalue) != 1 || bvalue[0] != "true" {
					continue
				}
			}
			if len(v) == 1 {
				target[k] = v[0]
			} else if len(v) > 1 {
				target[k] = v
			}
		}
	}
	// Build layers
	for _, layer := range layerInfo {
		// Create a compact jsonld
		layerMap := make(map[string]interface{})
		if context != nil {
			for k, v := range context {
				layerMap[k] = v
			}
		}
		copyMap(layerMap, layer, map[string][]string{"entityIdFields": []string{"true"}, ls.EntityIDFieldsTerm: []string{"true"}})

		layerNode := make(map[string]interface{})
		layerMap[ls.LayerRootTerm] = layerNode
		objectNode := []interface{}{}
		// Attributes
		for attrIndex, attrRow := range attrRows {
			if attrIndex == 0 {
				// Must be object
				if attrRow["@type"][0] != "Object" && attrRow["@type"][0] != ls.AttributeTypeObject {
					return nil, fmt.Errorf("First attribute row must define an object")
				}
				copyMap(layerNode, attrRow, layer)
			} else {
				attrNode := make(map[string]interface{})
				copyMap(attrNode, attrRow, layer)
				objectNode = append(objectNode, attrNode)
			}
		}
		layerNode[ls.ObjectAttributeListTerm] = objectNode

		l, err := ls.UnmarshalLayer(layerMap, ctx.GetInterner())
		if err != nil {
			return nil, fmt.Errorf("Cannot create layer %s: %w", layer["@id"][0], err)
		}
		ret = append(ret, l)
	}
	return ret, nil
}
