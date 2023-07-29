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
	"reflect"
)

// Composer interface represents term composition algorithm. During
// layer composition, any term metadata that implements Composer
// interface will be composed using the customized implementation. If
// the term does not implement the Composer interface, Setcomposition
// will be used
type Composer interface {
	Compose(v1, v2 PropertyValue) (PropertyValue, error)
}

// CompositionType determines the composition semantics for the term
type CompositionType string

const (
	// SetComposition means when two terms are composed, set-union of the values are taken
	SetComposition CompositionType = "set"
	// ListComposition means when two terms are composed, their values are appended
	ListComposition CompositionType = "list"
	// OverrideComposition means when two terms are composed, the new one replaces the old one
	OverrideComposition CompositionType = "override"
	// NoComposition means when two terms are composed, the original remains
	NoComposition CompositionType = "nocompose"
	// ErrorComposition means if two terms are composed and they are different, composition fails
	ErrorComposition CompositionType = "error"
)

// GetComposerForTerm returns a term composer
func GetComposerForTerm(term string) Composer {
	info := GetTerm(term)
	c, ok := info.Metadata.(Composer)
	if ok {
		return c
	}
	return info.Composition
}

// GetComposerForProperty returns the composer for the given
// property. Never returns nil
func GetComposerForProperty(p PropertyValue) Composer {
	if p.Value() == nil {
		return nil
	}
	s := p.Sem()
	c, ok := s.Metadata.(Composer)
	if ok {
		return c
	}
	return s.Composition
}

// Compose target and src based on the composition type
func (c CompositionType) Compose(target, src PropertyValue) (PropertyValue, error) {
	switch c {
	case SetComposition:
		return SetUnion(target, src), nil
	case OverrideComposition:
		if src.Value() == nil {
			return target, nil
		}
		return src, nil
	case ListComposition:
		return ListAppend(target, src), nil
	case NoComposition:
		if target.Value() == nil {
			return src, nil
		}
		return target, nil
	case ErrorComposition:
		if target.Value() != nil && src.Value() != nil {
			// Composition is valid if values are the same
			if target.Equal(src) {
				return target, nil
			}
			return PropertyValue{}, ErrInvalidComposition
		}
		if target.Value() != nil {
			return target, nil
		}
		return src, nil
	}
	return SetUnion(target, src), nil
}

// SetUnion computes the set union of properties v1 and v2
func SetUnion(v1, v2 PropertyValue) PropertyValue {
	if v1.Value() == nil {
		return v2
	}
	if v2.Value() == nil {
		return v1
	}
	if v1.Value() == v2.Value() {
		return v1
	}
	values := make(map[reflect.Value]struct{})
	val1 := reflect.ValueOf(v1.Value())
	val2 := reflect.ValueOf(v2.Value())
	vtomap := func(value reflect.Value) {
		if value.Type().Kind() == reflect.Slice || value.Type().Kind() == reflect.Array {
			n := value.Len()
			for i := 0; i < n; i++ {
				values[value.Index(i)] = struct{}{}
			}
		} else {
			values[value] = struct{}{}
		}
	}
	vtomap(val1)
	vtomap(val2)

	makeSlice := func(elemType reflect.Type) PropertyValue {
		result := reflect.MakeSlice(reflect.SliceOf(elemType), 0, len(values))
		for k := range values {
			reflect.AppendSlice(result, k)
		}
		return NewPropertyValue(v1.Term(), result.Interface())
	}
	makeAnySlice := func() PropertyValue {
		result := make([]any, 0, len(values))
		for k := range values {
			result = append(result, k.Interface())
		}
		return NewPropertyValue(v1.Term(), result)
	}

	// If the union is between two same types, or slices/arrays of two
	// same types, or one slice/array and one compatible type, the
	// result is the slice of that type
	if val1.Type().Kind() == reflect.Slice || val1.Type().Kind() == reflect.Array {
		if val2.Type().Kind() == reflect.Slice || val2.Type().Kind() == reflect.Array {
			// Both are slice/array. If compatible elements, result is the same type
			if val1.Type().Elem() == val2.Type().Elem() {
				return makeSlice(val1.Type().Elem())
			}
			return makeAnySlice()
		}
		// val1 is a slice/array. val2 is a value
		if val1.Type().Elem() == val2.Type() {
			return makeSlice(val1.Type().Elem())
		}
		return makeAnySlice()
	}
	if val2.Type().Kind() == reflect.Slice || val2.Type().Kind() == reflect.Array {
		if val1.Type() == val2.Type().Elem() {
			return makeSlice(val2.Type().Elem())
		}
		return makeAnySlice()
	}
	return makeAnySlice()
}

// ListAppend appends v2 and v1
func ListAppend(v1, v2 PropertyValue) PropertyValue {
	val := GenericListAppend(v1.Value(), v2.Value())
	return NewPropertyValue(v1.Term(), val)
}
