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
package json

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type composeSchemaTest struct {
	Name     string            `json:"name"`
	Schemas  []json.RawMessage `json:"schemas"`
	Expected json.RawMessage   `json:"expected"`
}

func TestComposeSchema(t *testing.T) {
	data, err := os.ReadFile("testdata/composeSchemaTests.json")
	if err != nil {
		panic(err)
	}
	var testCases []composeSchemaTest
	if err := json.Unmarshal(data, &testCases); err != nil {
		panic(err)
	}
	for _, testCase := range testCases {
		t.Logf("Running %s", testCase.Name)

		names := make([]string, 0)
		for i := 0; i < len(testCase.Schemas); i++ {
			names = append(names, strconv.Itoa(i))
		}

		node, err := ComposeSchema(ls.DefaultContext(), names[0], names[1:], func(_ *ls.Context, name string) (io.ReadCloser, error) {
			n, _ := strconv.Atoi(name)
			return io.NopCloser(bytes.NewReader(testCase.Schemas[n])), nil
		})
		if err != nil {
			t.Errorf("In %s: %s", testCase.Name, err)
			continue
		}
		var expected map[string]interface{}
		if err := json.Unmarshal(testCase.Expected, &expected); err != nil {
			panic(err)
		}
		if !reflect.DeepEqual(node.Marshal(), expected) {
			t.Errorf("Not equal, got: %v, expected: %v", node.Marshal(), expected)
		}
	}
}
