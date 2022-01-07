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
	SchemaTerm = NewTerm(LS+"Schema", false, false, NoComposition, nil)

	// OverlayTerm is the layer type for overlays
	OverlayTerm = NewTerm(LS+"Overlay", false, false, NoComposition, nil)

	// SchemaManifestTerm is the schema manifest type
	SchemaManifestTerm = NewTerm(LS+"SchemaManifest", false, false, NoComposition, nil)

	// TargetType is the term specifying the data type for the attribute defined
	TargetType = NewTerm(LS+"targetType", true, false, SetComposition, nil)

	// DescriptionTerm is used for comments/descriptions
	DescriptionTerm = NewTerm(LS+"description", false, false, SetComposition, nil)

	// AttributeNameTerm represents the name of an attribute
	AttributeNameTerm = NewTerm(LS+"attributeName", false, false, OverrideComposition, nil)

	// AttributeIndexTerm represents the index of an array element
	AttributeIndexTerm = NewTerm(LS+"attributeIndex", false, false, NoComposition, nil)

	// AttributeValueTerm represents the value of an attribute
	AttributeValueTerm = NewTerm(LS+"attributeValue", false, false, ErrorComposition, nil)

	// LayerRootTerm is an edge term that connects layer node to the root node of the schema
	LayerRootTerm = NewTerm(LS+"layer", false, false, ErrorComposition, nil)

	// DefaultValueTerm is the default value for an attribute if attribute is not present
	DefaultValueTerm = NewTerm(LS+"defaultValue", false, false, OverrideComposition, nil)

	// EntityIDTerm marks a field of an entity as the entity unique
	// ID. Fields contained within an entity will get IDs relative to
	// this ID.
	//
	// This is a marker term. The contents are ignored. Existance of
	// entityId term on a field marks it as entity id.
	EntityIDTerm = NewTerm(LS+"entityId", false, false, OverrideComposition, nil)

	// EntitySchemaTerm is inserted by the schema compilation to mark
	// entity roots. It records the schema ID containing the entity
	// definition.
	EntitySchemaTerm = NewTerm(LS+"entitySchema", false, false, ErrorComposition, nil)
)

// AttributeTypes defines the terms describing attribute types. Each
// attribute must have one of the attribute types plus the Attribute
// type, marking the object as an attribute.
var AttributeTypes = struct {
	Value       string
	Object      string
	Array       string
	Reference   string
	Composite   string
	Polymorphic string
	Attribute   string
}{
	Value:       NewTerm(LS+"Value", false, false, OverrideComposition, nil),
	Object:      NewTerm(LS+"Object", false, false, OverrideComposition, nil),
	Array:       NewTerm(LS+"Array", false, false, OverrideComposition, nil),
	Reference:   NewTerm(LS+"Reference", false, false, OverrideComposition, nil),
	Composite:   NewTerm(LS+"Composite", false, false, OverrideComposition, nil),
	Polymorphic: NewTerm(LS+"Polymorphic", false, false, OverrideComposition, nil),
	Attribute:   NewTerm(LS+"Attribute", false, false, OverrideComposition, nil),
}

// LayerTerms includes type specific terms recognized by the schema
// compiler. These are terms used to define elements of an attribute.
var LayerTerms = struct {
	// Unordered named attributes (json object)
	Attributes string
	// Ordered named attributes (json object, xml elements)
	AttributeList string
	// Reference to another schema. This will be resolved to another
	// schema during compilation
	Reference string
	// ArrayItems contains the definition for the items of the array
	ArrayItems string
	// All components of a composite attribute
	AllOf string
	// All options of a polymorphic attribute
	OneOf string
}{
	Attributes:    NewTerm(LS+"Object/attributes", false, false, ErrorComposition, nil),
	AttributeList: NewTerm(LS+"Object/attributeList", false, true, ErrorComposition, nil),
	Reference:     NewTerm(LS+"Reference/ref", true, false, ErrorComposition, nil),
	ArrayItems:    NewTerm(LS+"Array/elements", false, false, ErrorComposition, nil),
	AllOf:         NewTerm(LS+"Composite/allOf", false, true, ErrorComposition, nil),
	OneOf:         NewTerm(LS+"Polymorphic/oneOf", false, true, ErrorComposition, nil),
}

// DocumentNodeTerm is ithe type of document nodes
var DocumentNodeTerm = NewTerm(LS+"DocumentNode", false, false, ErrorComposition, nil)

// HasTerm is an edge term for linking document elements
var HasTerm = NewTerm(LS+"data/has", false, false, ErrorComposition, nil)

// FilterAttributeTypes returns all recognized attribute types from
// the given types array. This is mainly used for validation, to
// ensure there is only one attribute type
func FilterAttributeTypes(types []string) []string {
	ret := make([]string, 0, len(types))
	for _, x := range types {
		if x == AttributeTypes.Value ||
			x == AttributeTypes.Object ||
			x == AttributeTypes.Array ||
			x == AttributeTypes.Reference ||
			x == AttributeTypes.Composite ||
			x == AttributeTypes.Polymorphic {
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
		if x != AttributeTypes.Value &&
			x != AttributeTypes.Object &&
			x != AttributeTypes.Array &&
			x != AttributeTypes.Reference &&
			x != AttributeTypes.Composite &&
			x != AttributeTypes.Polymorphic &&
			x != AttributeTypes.Attribute {
			ret = append(ret, x)
		}
	}
	return ret
}

var (
	// CharacterEncodingTerm is used to specify a character encoding for
	// the data processed with the layer
	CharacterEncodingTerm = NewTerm(LS+"characterEncoding", false, false, OverrideComposition, nil)

	// InstanceOfTerm is an edge term that is used to connect values with
	// their schema specifications
	InstanceOfTerm = NewTerm(LS+"data/instanceOf", false, false, ErrorComposition, nil)

	// AsPropertyOfTerm is optional. If specified, it gives the nearest
	// node that is an instance of the given type. If not, it is the
	// nearest document node
	AsPropertyOfTerm = NewTerm(LS+"asPropertyOf", false, false, OverrideComposition, nil)
	// AsPropertyTerm specifies the property name the data point
	// should be added in the parent node
	AsPropertyTerm = NewTerm(LS+"asProperty", false, false, OverrideComposition, nil)

	BundleTerm     = NewTerm(LS+"SchemaManifest/bundle", false, false, ErrorComposition, nil)
	SchemaBaseTerm = NewTerm(LS+"SchemaManifest/schema", true, false, ErrorComposition, nil)
	OverlaysTerm   = NewTerm(LS+"SchemaManifest/overlays", true, true, ErrorComposition, nil)
)

var registeredTerms = map[string]TermSemantics{}

// If a term is know, using this function avoids duplicate string
// copies
func knownTerm(s string) string {
	x, ok := registeredTerms[s]
	if ok {
		return x.Term
	}
	return s
}

func RegisterTerm(t TermSemantics) {
	_, ok := registeredTerms[t.Term]
	if ok {
		panic("Duplicate term :" + t.Term)
	}
	registeredTerms[t.Term] = t
}

func GetTermInfo(term string) TermSemantics {
	t, ok := registeredTerms[term]
	if !ok {
		return TermSemantics{Term: term, Composition: SetComposition}
	}
	return t
}

// GetTermMetadata returns metadata about a term
func GetTermMetadata(term string) interface{} {
	t := GetTermInfo(term)
	return t.Metadata
}
