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

package jsonld

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type marshalTestCase struct {
	Name  string          `json:"name"`
	Input any             `json:"input"`
	Graph json.RawMessage `json:"graph"`
}

func (tc marshalTestCase) GetName() string { return tc.Name }

func (tc marshalTestCase) Run(t *testing.T) {
	t.Logf("Running %s", tc.Name)
	layer, err := UnmarshalLayer(tc.Input, nil)
	if err != nil {
		t.Errorf("%s: Cannot unmarshal layer: %v", tc.Name, err)
		return
	}
	m := ls.NewJSONMarshaler(nil)
	g := lpg.NewGraph()
	if err := m.Unmarshal(tc.Graph, g); err != nil {
		t.Error(err)
		return
	}
	if !lpg.CheckIsomorphism(layer.Graph, g, func(n1, n2 *lpg.Node) bool {
		if !n1.GetLabels().IsEqual(n2.GetLabels()) {
			return false
		}
		if !ls.IsPropertiesEqual(ls.PropertiesAsMap(n1), ls.PropertiesAsMap(n2)) {
			return false
		}
		return true
	}, func(e1, e2 *lpg.Edge) bool {
		if e1.GetLabel() != e2.GetLabel() {
			return false
		}
		if !ls.IsPropertiesEqual(ls.PropertiesAsMap(e1), ls.PropertiesAsMap(e2)) {
			return false
		}
		return true
	}) {
		got, _ := m.Marshal(layer.Graph)
		dst := bytes.Buffer{}
		json.Indent(&dst, got, "", "  ")
		t.Errorf("%s: Got:\n%s\n", tc.Name, dst.String())
	}
}

func TestMarshaling(t *testing.T) {
	ls.RunTestsFromFile(t, "testdata/marshalcases.json", func(in json.RawMessage) (ls.TestCase, error) {
		var c marshalTestCase
		err := json.Unmarshal(in, &c)
		return c, err
	})
}

// // marshal_test.go:66: Invalid input: http://hl7.org/fhir/InsurancePlan#contact.* - Cannot follow link in attribute list:
// func TestCannotFollowLinkError(t *testing.T) {
// 	d, err := ioutil.ReadFile("testdata/marshal_error_test1.json")
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	var intf interface{}
// 	if err := json.Unmarshal(d, &intf); err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	_, err = UnmarshalLayer(intf, nil)
// 	if err != nil {
// 		t.Error(err)
// 	}
// }
