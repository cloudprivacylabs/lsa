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

// Attribute contains the structural elements of attributes as well as
// additional attributes
type Attribute struct {
	ID     string
	Values map[string]interface{}
	Type   AttributeType

	parent *Attribute
}

type AttributeType interface {
	GetType() string
	Clone(*Attribute) AttributeType
	Iterate(func(*Attribute) bool) bool
	UnmarshalExpanded(map[string]interface{}, *Attribute) error
	MarshalExpanded(map[string]interface{})
}

var AttributeTypes = struct {
	Value       string
	Object      string
	Reference   string
	Array       string
	Composite   string
	Polymorphic string
}{
	Value:       LS + "/Value",
	Object:      LS + "/Object",
	Reference:   LS + "/Reference",
	Array:       LS + "/Array",
	Composite:   LS + "/Composite",
	Polymorphic: LS + "/Polymorphic",
}

// NewAttribute returns a new value attribute
func NewAttribute(parent *Attribute) *Attribute {
	return &Attribute{parent: parent,
		Values: make(map[string]interface{}),
	}
}

// Clone returns a deep-copy of an attribute with a different parent
func (attribute *Attribute) Clone(parent *Attribute) *Attribute {
	ret := &Attribute{ID: attribute.ID,
		Values: ld.CloneDocument(attribute.Values).(map[string]interface{}),
		parent: parent,
	}
	if attribute.Type != nil {
		ret.Type = attribute.Type.Clone(ret)
	}
	return ret
}

// GetParent returns the parent layer object
func (attribute *Attribute) GetParent() *Attribute {
	return attribute.parent
}

// GetPath returns all objects from root to this attribute
func (attribute *Attribute) GetPath() []*Attribute {
	if attribute.parent == nil {
		return []*Attribute{attribute}
	}
	return append(attribute.parent.GetPath(), attribute)
}

// ComposeOptions composes the options of this composite attribute and makes it into a new
// object attribute
func (attribute *Attribute) ComposeOptions(layer *Layer) error {
	if attribute.Type == nil {
		return ErrNotACompositeType(attribute.ID)
	}
	composition, ok := attribute.Type.(*CompositeType)
	if !ok {
		return ErrNotACompositeType(attribute.ID)
	}
	result := NewObjectType(attribute)
	// All options must be an attribute or object
	for _, option := range composition.Options {
		if option.Type == nil {
			return ErrInvalidCompositeType(attribute.ID)
		}
	redo:
		switch optionType := option.Type.(type) {
		case ValueType:
			result.Add(option.Clone(attribute), layer)
		case *CompositeType:
			if err := option.ComposeOptions(layer); err != nil {
				return err
			}
			if _, ok := option.Type.(*CompositeType); ok {
				return ErrInvalidCompositeType(attribute.ID)
			}
			goto redo
		case *ObjectType:
			for i := 0; i < optionType.Len(); i++ {
				result.Add(optionType.Get(i).Clone(attribute), layer)
			}
		case *PolymorphicType, *ReferenceType, *ArrayType:
			return ErrInvalidCompositeType(attribute.ID)
		}
	}
	attribute.Type = result
	return nil
}

