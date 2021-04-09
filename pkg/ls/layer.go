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
	"github.com/piprate/json-gold/ld"

	"github.com/cloudprivacylabs/lsa/pkg/terms"
)

// TermSchemaType is the object type for schema
const TermSchemaType = LS + "/Schema"

// TermOverlayType is the object type for layers
const TermOverlayType = LS + "/Overlay"

var LayerTerms = struct {
	ObjectType    terms.ValueTerm
	ObjectVersion terms.ValueTerm
	Attributes    terms.ObjectSetTerm
	Reference     terms.IDTerm
	ArrayItems    terms.ObjectTerm
	AllOf         terms.ObjectListTerm
	OneOf         terms.ObjectListTerm
}{
	ObjectType:    terms.ValueTerm(LS + "/Layer/objectType"),
	ObjectVersion: terms.ValueTerm(LS + "/Layer/objectVersion"),
	Attributes:    terms.ObjectSetTerm(LS + "/Object/attributes"),

	// Reference is an IRI that points to another object
	Reference: terms.IDTerm(LS + "/Reference/reference"),

	// ArrayItems defines the items of an array object. ArrayItems is an
	// attribute that can contain all attribute related terms
	ArrayItems: terms.ObjectTerm(LS + "/Array/items"),

	// AllOf is a list that denotes composition. The resulting object is
	// a composition of the elements of the list
	AllOf: terms.ObjectListTerm(LS + "/Composite/allOf"),

	// OneOf is a list that denotes polymorphism. The resulting object
	// can be one of the objects listed.
	OneOf: terms.ObjectListTerm(LS + "/Polymorphic/oneOf"),
}

// Layer can be a Schema or an Overlay
type Layer struct {
	ID            string
	Type          string
	ObjectType    string
	ObjectVersion string
	// Root is an object attribute
	Root *Attribute

	Index map[string]*Attribute
}

func (layer *Layer) Clone() *Layer {
	ret := &Layer{ID: layer.ID,
		Type:       layer.Type,
		ObjectType: layer.ObjectType,
		Root:       layer.Root.Clone(nil),
		Index:      make(map[string]*Attribute),
	}
	ret.Root.GetAttributes().Iterate(func(a *Attribute) bool {
		if len(a.ID) > 0 {
			ret.Index[a.ID] = a
		}
		return true
	})
	return ret
}

// NewLayer returns an empty schema layer
func NewLayer() *Layer {
	root := NewAttribute(nil)
	root.Type = NewObjectType(root)
	return &Layer{Root: root,
		Index: make(map[string]*Attribute),
	}
}

// LayerFromLD expands the jsonld input and creates a new schema layer
func LayerFromLD(jsonldInput interface{}) (*Layer, error) {
	proc := ld.NewJsonLdProcessor()
	expanded, err := proc.Expand(jsonldInput, nil)
	if err != nil {
		return nil, err
	}
	ret := Layer{}
	if err := ret.UnmarshalExpanded(expanded); err != nil {
		return nil, err
	}

	return &ret, nil
}

// UnmarshalExpanded unmarshals an expanded json-ld schema or overlay. The input
// must be a []interface{}
func (layer *Layer) UnmarshalExpanded(in interface{}) error {
	arr, _ := in.([]interface{})
	if len(arr) != 1 {
		return ErrInvalidInput("Invalid layer")
	}
	layer.ID = ""
	layer.Type = ""
	layer.Root = NewAttribute(nil)
	layer.Root.Type = NewObjectType(layer.Root)

	m := arr[0].(map[string]interface{})
	layer.ObjectType = LayerTerms.ObjectType.GetExpanded(m)
	layer.ObjectVersion = LayerTerms.ObjectVersion.GetExpanded(m)
	layer.Index = make(map[string]*Attribute)
	layer.Type = GetNodeType(arr[0])
	if layer.Type != TermSchemaType && layer.Type != TermOverlayType {
		return ErrInvalidLayerType(layer.Type)
	}
	if err := layer.Root.GetAttributes().UnmarshalExpanded(arr[0].(map[string]interface{}), nil); err != nil {
		return err
	}
	var err error
	layer.Root.GetAttributes().Iterate(func(a *Attribute) bool {
		if len(a.ID) > 0 {
			if _, ok := layer.Index[a.ID]; ok {
				err = ErrDuplicateAttribute(a.ID)
				return false
			}
			layer.Index[a.ID] = a
		}
		return true
	})
	if err != nil {
		return err
	}
	for k, v := range arr[0].(map[string]interface{}) {
		switch k {
		case "@id":
			layer.ID = v.(string)
		case "@type":
		case string(LayerTerms.Attributes),
			string(LayerTerms.ObjectType),
			string(LayerTerms.ObjectVersion):
		default:
			layer.Root.Values[k] = v
		}
	}
	return nil
}

// MarshalExpanded returns marshaled layer in expanded json-ld format
func (layer *Layer) MarshalExpanded() interface{} {
	ret := map[string]interface{}{}
	if len(layer.ID) > 0 {
		ret["@id"] = layer.ID
	}
	ret["@type"] = []interface{}{layer.Type}
	LayerTerms.ObjectType.PutExpanded(ret, layer.ObjectType)
	LayerTerms.ObjectVersion.PutExpanded(ret, layer.ObjectVersion)
	for k, v := range layer.Root.Values {
		ret[k] = v
	}
	layer.Root.GetAttributes().MarshalExpanded(ret)
	return []interface{}{ret}
}
