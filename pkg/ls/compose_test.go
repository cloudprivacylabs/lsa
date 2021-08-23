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
package ls

import (
	"encoding/json"
	"testing"
)

type composeTestCase struct {
	Name     string        `json:"name"`
	Base     interface{}   `json:"base"`
	Overlays []interface{} `json:"overlays"`
	Expected interface{}   `json:"expected"`
}

func (tc composeTestCase) GetName() string { return tc.Name }

func (tc composeTestCase) Run(t *testing.T) {
	t.Logf("Running %s", tc.Name)
	base, err := UnmarshalLayer(tc.Base)
	if err != nil {
		t.Errorf("%s: Cannot unmarshal layer: %v", tc.Name, err)
		return
	}

	for i, o := range tc.Overlays {
		ovl, err := UnmarshalLayer(o)
		if err != nil {
			t.Errorf("%s: Cannot unmarshal overlay %d: %v", tc.Name, i, err)
			return
		}
		err = base.Compose(ovl)
		if err != nil {
			t.Errorf("%s: Compose error: %v", tc.Name, err)
			return
		}
	}

	marshaled := MarshalLayer(base)
	if err := DeepEqual(ToMap(marshaled), ToMap(tc.Expected)); err != nil {
		expected, _ := json.MarshalIndent(ToMap(tc.Expected), "", "  ")
		got, _ := json.MarshalIndent(ToMap(marshaled), "", "  ")
		t.Errorf("%v %s: Expected:\n%s\nGot:\n%s\n", err, tc.Name, string(expected), string(got))
	}
}

func TestCompose(t *testing.T) {
	RunTestsFromFile(t, "testdata/composecases.json", func(in json.RawMessage) (testCase, error) {
		var c composeTestCase
		err := json.Unmarshal(in, &c)
		return c, err
	})
}

// func TestOverride(t *testing.T) {
// 	base := expand(t, `{
// "@context": "../../schemas/ls.jsonld",
// "@type":"Schema",
// "attributes": [
// 	{
//  	"@id":  "attr1" ,
//   "@type": "Value",
// 	"type":"string"
//  }
//  ]
// }`)
// 	ovl := expand(t, `{
// "@context": "../../schemas/ls.jsonld",
// "@type":"Overlay",
// "attributes": [
// 	{
//  	"@id":  "attr1" ,
//   "@type": "Value",
// 	"type":"int"
//  }
//  ]
// }`)
// 	baseattr := getLayer(base)
// 	ovlattr := getLayer(ovl)
// 	err := baseattr.Compose(ComposeOptions{}, Terms, ovlattr)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	base = baseattr.MarshalExpanded().([]interface{})
// 	t.Logf("%+v", base[0])
// 	//attrBase := base[0].(map[string]interface{})[LayerTerms.Attributes.GetTerm()]
// 	item := baseattr.Index["attr1"]
// 	if GetNodeValue(item.Values[AttributeAnnotations.Type.GetTerm()].([]interface{})[0]) != "int" {
// 		t.Errorf("Expecting int")
// 	}
// 	if len(item.Values[AttributeAnnotations.Type.GetTerm()].([]interface{})) != 1 {
// 		t.Errorf("Expecting 1 elements")
// 	}
// }
