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

// TermSchemaManifestType is the object type for a schema manifest
const TermSchemaManifestType = LS + "/SchemaManifest"

// SchemaManifest contains the schema metadata and layer pointers
type SchemaManifest struct {
	ID            string
	PublishedAt   string
	Bundle        string
	ObjectType    string
	ObjectVersion string
	Schema        string
	Overlays      []string
	Values        map[string]interface{}
}

var SchemaManifestTerms = struct {
	PublishedAt terms.ValueTerm
	Bundle      terms.IDTerm
	Schema      terms.IDTerm
	Overlays    terms.IDListTerm
}{
	PublishedAt: terms.ValueTerm(LS + "/SchemaManifest/publishedAt"),
	Bundle:      terms.IDTerm(LS + "/SchemaManifest/bundle"),
	Schema:      terms.IDTerm(LS + "/SchemaManifest/schema"),
	Overlays:    terms.IDListTerm(LS + "/SchemaManifest/overlays"),
}

// SchemaManifestFromLD expands the jsonld input and creates a new
// schema manifest
func SchemaManifestFromLD(jsonldInput interface{}) (*SchemaManifest, error) {
	proc := ld.NewJsonLdProcessor()
	expanded, err := proc.Expand(jsonldInput, nil)
	if err != nil {
		return nil, err
	}
	ret := SchemaManifest{Values: make(map[string]interface{})}
	if err := ret.UnmarshalExpanded(expanded); err != nil {
		return nil, err
	}
	return &ret, nil
}

// UnmarshalExpanded uses the expanded input to populate the schema manifest
func (schema *SchemaManifest) UnmarshalExpanded(in interface{}) error {
	arr, _ := in.([]interface{})
	if len(arr) != 1 {
		return ErrInvalidInput("Invalid schema manifest")
	}
	m, _ := arr[0].(map[string]interface{})
	if m == nil {
		return ErrInvalidInput("Invalid schema manifest")
	}
	if t := GetNodeType(m); t != TermSchemaManifestType {
		return ErrNotASchema(t)
	}
	for k, v := range m {
		switch k {
		case "@type":
		case "@id":
			schema.ID = GetNodeID(m)
		case string(SchemaManifestTerms.PublishedAt):
			schema.PublishedAt = SchemaManifestTerms.PublishedAt.StringFromExpanded(v)
		case string(LayerTerms.ObjectType):
			schema.ObjectType = LayerTerms.ObjectType.StringFromExpanded(v)
		case string(LayerTerms.ObjectVersion):
			schema.ObjectVersion = LayerTerms.ObjectVersion.StringFromExpanded(v)
		case string(SchemaManifestTerms.Bundle):
			schema.Bundle = SchemaManifestTerms.Bundle.StringFromExpanded(v)
		case string(SchemaManifestTerms.Schema):
			schema.Schema = SchemaManifestTerms.Schema.StringFromExpanded(v)
		case string(SchemaManifestTerms.Overlays):
			for _, el := range SchemaManifestTerms.Overlays.ElementValuesFromExpanded(v) {
				schema.Overlays = append(schema.Overlays, el)
			}
		default:
			schema.Values[k] = v
		}
	}

	if len(schema.ObjectType) == 0 {
		return ErrInvalidInput("Empty object type in schema manifest")
	}
	if len(schema.Schema) == 0 {
		return ErrInvalidInput("Empty schema base in schema manifest")
	}
	return nil
}

func (schema *SchemaManifest) MarshalExpanded() interface{} {
	ret := map[string]interface{}{
		"@type": []interface{}{TermSchemaManifestType},
	}
	if len(schema.ID) > 0 {
		ret["@id"] = schema.ID
	}
	for k, v := range schema.Values {
		ret[k] = v
	}
	SchemaManifestTerms.PublishedAt.PutExpanded(ret, schema.PublishedAt)
	LayerTerms.ObjectType.PutExpanded(ret, schema.ObjectType)
	LayerTerms.ObjectVersion.PutExpanded(ret, schema.ObjectVersion)
	SchemaManifestTerms.Bundle.PutExpanded(ret, schema.Bundle)
	SchemaManifestTerms.Schema.PutExpanded(ret, schema.Schema)
	if len(schema.Overlays) != 0 {
		ret[SchemaManifestTerms.Overlays.GetTerm()] = SchemaManifestTerms.Overlays.MakeExpandedContainerFromValues(schema.Overlays)
	}
	return ret
}
