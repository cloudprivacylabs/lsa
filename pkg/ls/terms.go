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

type TermType int

const (
	TermTypeUnknown TermType = iota
	TermTypeValue
	TermTypeID
	TermTypeList
	TermTypeIDList
	TermTypeSet
	TermTypeIDSet
)

// Term contains the URI for the term and term metadata
type Term struct {
	ID   string
	Type TermType

	// Compose function composes two term values and returns the compose
	// value. t1 and t2 are schema values for the term.
	Compose func(options ComposeOptions, t1, t2 interface{}) (interface{}, error)

	// Validate gets the expanded value for the term in the schema (most
	// likely a []interface{}), and the corresponding expanded value in
	// the document, and returns error if validation fails
	Validate func(schemaTermValue interface{}, docValue interface{}) error
}

const uriBase = "http://schemas.cloudprivacylabs.com"

// TermSchemaType is the object type for a schema
const TermSchemaType = uriBase + "/Schema"

// TermLayerType is the object type for layers
const TermLayerType = uriBase + "/Layer"

// AttributeStructure defines the terms used in defining the structure of
// attributes in schema layers
var AttributeStructure = struct {
	Attributes Term
	Reference  Term
	ArrayItems Term
	AllOf      Term
	OneOf      Term
}{
	// Attributes is an id map containing the attribute ID and
	// annotations. Each attribute can be one of:
	//
	//   * Reference: the attribute must have a `reference` term
	//   * Array: the attribute must have an `arrayItems` term
	//   * Composition: the attribute must have an `allOf` term
	//   * Polymorphism: the attribute must have a `oneOf` term
	//   * Value: If the attribute has none of the above, then it is a
	//     value.
	Attributes: Term{
		ID: uriBase + "/attributes",
	},

	// Reference is an IRI that points to another object
	Reference: Term{
		ID:      uriBase + "/attribute/reference",
		Type:    TermTypeID,
		Compose: OverrideComposeTerm,
	},

	// ArrayItems defines the items of an array object. ArrayItems is an
	// attribute that can contain all attribute related terms
	ArrayItems: Term{
		ID: uriBase + "/attribute/arrayItems",
	},

	// AllOf is a list that denotes composition. The resulting object is
	// a composition of the elements of the list
	AllOf: Term{
		ID:   uriBase + "/attribute/allOf",
		Type: TermTypeList,
	},

	// OneOf is a list that denotes polymorphism. The resulting object
	// can be one of the objects listed.
	OneOf: Term{
		ID:   uriBase + "/attribute/oneOf",
		Type: TermTypeList,
	},
}

// AttributeAnnotations includes the terms used to annotation attributes
var AttributeAnnotations = struct {
	Name        Term
	Privacy     Term
	Information Term
	Encoding    Term
	Type        Term
	Format      Term
	Pattern     Term
	Label       Term
	Enumeration Term
	Required    Term
}{
	// Name defines the name of the attribue at the ingestion
	// stage or at the output stage. This can be a column name for
	// tabular data, JSON key, or XML name
	Name: Term{
		ID:      uriBase + "/attribute/name",
		Type:    TermTypeValue,
		Compose: OverrideComposeTerm,
	},
	Privacy: Term{
		ID:   uriBase + "/attribute/privacyClassification",
		Type: TermTypeSet,
	},
	Information: Term{
		ID:   uriBase + "/attribute/information",
		Type: TermTypeValue,
	},
	Encoding: Term{
		ID:      uriBase + "/attribute/encoding",
		Type:    TermTypeValue,
		Compose: OverrideComposeTerm,
	},
	Type: Term{
		ID:      uriBase + "/attribute/type",
		Type:    TermTypeValue,
		Compose: OverrideComposeTerm,
	},
	Format: Term{
		ID:      uriBase + "/attribute/format",
		Type:    TermTypeValue,
		Compose: OverrideComposeTerm,
	},
	Pattern: Term{
		ID:      uriBase + "/attribute/pattern",
		Type:    TermTypeValue,
		Compose: OverrideComposeTerm,
	},
	Label: Term{
		ID:      uriBase + "/attribute/label",
		Type:    TermTypeValue,
		Compose: OverrideComposeTerm,
	},
	Enumeration: Term{
		ID:      uriBase + "/attribute/enumeration",
		Type:    TermTypeValue,
		Compose: OverrideComposeTerm,
		Validate: func(schemaTermValue, docValue interface{}) error {
			if docValue == nil {
				return nil
			}
			for _, el := range GetListElements(schemaTermValue) {
				value := GetNodeValue(el)
				if docValue == value {
					return nil
				}
			}
			return ErrValidation("Value not allowed by enumeration constraints")
		},
	},
	Required: Term{
		ID:      uriBase + "/attribute/required",
		Type:    TermTypeValue,
		Compose: OverrideComposeTerm,
		Validate: func(schemaTermValue, docValue interface{}) error {
			if arr, _ := schemaTermValue.([]interface{}); len(arr) == 1 {
				if required, _ := GetNodeValue(arr[0]).(bool); required {
					if docValue == nil {
						return ErrValidation("Required value missing")
					}
				}
			}
			return nil
		},
	},
}

