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
package terms

import (
	"github.com/piprate/json-gold/ld"
)

// IDTerm is a term whose value is an @id
type IDTerm string

func (t IDTerm) GetTerm() string                 { return string(t) }
func (t IDTerm) GetValueType() ValueType         { return IDTermType }
func (t IDTerm) GetContainerType() ContainerType { return MonadicTermType }

// Compose the term values. This implements the default composition
// where t2 overrides the value of t1. The returned value is a copy of t1 or t2.
func (t IDTerm) Compose(t1, t2 interface{}) (interface{}, error) {
	if t2 == nil {
		return ld.CloneDocument(t1), nil
	}
	return ld.CloneDocument(t2), nil
}

// FromExpanded returns the  value from an expanded value. The
// input must be []interface{}
func (t IDTerm) FromExpanded(in interface{}) interface{} {
	if arr, ok := in.([]interface{}); ok {
		if len(arr) == 1 {
			if m, ok := arr[0].(map[string]interface{}); ok {
				return m["@id"]
			}
		}
	}
	return nil
}

// StringFromExpanded returns the string value from an expanded value. The input must
// be []interface{}
func (t IDTerm) StringFromExpanded(in interface{}) string {
	return GetKeyValueFromExpanded("@id", in)
}

// MakeExpanded creates a new expanded value from the given value. The
// retuned object is
//
//  [ { @id: ID } ]
func (t IDTerm) MakeExpanded(ID interface{}) []interface{} {
	return []interface{}{map[string]interface{}{"@id": ID}}
}

// PutExpanded sets dest[t] to ID in expanded form
func (t IDTerm) PutExpanded(dest map[string]interface{}, ID interface{}) {
	dest[string(t)] = t.MakeExpanded(ID)
}

// GetExpandedString gets the term from the map
func (t IDTerm) GetExpandedString(src map[string]interface{}) string {
	return t.StringFromExpanded(src[string(t)])
}

// ValueTerm is a term whose value is a @value
type ValueTerm string

func (t ValueTerm) GetTerm() string                 { return string(t) }
func (t ValueTerm) GetValueType() ValueType         { return ValueTermType }
func (t ValueTerm) GetContainerType() ContainerType { return MonadicTermType }

// Compose the term values. This implements the default composition
// where t2 overrides the value of t1. The returned value is a copy of t1 or t2.
func (t ValueTerm) Compose(t1, t2 interface{}) (interface{}, error) {
	if t2 == nil {
		return ld.CloneDocument(t1), nil
	}
	return ld.CloneDocument(t2), nil
}

// StringFromExpanded returns the string value from an expanded value. The
// input must be []interface{}
func (t ValueTerm) StringFromExpanded(in interface{}) string {
	return GetKeyValueFromExpanded("@value", in)
}

// FromExpanded returns the  value from an expanded value. The
// input must be []interface{}
func (t ValueTerm) FromExpanded(in interface{}) interface{} {
	if arr, ok := in.([]interface{}); ok {
		if len(arr) == 1 {
			return arr[0]
		}
	}
	return nil
}

// MakeExpanded creates a new expanded value from the given value. The returned
// object is
//
//  [ { @value: value} ]
func (t ValueTerm) MakeExpanded(value interface{}) []interface{} {
	return []interface{}{map[string]interface{}{"@value": value}}
}

// PutExpanded sets dest[t] to vakue in expanded form
func (t ValueTerm) PutExpanded(dest map[string]interface{}, value interface{}) {
	dest[string(t)] = t.MakeExpanded(value)
}

// GetExpandedString gets the string term value from the map
func (t ValueTerm) GetExpandedString(src map[string]interface{}) string {
	return t.StringFromExpanded(src[string(t)])
}

// An object term is a term whose value is not a simple value, but can
// be an array or object
type ObjectTerm string

func (t ObjectTerm) GetTerm() string                 { return string(t) }
func (t ObjectTerm) GetValueType() ValueType         { return ObjectTermType }
func (t ObjectTerm) GetContainerType() ContainerType { return MonadicTermType }

// Compose the term values. This implements the default composition
// where t2 overrides the value of t1. The returned object is a copy of t1 or t2.
func (t ObjectTerm) Compose(t1, t2 interface{}) (interface{}, error) {
	if t2 == nil {
		return ld.CloneDocument(t1), nil
	}
	return ld.CloneDocument(t2), nil
}

// FromExpanded returns the first element of in if in is an array of
// 1.
func (t ObjectTerm) FromExpanded(in interface{}) interface{} {
	if arr, ok := in.([]interface{}); ok {
		if len(arr) == 1 {
			return arr[0]
		}
	}
	return nil
}