// UnmarshalExpanded unmarshals an attribute. The input is a
// map[string]interface{}. The attribute may or may not have an ID.
// It must have a type
func (attribute *Attribute) UnmarshalExpanded(in interface{}, parent *Attribute) error {
	m, ok := in.(map[string]interface{})
	if !ok {
		return ErrInvalidInput("Invalid attribute")
	}
	attribute.ID = GetNodeID(in)
	attribute.parent = parent
	attribute.Values = make(map[string]interface{}, len(m))
	t := GetNodeType(in)
	switch t {
	case AttributeTypes.Value:
		attribute.Type = ValueType{}
	case AttributeTypes.Object:
		attribute.Type = NewObjectType(attribute)
	case AttributeTypes.Reference:
		attribute.Type = &ReferenceType{}
	case AttributeTypes.Array:
		attribute.Type = NewArrayType(attribute)
	case AttributeTypes.Composite:
		attribute.Type = &CompositeType{}
	case AttributeTypes.Polymorphic:
		attribute.Type = &PolymorphicType{}
	case "":
		// Infer attribute type
		n := 0
		for k := range m {
			switch k {
			case string(LayerTerms.Attributes):
				n++
				attribute.Type = NewObjectType(attribute)
			case string(LayerTerms.Reference):
				n++
				attribute.Type = &ReferenceType{}
			case string(LayerTerms.ArrayItems):
				n++
				attribute.Type = NewArrayType(attribute)
			case string(LayerTerms.AllOf):
				n++
				attribute.Type = &CompositeType{}
			case string(LayerTerms.OneOf):
				n++
				attribute.Type = &PolymorphicType{}
			}
		}
		if n == 0 {
			attribute.Type = ValueType{}
		} else if n > 1 {
			return ErrInvalidInput("Cannot determine type of attribute:" + attribute.ID)
		}
	default:
		return ErrInvalidAttributeType(attribute.ID + "/" + t)
	}
	if err := attribute.Type.UnmarshalExpanded(m, attribute); err != nil {
		return err
	}
	for k, v := range m {
		switch k {
		case "@id",
			"@type",
			string(LayerTerms.Attributes),
			string(LayerTerms.Reference),
			string(LayerTerms.ArrayItems),
			string(LayerTerms.AllOf),
			string(LayerTerms.OneOf):
		default:
			attribute.Values[k] = v
		}
	}
	return nil
}

// MarshalExpanded marshals the attribute as an expanded JSON-LD document
func (attribute *Attribute) MarshalExpanded() map[string]interface{} {
	ret := make(map[string]interface{})

	if len(attribute.ID) != 0 {
		ret["@id"] = attribute.ID
	}
	if attribute.Type != nil {
		ret["@type"] = []interface{}{attribute.Type.GetType()}
		attribute.Type.MarshalExpanded(ret)
	}
	for k, v := range attribute.Values {
		ret[k] = v
	}
	return ret
}

// Iterates the child attributes depth first, calls f for each
// attribute until f returns false
func (attribute *Attribute) Iterate(f func(*Attribute) bool) bool {
	if !f(attribute) {
		return false
	}
	if attribute.Type != nil {
		if !attribute.Type.Iterate(f) {
			return false
		}
	}
	return true
}

func (attribute *Attribute) IsValue() bool {
	_, ok := attribute.Type.(ValueType)
	return ok
}

func (attribute *Attribute) GetObjectType() *ObjectType {
	x, _ := attribute.Type.(*ObjectType)
	return x
}

func (attribute *Attribute) GetArrayItems() *Attribute {
	x, ok := attribute.Type.(*ArrayType)
	if ok {
		return x.Attribute
	}
	return nil
}

func (attribute *Attribute) GetPolymorphicOptions() []*Attribute {
	x, ok := attribute.Type.(*PolymorphicType)
	if ok {
		return x.Options
	}
	return nil
}

func (attribute *Attribute) GetCompositionOptions() []*Attribute {
	x, ok := attribute.Type.(*CompositeType)
	if ok {
		return x.Options
	}
	return nil
}

func (attribute *Attribute) GetAttributes() *ObjectType {
	x, _ := attribute.Type.(*ObjectType)
	return x
}

type ValueType struct{}

func (v ValueType) GetType() string                                            { return AttributeTypes.Value }
func (v ValueType) Iterate(func(*Attribute) bool) bool                         { return true }
func (v ValueType) Clone(*Attribute) AttributeType                             { return ValueType{} }
func (v ValueType) UnmarshalExpanded(map[string]interface{}, *Attribute) error { return nil }
func (v ValueType) MarshalExpanded(map[string]interface{})                     {}

type ReferenceType struct {
	Reference string
}

func (r ReferenceType) GetType() string                    { return AttributeTypes.Reference }
func (r ReferenceType) Iterate(func(*Attribute) bool) bool { return true }
func (r ReferenceType) Clone(*Attribute) AttributeType     { return &ReferenceType{r.Reference} }
func (r *ReferenceType) UnmarshalExpanded(input map[string]interface{}, parent *Attribute) error {
	r.Reference = LayerTerms.Reference.GetExpandedString(input)
	if len(r.Reference) == 0 {
		return ErrInvalidInput("Empty reference")
	}
	return nil
}
func (r ReferenceType) MarshalExpanded(out map[string]interface{}) {
	LayerTerms.Reference.PutExpanded(out, r.Reference)
}

