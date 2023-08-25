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
	"fmt"
	"reflect"
	"testing"

	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type marshalTestCase struct {
	Name      string          `json:"name"`
	Input     any             `json:"input"`
	Graph     json.RawMessage `json:"graph"`
	Compacted any             `json:"compacted"`
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

	if tc.Compacted != nil {
		marshaled, err := MarshalLayer(layer)
		if err != nil {
			t.Error(err)
			return
		}
		if err := deepCompare(tc.Compacted, marshaled); err != nil {
			t.Errorf("%v: Expected: %v\nGot: %v", err, tc.Compacted, marshaled)
		}

	}
}

func deepCompare(v1, v2 any) error {
	if v1 == nil && v2 == nil {
		return nil
	}
	arr1, ok := v1.([]any)
	if ok {
		arr2, ok := v2.([]any)
		if !ok {
			return fmt.Errorf("Expecting array %v, got %v", v1, v2)
		}
		if len(arr1) != len(arr2) {
			return fmt.Errorf("Different lengths: %v vs %v", v1, v2)
		}
		for i := range arr1 {
			if err := deepCompare(arr1[i], arr2[i]); err != nil {
				return err
			}
		}
		return nil
	}
	obj1, ok := v1.(map[string]any)
	if ok {
		obj2, ok := v2.(map[string]any)
		if !ok {
			return fmt.Errorf("Expecting obj %v, got %v", v1, v2)
		}
		if len(obj1) != len(obj2) {
			return fmt.Errorf("Different lengths: %v vs %v", v1, v2)
		}
		for k, val1 := range obj1 {
			val2, ok := obj2[k]
			if !ok {
				return fmt.Errorf("Missing key %s", k)
			}
			if k == "@type" {
				set1 := lpg.NewStringSet()
				set2 := lpg.NewStringSet()
				if a, ok := val1.([]any); ok {
					for _, x := range a {
						set1.Add(x.(string))
					}
				} else {
					set1.Add(val1.(string))
				}
				if a, ok := val2.([]any); ok {
					for _, x := range a {
						set2.Add(x.(string))
					}
				} else {
					set2.Add(val2.(string))
				}
				if !set1.IsEqual(set2) {
					return fmt.Errorf("Expected type %v, got %v", set1, set2)
				}
			} else {
				if err := deepCompare(val1, val2); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if reflect.DeepEqual(v1, v2) {
		return nil
	}
	return fmt.Errorf("Expecting %v got %v", v1, v2)
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