// MakeExpanded returns [obj]
func (t ObjectTerm) MakeExpanded(obj interface{}) []interface{} {
	return []interface{}{obj}
}

// PutExpanded sets dest[t] to [obj]
func (t ObjectTerm) PutExpanded(dest map[string]interface{}, obj interface{}) {
	dest[string(t)] = t.MakeExpanded(obj)
}

// IDListTerm is a term of the form
//
// [
//   { "@list": [
//     { "@id": value },
//      ...
//   ]}
// ]
type IDListTerm string

func (t IDListTerm) GetTerm() string                 { return string(t) }
func (t IDListTerm) GetValueType() ValueType         { return IDTermType }
func (t IDListTerm) GetContainerType() ContainerType { return ListTermType }

// MakeExpandedContainer returns an expanded list from the given elements
func (t IDListTerm) MakeExpandedContainer(elements []interface{}) interface{} {
	return []interface{}{map[string]interface{}{"@list": elements}}
}

// MakeExpandedContainerFromValues creates a new container from the given values
func (t IDListTerm) MakeExpandedContainerFromValues(ids []string) interface{} {
	ret := make([]interface{}, 0, len(ids))
	for _, x := range ids {
		ret = append(ret, t.MakeExpandedElement(x))
	}
	return t.MakeExpandedContainer(ret)
}

// MakeExpandedElement makes an element of the list
func (t IDListTerm) MakeExpandedElement(ID string) interface{} {
	return map[string]interface{}{"@id": ID}
}

// Compose the term values. The default composition for ID list is the
// concatenation of the two lists. The returned object is a new object
// containing copies of t1 and t2
func (t IDListTerm) Compose(t1, t2 interface{}) (interface{}, error) {
	return AppendLists(t, t1, t2), nil
}

// ElementsFromExpanded returns the elements list of the term
func (t IDListTerm) ElementsFromExpanded(in interface{}) []interface{} {
	return GetListElementsFromExpanded(in)
}

// ElementValuesFromExpanded returns the element values from an expanded object
//
//  [
//    {"@list": [ elements ] }
//  ]
func (t IDListTerm) ElementValuesFromExpanded(in interface{}) []string {
	return GetStringsFromArray("@id", t.ElementsFromExpanded(in))
}

// ValueListTerm is a term of the form
//
// [
//   { "@list": [
//     { "@id": value },
//      ...
//   ]}
// ]
type ValueListTerm string

func (t ValueListTerm) GetTerm() string                 { return string(t) }
func (t ValueListTerm) GetValueType() ValueType         { return ValueTermType }
func (t ValueListTerm) GetContainerType() ContainerType { return ListTermType }

// MakeExpandedContainerFromValues creates a new container from the given values
func (t ValueListTerm) MakeExpandedContainerFromValues(values []string) interface{} {
	ret := make([]interface{}, 0, len(values))
	for _, x := range values {
		ret = append(ret, t.MakeExpandedElement(x))
	}
	return t.MakeExpandedContainer(ret)
}

// MakeExpandedContainer makes a @list container from the given elements
func (t ValueListTerm) MakeExpandedContainer(elements []interface{}) interface{} {
	return []interface{}{map[string]interface{}{"@list": elements}}
}

// MakeExpandedElement makes an element of the list
func (t ValueListTerm) MakeExpandedElement(ID string) interface{} {
	return map[string]interface{}{"@value": ID}
}

// Compose the term values. The default composition for value list is the
// concatenation of the two lists. The returned object is a new object
// containing copies of t1 and t2
func (t ValueListTerm) Compose(t1, t2 interface{}) (interface{}, error) {
	return AppendLists(t, t1, t2), nil
}

// ElementsFromExpanded returns the elements list of the term
func (t ValueListTerm) ElementsFromExpanded(in interface{}) []interface{} {
	return GetListElementsFromExpanded(in)
}

// ElementValuesFromExpanded returns the element values from an expanded object
//
//  [
//    {"@list": [ elements ] }
//  ]
func (t ValueListTerm) ElementValuesFromExpanded(in interface{}) []string {
	return GetStringsFromArray("@value", t.ElementsFromExpanded(in))
}

// IDSetTerm is a set of @ids
type IDSetTerm string

func (t IDSetTerm) GetTerm() string                 { return string(t) }
func (t IDSetTerm) GetValueType() ValueType         { return IDTermType }
func (t IDSetTerm) GetContainerType() ContainerType { return SetTermType }

