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
	Compose(interface{}, interface{}) (interface{}, error)
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
	md := GetTermMetadata(term)
	if md == nil {
		return SetComposition
	}
	c, ok := md.(Composer)
	if ok {
		return c
	}
	return SetComposition
}

// Compose target and src based on the composition type
func (c CompositionType) Compose(target, src interface{}) (interface{}, error) {
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
		return target, nil
	case ErrorComposition:
		if target != src && src != nil {
			return nil, ErrInvalidComposition
		}
		return target, nil
	}
	return SetUnion(target, src), nil
}

// SetUnion computes the set union of properties v1 and v2
func SetUnion(v1, v2 interface{}) interface{} {
	if v1 == nil {
		return v2
	}
	if v2 == nil {
		return v1
	}
	switch e := v1.(type) {
	case []interface{}:
		values := make(map[interface{}]struct{})
		for _, k := range e {
			values[k] = struct{}{}
		}
		ret := e
		if n, ok := v2.([]interface{}); ok {
			for _, item := range n {
				if _, exists := values[item]; !exists {
					values[item] = struct{}{}
					ret = append(ret, item)
				}
			}
			return ret
		}
		if _, exists := values[v2]; !exists {
			return append(e, v2)
		}
		return e
	default:
		ret := []interface{}{e}
		if n, ok := v2.([]interface{}); ok {
			for _, item := range n {
				if item != e {
					ret = append(ret, item)
				}
			}
			if len(ret) == 1 {
				return ret[0]
			}
			return ret
		}
		if e != v2 {
			return []interface{}{e, v2}
		}
		return e
	}
}

// ListAppend appends v2 and v1
func ListAppend(v1, v2 interface{}) interface{} {
	if v1 == nil {
		return v2
	}
	if v2 == nil {
		return v1
	}
	switch e := v1.(type) {
	case []interface{}:
		ret := e
		if n, ok := v2.([]interface{}); ok {
			return append(ret, n...)
		}
		return append(e, v2)
	default:
		ret := []interface{}{e}
		if n, ok := v2.([]interface{}); ok {
			return append(ret, n...)
		}
		return []interface{}{e, v2}
	}
}
