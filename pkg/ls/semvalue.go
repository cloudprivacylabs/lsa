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

import ()

// A ValueFilter applies a filter to the node value
type ValueFilter interface {
	FilterValue(interface{}, Node) interface{}
}

// NopFilter does not modify the underlying value
type NopFilter struct{}

func (NopFilter) FilterValue(in interface{}, _ Node) interface{} { return in }

// GetValueFilter returns the value filter for the term. If the term has none, returns NopFilter
func GetValueFilter(term string) ValueFilter {
	flt, ok := GetTermMetadata(term).(ValueFilter)
	if ok {
		return flt
	}
	return NopFilter{}
}

// FilterValue computes the processed node value
func FilterValue(value interface{}, docnode Node, properties map[string]*PropertyValue) interface{} {
	for k := range properties {
		value = GetValueFilter(k).FilterValue(value, docnode)
	}
	return value
}
