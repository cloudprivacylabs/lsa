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

// SchemaVAriant contains the minimal information to define a schema variant with an optional bundle
type SchemaVariant struct {
	ID        string
	Type      string
	ValueType string
	Schema    string
	Overlays  []string
}

// GetID returns the schema variant ID
func (m *SchemaVariant) GetID() string { return m.ID }

var ErrNotASchemaVariant = errors.New("Not a schema variant")

// Unmarshals the given jsonld document into a schema variant
func UnmarshalSchemaVariant(in interface{}) (*SchemaVariant, error) {
	proc := ld.NewJsonLdProcessor()
	compacted, err := proc.Compact(in, map[string]interface{}{}, nil)
	if err != nil {
		return nil, err
	}
	ret := SchemaVariant{}
	for k, v := range compacted {
		switch k {
		case "@id":
			ret.ID = v.(string)
		case "@type":
			ret.Type = v.(string)
		case ValueTypeTerm:
			ret.ValueType = LDGetNodeID(v)
		case SchemaBaseTerm:
			ret.Schema = LDGetNodeID(v)
		case OverlaysTerm:
			for _, x := range LDGetListElements(v) {
				ret.Overlays = append(ret.Overlays, LDGetNodeID(x))
			}
		}
	}
	if ret.Type != SchemaVariantTerm {
		return nil, ErrNotASchemaVariant
	}
	if len(ret.ID) == 0 || ret.ID == "./" || strings.HasPrefix(ret.ID, "_") {
		return nil, MakeErrInvalidInput("No schema variant  @id")
	}
	return &ret, nil
}

// MarshalSchemaVariant returns a compact jsonld document for the variant
func MarshalSchemaVariant(variant *SchemaVariant) interface{} {
	m := make(map[string]interface{})
	m["@id"] = variant.ID
	m["@type"] = SchemaVariantTerm
	m[ValueTypeTerm] = map[string]interface{}{"@id": variant.ValueType}
	m[SchemaBaseTerm] = map[string]interface{}{"@id": variant.Schema}
	if len(variant.Overlays) > 0 {
		arr := make([]interface{}, 0, len(variant.Overlays))
		for _, x := range variant.Overlays {
			arr = append(arr, map[string]interface{}{"@id": x})
		}
		m[OverlaysTerm] = map[string]interface{}{"@list": arr}
	}
	return m
}
