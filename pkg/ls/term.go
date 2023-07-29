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
	"strconv"

	"github.com/cloudprivacylabs/lpg/v2"
)

// Term provides the semantics of a term within the layered schemas
// framework.
//
// To create and register a new Term, use
//
//   	SchemaTerm = NewTerm(LS, "Schema").SetComposition(NoComposition).SetTags(SchemaElementTag).Register()
//
// The following creates and registers a term with a specific value accessor
//
//   var EntitySchemaTerm = StringTerm{
//        Term: NewTerm(LS, "entitySchema").
//        SetType(StringType{}).
//        SetComposition(ErrorComposition).
//        SetTags(SchemaElementTag).
//       	Register(),
//   }
//
// This provides
//
//     sch:=EntitySchemaTerm.PropertyValue(node)

type Term struct {
	// The term
	Name string

	// Namespace of the term.
	Namespace string
	// Local name (excluding the namespace)
	LName string
	// Aliases of the term
	Aliases []string

	// If true, the term value is an @id (IRI). In JSON-LD, the values for
	// this term will be marshalled as @id
	IsID bool

	// If true, the term is a list. In JSON-LD, its elements will be
	// marshaled under @list
	IsList bool

	// Value type for the term.
	Type ValueType

	// Composition semantics of the term
	Composition CompositionType

	// Tags define additional metadata about a term
	Tags map[string]struct{}

	Metadata interface{}
}

// Known tags for term semantics
const (
	// SchemaElementTag means that the term is used for schema definitions only
	SchemaElementTag = "schemaElement"

	// ProvenanceTag means that the term is provenance related
	ProvenanceTag = "provenance"

	// ValidationTag means that the tag is about validation
	ValidationTag = "validation"
)

// NewTerm create a new term.
func NewTerm(ns, lname string, aliases ...string) Term {
	t := Term{Name: ns + lname,
		Namespace:   ns,
		LName:       lname,
		Aliases:     aliases,
		Composition: OverrideComposition,
		Type:        AnyType{},
		Tags:        make(map[string]struct{}),
	}
	return t
}

// Register a term and return its name
func (t Term) Register() Term {
	RegisterTerm(t)
	return t
}

func (t Term) SetID(v bool) Term {
	t.IsID = v
	return t
}

func (t Term) SetType(typ ValueType) Term {
	t.Type = typ
	return t
}

func (t Term) SetList(v bool) Term {
	t.IsList = v
	return t
}

func (t Term) SetComposition(comp CompositionType) Term {
	t.Composition = comp
	return t
}

func (t Term) SetAliases(aliases ...string) Term {
	t.Aliases = aliases
	return t
}

func (t Term) SetMetadata(md any) Term {
	t.Metadata = md
	return t
}

func (t Term) SetTags(tags ...string) Term {
	for _, tag := range tags {
		t.Tags[tag] = struct{}{}
	}
	return t
}

func (t Term) Compose(target, src PropertyValue) (PropertyValue, error) {
	return t.Composition.Compose(target, src)
}

// NewPropertyValue creates a new property value by coercing the value to the property value type
func (t Term) NewPropertyValue(value any) (PropertyValue, error) {
	v, err := t.Type.Coerce(value)
	if err != nil {
		return PropertyValue{}, err
	}
	return NewPropertyValue(t.Name, v), nil
}

// MustPropertyValue creates a new property value by coercing the
// value to the property value type, and panics if coercion fails
func (t Term) MustPropertyValue(value any) PropertyValue {
	pv, err := t.NewPropertyValue(value)
	if err != nil {
		panic(err)
	}
	return pv
}

var registeredTerms = map[string]Term{}

// If a term is known, using this function avoids duplicate string
// copies
func knownTerm(s string) string {
	x, ok := registeredTerms[s]
	if ok {
		return x.Name
	}
	return s
}

func RegisterTerm(t Term) {
	reg := func(s string) {
		_, ok := registeredTerms[s]
		if ok {
			panic("Duplicate term :" + t.Name)
		}
		registeredTerms[s] = t
	}
	reg(t.Name)
	for _, alias := range t.Aliases {
		reg(alias)
	}
}

func GetTerm(term string) Term {
	t, ok := registeredTerms[term]
	if !ok {
		return Term{
			Name:        term,
			Composition: SetComposition,
			Type:        AnyType{},
		}
	}
	return t
}

// GetTermMetadata returns metadata about a term
func GetTermMetadata(term string) interface{} {
	t := GetTerm(term)
	return t.Metadata
}

func IsTermRegistered(term string) bool {
	_, ok := registeredTerms[term]
	return ok
}

// SameTerm returns true if term1 is an alias of term2 or vice versa
func SameTerm(term1, term2 string) bool {
	if term1 == term2 {
		return true
	}
	s1, ok1 := registeredTerms[term1]
	s2, ok2 := registeredTerms[term2]
	if !ok1 && !ok2 {
		return false
	}
	return s1.Name == s2.Name
}

// ValueType represents the expected value type of a term. When
// putting values into the graph, you should "coerce" the values into
// the expected type.
type ValueType interface {
	// Coerce a value to the correct type
	Coerce(input any) (any, error)
}

