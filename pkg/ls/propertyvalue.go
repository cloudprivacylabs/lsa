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
	"fmt"
	"strconv"
	"strings"
)

// PropertyValue can be a string or []string. It is an immutable value object
type PropertyValue struct {
	value interface{}
}

// IntPropertyValue converts the int value to string, and returns a string value
func IntPropertyValue(i int) *PropertyValue { return &PropertyValue{value: fmt.Sprint(i)} }

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

// Has checks if  p has the given string or is equal to it
func (p *PropertyValue) Has(s string) bool {
	if p.IsString() {
		return p.AsString() == s
	}
	if p.IsStringSlice() {
		for _, x := range p.AsStringSlice() {
			if x == s {
				return true
			}
		}
	}
	return false
}

// Equal compares two property values, and returns true if they are equal
func (p *PropertyValue) Equal(v *PropertyValue) bool {
	if p.IsString() && v.IsString() && p.AsString() == v.AsString() {
		return true
	}
	if p.IsStringSlice() && v.IsStringSlice() {
		s1 := p.AsStringSlice()
		s2 := v.AsStringSlice()
		if len(s1) == len(s2) {
			for i, x := range s1 {
				if s2[i] != x {
					return false
				}
			}
			return true
		}
	}
	return false
}

// Returns true if the underlying value is a string, and that string can be converted to int
func (p *PropertyValue) IsInt() bool {
	s, ok := p.value.(string)
	if !ok {
		return false
	}
	_, err := strconv.Atoi(s)
	return err == nil
}

// AsInt attempts to return the underlying string value as integer
func (p *PropertyValue) AsInt() int {
	i, _ := strconv.Atoi(p.value.(string))
	return i
}

func (p PropertyValue) Clone() *PropertyValue {
	return &PropertyValue{value: p.value}
}

func (p PropertyValue) String() string {
	if p.IsString() {
		return p.AsString()
	}
	if p.IsStringSlice() {
		return strings.Join(p.AsStringSlice(), ",")
	}
	return fmt.Sprint(p.value)
}

// CopyPropertyMap returns a copy of the property map
func CopyPropertyMap(m map[string]*PropertyValue) map[string]*PropertyValue {
	ret := make(map[string]*PropertyValue, len(m))
	for k, v := range m {
		ret[k] = v.Clone()
	}
	return ret
}
