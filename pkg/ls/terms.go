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
const AttributeNameTerm = LS + "attributeName"

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

// CharacterEncodingTerm is used to specify a character encoding for
// the data processed with the layer
const CharacterEncodingTerm = LS + "characterEncoding"

// InstanceOfTerm is an edge term that is used to connect values with
// their schema specifications
const InstanceOfTerm = LS + "data#instanceOf"

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
	}{},

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

	TypeTerms.Attributes: struct {
		CompositionType
	}{
		ErrorComposition,
	},
	TypeTerms.AttributeList: struct {
		CompositionType
	}{
		ErrorComposition,
	},
	TypeTerms.Reference: struct {
		CompositionType
	}{
		ErrorComposition,
	},
	TypeTerms.ArrayItems: struct {
		CompositionType
	}{
		ErrorComposition,
	},
	TypeTerms.AllOf: struct {
		CompositionType
	}{
		ErrorComposition,
	},
	TypeTerms.OneOf: struct {
		CompositionType
	}{
		ErrorComposition,
	},

	CharacterEncodingTerm: struct{}{},
	InstanceOfTerm:        struct{}{},
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

// NodeCompiler interface represents term compilation algorithm when the term is a node
type NodeCompiler interface {
	// CompileNode gets a node and compiles the associated term on that
	// node. It should store the compiled state into node.Compiled with
	// the an opaque key
	CompileNode(*SchemaNode) error
}

// EdgeCompiler interface represents term compilation algorithm when the term is an edge
type EdgeCompiler interface {
	// CompileEdge gets an edge and compiles the associated term on that
	// edge. It should store tje compiled state into edge.Compiled with
	// an opaque key
	CompileEdge(*SchemaEdge) error
}

// TermCompiler interface represents term compilation algorithm
type TermCompiler interface {
	// CompileTerm gets a term and its value, and returns an object that
	// will be placed in the Compiled map of the node or the edge.
	CompileTerm(string, interface{}) (interface{}, error)
}

type emptyCompiler struct{}

// CompileNode returns the value unmodified
func (emptyCompiler) CompileNode(*SchemaNode) error                        { return nil }
func (emptyCompiler) CompileEdge(*SchemaEdge) error                        { return nil }
func (emptyCompiler) CompileTerm(string, interface{}) (interface{}, error) { return nil, nil }

// GetNodeCompiler return a compiler that will compile the value
func GetNodeCompiler(term string) NodeCompiler {
	md := GetTermMetadata(term)
	if md == nil {
		return emptyCompiler{}
	}
	c, ok := md.(NodeCompiler)
	if ok {
		return c
	}
	return emptyCompiler{}
}

// GetEdgeCompiler return a compiler that will compile the value
func GetEdgeCompiler(term string) EdgeCompiler {
	md := GetTermMetadata(term)
	if md == nil {
		return emptyCompiler{}
	}
	c, ok := md.(EdgeCompiler)
	if ok {
		return c
	}
	return emptyCompiler{}
}

// GetTermCompiler return a compiler that will compile the value
func GetTermCompiler(term string) TermCompiler {
	md := GetTermMetadata(term)
	if md == nil {
		return emptyCompiler{}
	}
	c, ok := md.(TermCompiler)
	if ok {
		return c
	}
	return emptyCompiler{}
}

// Composer interface represents term composition algorithm
type Composer interface {
	Compose(interface{}, interface{}) (interface{}, error)
}

// CompositionType determines the composition semantics for the term
type CompositionType string

const (
	// SetComposition means when two terms are composed, set-union of the values are taken
	SetComposition CompositionType = "set"
	// ListComposition means when two terms are composed, their values are appended
	ListComposition CompositionType = "list"
	// OverrideComposition means when two terms are composed, the new one replaces the old one
	OverrideComposition CompositionType = "override"
	// NoComposition means when two terms are composed, the original remains
	NoComposition CompositionType = "nocompose"
	// ErrorComposition means if two terms are composed and they are different, composition fails
	ErrorComposition CompositionType = "error"
)

// GetComposerForTerm returns a term composer
func GetComposerForTerm(term string) Composer {
	md := GetTermMetadata(term)
	if md == nil {
		return SetComposition
	}
	c, ok := md.(Composer)
	if ok {
		return c
	}
	return SetComposition
}

// Compose target and src based on the composition type
func (c CompositionType) Compose(target, src interface{}) (interface{}, error) {
	switch c {
	case SetComposition:
		return SetUnion(target, src), nil
	case OverrideComposition:
		if src == nil {
			return target, nil
		}
		return src, nil
	case ListComposition:
		return ListAppend(target, src), nil
	case NoComposition:
		return target, nil
	case ErrorComposition:
		if target != src && src != nil {
			return nil, ErrInvalidComposition
		}
		return target, nil
	}
	return SetUnion(target, src), nil
}

// SetUnion computes the set union of properties v1 and v2
func SetUnion(v1, v2 interface{}) interface{} {
	if v1 == nil {
		return v2
	}
	if v2 == nil {
		return v1
	}
	switch e := v1.(type) {
	case []interface{}:
		values := make(map[interface{}]struct{})
		for _, k := range e {
			values[k] = struct{}{}
		}
		ret := e
		if n, ok := v2.([]interface{}); ok {
			for _, item := range n {
				if _, exists := values[item]; !exists {
					values[item] = struct{}{}
					ret = append(ret, item)
				}
			}
			return ret
		}
		if _, exists := values[v2]; !exists {
			return append(e, v2)
		}
		return e
	default:
		ret := []interface{}{e}
		if n, ok := v2.([]interface{}); ok {
			for _, item := range n {
				if item != e {
					ret = append(ret, item)
				}
			}
			if len(ret) == 1 {
				return ret[0]
			}
			return ret
		}
		if e != v2 {
			return []interface{}{e, v2}
		}
		return e
	}
}

// ListAppend appends v2 and v1
func ListAppend(v1, v2 interface{}) interface{} {
	if v1 == nil {
		return v2
	}
	if v2 == nil {
		return v1
	}
	switch e := v1.(type) {
	case []interface{}:
		ret := e
		if n, ok := v2.([]interface{}); ok {
			return append(ret, n...)
		}
		return append(e, v2)
	default:
		ret := []interface{}{e}
		if n, ok := v2.([]interface{}); ok {
			return append(ret, n...)
		}
		return []interface{}{e, v2}
	}
}
