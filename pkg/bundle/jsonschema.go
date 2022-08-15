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
	"io"
	"strings"

	"github.com/bserdar/jsonom"

	"github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// A JSONSchema is a reference to a JSON schema document, with
// optional overlays. The JSON schema is loaded, and combined with the
// specified overlays. Then, it is made available to the bundle using
// the ID
type JSONSchema struct {
	// NAme of the JSON schema. This name is passed to the json schema
	// loader to load the schema
	Name string `json:"name" yaml:"name"`

	// The ID of the JSON schema. This ID will be used to reference to
	// this schema within the bundle
	ID string `json:"id" yaml:"id"`

	// JSON schema overlays. These overlays are also JSON schemas
	Overlays []string `json:"overlays" yaml:"overlays"`
}

func (sch JSONSchema) Load(ctx *ls.Context, loader func(*ls.Context, string) (io.ReadCloser, error)) (jsonom.Node, error) {
	ctx.GetLogger().Debug(map[string]interface{}{"jsonSchema.load": sch.Name, "stage": "begin"})
	node, err := json.ComposeSchema(ctx, sch.Name, sch.Overlays, loader)
	if err != nil {
		ctx.GetLogger().Debug(map[string]interface{}{"jsonSchema.load": sch.Name, "stage": "cannot load", "err": err})
		return nil, err
	}
	return node, nil
}

type JSONSchemaReference struct {
	// Refer to a layer by ID. Layer can be imported as a spreadsheet
	LayerID string `json:"layerId" yaml:"layerId" bson:"layerId"`
	// This is the reference to the JSON schema defined in the bundle, with an optional fragment
	Ref       string `json:"ref" yaml:"ref" bson:"ref"`
	Namespace string `json:"namespace" yaml:"namespace" bson:"namespace"`
}

// GetSchemaBase returns the JSON schema base name, without the fragment
func (j JSONSchemaReference) GetSchemaBase() string {
	return getSchemaBase(j.Ref)
}

func getSchemaBase(ref string) string {
	ix := strings.Index(ref, "#")
	if ix == -1 {
		return ref
	}
	return ref[:ix]
}
