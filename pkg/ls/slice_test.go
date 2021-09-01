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

type sliceTestCase struct {
	Name     string      `json:"name"`
	Schema   interface{} `json:"schema"`
	Terms    []string    `json:"terms"`
	Expected interface{} `json:"expected"`
}

func (tc sliceTestCase) GetName() string { return tc.Name }

func (tc sliceTestCase) Run(t *testing.T) {
	t.Logf("Running %s", tc.Name)
	sch, err := UnmarshalLayer(tc.Schema)
	if err != nil {
		t.Errorf("%s: Cannot unmarshal layer: %v", tc.Name, err)
		return
	}
	slice := sch.Slice(OverlayTerm, GetSliceByTermsFunc(tc.Terms, false))
	marshaled := MarshalLayer(slice)
	if err := DeepEqual(ToMap(marshaled), ToMap(tc.Expected)); err != nil {
		expected, _ := json.MarshalIndent(ToMap(tc.Expected), "", "  ")
		got, _ := json.MarshalIndent(ToMap(marshaled), "", "  ")
		t.Errorf("%v %s: Expected:\n%s\nGot:\n%s\n", err, tc.Name, string(expected), string(got))
	}
}

func TestSlice(t *testing.T) {
	RunTestsFromFile(t, "testdata/slicecases.json", func(in json.RawMessage) (TestCase, error) {
		var c sliceTestCase
		err := json.Unmarshal(in, &c)
		return c, err
	})
}