type ErrInvalidIntegerValue struct {
	Value any
}

func (e ErrInvalidIntegerValue) Error() string {
	return fmt.Sprintf("Invalid integer value: %v", e.Value)
}

type ErrInvalidFloatValue struct {
	Value any
}

func (e ErrInvalidFloatValue) Error() string {
	return fmt.Sprintf("Invalid floating point value: %v", e.Value)
}

type ErrInvalidStringSliceValue struct {
	Value any
}

func (e ErrInvalidStringSliceValue) Error() string {
	return fmt.Sprintf("Invalid string slice value: %v", e.Value)
}

// AnyType is the default type that does not do any translation
type AnyType struct{}

func (AnyType) Coerce(input any) (any, error) { return input, nil }

type StringType struct{}

// Coerce an input value to string
func (StringType) Coerce(input any) (any, error) {
	if str, ok := input.(string); ok {
		return str, nil
	}
	if input == nil {
		return "", nil
	}
	return fmt.Sprint(input), nil
}

// StringTerm is a wrapper for string values. Use it to declare string terms:
//
//	var strTerm = StringTerm { Term:NewTerm(...) }.Register()
//
// Then you can use it as:
//
//	str:=strTerm.PropertyValue(node)
type StringTerm struct {
	Term
}

// PropertyValue returns the value of the property in the node or edge as a string
func (s StringTerm) PropertyValue(source lpg.WithProperties) string {
	sx, _ := GetPropertyValueAs[string](source, s.Name)
	return sx
}

type StringSliceType struct{}

// Coerce an input value to []string
func (StringSliceType) Coerce(input any) (any, error) {
	if arr, ok := input.([]string); ok {
		return arr, nil
	}
	if str, ok := input.(string); ok {
		return []string{str}, nil
	}
	if arr, ok := input.([]any); ok {
		ret := make([]string, 0, len(arr))
		for _, x := range arr {
			ret = append(ret, fmt.Sprint(x))
		}
		return ret, nil
	}
	return nil, ErrInvalidStringSliceValue{input}
}

// StringSliceTerm is a wrapper for []string values. Use it to declare []string terms:
//
//	var strSliceTerm = StringSliceTerm { Term:NewTerm(...) }.Register()
//
// Then you can use it as:
//
//	slice:=strSliceTerm.PropertyValue(node)
type StringSliceTerm struct {
	Term
}

// PropertyValue returns the value of the property in the node or edge as a []string
func (s StringSliceTerm) PropertyValue(source lpg.WithProperties) []string {
	pv, ok := GetPropertyValue(source, s.Name)
	if !ok {
		return []string{}
	}
	return pv.AsStringSlice()
}

type IntegerType struct{}

// Coerce an input value to int
func (IntegerType) Coerce(input any) (any, error) {
	if i, ok := input.(int); ok {
		return i, nil
	}
	if input == nil {
		return 0, nil
	}
	switch k := input.(type) {
	case int8:
		return int(k), nil
	case int16:
		return int(k), nil
	case int32:
		return int(k), nil
	case int64:
		return int(k), nil
	case uint8:
		return int(k), nil
	case uint16:
		return int(k), nil
	case uint32:
		return int(k), nil
	case uint64:
		return int(k), nil
	case uint:
		return int(k), nil
	case float64:
		return int(k), nil
	case float32:
		return int(k), nil
	case bool:
		if k {
			return 1, nil
		}
		return 0, nil
	case string:
		return strconv.Atoi(k)
	case json.Number:
		v, err := k.Int64()
		return int(v), err
	}
	return nil, ErrInvalidIntegerValue{input}
}

// IntegerTerm is a wrapper for int values. Use it to declare int terms:
//
//	var intTerm = IntegerTerm { Term:NewTerm(...) }.Register()
//
// Then you can use it as:
//
//	val:=intTerm.PropertyValue(node)
type IntegerTerm struct {
	Term
}

// PropertyValue returns the value of the property in the node or edge as an int
func (s IntegerTerm) PropertyValue(source lpg.WithProperties) int {
	pv, _ := GetPropertyValueAs[int](source, s.Name)
	return pv
}

type FloatType struct{}

// Coerce an input value to float64
func (FloatType) Coerce(input any) (any, error) {
	if i, ok := input.(float64); ok {
		return i, nil
	}
	if input == nil {
		return float64(0), nil
	}
	switch k := input.(type) {
	case int8:
		return float64(k), nil
	case int16:
		return float64(k), nil
	case int32:
		return float64(k), nil
	case int64:
		return float64(k), nil
	case int:
		return float64(k), nil
	case uint8:
		return float64(k), nil
	case uint16:
		return float64(k), nil
	case uint32:
		return float64(k), nil
	case uint64:
		return float64(k), nil
	case uint:
		return float64(k), nil
	case float32:
		return float64(k), nil
	case bool:
		if k {
			return 1.0, nil
		}
		return 0.0, nil
	case string:
		return strconv.ParseFloat(k, 64)
	case json.Number:
		v, err := k.Float64()
		return float64(v), err
	}
	return nil, ErrInvalidFloatValue{input}
}
