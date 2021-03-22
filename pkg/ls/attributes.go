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

// Attributes describes an object
type Attributes struct {
	parent       SchemaObject
	attributes   []*Attribute
	attributeMap map[string]*Attribute
}

// NewAttributes returns a new empty attribute object with the given parent
func NewAttributes(parent SchemaObject) *Attributes {
	return &Attributes{parent: parent,
		attributes:   make([]*Attribute, 0),
		attributeMap: make(map[string]*Attribute),
	}
}

func (attributes *Attributes) Clone(parent SchemaObject) *Attributes {
	ret := &Attributes{parent: parent,
		attributes:   make([]*Attribute, len(attributes.attributes)),
		attributeMap: make(map[string]*Attribute, len(attributes.attributes)),
	}
	for i, a := range attributes.attributes {
		newNode := a.Clone(ret)
		ret.attributes[i] = newNode
		ret.attributeMap[newNode.ID] = newNode
	}
	return ret
}

// GetParent returns the parent object.
func (attributes *Attributes) GetParent() SchemaObject {
	return attributes.parent
}

// Len returns the number of attributes included in this Attributes object
func (attributes *Attributes) Len() int {
	return len(attributes.attributes)
}

// Get returns the n'th attribute
func (attributes *Attributes) Get(n int) *Attribute {
	return attributes.attributes[n]
}

// GetByID returns the attribute by its ID. The returned attribute is
// an immediate child of this Attributed object
func (attributes *Attributes) GetByID(ID string) *Attribute {
	return attributes.attributeMap[ID]
}

// FindByID searches the attribute with the given ID under this
// attributes
func (attributes *Attributes) FindByID(ID string) *Attribute {
	var ret *Attribute
	attributes.Iterate(func(a *Attribute) bool {
		if a.ID == ID {
			ret = a
			return false
		}
		return true
	})
	return ret
}

// Iterates the child attributes depth first, calls f for each
// attribute until f returns false
func (attributes *Attributes) Iterate(f func(*Attribute) bool) bool {
	for _, x := range attributes.attributes {
		if !x.Iterate(f) {
			return false
		}
	}
	return true
}

// Add a new attribute to this attributes. If the same ID was used or
// if the attribute does not have an ID, returns error
func (attributes *Attributes) Add(attribute *Attribute) error {
	if len(attribute.ID) == 0 {
		return ErrAttributeWithoutID
	}
	if _, exists := attributes.attributeMap[attribute.ID]; exists {
		return ErrDuplicateAttribute(attribute.ID)
	}
	attribute.parent = attributes
	attributes.attributes = append(attributes.attributes, attribute)
	attributes.attributeMap[attribute.ID] = attribute
	return nil
}

// UmarshalExpanded unmarshals an expanded jsonld document to
// attributes. The input is []interface{} where each element is an
// attribute
func (attributes *Attributes) UnmarshalExpanded(in interface{}) error {
	arr, ok := in.([]interface{})
	if !ok {
		return ErrInvalidInput
	}
	attributes.attributes = make([]*Attribute, 0, len(arr))
	attributes.attributeMap = make(map[string]*Attribute, len(arr))
	for _, attr := range arr {
		a := Attribute{}
		if err := a.UnmarshalExpanded(attr); err != nil {
			return err
		}
		if len(a.ID) == 0 {
			return ErrAttributeWithoutID
		}
		a.parent = attributes
		attributes.attributes = append(attributes.attributes, &a)
		attributes.attributeMap[a.ID] = &a
	}
	return nil
}

// Marshal attributes as an expanded JSON-LD object
func (attributes *Attributes) MarshalExpanded() []interface{} {
	out := make([]interface{}, 0, len(attributes.attributes))
	for _, x := range attributes.attributes {
		out = append(out, x.MarshalExpanded())
	}
	return out
}
