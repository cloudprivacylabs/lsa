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

// SchemaLayer contains the schema/overlay model
type SchemaLayer struct {
	ID         string
	Type       string
	ObjectType string
	Attributes
	Values map[string]interface{}
}

func (layer *SchemaLayer) Clone() *SchemaLayer {
	return &SchemaLayer{ID: layer.ID,
		Type:       layer.Type,
		ObjectType: layer.ObjectType,
		Attributes: *layer.Attributes.Clone(nil),
		Values:     ld.CloneDocument(layer.Values).(map[string]interface{}),
	}
}

// NewEmptySchemaLayer returns an empty schema kayer
func NewEmptySchemaLayer() *SchemaLayer {
	return &SchemaLayer{Attributes: *NewAttributes(nil),
		Values: make(map[string]interface{}),
	}
}

// NewSchemaLayer expands the jsonld input and creates a new schema layer
func NewSchemaLayer(jsonldInput interface{}) (*SchemaLayer, error) {
	proc := ld.NewJsonLdProcessor()
	expanded, err := proc.Expand(jsonldInput, nil)
	if err != nil {
		return nil, err
	}
	ret := SchemaLayer{}
	if err := ret.UnmarshalExpanded(expanded); err != nil {
		return nil, err
	}
	return &ret, nil
}

// UnmarshalExpanded unmarshals an expanded json-ld schema or overlay. The input
// must be a []interface{}
func (layer *SchemaLayer) UnmarshalExpanded(in interface{}) error {
	arr, _ := in.([]interface{})
	if len(arr) != 1 {
		return ErrInvalidInput
	}
	layer.ID = ""
	layer.Type = ""
	layer.Values = make(map[string]interface{})
	layer.Type = GetNodeType(arr[0])
	for k, v := range arr[0].(map[string]interface{}) {
		switch k {
		case "@id":
			layer.ID = v.(string)
		case AttributeStructure.Attributes.ID:
			if err := layer.Attributes.UnmarshalExpanded(v); err != nil {
				return err
			}
		case SchemaTerms.ObjectType.ID:
			layer.ObjectType = GetStringValue("@value", v)
		default:
			layer.Values[k] = v
		}
	}
	if layer.Type != TermSchemaBaseType && layer.Type != TermOverlayType {
		return ErrInvalidLayerType(layer.Type)
	}
	return nil
}

// MarshalExpanded returns marshaled layer in expanded json-ld format
func (layer *SchemaLayer) MarshalExpanded() interface{} {
	ret := map[string]interface{}{}
	if len(layer.ID) > 0 {
		ret["@id"] = layer.ID
	}
	if len(layer.Type) > 0 {
		ret["@type"] = layer.Type
	}
	if len(layer.ObjectType) > 0 {
		ret[SchemaTerms.ObjectType.ID] = []interface{}{map[string]interface{}{"@value": layer.ObjectType}}
	}
	for k, v := range layer.Values {
		ret[k] = v
	}
	ret[AttributeStructure.Attributes.ID] = layer.Attributes.MarshalExpanded()
	return []interface{}{ret}
}

// Validate the schema layer. Checks for duplicate attributes
func (layer *SchemaLayer) Validate() error {
	var err error
	ids := map[string]struct{}{}
	layer.Attributes.Iterate(func(a *Attribute) bool {
		if len(a.ID) > 0 {
			if _, exists := ids[a.ID]; exists {
				err = ErrDuplicateAttribute(a.ID)
				return false
			}
			ids[a.ID] = struct{}{}
		}
		return true
	})
	return err
}

// A SchemaObject is either Attributes or Attribute
type SchemaObject interface {
	GetParent() SchemaObject
}

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

// Attribute contains the structural elements of attributes as well as
// additional attributes
type Attribute struct {
	ID     string
	Values map[string]interface{}

	parent     SchemaObject
	attributes *Attributes
	reference  string
	arrayItems *Attribute
	allOf      []*Attribute
	oneOf      []*Attribute
}

// NewAttribute returns a new value attribute
func NewAttribute(parent SchemaObject) *Attribute {
	return &Attribute{parent: parent,
		Values: make(map[string]interface{}),
	}
}

