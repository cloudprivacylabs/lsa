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
package terms

import (
	"github.com/piprate/json-gold/ld"
)

// GetKeyValueFromExpanded returns the value of the given key from an
// expanded JSON-LD object. The input is
//
//   [ { key: value } ]
//
func GetKeyValueFromExpanded(key string, in interface{}) string {
	arr, _ := in.([]interface{})
	if len(arr) != 1 {
		return ""
	}
	return GetKeyValueFromMap(key, arr[0])
}

// GetKeyValueFromMap returns the value of the given key from an
// expanded JSON-LD object. The input is
//
//  { key: value }
//
func GetKeyValueFromMap(key string, in interface{}) string {
	if m, ok := in.(map[string]interface{}); ok {
		x, _ := m[key].(string)
		return x
	}
	return ""
}

// GetListElementsFromExpanded returns the elements of an expanded list
//
// The input is
//
//   [   { "@list": elements } ]
//
// The output is elements
func GetListElementsFromExpanded(in interface{}) []interface{} {
	arr, _ := in.([]interface{})
	if len(arr) != 1 {
		return nil
	}
	m, _ := arr[0].(map[string]interface{})
	if m == nil {
		return nil
	}
	elements, _ := m["@list"]
	if elements == nil {
		return nil
	}
	arr, _ = elements.([]interface{})
	return arr
}

// GetSetElementsFromExpanded returns the elements from an expanded
// set. The input is
//
//    [ { "@set": elements } ]
//
// or
//
//    elements
//
// The output is elements
func GetSetElementsFromExpanded(in interface{}) []interface{} {
	arr, _ := in.([]interface{})
	if len(arr) == 1 {
		if m, ok := arr[0].(map[string]interface{}); ok {
			if elements, ok := m["@set"]; ok {
				arr, _ = elements.([]interface{})
			}
		}
	}
	return arr
}

// GetStringsFromArray returns string values from a value or id
// array. The input is
//
//   [ { key: value }, { key:value } ... ]
//
// The output is
//
//   [ value, value, ... ]
//
func GetStringsFromArray(key string, in []interface{}) []string {
	ret := make([]string, 0, len(in))
	for _, x := range in {
		if m, ok := x.(map[string]interface{}); ok {
			s, _ := m[key].(string)
			ret = append(ret, s)
		}
	}
	return ret
}

// AppendLists appends t2 to t1. Returns copies of values.
func AppendLists(term ContainerTerm, t1, t2 interface{}) interface{} {
	el1 := term.ElementsFromExpanded(t1)
	el2 := term.ElementsFromExpanded(t2)
	ret := make([]interface{}, 0, len(el1)+len(el2))
	for _, x := range el1 {
		ret = append(ret, ld.CloneDocument(x))
	}
	for _, x := range el2 {
		ret = append(ret, ld.CloneDocument(x))
	}
	return term.MakeExpandedContainer(ret)
}

// UnionStringSets computes the set union of the string values t1 and t2.
func UnionStringSets(term StringContainerTerm, t1, t2 interface{}) interface{} {
	v1 := term.ElementValuesFromExpanded(t1)
	v2 := term.ElementValuesFromExpanded(t2)
	set := make(map[string]struct{}, len(v1)+len(v2))
	ret := make([]interface{}, 0, len(v1)+len(v2))
	for _, x := range v1 {
		if _, ok := set[x]; !ok {
			set[x] = struct{}{}
			ret = append(ret, term.MakeExpandedElement(x))
		}
	}
	for _, x := range v2 {
		if _, ok := set[x]; !ok {
			set[x] = struct{}{}
			ret = append(ret, term.MakeExpandedElement(x))
		}
	}
	return term.MakeExpandedContainer(ret)
}
