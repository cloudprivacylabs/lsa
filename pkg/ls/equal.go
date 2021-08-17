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
	"fmt"
)

// ToMap converts the input to a map
func ToMap(in interface{}) interface{} {
	data, _ := json.Marshal(in)
	var v interface{}
	json.Unmarshal(data, &v)
	return v
}

// DeepEqual test equivalence between two JSON trees
func DeepEqual(i1, i2 interface{}) error {
	var deepEqualArray func([]interface{}, []interface{}) error
	var deepEqualMap func(map[string]interface{}, map[string]interface{}) error
	var deepEqualValue func(interface{}, interface{}) error

	toStr := func(in interface{}) string {
		x, _ := json.MarshalIndent(in, "", "  ")
		return string(x)
	}

	deepEqualArray = func(a1, a2 []interface{}) error {
		if len(a1) != len(a2) {
			return fmt.Errorf("Different lengths: %s\n %s", toStr(a1), toStr(a2))
		}
		for i := range a1 {
			if err := DeepEqual(a1[i], a2[i]); err != nil {
				return err
			}
		}
		return nil
	}

	deepEqualMap = func(m1, m2 map[string]interface{}) error {
		if len(m1) != len(m2) {
			return fmt.Errorf("Different lengths: %d vs %d\n first: %s\n second: %s", len(m1), len(m2), toStr(m1), toStr(m2))
		}
		for k, v := range m1 {
			val, exists := m2[k]
			if !exists {
				return fmt.Errorf("Missing key %s in %v/%v", k, m1, m2)
			}
			if err := DeepEqual(v, val); err != nil {
				return err
			}
		}
		return nil
	}

	deepEqualValue = func(v1, v2 interface{}) error {
		if a1, ok := i1.([]interface{}); ok {
			a2, ok := i2.([]interface{})
			if ok {
				return deepEqualArray(a1, a2)
			}
			return fmt.Errorf("1 array 2 not: 1: %s %T\n 2: %s %T\n", toStr(a1), a1, toStr(a2), a2)
		}
		if m1, ok := i1.(map[string]interface{}); ok {
			m2, ok := i2.(map[string]interface{})
			if ok {
				return deepEqualMap(m1, m2)
			}
			return fmt.Errorf("1 map 2 not: %v %T %v %T", m1, m1, m2, m2)
		}
		if v1 != v2 {
			return fmt.Errorf("Wrong value %v %T %v %T", v1, v1, v2, v2)
		}
		return nil
	}

	return deepEqualValue(i1, i2)
}
