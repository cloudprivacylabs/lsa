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
package jsonld

import (
	"encoding/json"

	"github.com/cloudprivacylabs/lsa/pkg/layers"
	"github.com/piprate/json-gold/ld"
)

type SchemaManifest struct {
	ID         string   `json:"@id"`
	TargetType string   `json:"https://lschema.org/targetType"`
	Bundle     string   `json:"https://lschema.org/SchemaManifest#bundle,omitempty"`
	Schema     string   `json:"https://lschema.org/SchemaManifest#schema"`
	Overlays   []string `json:"https://lschema.org/SchemaManifest#overlays,omitempty"`
}

// Unmarshals the given jsonld document into a schema manifest
func UnmarshalSchemaManifest(in interface{}) (*layers.SchemaManifest, error) {
	var m SchemaManifest
	proc := ld.NewJsonLdProcessor()
	compacted, err := proc.Compact(in, map[string]interface{}{}, nil)
	if err != nil {
		return nil, err
	}
	d, _ := json.Marshal(compacted)
	err = json.Unmarshal(d, &m)
	if err != nil {
		return nil, err
	}
	return (*layers.SchemaManifest)(&m), nil
}

// MarshalSchemaNanifest returns a compact jsonld document for the manifest
func MarshalSchemaManifest(manifest *layers.SchemaManifest) interface{} {
	m := (*SchemaManifest)(manifest)
	d, _ := json.Marshal(m)
	var v interface{}
	json.Unmarshal(d, &v)
	return v
}
