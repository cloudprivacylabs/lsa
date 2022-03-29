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

type compileTestCase struct {
	Name     string          `json:"name"`
	Schemas  []interface{}   `json:"schemas"`
	Expected json.RawMessage `json:"expected"`
}

func (tc compileTestCase) GetName() string { return tc.Name }

func (tc compileTestCase) Run(t *testing.T) {
	t.Logf("Running %s", tc.Name)
	schemas := make([]*Layer, 0)
	for _, sch := range tc.Schemas {
		l, err := UnmarshalLayer(sch, nil)
		if err != nil {
			t.Errorf("%s: Cannot unmarshal layer: %v", tc.Name, err)
			return
		}
		schemas = append(schemas, l)
	}
	compiler := Compiler{
		Loader: SchemaLoaderFunc(func(x string) (*Layer, error) {
			for _, s := range schemas {
				if s.GetID() == x {
					return s, nil
				}
			}
			return nil, nil
		}),
	}
	result, err := compiler.Compile(DefaultContext(), schemas[0].GetID())
	if err != nil {
		t.Errorf("Cannot compile %v", err)
		return
	}
	m := JSONMarshaler{}
	data, _ := m.Marshal(result.Graph)
	t.Logf(string(data))
}

func TestCompile(t *testing.T) {
	RunTestsFromFile(t, "testdata/compilecases.json", func(in json.RawMessage) (TestCase, error) {
		var c compileTestCase
		err := json.Unmarshal(in, &c)
		return c, err
	})
}