var SchemaTerms = struct {
	IssuedBy       Term
	IssuerRole     Term
	IssuedAt       Term
	Purpose        Term
	Classification Term
	ObjectType     Term
	ObjectVersion  Term
	Layers         Term
}{
	IssuedBy: Term{
		ID:   uriBase + "/Schema/issuedBy",
		Type: TermTypeValue,
	},
	IssuerRole: Term{
		ID:   uriBase + "/Schema/issuerRole",
		Type: TermTypeValue,
	},
	IssuedAt: Term{
		ID:   uriBase + "/Schema/issuedAt",
		Type: TermTypeValue,
	},
	Purpose: Term{
		ID:   uriBase + "/Schema/purpose",
		Type: TermTypeValue,
	},
	Classification: Term{
		ID:   uriBase + "/Schema/classification",
		Type: TermTypeValue,
	},
	ObjectType: Term{
		ID:   uriBase + "/Schema/objectType",
		Type: TermTypeValue,
	},
	ObjectVersion: Term{
		ID:   uriBase + "/Schema/objectVersion",
		Type: TermTypeValue,
	},
	Layers: Term{
		ID:   uriBase + "/Schema/layers",
		Type: TermTypeIDList,
	},
}

var DocTerms = struct {
	// Value of the attribute from the input document
	Value Term
	// Attributes, for the ingested document
	Attributes Term
	// ArrayElements, for the ingested document
	ArrayElements Term
	// A reference to the schema attribute
	SchemaAttributeID Term
}{
	Value: Term{
		ID:   uriBase + "/doc/value",
		Type: TermTypeValue,
	},
	Attributes: Term{
		ID: uriBase + "/doc/attributes",
	},
	ArrayElements: Term{
		ID: uriBase + "/doc/arrayElements",
	},
	SchemaAttributeID: Term{
		ID: uriBase + "/doc/attributeId",
	},
}

// Terms contains the registered terms
var Terms = map[string]*Term{
	AttributeStructure.Attributes.ID: &AttributeStructure.Attributes,
	AttributeStructure.Reference.ID:  &AttributeStructure.Reference,
	AttributeStructure.ArrayItems.ID: &AttributeStructure.ArrayItems,
	AttributeStructure.AllOf.ID:      &AttributeStructure.AllOf,
	AttributeStructure.OneOf.ID:      &AttributeStructure.OneOf,

	AttributeAnnotations.Name.ID:        &AttributeAnnotations.Name,
	AttributeAnnotations.Privacy.ID:     &AttributeAnnotations.Privacy,
	AttributeAnnotations.Information.ID: &AttributeAnnotations.Information,
	AttributeAnnotations.Encoding.ID:    &AttributeAnnotations.Encoding,
	AttributeAnnotations.Type.ID:        &AttributeAnnotations.Type,
	AttributeAnnotations.Format.ID:      &AttributeAnnotations.Format,
	AttributeAnnotations.Pattern.ID:     &AttributeAnnotations.Pattern,
	AttributeAnnotations.Label.ID:       &AttributeAnnotations.Label,
	AttributeAnnotations.Enumeration.ID: &AttributeAnnotations.Enumeration,
	AttributeAnnotations.Required.ID:    &AttributeAnnotations.Required,

	SchemaTerms.IssuedBy.ID:       &SchemaTerms.IssuedBy,
	SchemaTerms.IssuerRole.ID:     &SchemaTerms.IssuerRole,
	SchemaTerms.IssuedAt.ID:       &SchemaTerms.IssuedAt,
	SchemaTerms.Purpose.ID:        &SchemaTerms.Purpose,
	SchemaTerms.Classification.ID: &SchemaTerms.Classification,
	SchemaTerms.ObjectType.ID:     &SchemaTerms.ObjectType,
	SchemaTerms.ObjectVersion.ID:  &SchemaTerms.ObjectVersion,
	SchemaTerms.Layers.ID:         &SchemaTerms.Layers,
}
