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

// SchemaTerm is the layer type for schemas
const SchemaTerm = LS + "Schema"

// OverlayTerm is the layer type for overlays
const OverlayTerm = LS + "Overlay"

// SchemaManifestTerm is the schema manifest type
const SchemaManifestTerm = LS + "SchemaManifest"

// TargetType is the term specifying the data type for the attribute defined
const TargetType = LS + "targetType"

// DescriptionTerm is used for comments/descriptions
const DescriptionTerm = LS + "description"

// AttributeNameTerm represents the name of an attribute
const AttributeNameTerm = LS + "attributeName"

// AttributeIndexTerm represents the index of an array element
const AttributeIndexTerm = LS + "attributeIndex"

// LayerRootTerm is an edge term that connects layer node to the root node of the schema
const LayerRootTerm = LS + "layer"

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
	Value:       LS + "Value",
	Object:      LS + "Object",
	Array:       LS + "Array",
	Reference:   LS + "Reference",
	Composite:   LS + "Composite",
	Polymorphic: LS + "Polymorphic",
	Attribute:   LS + "Attribute",
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
	Attributes:    LS + "Object#attributes",
	AttributeList: LS + "Object#attributeList",
	Reference:     LS + "Reference#reference",
	ArrayItems:    LS + "Array#items",
	AllOf:         LS + "Composite#allOf",
	OneOf:         LS + "Polymorphic#oneOf",
}

var DataEdgeTerms = struct {
	// Edge label linking attribute nodes to an object node
	ObjectAttributes string
	// Edge label linking array element nodes to an array node
	ArrayElements string
}{
	ObjectAttributes: LS + "data/object#attributes",
	ArrayElements:    LS + "data/array#elements",
}

// FilterAttributeTypes returns all recognized attribute types from
// the given types array. This is mainly used for validation, to
// ensure there is only one attribute type
func FilterAttributeTypes(types []string) []string {
	ret := make([]string, 0)
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

// CharacterEncodingTerm is used to specify a character encoding for
// the data processed with the layer
const CharacterEncodingTerm = LS + "characterEncoding"

// InstanceOfTerm is an edge term that is used to connect values with
// their schema specifications
const InstanceOfTerm = LS + "data#instanceOf"

const BundleTerm = LS + "SchemaManifest#bundle"
const SchemaBaseTerm = LS + "SchemaManifest#schema"
const OverlaysTerm = LS + "SchemaManifest#overlays"

// All registered terms and their associated metadata
var termMetadata = map[string]interface{}{
	SchemaTerm: struct {
		CompositionType
	}{
		NoComposition,
	},
	OverlayTerm: struct {
		CompositionType
	}{
		NoComposition,
	},
	TargetType: struct {
		CompositionType
	}{
		SetComposition,
	},

	DescriptionTerm: struct {
		CompositionType
	}{
		SetComposition,
	},
	AttributeNameTerm: struct {
		CompositionType
	}{
		ErrorComposition,
	},
	AttributeIndexTerm: struct {
		CompositionType
	}{
		ErrorComposition,
	},

	AttributeTypes.Value: struct {
		CompositionType
	}{
		OverrideComposition,
	},
	AttributeTypes.Object: struct {
		CompositionType
	}{
		OverrideComposition,
	},
	AttributeTypes.Array: struct {
		CompositionType
	}{
		OverrideComposition,
	},
	AttributeTypes.Reference: struct {
		CompositionType
	}{
		OverrideComposition,
	},
	AttributeTypes.Composite: struct {
		CompositionType
	}{
		OverrideComposition,
	},
	AttributeTypes.Polymorphic: struct {
		CompositionType
	}{
		OverrideComposition,
	},
	AttributeTypes.Attribute: struct {
		CompositionType
	}{
		OverrideComposition,
	},

	LayerTerms.Attributes: struct {
		CompositionType
	}{
		ErrorComposition,
	},
	LayerTerms.AttributeList: struct {
		CompositionType
	}{
		ErrorComposition,
	},
	LayerTerms.Reference: struct {
		CompositionType
	}{
		ErrorComposition,
	},
	LayerTerms.ArrayItems: struct {
		CompositionType
	}{
		ErrorComposition,
	},
	LayerTerms.AllOf: struct {
		CompositionType
	}{
		ErrorComposition,
	},
	LayerTerms.OneOf: struct {
		CompositionType
	}{
		ErrorComposition,
	},

	CharacterEncodingTerm: struct {
		CompositionType
	}{
		OverrideComposition,
	},
	InstanceOfTerm: struct {
		CompositionType
	}{
		ErrorComposition,
	},

	DataEdgeTerms.ObjectAttributes: struct {
		CompositionType
	}{
		ErrorComposition,
	},
	DataEdgeTerms.ArrayElements: struct {
		CompositionType
	}{
		ErrorComposition,
	},
}

// GetTermMetadata returns metadata about a term
func GetTermMetadata(term string) interface{} {
	return termMetadata[term]
}

// RegisterTermMetadata registers metadata for a term
func RegisterTermMetadata(term string, md interface{}) {
	if _, ok := termMetadata[term]; ok {
		panic("Duplicate term: " + term)
	}
	termMetadata[term] = md
}
