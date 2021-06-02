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

// TargetType is the term specifying the data type for the attribute defined
const TargetType = LS + "targetType"

// DescriptionTerm is used for comments/descriptions
const DescriptionTerm = LS + "description"

// AttributeNameTerm represents the name of an attribute
const AttributeNameTerm = LS + "attrName"

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

// TypeTerms includes type specific terms recognized by the schema
// compiler. These are terms used to define elements of an attribute.
var TypeTerms = struct {
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
