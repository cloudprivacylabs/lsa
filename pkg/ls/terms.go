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
	SchemaTerm = NewTerm(LS, "Schema").SetComposition(NoComposition).SetTags(SchemaElementTag).Term

	// OverlayTerm is the layer type for overlays
	OverlayTerm = NewTerm(LS, "Overlay").SetComposition(NoComposition).SetTags(SchemaElementTag).Term

	// ComposeTerm is used for overlays to redefine term compositions. One of CompositionType constants
	ComposeTerm = NewTerm(LS, "compose").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term

	// AttributeOverlaysTerm lists the overlays for schema attributes
	// that are matched by ID, as opposed to matching by ID and their
	// place in the layer
	AttributeOverlaysTerm = NewTerm(LS, "attributeOverlays").SetList(true).SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term

	// NSMapTerm specifies a namespace map for an overlay. A Namespace
	// map includes one or more expressions of the form:
	//
	//    from -> to
	//
	// where from and to are attribute id prefixes. All the prefixes of
	// attributes that match from are converted to to.
	//
	// This is necessary when a different variants of a schema is used in
	// a complex schema. Each variant gets its own namespace.
	NSMapTerm = NewTerm(LS, "nsMap").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term

	// CharacterEncodingTerm is used to specify a character encoding for
	// the data processed with the layer
	CharacterEncodingTerm = NewTerm(LS, "characterEncoding").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term

	// InstanceOfTerm is an edge term that is used to connect values with
	// their schema specifications
	InstanceOfTerm = NewTerm(LS, "instanceOf").SetComposition(ErrorComposition).Term

	// SchemaNodeIDTerm denotes the schema node ID for ingested nodes
	SchemaNodeIDTerm = NewTerm(LS, "schemaNodeId").SetComposition(ErrorComposition).SetTags(SchemaElementTag).Term

	// SchemaVariantTerm is the schema variant type
	SchemaVariantTerm = NewTerm(LS, "SchemaVariant").SetComposition(NoComposition).SetTags(SchemaElementTag).Term

	// DescriptionTerm is used for comments/descriptions
	DescriptionTerm = NewTerm(LS, "description").SetComposition(SetComposition).SetTags(SchemaElementTag).Term

	// AttributeNameTerm represents the name of an attribute
	AttributeNameTerm = NewTerm(LS, "attributeName").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term

	// AttributeIndexTerm represents the index of an array element
	AttributeIndexTerm = NewTerm(LS, "attributeIndex").SetComposition(NoComposition).SetTags(SchemaElementTag).Term

	// ConditionalTerm specifies conditions for ingestion
	ConditionalTerm = NewTerm(LS, "conditional").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term

	// LayerRootTerm is an edge term that connects layer node to the root node of the schema
	LayerRootTerm = NewTerm(LS, "layer").SetComposition(ErrorComposition).SetTags(SchemaElementTag).Term

	// DefaultValueTerm is the default value for an attribute if attribute is not present
	DefaultValueTerm = NewTerm(LS, "defaultValue").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term

	// Format specifies a type-specific formatting directive, such as a date format
	FormatTerm = NewTerm(LS, "format").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term
	// EntitySchemaTerm is inserted by the schema compilation to mark
	// entity roots. It records the schema ID containing the entity
	// definition.
	EntitySchemaTerm = NewTerm(LS, "entitySchema").SetComposition(ErrorComposition).SetTags(SchemaElementTag).Term

	// NodeIDTerm keeps the node ID or the attribute ID
	NodeIDTerm = NewTerm(LS, "nodeId").SetComposition(ErrorComposition).SetTags(SchemaElementTag).Term

	// IngestAsTerm ingests value as an edge, node, or property
	IngestAsTerm = NewTerm(LS, "ingestAs").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term

	// AsPropertyOfTerm is optional. If specified, it gives the nearest
	// ancestor node that is an instance of the given type. If not, it
	// is the parent document node
	AsPropertyOfTerm = NewTerm(LS, "asPropertyOf").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term

	// EdgeLabelTerm represents the value used as an edge label, when ingesting an edge
	EdgeLabelTerm = NewTerm(LS, "edgeLabel").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term

	// OutputEdgeLabelTerm determines the labels of the output edges
	OutputEdgeLabelTerm = NewTerm(LS, "outputEdgeLabel").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term

	// PropertyNameTerm represents the value used as a property name when ingesting a property
	PropertyNameTerm = NewTerm(LS, "propertyName").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term

	// DocumentNodeTerm is the type of document nodes
	DocumentNodeTerm = NewTerm(LS, "DocumentNode").SetComposition(ErrorComposition).Term

	// NodeValueTerm is the property key used to keep node value
	NodeValueTerm = NewTerm(LS, "value").SetComposition(ErrorComposition).Term

	// ValueTypeTerm defines the type of a value
	ValueTypeTerm = NewTerm(LS, "valueType").SetComposition(OverrideComposition).Term

	// HasTerm is an edge term for linking document elements
	HasTerm = NewTerm(LS, "has").SetComposition(ErrorComposition).Term

	// EntityIDFieldsTerm is a string or []string that lists the attribute IDs
	// for entity ID. It is defined at the root node of a layer. All
	// attribute IDs must refer to value nodes.
	EntityIDFieldsTerm = NewTerm(LS, "entityIdFields").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term

	// EntityIDTerm is a string or []string that gives the unique ID of
	// an entity. This is a node property at the root node of an entity
	EntityIDTerm = NewTerm(LS, "entityId").SetComposition(OverrideComposition).Term

	// LabeledAsTerm adds labels to JSON schemas
	LabeledAsTerm = NewTerm(LS, "labeledAs").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term

	// TypeDiscriminatorTerm represents a set of schema field hints for defining polymorphic objects
	TypeDiscriminatorTerm = NewTerm(LS, "typeDiscriminator").SetComposition(NoComposition).SetTags(SchemaElementTag).Term

	// IncludeSchemaTerm represents another schema to replace and copy its contents
	IncludeSchemaTerm = NewTerm(LS, "include").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term

	// Namespace defines the namespace prefix
	NamespaceTerm = NewTerm(LS, "namespace").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term

	// SourceTerm gives the source information for the data element
	SourceTerm = NewTerm(LS, "provenance/source").SetComposition(OverrideComposition).SetTags(ProvenanceTag).Term
)

