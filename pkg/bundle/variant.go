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

package bundle

import (
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// Variant combines a schema and overlays
type Variant struct {
	SchemaRef `yaml:",inline"`
	Overlays  []SchemaRef `json:"overlays" yaml:"overlays"`
}

type SchemaRef struct {
	Schema     string               `json:"schema,omitempty" yaml:"schema,omitempty"`
	LayerID    string               `json:"layerId,omitempty" yaml:"layerId,omitempty"`
	JSONSchema *JSONSchemaReference `json:"jsonSchema" yaml:"jsonSchema"`

	layer *ls.Layer
}

// Merge resets b with ref if ref is nonempty
func (ref *SchemaRef) Merge(b SchemaRef) {
	if len(b.Schema) > 0 || b.JSONSchema != nil {
		*ref = b
	}
}

// Merge variant into b
func (variant *Variant) Merge(b Variant) {
	variant.SchemaRef.Merge(b.SchemaRef)
	variant.Overlays = append(variant.Overlays, b.Overlays...)
}
