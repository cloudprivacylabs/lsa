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
	"github.com/cloudprivacylabs/lsa/pkg/terms"
)

// LS is the namespace for layeres schemas ontology
const LS = "http://layeredschemas.org/v1.0"

// EnumerationTerm is the term definition for enumeration with a custom validator
type EnumerationTerm struct {
	terms.ValueListTerm
}

// Validate checks if the doc value is one of the schema values
func (t EnumerationTerm) Validate(schemaTermValue, docValue interface{}) error {
	if docValue == nil {
		return nil
	}
	for _, el := range t.ElementValuesFromExpanded(schemaTermValue) {
		if docValue == el {
			return nil
		}
	}
	return ErrValidation("Value not allowed by enumeration constraints")
}

// RequiredTerm is the term definition for boolean required flag with a custom validator
type RequiredTerm struct {
	terms.ValueTerm
}

// Validate checks if the doc value exists or not
func (t RequiredTerm) Validate(schemaTermValue, docValue interface{}) error {
	if required, _ := t.FromExpanded(schemaTermValue).(bool); required {
		if docValue == nil {
			return ErrValidation("Required value missing")
		}
	}
	return nil
}

// AttributeAnnotations includes the terms used to annotation attributes
var AttributeAnnotations = struct {
	Name        terms.ValueTerm
	Privacy     terms.ValueSetTerm
	Information terms.ValueSetTerm
	Encoding    terms.ValueTerm
	Type        terms.ValueTerm
	Format      terms.ValueTerm
	Pattern     terms.ValueTerm
	Label       terms.ValueTerm
	Enumeration EnumerationTerm
	Required    RequiredTerm
}{
	// Name defines the name of the attribue at the ingestion
	// stage or at the output stage. This can be a column name for
	// tabular data, JSON key, or XML name
	Name:        terms.ValueTerm(LS + "/attr/name"),
	Privacy:     terms.ValueSetTerm(LS + "/attr/privacyClassification"),
	Information: terms.ValueSetTerm(LS + "/attr/information"),
	Encoding:    terms.ValueTerm(LS + "/attr/encoding"),
	Type:        terms.ValueTerm(LS + "/attr/type"),
	Format:      terms.ValueTerm(LS + "/attr/format"),
	Pattern:     terms.ValueTerm(LS + "/attr/pattern"),
	Label:       terms.ValueTerm(LS + "/attr/label"),
	Enumeration: EnumerationTerm{LS + "/attr/enumeration"},
	Required:    RequiredTerm{LS + "/attribute/required"},
}

var DocTerms = struct {
	// Value of the attribute from the input document
	Value terms.ValueTerm
	// Attributes, for the ingested document
	Attributes terms.ObjectListTerm
	// ArrayElements, for the ingested document
	ArrayElements terms.ObjectListTerm
	// A reference to the schema attribute
	SchemaAttributeID terms.IDTerm
}{
	Value:             terms.ValueTerm(LS + "/doc/value"),
	Attributes:        terms.ObjectListTerm(LS + "/doc/attributes"),
	ArrayElements:     terms.ObjectListTerm(LS + "/doc/arrayElements"),
	SchemaAttributeID: terms.IDTerm(LS + "/doc/attributeId"),
}

// Terms contains the registered terms
var Terms = terms.NewVocabulary(LayerTerms.Attributes,
	LayerTerms.Reference,
	LayerTerms.ArrayItems,
	LayerTerms.AllOf,
	LayerTerms.OneOf,

	AttributeAnnotations.Name,
	AttributeAnnotations.Privacy,
	AttributeAnnotations.Information,
	AttributeAnnotations.Encoding,
	AttributeAnnotations.Type,
	AttributeAnnotations.Format,
	AttributeAnnotations.Pattern,
	AttributeAnnotations.Label,
	AttributeAnnotations.Enumeration,
	AttributeAnnotations.Required,

	SchemaManifestTerms.PublishedAt,
	LayerTerms.ObjectType,
	LayerTerms.ObjectVersion,
	SchemaManifestTerms.Bundle,
	SchemaManifestTerms.Schema,
	SchemaManifestTerms.Overlays)
