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
	SchemaTerm = NewTerm(LS, "Schema", false, false, NoComposition, nil)

	// OverlayTerm is the layer type for overlays
	OverlayTerm = NewTerm(LS, "Overlay", false, false, NoComposition, nil)

	// ComposeTerm is used for overlays to redefine term compositions. One of CompositionType constants
	ComposeTerm = NewTerm(LS, "compose", false, false, OverrideComposition, nil)

	// AttributeOverlaysTerm lists the overlays for schema attributes
	// that are matched by ID, as opposed to matching by ID and their
	// place in the layer
	AttributeOverlaysTerm = NewTerm(LS, "attributeOverlays", false, true, OverrideComposition, nil)

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
	NSMapTerm = NewTerm(LS, "nsMap", false, false, OverrideComposition, nil)

	// CharacterEncodingTerm is used to specify a character encoding for
	// the data processed with the layer
	CharacterEncodingTerm = NewTerm(LS, "characterEncoding", false, false, OverrideComposition, nil)

	// InstanceOfTerm is an edge term that is used to connect values with
	// their schema specifications
	InstanceOfTerm = NewTerm(LS, "instanceOf", false, false, ErrorComposition, nil)

	// SchemaNodeIDTerm denotes the schema node ID for ingested nodes
	SchemaNodeIDTerm = NewTerm(LS, "schemaNodeId", false, false, ErrorComposition, nil)

	// SchemaVariantTerm is the schema variant type
	SchemaVariantTerm = NewTerm(LS, "SchemaVariant", false, false, NoComposition, nil)

	// DescriptionTerm is used for comments/descriptions
	DescriptionTerm = NewTerm(LS, "description", false, false, SetComposition, nil)

	// AttributeNameTerm represents the name of an attribute
	AttributeNameTerm = NewTerm(LS, "attributeName", false, false, OverrideComposition, nil)

	// AttributeIndexTerm represents the index of an array element
	AttributeIndexTerm = NewTerm(LS, "attributeIndex", false, false, NoComposition, nil)

	// ConditionalTerm specifies conditions for ingestion
	ConditionalTerm = NewTerm(LS, "conditional", false, false, OverrideComposition, nil)

	// LayerRootTerm is an edge term that connects layer node to the root node of the schema
	LayerRootTerm = NewTerm(LS, "layer", false, false, ErrorComposition, nil)

	// DefaultValueTerm is the default value for an attribute if attribute is not present
	DefaultValueTerm = NewTerm(LS, "defaultValue", false, false, OverrideComposition, nil)

	// Format specifies a type-specific formatting directive, such as a date format
	FormatTerm = NewTerm(LS, "format", false, false, OverrideComposition, nil)
	// EntitySchemaTerm is inserted by the schema compilation to mark
	// entity roots. It records the schema ID containing the entity
	// definition.
	EntitySchemaTerm = NewTerm(LS, "entitySchema", false, false, ErrorComposition, nil)

	// NodeIDTerm keeps the node ID or the attribute ID
	NodeIDTerm = NewTerm(LS, "nodeId", false, false, ErrorComposition, nil)

	// IngestAsTerm ingests value as an edge, node, or property
	IngestAsTerm = NewTerm(LS, "ingestAs", false, false, OverrideComposition, nil)

	// AsPropertyOfTerm is optional. If specified, it gives the nearest
	// ancestor node that is an instance of the given type. If not, it
	// is the parent document node
	AsPropertyOfTerm = NewTerm(LS, "asPropertyOf", false, false, OverrideComposition, nil)

	// EdgeLabelTerm represents the value used as an edge label, when ingesting an edge
	EdgeLabelTerm = NewTerm(LS, "edgeLabel", false, false, OverrideComposition, nil)

	// PropertyNameTerm represents the value used as a property name when ingesting a property
	PropertyNameTerm = NewTerm(LS, "propertyName", false, false, OverrideComposition, nil)

	// DocumentNodeTerm is the type of document nodes
	DocumentNodeTerm = NewTerm(LS, "DocumentNode", false, false, ErrorComposition, nil)

	// NodeValueTerm is the property key used to keep node value
	NodeValueTerm = NewTerm(LS, "value", false, false, ErrorComposition, nil)

	// ValueTypeTerm defines the type of a value
	ValueTypeTerm = NewTerm(LS, "valueType", false, false, OverrideComposition, nil)

	// HasTerm is an edge term for linking document elements
	HasTerm = NewTerm(LS, "has", false, false, ErrorComposition, nil)

	// EntityIDFieldsTerm is a string or []string that lists the attribute IDs
	// for entity ID. It is defined at the root node of a layer. All
	// attribute IDs must refer to value nodes.
	EntityIDFieldsTerm = NewTerm(LS, "entityIdFields", false, false, OverrideComposition, nil)

	// EntityIDTerm is a string or []string that gives the unique ID of
	// an entity. This is a node property at the root node of an entity
	EntityIDTerm = NewTerm(LS, "entityId", false, false, OverrideComposition, nil)

	// LabeledAsTerm adds labels to JSON schemas
	LabeledAsTerm = NewTerm(LS, "labeledAs", false, false, OverrideComposition, nil)

	// TypeDiscriminatorTerm represents a set of schema field hints for defining polymorphic objects
	TypeDiscriminatorTerm = NewTerm(LS, "typeDiscriminator", false, false, NoComposition, nil)

	// AttributeIncludeTerm represents another schema to replace and copy its contents
	AttributeIncludeTerm = NewTerm(LS, "include", false, false, OverrideComposition, nil)
)

// Attribute types defines the terms describing attribute types. Each
// attribute must have one of the attribute types plus the Attribute
// type, marking the object as an attribute.
var (
	AttributeTypeValue       = NewTerm(LS, "Value", false, false, OverrideComposition, nil)
	AttributeTypeObject      = NewTerm(LS, "Object", false, false, OverrideComposition, nil)
	AttributeTypeArray       = NewTerm(LS, "Array", false, false, OverrideComposition, nil)
	AttributeTypeReference   = NewTerm(LS, "Reference", false, false, OverrideComposition, nil)
	AttributeTypeComposite   = NewTerm(LS, "Composite", false, false, OverrideComposition, nil)
	AttributeTypePolymorphic = NewTerm(LS, "Polymorphic", false, false, OverrideComposition, nil)
	AttributeNodeTerm        = NewTerm(LS, "Attribute", false, false, OverrideComposition, nil)
)

// Layer terms includes type specific terms recognized by the schema
// compiler. These are terms used to define elements of an attribute.
var (
	// Unordered named attributes (json object)
	ObjectAttributesTerm = NewTerm(LS, "Object/attributes", false, false, ErrorComposition, nil)
	// Ordered named attributes (json object, xml elements)
	ObjectAttributeListTerm = NewTerm(LS, "Object/attributeList", false, true, ErrorComposition, nil)
	// Reference to another schema. This will be resolved to another
	// schema during compilation
	ReferenceTerm = NewTerm(LS, "Reference/ref", false, false, OverrideComposition, nil)
	// ArrayItems contains the definition for the items of the array
	ArrayItemsTerm = NewTerm(LS, "Array/elements", false, false, ErrorComposition, nil)
	// All components of a composite attribute
	AllOfTerm = NewTerm(LS, "Composite/allOf", false, true, ErrorComposition, nil)
	// All options of a polymorphic attribute
	OneOfTerm = NewTerm(LS, "Polymorphic/oneOf", false, true, ErrorComposition, nil)
)

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
