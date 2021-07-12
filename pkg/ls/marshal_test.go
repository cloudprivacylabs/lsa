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
	"io/ioutil"
	"testing"
)

type marshalTestCase struct {
	Name      string      `json:"name"`
	Input     interface{} `json:"input"`
	Marshaled interface{} `json:"marshaled"`
}

func (tc marshalTestCase) getName() string { return tc.Name }

func (tc marshalTestCase) run(t *testing.T) {
	t.Logf("Running %s", tc.Name)
	layer, err := UnmarshalLayer(tc.Input)
	if err != nil {
		t.Errorf("%s: Cannot unmarshal layer: %v", tc.Name, err)
		return
	}
	marshaled := MarshalLayer(layer)
	if err := deepEqual(toMap(marshaled), toMap(tc.Marshaled)); err != nil {
		expected, _ := json.MarshalIndent(toMap(tc.Marshaled), "", "  ")
		got, _ := json.MarshalIndent(toMap(marshaled), "", "  ")
		t.Errorf("%v %s: Expected:\n%s\nGot:\n%s\n", err, tc.Name, string(expected), string(got))
	}
}

func TestMarshaling(t *testing.T) {
	runTestsFromFile(t, "testdata/marshalcases.json", func(in json.RawMessage) (testCase, error) {
		var c marshalTestCase
		err := json.Unmarshal(in, &c)
		return c, err
	})
}

//  marshal_test.go:66: Invalid input: http://hl7.org/fhir/InsurancePlan#contact.* - Cannot follow link in attribute list:
func TestCannotFollowLinkError(t *testing.T) {
	d, err := ioutil.ReadFile("testdata/marshal_error_test1.json")
	if err != nil {
		t.Error(err)
		return
	}
	var intf interface{}
	if err := json.Unmarshal(d, &intf); err != nil {
		t.Error(err)
		return
	}
	_, err = UnmarshalLayer(intf)
	if err != nil {
		t.Error(err)
	}
}
