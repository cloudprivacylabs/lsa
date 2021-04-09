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
//
// parent is the parent of the new attributes in the sliced layer
func (attributes *ObjectType) Slice(filter func(string, *Attribute) bool, parent *Attribute, newLayer *Layer) *ObjectType {
	ret := NewObjectType(parent)
	for _, attr := range attributes.attributes {
		newAttr := attr.Slice(filter, parent, newLayer)
		if newAttr != nil {
			ret.Add(newAttr, newLayer)
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
//
// parent is the parent of the new attribute in the sliced layer
func (attribute *Attribute) Slice(filter func(string, *Attribute) bool, parent *Attribute, newLayer *Layer) *Attribute {
	newAttr := NewAttribute(parent)
	newAttr.ID = attribute.ID
	full := filter("", attribute)

	// Any term in attribute selected?
	for k, v := range attribute.Values {
		if filter(k, attribute) {
			newAttr.Values[k] = ld.CloneDocument(v)
			full = true
		}
	}

	fillOptions := func(options []*Attribute) []*Attribute {
		ret := make([]*Attribute, 0, len(options))
		for _, x := range options {
			attr := x.Slice(filter, attribute, newLayer)
			if attr != nil {
				ret = append(ret, attr)
			}
		}
		return ret
	}
	switch t := attribute.Type.(type) {
	case *ObjectType:
		n := t.Slice(filter, newAttr, newLayer)
		if n != nil {
			newAttr.Type = n
			full = true
		} else {
			newAttr.Type = NewObjectType(newAttr)
		}

	case ValueType, *ReferenceType:
		newAttr.Type = attribute.Type
	case *ArrayType:
		n := t.Attribute.Slice(filter, newAttr, newLayer)
		if n != nil {
			newAttr.Type = &ArrayType{n}
			full = true
		} else {
			newAttr.Type = NewArrayType(newAttr)
		}
	case *CompositeType:
		o := fillOptions(t.Options)
		if len(o) > 0 {
			newAttr.Type = &CompositeType{o}
			full = true
		} else {
			newAttr.Type = &CompositeType{}
		}
	case *PolymorphicType:
		o := fillOptions(t.Options)
		if len(o) > 0 {
			newAttr.Type = &PolymorphicType{o}
			full = true
		} else {
			newAttr.Type = &PolymorphicType{}
		}
	}

	if !full {
		return nil
	}

	return newAttr
}