func (attribute *Attribute) Clone(parent SchemaObject) *Attribute {
	ret := &Attribute{ID: attribute.ID,
		Values:    ld.CloneDocument(attribute.Values).(map[string]interface{}),
		parent:    parent,
		reference: attribute.reference,
	}
	if attribute.attributes != nil {
		ret.attributes = attribute.attributes.Clone(ret)
	}
	if attribute.arrayItems != nil {
		ret.arrayItems = attribute.arrayItems.Clone(ret)
	}
	if attribute.allOf != nil {
		ret.allOf = make([]*Attribute, len(attribute.allOf))
		for i, x := range attribute.allOf {
			ret.allOf[i] = x.Clone(ret)
		}
	}
	if attribute.oneOf != nil {
		ret.oneOf = make([]*Attribute, len(attribute.oneOf))
		for i, x := range attribute.oneOf {
			ret.oneOf[i] = x.Clone(ret)
		}
	}
	return ret
}

// GetParent returns the parent schema object
func (attribute *Attribute) GetParent() SchemaObject {
	return attribute.parent
}

// IsValue is true if the attribute denotes a value
func (attribute *Attribute) IsValue() bool {
	return attribute.attributes == nil &&
		len(attribute.reference) == 0 &&
		attribute.arrayItems == nil &&
		attribute.allOf == nil &&
		attribute.oneOf == nil
}

// IsReference returns true if attribute is a reference
func (attribute *Attribute) IsReference() bool {
	return len(attribute.reference) != 0
}

// GetReference returns the reference
func (attribute *Attribute) GetReference() string {
	return attribute.reference
}

// IsObject returns true if attribute is an object, and thus has attributes
func (attribute *Attribute) IsObject() bool {
	return attribute.attributes != nil
}

// GetAttributes returns the object attributes
func (attribute *Attribute) GetAttributes() *Attributes {
	return attribute.attributes
}

// IsArray returns if this attribute is an array
func (attribute *Attribute) IsArray() bool {
	return attribute.arrayItems != nil
}

// GetArrayItems returns the items descriptor for the array
func (attribute *Attribute) GetArrayItems() *Attribute {
	return attribute.arrayItems
}

// IsComposition returns if the attribute has allOf
func (attribute *Attribute) IsComposition() bool {
	return attribute.allOf != nil
}

// GetCompositionOptions returns the allOf options
func (attribute *Attribute) GetCompositionOptions() []*Attribute {
	return attribute.allOf
}

// ComposeOptions composes the options of this attribute and makes it into a new
// object
func (attribute *Attribute) ComposeOptions() error {
	if attribute.allOf == nil {
		return nil
	}
	ret := &Attributes{parent: attribute,
		attributes:   make([]*Attribute, 0),
		attributeMap: make(map[string]*Attribute)}
	// All options must be an attribute
	for _, option := range attribute.allOf {
		if option.IsComposition() {
			if err := option.ComposeOptions(); err != nil {
				return err
			}
		}
		if err := ret.Add(option); err != nil {
			return err
		}
	}
	attribute.MakeObject(ret)
	return nil
}

// IsPolyorphic returns if attribute has oneOf
func (attribute *Attribute) IsPolymorphic() bool {
	return attribute.oneOf != nil
}

// GetPolymorphicOptions returns the options for oneOf
func (attribute *Attribute) GetPolymorphicOptions() []*Attribute {
	return attribute.oneOf
}

// MakeObject converts this attribute to an object by setting its
// attributes.
func (attribute *Attribute) MakeObject(attributes *Attributes) {
	attribute.attributes = attributes
	attributes.parent = attribute
	attribute.reference = ""
	attribute.arrayItems = nil
	attribute.allOf = nil
	attribute.oneOf = nil
}

// MakeReference converts this attribute to a reference
func (attribute *Attribute) MakeReference(ref string) {
	attribute.attributes = nil
	attribute.reference = ref
	attribute.arrayItems = nil
	attribute.allOf = nil
	attribute.oneOf = nil
}

// MakeValue converts this attribute to a value
func (attribute *Attribute) MakeValue() {
	attribute.attributes = nil
	attribute.reference = ""
	attribute.arrayItems = nil
	attribute.allOf = nil
	attribute.oneOf = nil
}

// MakeArray converts this attribute to an array
func (attribute *Attribute) MakeArray(arrayItems *Attribute) {
	arrayItems.parent = attribute
	attribute.attributes = nil
	attribute.reference = ""
	attribute.arrayItems = arrayItems
	attribute.allOf = nil
	attribute.oneOf = nil
}

// MakeComposition converts this attribute to a composition
func (attribute *Attribute) MakeComposition(items []*Attribute) {
	attribute.attributes = nil
	attribute.reference = ""
	attribute.arrayItems = nil
	attribute.allOf = items
	for _, x := range items {
		x.parent = attribute
	}
	attribute.oneOf = nil
}