// Attribute types defines the terms describing attribute types. Each
// attribute must have one of the attribute types plus the Attribute
// type, marking the object as an attribute.
var (
	AttributeTypeValue       = NewTerm(LS, "Value").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term
	AttributeTypeObject      = NewTerm(LS, "Object").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term
	AttributeTypeArray       = NewTerm(LS, "Array").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term
	AttributeTypeReference   = NewTerm(LS, "Reference").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term
	AttributeTypeComposite   = NewTerm(LS, "Composite").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term
	AttributeTypePolymorphic = NewTerm(LS, "Polymorphic").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term
	AttributeNodeTerm        = NewTerm(LS, "Attribute").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term
)

// Layer terms includes type specific terms recognized by the schema
// compiler. These are terms used to define elements of an attribute.
var (
	// Unordered named attributes (json object)
	ObjectAttributesTerm = NewTerm(LS, "Object/attributes").SetComposition(ErrorComposition).SetTags(SchemaElementTag).Term
	// Ordered named attributes (json object, xml elements)
	ObjectAttributeListTerm = NewTerm(LS, "Object/attributeList").SetComposition(ErrorComposition).SetTags(SchemaElementTag).Term
	// Reference to another schema. This will be resolved to another
	// schema during compilation
	ReferenceTerm = NewTerm(LS, "Reference/ref").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Term
	// ArrayItems contains the definition for the items of the array
	ArrayItemsTerm = NewTerm(LS, "Array/elements").SetComposition(ErrorComposition).SetTags(SchemaElementTag).Term
	// All components of a composite attribute
	AllOfTerm = NewTerm(LS, "Composite/allOf").SetComposition(ErrorComposition).SetTags(SchemaElementTag).Term
	// All options of a polymorphic attribute
	OneOfTerm = NewTerm(LS, "Polymorphic/oneOf").SetComposition(ErrorComposition).SetTags(SchemaElementTag).Term
)

// IsAttributeType returns true if the term is one of the attribute types
func IsAttributeType(typeName string) bool {
	return typeName == AttributeTypeValue ||
		typeName == AttributeNodeTerm ||
		typeName == AttributeTypeObject ||
		typeName == AttributeTypeArray ||
		typeName == AttributeTypeReference ||
		typeName == AttributeTypeComposite ||
		typeName == AttributeTypePolymorphic
}

// FilterAttributeTypes returns all recognized attribute types from
// the given types array. This is mainly used for validation, to
// ensure there is only one attribute type
func FilterAttributeTypes(types []string) []string {
	ret := make([]string, 0, len(types))
	for _, x := range types {
		if x == AttributeTypeValue ||
			x == AttributeTypeObject ||
			x == AttributeTypeArray ||
			x == AttributeTypeReference ||
			x == AttributeTypeComposite ||
			x == AttributeTypePolymorphic {
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
		if x != AttributeTypeValue &&
			x != AttributeTypeObject &&
			x != AttributeTypeArray &&
			x != AttributeTypeReference &&
			x != AttributeTypeComposite &&
			x != AttributeTypePolymorphic &&
			x != AttributeNodeTerm {
			ret = append(ret, x)
		}
	}
	return ret
}
