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
	"github.com/piprate/json-gold/ld"
)

// Slice creates a copy of the attributes object by including terms
// selected by the filter function. All attribute nodes are processed,
// thus the filter function may return false for a node A and then
// true for a descendant of A, which causes A to be included. If the
// slice is empty, returns nil. This will call filter once with empty
// term, and f should return if the attribute should be included or
// not.
func (attributes *Attributes) Slice(filter func(string, *Attribute) bool) *Attributes {
	ret := NewAttributes(nil)
	for _, attr := range attributes.attributes {
		newAttr := attr.Slice(filter)
		if newAttr != nil {
			ret.Add(newAttr)
		}
	}
	if ret.Len() > 0 {
		return ret
	}
	return nil
}

// Slice creates a copy of the attribute if either filter selects this
// attribute, or any term in this attribute, or any attributes under
// this attribute. This will call filter once with empty term, and f
// should return if the attribute should be included or not.
func (attribute *Attribute) Slice(filter func(string, *Attribute) bool) *Attribute {
	newAttr := NewAttribute(nil)
	newAttr.ID = attribute.ID
	empty := !filter("", attribute)

	// Any term in attribute selected?
	for k, v := range attribute.Values {
		if filter(k, attribute) {
			newAttr.Values[k] = ld.CloneDocument(v)
			empty = false
		}
	}
	if attribute.attributes != nil {
		newAttr.attributes = attribute.attributes.Slice(filter)
		if newAttr.attributes != nil {
			empty = false
		}
	} else if attribute.arrayItems != nil {
		newAttr.arrayItems = attribute.arrayItems.Slice(filter)
		if newAttr.arrayItems != nil {
			empty = false
		}
	} else if attribute.allOf != nil {
		for _, x := range attribute.allOf {
			n := x.Slice(filter)
			if n != nil {
				newAttr.allOf = append(newAttr.allOf, n)
				empty = false
			}
		}
	} else if attribute.oneOf != nil {
		for _, x := range attribute.oneOf {
			n := x.Slice(filter)
			if n != nil {
				newAttr.oneOf = append(newAttr.oneOf, n)
				empty = false
			}
		}
	} else if len(attribute.reference) > 0 {
		if !empty {
			newAttr.reference = attribute.reference
		}
	}

	if empty {
		return nil
	}
	return newAttr
}