// MakePolymorphic converts this attribute to a polymorphic attribute
func (attribute *Attribute) MakePolymorphic(items []*Attribute) {
	attribute.attributes = nil
	attribute.reference = ""
	attribute.arrayItems = nil
	attribute.allOf = nil
	attribute.oneOf = items
	for _, x := range items {
		x.parent = attribute
	}
}

// UnmarshalExpanded unmarshals an attribute. The input is a
// map[string]interface{}. The attribute may or may not have an ID
func (attribute *Attribute) UnmarshalExpanded(in interface{}) error {
	m, ok := in.(map[string]interface{})
	if !ok {
		return ErrInvalidInput
	}
	attribute.ID = ""
	attribute.attributes = nil
	attribute.reference = ""
	attribute.arrayItems = nil
	attribute.allOf = nil
	attribute.oneOf = nil
	attribute.Values = make(map[string]interface{}, len(m))
	elems := 0
	for k, v := range m {
		switch k {
		case "@id":
			attribute.ID = v.(string)
		case AttributeStructure.Attributes.ID:
			attribute.attributes = &Attributes{parent: attribute}
			if err := attribute.attributes.UnmarshalExpanded(v); err != nil {
				return err
			}
		case AttributeStructure.Reference.ID:
			elems++
			arr, ok := v.([]interface{})
			if !ok || len(arr) > 1 {
				return ErrInvalidInput
			}
			if len(arr) == 1 {
				attribute.reference = GetNodeValue(arr[0]).(string)
			}

		case AttributeStructure.ArrayItems.ID:
			elems++
			arr, ok := v.([]interface{})
			if !ok || len(arr) > 1 {
				return ErrInvalidInput
			}
			if len(arr) == 1 {
				attribute.arrayItems = &Attribute{parent: attribute}
				if err := attribute.arrayItems.UnmarshalExpanded(arr[0]); err != nil {
					return err
				}
			}

		case AttributeStructure.AllOf.ID, AttributeStructure.OneOf.ID:
			elems++
			arr, ok := v.([]interface{})
			if !ok || len(arr) != 1 {
				return ErrInvalidInput
			}
			m, ok := arr[0].(map[string]interface{})
			if !ok {
				return ErrInvalidInput
			}
			arr, ok = m["@list"].([]interface{})
			if !ok {
				return ErrInvalidInput
			}
			attrs := make([]*Attribute, len(arr))
			for i, x := range arr {
				attrs[i] = &Attribute{parent: attribute}
				if err := attrs[i].UnmarshalExpanded(x); err != nil {
					return err
				}
			}
			if k == AttributeStructure.AllOf.ID {
				attribute.allOf = attrs
			} else {
				attribute.oneOf = attrs
			}
		default:
			attribute.Values[k] = v
		}
	}
	if elems > 1 {
		return ErrInvalidInput
	}
	return nil
}

// MarshalExpanded marshals the attribute as an expanded JSON-LD document
func (attribute *Attribute) MarshalExpanded() map[string]interface{} {
	ret := make(map[string]interface{})

	if len(attribute.ID) != 0 {
		ret["@id"] = attribute.ID
	}
	if attribute.attributes != nil {
		ret[AttributeStructure.Attributes.ID] = attribute.attributes.MarshalExpanded()
	}
	if len(attribute.reference) != 0 {
		ret[AttributeStructure.Reference.ID] = []interface{}{map[string]interface{}{"@value": attribute.reference}}
	}
	if attribute.arrayItems != nil {
		ret[AttributeStructure.ArrayItems.ID] = []interface{}{attribute.arrayItems.MarshalExpanded()}
	}
	if attribute.allOf != nil {
		arr := make([]interface{}, len(attribute.allOf))
		for i, item := range attribute.allOf {
			arr[i] = item.MarshalExpanded()
		}
		ret[AttributeStructure.AllOf.ID] = []interface{}{map[string]interface{}{"@list": arr}}
	}
	if attribute.oneOf != nil {
		arr := make([]interface{}, len(attribute.oneOf))
		for i, item := range attribute.oneOf {
			arr[i] = item.MarshalExpanded()
		}
		ret[AttributeStructure.OneOf.ID] = []interface{}{map[string]interface{}{"@list": arr}}
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
	if attribute.attributes != nil {
		if !attribute.attributes.Iterate(f) {
			return false
		}
	}
	if attribute.arrayItems != nil {
		if !attribute.arrayItems.Iterate(f) {
			return false
		}
	}
	for _, x := range attribute.allOf {
		if !x.Iterate(f) {
			return false
		}
	}
	for _, x := range attribute.oneOf {
		if !x.Iterate(f) {
			return false
		}
	}
	return true
}
