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

// PropertyValue can be a string or []string. It is an immutable value object
type PropertyValue struct {
	value interface{}
}

// StringPropertyValue creates a string value
func StringPropertyValue(s string) *PropertyValue { return &PropertyValue{value: s} }

// StringSlicePropertyValue creates a []string value. If s is nil, it creates an empty slice
func StringSlicePropertyValue(s []string) *PropertyValue {
	if s == nil {
		return &PropertyValue{value: []string{}}
	}
	return &PropertyValue{value: s}
}

// AsString returns the value as string
func (p *PropertyValue) AsString() string {
	if p == nil {
		return ""
	}
	if s, ok := p.value.(string); ok {
		return s
	}
	return ""
}

// AsStringSlice returns the value as string slice
func (p *PropertyValue) AsStringSlice() []string {
	if p == nil {
		return nil
	}
	if s, ok := p.value.([]string); ok {
		return s
	}
	return nil
}

// AsInterfaceSlice returns an interface slice of the underlying value if it is a []string
func (p *PropertyValue) AsInterfaceSlice() []interface{} {
	if !p.IsStringSlice() {
		return nil
	}
	sl := p.AsStringSlice()
	ret := make([]interface{}, 0, len(sl))
	for _, x := range sl {
		ret = append(ret, x)
	}
	return ret
}

// IsString returns true if the underlying value is a string
func (p *PropertyValue) IsString() bool {
	if p == nil {
		return false
	}
	_, ok := p.value.(string)
	return ok
}

// IsStringSlice returns true if the underlying value is a string slice
func (p *PropertyValue) IsStringSlice() bool {
	if p == nil {
		return false
	}
	_, ok := p.value.([]string)
	return ok
}

func (p PropertyValue) Clone() *PropertyValue {
	return &PropertyValue{value: p.value}
}

// CopyPropertyMap returns a copy of the property map
func CopyPropertyMap(m map[string]*PropertyValue) map[string]*PropertyValue {
	ret := make(map[string]*PropertyValue, len(m))
	for k, v := range m {
		ret[k] = v.Clone()
	}
	return ret
}
