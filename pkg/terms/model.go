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

// ValueType specifies the type of values stored under a term
type ValueType string

const (
	// ValueTerm is a term that contains @value elements
	ValueTermType ValueType = "@value"
	// IDTerm is a term that contains @id elements
	IDTermType ValueType = "@id"
	// ObjectTerm is a term that contains other terms
	ObjectTermType ValueType = "object"
)

// ContainerType specifies the container characteristics of a term
type ContainerType string

const (
	// The term is a @set
	SetTermType ContainerType = "@set"
	// The term is a @list
	ListTermType ContainerType = "@list"
	// The term accepts only single values (in json-ld terms, it is a
	// @set with cardinality 1)
	MonadicTermType ContainerType = "monadic"
)

// Term interface provides access to term ID and metadata
type Term interface {
	GetTerm() string
	GetValueType() ValueType
	GetContainerType() ContainerType
}

// Composable term implements the compose function that defines how
// the values in a document can be composed
type Composable interface {
	Compose(t1, t2 interface{}) (interface{}, error)
}

// A Validator term implements value validator
type Validator interface {
	// Validate gets the expanded value for the term in the schema (most
	// likely a []interface{}), and the corresponding expanded value in
	// the document, and returns error if validation fails
	Validate(schemaTermValue interface{}, docValue interface{}) error
}

type MonadicValueTerm interface {
	Term

	// FromExpanded returns the interface value of the term from an
	// expanded value. The input must be an []interface{} with one
	// element
	FromExpanded(in interface{}) interface{}

	// Makes an expanded  value
	MakeExpanded(interface{}) []interface{}

	// PutExpanded sets the value of the term in the given map
	PutExpanded(map[string]interface{}, interface{})
}

// StringTerm is a term that has a single string value
type StringTerm interface {
	MonadicValueTerm

	// StringFromExpanded returns the string value of the term from an
	// expanded value. The input must be an []interface{} with one
	// element
	StringFromExpanded(in interface{}) string
}

// ContainerTerm is a term whose values are either a @set or @list
type ContainerTerm interface {
	Term
	ElementsFromExpanded(interface{}) []interface{}
	MakeExpandedContainer([]interface{}) interface{}
}

// StringContinerTerm is a term whose values are @set or @list of @id or @values
type StringContainerTerm interface {
	ContainerTerm
	ElementValuesFromExpanded(interface{}) []string
	MakeExpandedElement(string) interface{}
	MakeExpandedContainerFromValues([]string) interface{}
}

// TermModel defines the metadata for a term. This metadata controls
// how composition and validation for terms are done.
type TermModel struct {

	// The IRI for the term
	Term string `json:"term" yaml:"term"`

	// Value type for the term, @id, @value, or object
	Type ValueType `json:"valueType" yaml:"valueType"`

	// The type of the container for the term, @list, @set, or singleton
	Container ContainerType `json:"containerType" yaml:"containerType"`
}

func (t TermModel) MakeTerm() Term {
	switch t.Type {
	case ValueTermType:
		switch t.Container {
		case SetTermType:
			return ValueSetTerm(t.Term)
		case ListTermType:
			return ValueListTerm(t.Term)
		case MonadicTermType:
			return ValueTerm(t.Term)
		}
	case IDTermType:
		switch t.Container {
		case SetTermType:
			return IDSetTerm(t.Term)
		case ListTermType:
			return IDListTerm(t.Term)
		case MonadicTermType:
			return IDTerm(t.Term)
		}
	case ObjectTermType:
		switch t.Container {
		case SetTermType:
			return ObjectSetTerm(t.Term)
		case ListTermType:
			return ObjectListTerm(t.Term)
		case MonadicTermType:
			return ObjectTerm(t.Term)
		}
	}
	panic("Invalid term:" + t.Term)
}
