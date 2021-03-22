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

// Schema contains the schema metadata and layer pointers
type Schema struct {
	ID             string
	IssuedBy       string
	IssuerRole     string
	IssuedAt       string
	Purpose        string
	Classification string
	ObjectType     string
	ObjectVersion  string
	Layers         []string

	Composed *SchemaLayer
}

// NewSchema expands the jsonld input and creates a new schema
func NewSchema(jsonldInput interface{}) (*Schema, error) {
	proc := ld.NewJsonLdProcessor()
	expanded, err := proc.Expand(jsonldInput, nil)
	if err != nil {
		return nil, err
	}
	ret := Schema{}
	if err := ret.UnmarshalExpanded(expanded); err != nil {
		return nil, err
	}
	return &ret, nil
}

// UnmarshalExpanded uses the expanded input to populate the schema
func (schema *Schema) UnmarshalExpanded(in interface{}) error {
	arr, _ := in.([]interface{})
	if len(arr) != 1 {
		return ErrInvalidInput
	}
	m, _ := arr[0].(map[string]interface{})
	if m == nil {
		return ErrInvalidInput
	}
	if t := GetNodeType(m); t != TermSchemaType {
		return ErrNotASchema(t)
	}
	schema.ID = GetNodeID(m)
	schema.IssuedBy = GetStringValue("@value", m[SchemaTerms.IssuedBy.ID])
	schema.IssuerRole = GetStringValue("@value", m[SchemaTerms.IssuerRole.ID])
	schema.IssuedAt = GetStringValue("@value", m[SchemaTerms.IssuedAt.ID])
	schema.Purpose = GetStringValue("@value", m[SchemaTerms.Purpose.ID])
	schema.Classification = GetStringValue("@value", m[SchemaTerms.Classification.ID])
	schema.ObjectType = GetStringValue("@value", m[SchemaTerms.ObjectType.ID])
	if len(schema.ObjectType) == 0 {
		return ErrInvalidInput
	}
	schema.ObjectVersion = GetStringValue("@value", m[SchemaTerms.ObjectVersion.ID])
	for _, el := range GetListElements(m[SchemaTerms.Layers.ID]) {
		schema.Layers = append(schema.Layers, GetNodeID(el))
	}
	return nil
}

func (schema *Schema) MarshalExpanded() interface{} {
	ret := map[string]interface{}{
		"@type": TermSchemaType,
	}
	if len(schema.ID) > 0 {
		ret["@id"] = schema.ID
	}
	if len(schema.IssuedBy) != 0 {
		ret[SchemaTerms.IssuedBy.ID] = []interface{}{map[string]interface{}{"@value": schema.IssuedBy}}
	}
	if len(schema.IssuerRole) != 0 {
		ret[SchemaTerms.IssuerRole.ID] = []interface{}{map[string]interface{}{"@value": schema.IssuerRole}}
	}
	if len(schema.IssuedAt) != 0 {
		ret[SchemaTerms.IssuedAt.ID] = []interface{}{map[string]interface{}{"@value": schema.IssuedAt}}
	}
	if len(schema.Purpose) != 0 {
		ret[SchemaTerms.Purpose.ID] = []interface{}{map[string]interface{}{"@value": schema.Purpose}}
	}
	if len(schema.Classification) != 0 {
		ret[SchemaTerms.Classification.ID] = []interface{}{map[string]interface{}{"@value": schema.Classification}}
	}
	if len(schema.ObjectType) != 0 {
		ret[SchemaTerms.ObjectType.ID] = []interface{}{map[string]interface{}{"@value": schema.ObjectType}}
	}
	if len(schema.ObjectVersion) != 0 {
		ret[SchemaTerms.ObjectVersion.ID] = []interface{}{map[string]interface{}{"@value": schema.ObjectVersion}}
	}
	if len(schema.Layers) != 0 {
		layers := make([]interface{}, 0, len(schema.Layers))
		for _, x := range schema.Layers {
			layers = append(layers, map[string]interface{}{"@id": x})
		}
		ret[SchemaTerms.Layers.ID] = []interface{}{map[string]interface{}{"@list": layers}}
	}
	return ret
}