// MakeExpandedContainerFromValues creates a new container from the given values
func (t IDSetTerm) MakeExpandedContainerFromValues(ids []string) interface{} {
	ret := make([]interface{}, 0, len(ids))
	for _, x := range ids {
		ret = append(ret, t.MakeExpandedElement(x))
	}
	return t.MakeExpandedContainer(ret)
}

// MakeExpandedContainer makes a @set container from the given elements
func (t IDSetTerm) MakeExpandedContainer(elements []interface{}) interface{} {
	return elements
}

// MakeExpandedElement makes an element of the list
func (t IDSetTerm) MakeExpandedElement(ID string) interface{} {
	return map[string]interface{}{"@id": ID}
}

// Compose the term values. The default composition for id set is the
// union of the two lists. The returned object is a new object
// containing copies of t1 and t2
func (t IDSetTerm) Compose(t1, t2 interface{}) (interface{}, error) {
	return UnionStringSets(t, t1, t2), nil
}

// ElementsFromExpanded returns the elements list of the term
func (t IDSetTerm) ElementsFromExpanded(in interface{}) []interface{} {
	return GetSetElementsFromExpanded(in)
}

// ElementValuesFromExpanded returns the element values from an expanded object
//
//  [
//    {"@set": [ elements ] }
//  ]
//
// or
//
//  [ elements ]
func (t IDSetTerm) ElementValuesFromExpanded(in interface{}) []string {
	return GetStringsFromArray("@id", GetSetElementsFromExpanded(in))
}

// ValueSetTerm is a set of @values
type ValueSetTerm string

func (t ValueSetTerm) GetTerm() string                 { return string(t) }
func (t ValueSetTerm) GetValueType() ValueType         { return ValueTermType }
func (t ValueSetTerm) GetContainerType() ContainerType { return SetTermType }

// MakeExpandedContainerFromValues creates a new container from the given values
func (t ValueSetTerm) MakeExpandedContainerFromValues(values []string) interface{} {
	ret := make([]interface{}, 0, len(values))
	for _, x := range values {
		ret = append(ret, t.MakeExpandedElement(x))
	}
	return t.MakeExpandedContainer(ret)
}

// MakeExpandedContainer makes a @set container from the given elements
func (t ValueSetTerm) MakeExpandedContainer(elements []interface{}) interface{} {
	return elements
}

// MakeExpandedElement makes an element of the list
func (t ValueSetTerm) MakeExpandedElement(value string) interface{} {
	return map[string]interface{}{"@value": value}
}

// Compose the term values. The default composition for value set is the
// union of the two lists. The returned object is a new object
// containing copies of t1 and t2
func (t ValueSetTerm) Compose(t1, t2 interface{}) (interface{}, error) {
	return UnionStringSets(t, t1, t2), nil
}

// ElementsFromExpanded returns the elements list of the term
func (t ValueSetTerm) ElementsFromExpanded(in interface{}) []interface{} {
	return GetSetElementsFromExpanded(in)
}

// ElementValuesFromExpanded returns the element values from an expanded object
//
//  [
//    {"@set": [ elements ] }
//  ]
//
// or
//
//  [ elements ]
func (t ValueSetTerm) ElementValuesFromExpanded(in interface{}) []string {
	return GetStringsFromArray("@value", GetSetElementsFromExpanded(in))
}

// ObjectListTerm is a list of objects
type ObjectListTerm string

func (t ObjectListTerm) GetTerm() string                 { return string(t) }
func (t ObjectListTerm) GetValueType() ValueType         { return ObjectTermType }
func (t ObjectListTerm) GetContainerType() ContainerType { return ListTermType }

// MakeExpandedContainer makes a @list container from the given elements
func (t ObjectListTerm) MakeExpandedContainer(elements []interface{}) interface{} {
	return []interface{}{map[string]interface{}{"@list": elements}}
}

// ElementsFromExpanded returns the elements from an expanded list
func (t ObjectListTerm) ElementsFromExpanded(in interface{}) []interface{} {
	return GetListElementsFromExpanded(in)
}

// ObjectSetTerm is a set of objects
type ObjectSetTerm string

func (t ObjectSetTerm) GetTerm() string                 { return string(t) }
func (t ObjectSetTerm) GetValueType() ValueType         { return ObjectTermType }
func (t ObjectSetTerm) GetContainerType() ContainerType { return SetTermType }

// MakeExpandedContainer makes a @set container from the given elements
func (t ObjectSetTerm) MakeExpandedContainer(elements []interface{}) interface{} {
	return elements
}

// ElementsFromExpanded returns the elements from an expanded set
func (t ObjectSetTerm) ElementsFromExpanded(in interface{}) []interface{} {
	return GetSetElementsFromExpanded(in)
}
