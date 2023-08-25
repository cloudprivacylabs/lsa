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

	"github.com/cloudprivacylabs/lpg/v2"
)

type sliceTestCase struct {
	Name     string   `json:"name"`
	Schema   any      `json:"schema"`
	Terms    []string `json:"terms"`
	Expected any      `json:"expected"`
}

func (tc sliceTestCase) GetName() string { return tc.Name }

func (tc sliceTestCase) Run(t *testing.T) {
	t.Logf("Running %s", tc.Name)
	sch, err := UnmarshalLayerFromTree(tc.Schema)
	if err != nil {
		t.Errorf("%s: Cannot unmarshal layer: %v", tc.Name, err)
		return
	}
	expected, err := GraphFromTree(tc.Expected)
	if err != nil {
		t.Errorf("Cannot parse expected: %s", err)
		return
	}
	slice := sch.Slice(OverlayTerm.Name, GetSliceByTermsFunc(tc.Terms, false))
	if !lpg.CheckIsomorphism(expected, slice.Graph, DefaultNodeEquivalenceFunc, DefaultEdgeEquivalenceFunc) {
		t.Errorf("%s: Expected:\n%s\nGot:\n%s\n", tc.Name, MarshalIndentGraph(expected), MarshalIndentGraph(slice.Graph))
	}
}

func TestSlice(t *testing.T) {
	RunTestsFromFile(t, "testdata/slicecases.json", func(in json.RawMessage) (TestCase, error) {
		var c sliceTestCase
		err := json.Unmarshal(in, &c)
		return c, err
	})
}
