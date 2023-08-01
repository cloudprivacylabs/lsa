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
	"text/template"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type CSVImportSpec struct {
	AttributeID  string     `json:"attributeId" yaml:"attributeId"`
	LayerType    string     `json:"layerType" yaml:"layerType"`
	LayerID      string     `json:"layerId" yaml:"layerId"`
	RootID       string     `json:"rootId" yaml:"rootId"`
	ValueType    string     `json:"valueType" yaml:"valueType"`
	EntityIDRows []int      `json:"entityIdRows" yaml:"entityIdRows"`
	EntityID     string     `json:"entityId" yaml:"entityId"`
	Required     string     `json:"required" yaml:"required"`
	StartRow     int        `json:"startRow" yaml:"startRow"`
	NRows        int        `json:"nrows" yaml:"nrows"`
	Terms        []TermSpec `json:"terms" yaml:"terms"`
}

func (spec CSVImportSpec) Import(records [][]string) (*ls.Layer, error) {
	exec := func(tmpl string, data interface{}) (string, error) {
		t, err := template.New("").Parse(tmpl)
		if err != nil {
			return "", err
		}
		var buf bytes.Buffer
		if err := t.Execute(&buf, data); err != nil {
			return "", err
		}
		return buf.String(), nil
	}
	rows := records[spec.StartRow:]
	if spec.NRows > 0 && spec.NRows > len(rows) {
		rows = rows[:spec.NRows]
	}
	layerType, err := exec(spec.LayerType, map[string]interface{}{"rows": rows})
	if err != nil {
		return nil, err
	}
	if layerType == "Overlay" {
		layerType = ls.OverlayTerm.Name
	} else if layerType == "Schema" {
		layerType = ls.SchemaTerm.Name
	}
	layer, err := Import(spec.AttributeID, spec.Terms, spec.StartRow, spec.NRows, spec.EntityIDRows, spec.EntityID, spec.Required, records)
	if err != nil {
		return nil, err
	}
	if len(layerType) > 0 {
		layer.SetLayerType(layerType)
	}
	layerID, err := exec(spec.LayerID, map[string]interface{}{"rows": rows})
	if err != nil {
		return nil, err
	}
	if len(layerID) > 0 {
		layer.SetID(layerID)
	}
	rootID, err := exec(spec.RootID, map[string]interface{}{"rows": rows})
	if err != nil {
		return nil, err
	}
	if len(rootID) > 0 {
		ls.SetNodeID(layer.GetSchemaRootNode(), rootID)
	}
	valueType, err := exec(spec.ValueType, map[string]interface{}{"rows": rows})
	if err != nil {
		return nil, err
	}
	if len(valueType) > 0 {
		layer.SetValueType(valueType)
	}
	return layer, nil
}
