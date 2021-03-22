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
)

// SchemaLayer contains the layer model
type SchemaLayer struct {
	ID         string
	ObjectType string
	Attributes
	Values map[string]interface{}
}

func (layer *SchemaLayer) Clone() *SchemaLayer {
	return &SchemaLayer{ID: layer.ID,
		ObjectType: layer.ObjectType,
		Attributes: *layer.Attributes.Clone(nil),
		Values:     ld.CloneDocument(layer.Values).(map[string]interface{}),
	}
}

// NewEmptySchemaLayer returns an empty schema kayer
func NewEmptySchemaLayer() *SchemaLayer {
	return &SchemaLayer{Attributes: *NewAttributes(nil),
		Values: make(map[string]interface{}),
	}
}

// NewSchemaLayer expands the jsonld input and creates a new schema layer
func NewSchemaLayer(jsonldInput interface{}) (*SchemaLayer, error) {
	proc := ld.NewJsonLdProcessor()
	expanded, err := proc.Expand(jsonldInput, nil)
	if err != nil {
		return nil, err
	}
	ret := SchemaLayer{}
	if err := ret.UnmarshalExpanded(expanded); err != nil {
		return nil, err
	}
	return &ret, nil
}

// UnmarshalExpanded unmarshals an expanded json-ld schema or overlay. The input
// must be a []interface{}
func (layer *SchemaLayer) UnmarshalExpanded(in interface{}) error {
	arr, _ := in.([]interface{})
	if len(arr) != 1 {
		return ErrInvalidInput
	}
	layer.ID = ""
	layer.Values = make(map[string]interface{})
	typ := GetNodeType(arr[0])
	if typ != TermLayerType {
		return ErrInvalidLayerType(typ)
	}
	for k, v := range arr[0].(map[string]interface{}) {
		switch k {
		case "@id":
			layer.ID = v.(string)
		case AttributeStructure.Attributes.ID:
			if err := layer.Attributes.UnmarshalExpanded(v); err != nil {
				return err
			}
		case SchemaTerms.ObjectType.ID:
			layer.ObjectType = GetStringValue("@value", v)
		case "@type":
		default:
			layer.Values[k] = v
		}
	}
	return nil
}

// MarshalExpanded returns marshaled layer in expanded json-ld format
func (layer *SchemaLayer) MarshalExpanded() interface{} {
	ret := map[string]interface{}{}
	if len(layer.ID) > 0 {
		ret["@id"] = layer.ID
	}
	ret["@type"] = TermLayerType
	if len(layer.ObjectType) > 0 {
		ret[SchemaTerms.ObjectType.ID] = []interface{}{map[string]interface{}{"@value": layer.ObjectType}}
	}
	for k, v := range layer.Values {
		ret[k] = v
	}
	ret[AttributeStructure.Attributes.ID] = layer.Attributes.MarshalExpanded()
	return []interface{}{ret}
}

// Validate the schema layer. Checks for duplicate attributes
func (layer *SchemaLayer) Validate() error {
	var err error
	ids := map[string]struct{}{}
	layer.Attributes.Iterate(func(a *Attribute) bool {
		if len(a.ID) > 0 {
			if _, exists := ids[a.ID]; exists {
				err = ErrDuplicateAttribute(a.ID)
				return false
			}
			ids[a.ID] = struct{}{}
		}
		return true
	})
	return err
}

// A SchemaObject is either Attributes or Attribute
type SchemaObject interface {
	GetParent() SchemaObject
}
