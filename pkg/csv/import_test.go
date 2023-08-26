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
	"encoding/csv"
	"os"
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func TestCSVImport(t *testing.T) {
	f, err := os.Open("testdata/GHP_data_capture_global.csv")
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Error(err)
		return
	}

	colSpecs := []TermSpec{
		{
			Term:         "https://lschema.org/validation#format",
			TermTemplate: "{{index .row 5}}",
		},
	}
	layer, err := Import("{{index .row 1}}", colSpecs, 3, 0, nil, "", "", records)
	if err != nil {
		t.Error(err)
		return
	}

	for _, s := range []string{"photo", "givenName", "middleName", "familyName", "birthDate", "linkedVaccineCertificate", "recipient", "disease", "vaccineDescription", "vaccineType", "medicinalProductName", "cvxCode", "marketingAuthorizationHolder", "doseNumber", "dosesPerCycle", "dateOfVaccination", "stateOfVaccination", "countryOfVaccination", "certificateIssuer", "certificateNumber"} {
		node, _ := layer.FindAttributeByID(s)
		if node == nil {
			t.Errorf("Cannot find %s", s)
		}
	}

}

func TestImportSchema(t *testing.T) {
	csv := [][]string{
		{"valueType", "type"},
		{"entityIdFields", "fld1", "fld2"},
		{"@id", "@type", "valueType", "https://dpv_something (,)"},
		{"https://sch", "Schema", "", ""},
		{"https://ovl", "Overlay", "true", "true"},
		{"testObj", "Object"},
		{"field1", "Value", "string", "dpv:value, dpv:otherValue"},
		{"field2", "Value", "xsd:date", ""},
	}
	layers, err := ImportSchema(ls.DefaultContext(), csv, map[string]interface{}{"@context": "../../schemas/ls.json"})
	if err != nil {
		t.Error(err)
	}
	if len(layers) != 2 {
		t.Errorf("Expecting 2 layers")
	}
	if layers[0].GetID() != "https://sch" {
		t.Errorf("Wrong id")
	}
	if layers[0].GetLayerType() != ls.SchemaTerm.Name {
		t.Errorf("Wrong type")
	}
	if layers[1].GetID() != "https://ovl" {
		t.Errorf("Wrong id")
	}
	if layers[1].GetLayerType() != ls.OverlayTerm.Name {
		t.Errorf("Wrong type")
	}
	if layers[0].GetValueType() != "type" {
		t.Errorf("Wrong value type: %v", layers[0].GetValueType())
	}
	if layers[1].GetValueType() != "type" {
		t.Errorf("Wrong value type: %s", layers[1].GetValueType())
	}
	if layers[0].GetEntityIDNodes()[0] != "fld1" {
		t.Errorf("Wrong id")
	}
	if layers[0].GetEntityIDNodes()[1] != "fld2" {
		t.Errorf("Wrong id")
	}
	m := ls.JSONMarshaler{}
	d, _ := m.Marshal(layers[0].Graph)
	t.Log(string(d))
	d, _ = m.Marshal(layers[1].Graph)
	t.Log(string(d))

	attr := layers[0].GetAttributeByID("field1")
	if !attr.HasLabel(ls.AttributeTypeValue.Name) {
		t.Errorf("Not value")
	}
	if s, _ := ls.GetPropertyValueAs[string](attr, "https://dpv_something"); s != "" {
		t.Error("Wrong dpv")
	}
	attr = layers[1].GetAttributeByID("field1")
	if s, _ := ls.GetPropertyValueAs[string](attr, ls.ValueTypeTerm.Name); s != "string" {
		t.Errorf("Wrong type")
	}
	if s, _ := ls.GetPropertyValueAs[[]any](attr, "https://dpv_something"); s[0] != "dpv:value" {
		t.Error("Wrong dpv")
	}
}
