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

// Composer interface represents term composition algorithm. During
// layer composition, any term metadata that implements Composer
// interface will be composed using the customized implementation. If
// the term does not implement the Composer interface, Setcomposition
// will be used
type Composer interface {
	Compose(v1, v2 *PropertyValue) (*PropertyValue, error)
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
	info := GetTermInfo(term)
	c, ok := info.Metadata.(Composer)
	if ok {
		return c
	}
	return info.Composition
}

// Compose target and src based on the composition type
func (c CompositionType) Compose(target, src *PropertyValue) (*PropertyValue, error) {
	switch c {
	case SetComposition:
		return SetUnion(target, src), nil
	case OverrideComposition:
		if src == nil {
			return target, nil
		}
		return src, nil
	case ListComposition:
		return ListAppend(target, src), nil
	case NoComposition:
		if target == nil {
			return src, nil
		}
		return target, nil
	case ErrorComposition:
		if target != nil && src != nil {
			// Composition is valid if values are the same
			if target.Equal(src) {
				return target, nil
			}
			return nil, ErrInvalidComposition
		}
		if target != nil {
			return target, nil
		}
		return src, nil
	}
	return SetUnion(target, src), nil
}

// SetUnion computes the set union of properties v1 and v2
func SetUnion(v1, v2 *PropertyValue) *PropertyValue {
	if v1 == nil {
		return v2
	}
	if v2 == nil {
		return v1
	}
	if v1.IsStringSlice() {
		slc := v1.AsStringSlice()
		values := make(map[string]struct{}, len(slc))
		ret := make([]string, 0, len(slc))
		for _, k := range slc {
			values[k] = struct{}{}
			ret = append(ret, k)
		}
		if v2.IsStringSlice() {
			for _, item := range v2.AsStringSlice() {
				if _, exists := values[item]; !exists {
					values[item] = struct{}{}
					ret = append(ret, item)
				}
			}
			return StringSlicePropertyValue(ret)
		}
		if _, exists := values[v2.AsString()]; !exists {
			ret = append(ret, v2.AsString())
		}
		return StringSlicePropertyValue(ret)
	}
	ret := []string{v1.AsString()}
	if v2.IsStringSlice() {
		for _, item := range v2.AsStringSlice() {
			if item != ret[0] {
				ret = append(ret, item)
			}
		}
		if len(ret) == 1 {
			return StringPropertyValue(ret[0])
		}
		return StringSlicePropertyValue(ret)
	}
	if ret[0] != v2.AsString() {
		ret = append(ret, v2.AsString())
	}
	if len(ret) == 1 {
		return StringPropertyValue(ret[0])
	}
	return StringSlicePropertyValue(ret)
}

// ListAppend appends v2 and v1
func ListAppend(v1, v2 *PropertyValue) *PropertyValue {
	if v1 == nil {
		return v2
	}
	if v2 == nil {
		return v1
	}
	if v1.IsStringSlice() {
		ret := v1.AsStringSlice()
		if v2.IsStringSlice() {
			return StringSlicePropertyValue(append(ret, v2.AsStringSlice()...))
		}
		return StringSlicePropertyValue(append(ret, v2.AsString()))
	}
	ret := []string{v1.AsString()}
	if v2.IsStringSlice() {
		return StringSlicePropertyValue(append(ret, v2.AsStringSlice()...))
	}
	return StringSlicePropertyValue(append(ret, v2.AsString()))
}
