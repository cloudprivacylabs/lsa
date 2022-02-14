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
	"errors"
	"strings"

	"github.com/piprate/json-gold/ld"
)

// SchemaManifest contains the minimal information to define a schema variant with an optional bundle
type SchemaManifest struct {
	ID         string
	Type       string
	TargetType string
	Bundle     string
	Schema     string
	Overlays   []string
}

// GetID returns the schema manifest ID
func (m *SchemaManifest) GetID() string { return m.ID }

var ErrNotASchemaManifest = errors.New("Not a schema manifest")

// Unmarshals the given jsonld document into a schema manifest
func UnmarshalSchemaManifest(in interface{}) (*SchemaManifest, error) {
	proc := ld.NewJsonLdProcessor()
	compacted, err := proc.Compact(in, map[string]interface{}{}, nil)
	if err != nil {
		return nil, err
	}
	ret := SchemaManifest{}
	for k, v := range compacted {
		switch k {
		case "@id":
			ret.ID = v.(string)
		case "@type":
			ret.Type = v.(string)
		case TargetType:
			ret.TargetType = LDGetNodeID(v)
		case BundleTerm:
			ret.Bundle = LDGetNodeID(v)
		case SchemaBaseTerm:
			ret.Schema = LDGetNodeID(v)
		case OverlaysTerm:
			for _, x := range LDGetListElements(v) {
				ret.Overlays = append(ret.Overlays, LDGetNodeID(x))
			}
		}
	}
	if ret.Type != SchemaManifestTerm {
		return nil, ErrNotASchemaManifest
	}
	if len(ret.ID) == 0 || ret.ID == "./" || strings.HasPrefix(ret.ID, "_") {
		return nil, MakeErrInvalidInput("No schema manifest  @id")
	}
	return &ret, nil
}

// MarshalSchemaNanifest returns a compact jsonld document for the manifest
func MarshalSchemaManifest(manifest *SchemaManifest) interface{} {
	m := make(map[string]interface{})
	m["@id"] = manifest.ID
	m["@type"] = SchemaManifestTerm
	m[TargetType] = map[string]interface{}{"@id": manifest.TargetType}
	if len(manifest.Bundle) > 0 {
		m[BundleTerm] = map[string]interface{}{"@id": manifest.Bundle}
	}
	m[SchemaBaseTerm] = map[string]interface{}{"@id": manifest.Schema}
	if len(manifest.Overlays) > 0 {
		arr := make([]interface{}, 0, len(manifest.Overlays))
		for _, x := range manifest.Overlays {
			arr = append(arr, map[string]interface{}{"@id": x})
		}
		m[OverlaysTerm] = map[string]interface{}{"@list": arr}
	}
	return m
}