type ArrayType struct {
	*Attribute
}

func NewArrayType(parent *Attribute) *ArrayType {
	return &ArrayType{Attribute: NewAttribute(parent)}
}

func (a ArrayType) GetType() string { return AttributeTypes.Array }

func (a ArrayType) Iterate(f func(*Attribute) bool) bool {
	if a.Type != nil {
		return a.Type.Iterate(f)
	}
	return true
}

func (a ArrayType) Clone(parent *Attribute) AttributeType {
	return &ArrayType{Attribute: a.Attribute.Clone(parent)}
}

func (a *ArrayType) UnmarshalExpanded(input map[string]interface{}, parent *Attribute) error {
	a.Attribute = NewAttribute(parent)
	items, _ := input[LayerTerms.ArrayItems.GetTerm()].([]interface{})
	if len(items) != 1 {
		return ErrInvalidInput("Empty array item")
	}
	if err := a.Attribute.UnmarshalExpanded(items[0], parent); err != nil {
		return err
	}
	return nil
}

func (a ArrayType) MarshalExpanded(out map[string]interface{}) {
	r := a.Attribute.MarshalExpanded()
	out[LayerTerms.ArrayItems.GetTerm()] = []interface{}{r}
}

type PolymorphicType struct {
	Options []*Attribute
}

func optionsToList(options []*Attribute) map[string]interface{} {
	out := make([]interface{}, 0, len(options))
	for _, x := range options {
		r := x.MarshalExpanded()
		out = append(out, r)
	}
	return map[string]interface{}{"@list": out}
}

func (p PolymorphicType) GetType() string { return AttributeTypes.Polymorphic }
func (p PolymorphicType) Iterate(f func(*Attribute) bool) bool {
	for _, attr := range p.Options {
		if !attr.Iterate(f) {
			return false
		}
	}
	return true
}

func (p PolymorphicType) Clone(parent *Attribute) AttributeType {
	ret := make([]*Attribute, len(p.Options))
	for i, x := range p.Options {
		ret[i] = x.Clone(parent)
	}
	return &PolymorphicType{Options: ret}
}

func (p *PolymorphicType) UnmarshalExpanded(input map[string]interface{}, parent *Attribute) error {
	allOf := LayerTerms.OneOf.ElementsFromExpanded(input[LayerTerms.OneOf.GetTerm()])
	if len(allOf) != 1 {
		return ErrInvalidInput("Invalid polymorphic type")
	}
	p.Options = make([]*Attribute, 0)
	for _, element := range GetListElements(allOf[0]) {
		attr := NewAttribute(parent)
		if err := attr.UnmarshalExpanded(element, parent); err != nil {
			return err
		}
		p.Options = append(p.Options, attr)
	}
	return nil
}

func (p PolymorphicType) MarshalExpanded(out map[string]interface{}) {
	out[LayerTerms.OneOf.GetTerm()] = []interface{}{optionsToList(p.Options)}
}

type CompositeType struct {
	Options []*Attribute
}

func (c CompositeType) GetType() string { return AttributeTypes.Composite }
func (c CompositeType) Iterate(f func(*Attribute) bool) bool {
	for _, attr := range c.Options {
		if !attr.Iterate(f) {
			return false
		}
	}
	return true
}

func (c CompositeType) Clone(parent *Attribute) AttributeType {
	ret := make([]*Attribute, len(c.Options))
	for i, x := range c.Options {
		ret[i] = x.Clone(parent)
	}
	return &CompositeType{ret}
}

func (c *CompositeType) UnmarshalExpanded(input map[string]interface{}, parent *Attribute) error {
	allOf := LayerTerms.OneOf.ElementsFromExpanded(input[LayerTerms.OneOf.GetTerm()])
	if len(allOf) != 1 {
		return ErrInvalidInput("Invalid composite type")
	}
	c.Options = make([]*Attribute, 0)
	for _, element := range GetListElements(allOf[0]) {
		attr := NewAttribute(parent)
		if err := attr.UnmarshalExpanded(element, parent); err != nil {
			return err
		}
		c.Options = append(c.Options, attr)
	}
	return nil
}

