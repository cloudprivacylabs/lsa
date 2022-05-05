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
		{"@id", "@type", "valueType", "entityIdFields", "https://dpv_something"},
		{"https://sch", "Schema", "", "field1", ""},
		{"https://ovl", "Overlay", "true", "", "true"},
		{"testObj", "Object"},
		{"field1", "Value", "string", "", "dpv:value"},
		{"field2", "Value", "xsd:date", "", ""},
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
	if layers[0].GetLayerType() != ls.SchemaTerm {
		t.Errorf("Wrong type")
	}
	if layers[1].GetID() != "https://ovl" {
		t.Errorf("Wrong id")
	}
	if layers[1].GetLayerType() != ls.OverlayTerm {
		t.Errorf("Wrong type")
	}
	m := ls.JSONMarshaler{}
	d, _ := m.Marshal(layers[0].Graph)
	t.Log(string(d))
	d, _ = m.Marshal(layers[1].Graph)
	t.Log(string(d))

	attr := layers[0].GetAttributeByID("field1")
	if !attr.HasLabel(ls.AttributeTypeValue) {
		t.Errorf("Not value")
	}
	if ls.AsPropertyValue(attr.GetProperty("https://dpv_something")).AsString() != "" {
		t.Error("Wrong dpv")
	}
	attr = layers[1].GetAttributeByID("field1")
	if ls.AsPropertyValue(attr.GetProperty(ls.ValueTypeTerm)).AsString() != "string" {
		t.Errorf("Wrong type")
	}
	if ls.AsPropertyValue(attr.GetProperty("https://dpv_something")).AsString() != "dpv:value" {
		t.Error("Wrong dpv")
	}
}
