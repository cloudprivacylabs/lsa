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
	"reflect"

	"github.com/cloudprivacylabs/lpg/v2"
)

type PropertyIterator interface {
	ForEachProperty(func(string, any) bool) bool
}

// A PropertyValue contains the value for a node or edge property,
// plus its term semantics. The object is read-only, so if you want to
// change a property, you have to create a new one and set it.
type PropertyValue struct {
	sem   Term
	value any
}

// NewPropertyValue create a new property value for the given term and value
func NewPropertyValue(term string, value any) PropertyValue {
	return PropertyValue{
		sem:   GetTerm(term),
		value: value,
	}
}

// GetNativeValue is used by the LPG expression evaluators to access
// the native value of the property
func (p PropertyValue) GetNativeValue() any { return p.value }

// Value returns the property value
func (p PropertyValue) Value() any { return p.value }

// Sem returns the term semantics for the property value
func (pv PropertyValue) Sem() Term { return pv.sem }

// Term returns the term for the property value
func (pv PropertyValue) Term() string { return pv.sem.Name }

func (pv PropertyValue) MarshalYAML() (any, error) {
	return pv.value, nil
}

func (pv PropertyValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(pv.value)
}

// Equal compares two property values, and returns true if they are equal
func (pv PropertyValue) Equal(v PropertyValue) bool {
	if pv.value == v.value {
		return true
	}
	return reflect.DeepEqual(pv.value, v.value)
}

func (pv PropertyValue) String() string {
	return fmt.Sprint(pv.value)
}

// AsStringSlice will return a string, []string, or []any as a string
// slice
func (pv PropertyValue) AsStringSlice() []string {
	str, _ := StringSliceType{}.Coerce(pv.Value())
	if str == nil {
		return []string{}
	}
	return str.([]string)
}

// CopyPropertyMap returns a copy of the property map
func CopyPropertyMap(m map[string]PropertyValue) map[string]PropertyValue {
	ret := make(map[string]PropertyValue, len(m))
	for k, v := range m {
		ret[k] = v
	}
	return ret
}

// IsPropertiesEqual compares two property maps and returns true if they are equal
func IsPropertiesEqual(p, q map[string]PropertyValue) bool {
	if len(p) != len(q) {
		return false
	}
	for k, v := range p {
		if !v.Equal(q[k]) {
			return false
		}
	}
	return true
}

// GetPropertyValue gets the property value from the node or the edge,
// and tries to convert it to PropertyValue.
func GetPropertyValue(source interface {
	GetProperty(string) (any, bool)
}, key string) (PropertyValue, bool) {
	if source == nil {
		return PropertyValue{}, false
	}
	p, ok := source.GetProperty(key)
	if !ok {
		return PropertyValue{}, false
	}
	pv, ok := p.(PropertyValue)
	return pv, ok
}

// GetPropertyValueAs gets the property value from the node or the edge,
// and tries to convert it to PropertyValue, and then gets a value of type T from it.
func GetPropertyValueAs[T any](source lpg.WithProperties, key string) (value T, ok bool) {
	p, k := GetPropertyValue(source, key)
	if !k {
		return
	}
	value, ok = p.value.(T)
	return
}

// CloneProperties can be used to clone edge and node properties
func CloneProperties(iterator PropertyIterator) map[string]any {
	newProperties := make(map[string]any)
	iterator.ForEachProperty(func(key string, value any) bool {
		newProperties[key] = value
		return true
	})
	return newProperties
}

func PropertiesAsMap(iterator PropertyIterator) map[string]PropertyValue {
	ret := make(map[string]PropertyValue)
	iterator.ForEachProperty(func(key string, value interface{}) bool {
		if p, ok := value.(PropertyValue); ok {
			ret[key] = p
		}
		return true
	})
	return ret
}

// ClonePropertyFunc can be used in graph copy functions
func ClonePropertyValueFunc(key string, value any) any {
	if p, ok := value.(PropertyValue); ok {
		return p
	}
	return value
}