func (c CompositeType) MarshalExpanded(out map[string]interface{}) {
	out[LayerTerms.AllOf.GetTerm()] = []interface{}{optionsToList(c.Options)}
}

// ObjectType describes an object
type ObjectType struct {
	attribute    *Attribute
	attributes   []*Attribute
	attributeMap map[string]*Attribute
}

// NewObjectType returns a new empty object for the given attribute
func NewObjectType(attr *Attribute) *ObjectType {
	return &ObjectType{
		attribute:    attr,
		attributes:   make([]*Attribute, 0),
		attributeMap: make(map[string]*Attribute),
	}
}

func (attributes ObjectType) GetType() string { return AttributeTypes.Object }

func (attributes *ObjectType) Clone(parent *Attribute) AttributeType {
	ret := &ObjectType{
		attribute:    parent,
		attributes:   make([]*Attribute, len(attributes.attributes)),
		attributeMap: make(map[string]*Attribute, len(attributes.attributes)),
	}
	for i, a := range attributes.attributes {
		newNode := a.Clone(parent)
		ret.attributes[i] = newNode
		ret.attributeMap[newNode.ID] = newNode
	}
	return ret
}

// Len returns the number of attributes included in this Attributes object
func (attributes *ObjectType) Len() int {
	return len(attributes.attributes)
}

// Get returns the n'th attribute
func (attributes *ObjectType) Get(n int) *Attribute {
	return attributes.attributes[n]
}

// GetByID returns the attribute by its ID. The returned attribute is
// an immediate child of this Attributed object
func (attributes *ObjectType) GetByID(ID string) *Attribute {
	return attributes.attributeMap[ID]
}

// Iterates the child attributes depth first, calls f for each
// attribute until f returns false
func (attributes *ObjectType) Iterate(f func(*Attribute) bool) bool {
	for _, x := range attributes.attributes {
		if !x.Iterate(f) {
			return false
		}
	}
	return true
}

// Add a new attribute to this attributes. If the same ID was used or
// if the attribute does not have an ID, returns error
func (attributes *ObjectType) Add(attribute *Attribute, layer *Layer) error {
	if len(attribute.ID) == 0 {
		return ErrAttributeWithoutID
	}
	if _, exists := layer.Index[attribute.ID]; exists {
		return ErrDuplicateAttribute(attribute.ID)
	}
	attribute.parent = attributes.attribute
	attributes.attributes = append(attributes.attributes, attribute)
	attributes.attributeMap[attribute.ID] = attribute
	layer.Index[attribute.ID] = attribute
	return nil
}

// UmarshalExpanded unmarshals an expanded jsonld document to
// attributes. The input is []interface{} where each element is an
// attribute
func (attributes *ObjectType) UnmarshalExpanded(in map[string]interface{}, parent *Attribute) error {
	arr := LayerTerms.Attributes.ElementsFromExpanded(in[LayerTerms.Attributes.GetTerm()])
	if arr == nil {
		return ErrInvalidInput("Invalid attributes")
	}
	attributes.attributes = make([]*Attribute, 0, len(arr))
	attributes.attributeMap = make(map[string]*Attribute, len(arr))
	attributes.attribute = parent
	for _, attr := range arr {
		a := Attribute{}
		if err := a.UnmarshalExpanded(attr, parent); err != nil {
			return err
		}
		if len(a.ID) == 0 {
			return ErrAttributeWithoutID
		}
		a.parent = parent
		attributes.attributes = append(attributes.attributes, &a)
		attributes.attributeMap[a.ID] = &a
	}
	return nil
}

// Marshal attributes as an expanded JSON-LD object
func (attributes *ObjectType) MarshalExpanded(out map[string]interface{}) {
	ret := make([]interface{}, 0, len(attributes.attributes))
	for _, x := range attributes.attributes {
		ret = append(ret, x.MarshalExpanded())
	}
	out[LayerTerms.Attributes.GetTerm()] = ret
}
