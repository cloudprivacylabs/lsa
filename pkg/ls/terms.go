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

const LS = "https://lschema.org/"

var (
	// SchemaTerm is the layer type for schemas
	SchemaTerm = NewTerm(LS, "Schema").SetComposition(NoComposition).SetTags(SchemaElementTag).Register()

	// OverlayTerm is the layer type for overlays
	OverlayTerm = NewTerm(LS, "Overlay").SetComposition(NoComposition).SetTags(SchemaElementTag).Register()

	// ComposeTerm is used for overlays to redefine term compositions. One of CompositionType constants
	ComposeTerm = RegisterStringTerm(NewTerm(LS, "compose").SetComposition(OverrideComposition).SetTags(SchemaElementTag))

	// AttributeOverlaysTerm lists the overlays for schema attributes
	// that are matched by ID, as opposed to matching by ID and their
	// place in the layer
	AttributeOverlaysTerm = NewTerm(LS, "attributeOverlays").SetList(true).SetComposition(OverrideComposition).SetTags(SchemaElementTag).Register()

	// CharacterEncodingTerm is used to specify a character encoding for
	// the data processed with the layer
	CharacterEncodingTerm = RegisterStringTerm(NewTerm(LS, "characterEncoding").SetComposition(OverrideComposition).SetTags(SchemaElementTag))

	// DescriptionTerm is used for comments/descriptions
	DescriptionTerm = NewTerm(LS, "description").SetComposition(SetComposition).SetTags(SchemaElementTag).Register()

	// AttributeNameTerm represents the name of an attribute
	AttributeNameTerm = RegisterStringTerm(NewTerm(LS, "attributeName").SetComposition(OverrideComposition).SetTags(SchemaElementTag))

	// AttributeIndexTerm represents the index of an array element
	AttributeIndexTerm = RegisterIntegerTerm(NewTerm(LS, "attributeIndex").SetComposition(NoComposition).SetTags(SchemaElementTag))

	// ConditionalTerm specifies conditions for ingestion
	ConditionalTerm = RegisterStringTerm(NewTerm(LS, "conditional").SetComposition(OverrideComposition).SetTags(SchemaElementTag))

	// LayerRootTerm is an edge term that connects layer node to the root node of the schema
	LayerRootTerm = NewTerm(LS, "layer").SetComposition(ErrorComposition).SetTags(SchemaElementTag).Register()

	// DefaultValueTerm is the default value for an attribute if attribute is not present
	DefaultValueTerm = NewTerm(LS, "defaultValue").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Register()

	// Format specifies a type-specific formatting directive, such as a date format
	FormatTerm = RegisterStringTerm(NewTerm(LS, "format").SetComposition(OverrideComposition).SetTags(SchemaElementTag))

	// NodeIDTerm keeps the node ID or the attribute ID
	NodeIDTerm = RegisterStringTerm(NewTerm(LS, "nodeId").SetComposition(NoComposition).SetTags(SchemaElementTag))

	// IngestAsTerm ingests value as an edge, node, or property
	IngestAsTerm = RegisterStringTerm(NewTerm(LS, "ingestAs").SetComposition(OverrideComposition).SetTags(SchemaElementTag))

	// AsPropertyOfTerm is optional. If specified, it gives the nearest
	// ancestor node that is an instance of the given type. If not, it
	// is the parent document node
	AsPropertyOfTerm = RegisterStringTerm(NewTerm(LS, "asPropertyOf").SetComposition(OverrideComposition).SetTags(SchemaElementTag))

	// EdgeLabelTerm represents the value used as an edge label, when ingesting an edge
	EdgeLabelTerm = RegisterStringTerm(NewTerm(LS, "edgeLabel").SetComposition(OverrideComposition).SetTags(SchemaElementTag))

	// OutputEdgeLabelTerm determines the labels of the output edges
	OutputEdgeLabelTerm = RegisterStringTerm(NewTerm(LS, "outputEdgeLabel").SetComposition(OverrideComposition).SetTags(SchemaElementTag))

	// PropertyNameTerm represents the value used as a property name when ingesting a property
	PropertyNameTerm = RegisterStringTerm(NewTerm(LS, "propertyName").SetComposition(OverrideComposition).SetTags(SchemaElementTag))

	// DocumentNodeTerm is the type of document nodes
	DocumentNodeTerm = NewTerm(LS, "DocumentNode").SetComposition(ErrorComposition).Register()

	// NodeValueTerm is the property key used to keep node value
	NodeValueTerm = NewTerm(LS, "value").SetComposition(ErrorComposition).Register()

	// ValueTypeTerm defines the type of a value
	ValueTypeTerm = NewTerm(LS, "valueType").SetComposition(OverrideComposition).Register()

	// HasTerm is an edge term for linking document elements
	HasTerm = NewTerm(LS, "has").SetComposition(ErrorComposition).Register()

	// LabeledAsTerm adds labels to the node it contained in
	LabeledAsTerm = RegisterStringSliceTerm(NewTerm(LS, "labeledAs").SetComposition(OverrideComposition).SetTags(SchemaElementTag))

	// TypeDiscriminatorTerm represents a set of schema field hints for defining polymorphic objects
	TypeDiscriminatorTerm = NewTerm(LS, "typeDiscriminator").SetComposition(NoComposition).SetTags(SchemaElementTag).Register()

	// IncludeSchemaTerm represents another schema to replace and copy its contents
	IncludeSchemaTerm = RegisterStringTerm(NewTerm(LS, "include").SetComposition(OverrideComposition).SetTags(SchemaElementTag))

	// Namespace defines the namespace prefix
	NamespaceTerm = RegisterStringTerm(NewTerm(LS, "namespace").SetComposition(OverrideComposition).SetTags(SchemaElementTag))

	// SourceTerm gives the source information for the data element
	SourceTerm = NewTerm(LS, "provenance/source").SetComposition(OverrideComposition).SetTags(ProvenanceTag).Register()
)

// Attribute types defines the terms describing attribute types. Each
// attribute must have one of the attribute types plus the Attribute
// type, marking the object as an attribute.
var (
	AttributeTypeValue       = NewTerm(LS, "Value").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Register()
	AttributeTypeObject      = NewTerm(LS, "Object").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Register()
	AttributeTypeArray       = NewTerm(LS, "Array").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Register()
	AttributeTypeReference   = NewTerm(LS, "Reference").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Register()
	AttributeTypeComposite   = NewTerm(LS, "Composite").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Register()
	AttributeTypePolymorphic = NewTerm(LS, "Polymorphic").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Register()
	AttributeNodeTerm        = NewTerm(LS, "Attribute").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Register()
)

// Layer terms includes type specific terms recognized by the schema
// compiler. These are terms used to define elements of an attribute.
var (
	// Unordered named attributes (json object)
	ObjectAttributesTerm = NewTerm(LS, "Object/attributes").SetComposition(ErrorComposition).SetTags(SchemaElementTag).Register()
	// Ordered named attributes (json object, xml elements)
	ObjectAttributeListTerm = NewTerm(LS, "Object/attributeList").SetComposition(ErrorComposition).SetTags(SchemaElementTag).Register()
	// Reference to another schema. This will be resolved to another
	// schema during compilation
	ReferenceTerm = RegisterStringTerm(NewTerm(LS, "Reference/ref").SetComposition(OverrideComposition).SetTags(SchemaElementTag))
	// ArrayItems contains the definition for the items of the array
	ArrayItemsTerm = NewTerm(LS, "Array/elements").SetComposition(ErrorComposition).SetTags(SchemaElementTag).Register()
	// All components of a composite attribute
	AllOfTerm = NewTerm(LS, "Composite/allOf").SetComposition(ErrorComposition).SetTags(SchemaElementTag).Register()
	// All options of a polymorphic attribute
	OneOfTerm = NewTerm(LS, "Polymorphic/oneOf").SetComposition(ErrorComposition).SetTags(SchemaElementTag).Register()
)

// IsAttributeType returns true if the term is one of the attribute types
func IsAttributeType(typeName string) bool {
	return typeName == AttributeTypeValue.Name ||
		typeName == AttributeNodeTerm.Name ||
		typeName == AttributeTypeObject.Name ||
		typeName == AttributeTypeArray.Name ||
		typeName == AttributeTypeReference.Name ||
		typeName == AttributeTypeComposite.Name ||
		typeName == AttributeTypePolymorphic.Name
}

// FilterAttributeTypes returns all recognized attribute types from
// the given types array. This is mainly used for validation, to
// ensure there is only one attribute type
func FilterAttributeTypes(types []string) []string {
	ret := make([]string, 0, len(types))
	for _, x := range types {
		if x == AttributeTypeValue.Name ||
			x == AttributeTypeObject.Name ||
			x == AttributeTypeArray.Name ||
			x == AttributeTypeReference.Name ||
			x == AttributeTypeComposite.Name ||
			x == AttributeTypePolymorphic.Name {
			ret = append(ret, x)
		}
	}
	return ret
}

// FilterNonLayerTypes returns the types that are not attribute or
// layer related
func FilterNonLayerTypes(types []string) []string {
	ret := make([]string, 0, len(types))
	for _, x := range types {
		if x != AttributeTypeValue.Name &&
			x != AttributeTypeObject.Name &&
			x != AttributeTypeArray.Name &&
			x != AttributeTypeReference.Name &&
			x != AttributeTypeComposite.Name &&
			x != AttributeTypePolymorphic.Name &&
			x != AttributeNodeTerm.Name {
			ret = append(ret, x)
		}
	}
	return ret
}
