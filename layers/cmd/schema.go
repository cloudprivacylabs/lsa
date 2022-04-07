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

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	jsonsch "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// Bundle defines type names for variants so references can be resolved
type Bundle struct {
	TypeNames map[string]BundleVariant `json:"typeNames" yaml:"typeNames"`
}

type BundleSchemaRef struct {
	LayerID    string               `json:"layerId,omitempty" yaml:"layerId,omitempty" bson:"layerId,omitempty"`
	JSONSchema *JSONSchemaReference `json:"jsonSchema" yaml:"jsonSchema" bson:"jsonSchema,omitempty"`
}

// GetLayerID returns the layer id for the schema reference
func (ref BundleSchemaRef) GetLayerID() string {
	if ref.JSONSchema != nil {
		return ref.JSONSchema.LayerID
	}
	return ref.LayerID
}

type JSONSchemaReference struct {
	LayerID string `json:"layerId" yaml:"layerId" bson:"layerId"`
	Ref     string `json:"ref" yaml:"ref" bson:"ref"`
}

// BundleVariant combines a schema and overlays
type BundleVariant struct {
	BundleSchemaRef
	Overlays []BundleSchemaRef `json:"overlays" yaml:"overlays"`
}

// ParseBundle parses a bundle from JSON
func ParseBundle(text string, contentType string) (*Bundle, error) {
	var ret Bundle
	if err := cmdutil.ReadJSON(contentType, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}

func (bundle *Bundle) importJSONSchema(ctx *ls.Context, typeTerm, path string, importEntities []jsonsch.Entity) (map[string]*ls.Layer, error) {
	compiler := jsonschema.NewCompiler()
	compiler.LoadURL = func(s string) (io.ReadCloser, error) {
		obj, err := cmdutil.ReadURL(s)
		if err != nil {
			return nil, err
		}
		return ioutil.NopCloser(bytes.NewReader(obj)), nil
	}
	// Import all JSON schemas into a graph
	jsonLayers := make(map[string]*ls.Layer)
	if len(importEntities) > 0 {
		compiled, err := jsonsch.CompileEntitiesWith(compiler, path, importEntities...)
		if err != nil {
			return nil, err
		}
		g := ls.NewLayerGraph()
		layers, err := jsonsch.BuildEntityGraph(g, typeTerm, jsonsch.LinkRefsByValueType, compiled...)
		if err != nil {
			return nil, err
		}
		for _, layer := range layers {
			jsonLayers[layer.Layer.GetID()] = layer.Layer
		}
	}
	return jsonLayers, nil
}

// GetLayers returns the layers of the bundle keyed by variant type
func (bundle *Bundle) GetLayers(ctx *ls.Context, relative string, loader func(file string) (*ls.Layer, error)) (map[string]*ls.Layer, error) {
	// layers keyed by layer id
	layers := make(map[string]*ls.Layer)
	// entities keyed by layer id
	schemaEntities := make(map[string]jsonsch.Entity)
	ovlEntities := make(map[string]jsonsch.Entity)

	processRef := func(variantType string, ref BundleSchemaRef, entitiesMap map[string]jsonsch.Entity) error {
		switch {
		case len(ref.LayerID) > 0:
			// Load the layer if not loaded before
			_, loaded := layers[ref.LayerID]
			if loaded {
				break
			}
			layer, err := loader(ref.LayerID)
			if err != nil {
				return err
			}
			layers[layer.GetID()] = layer

		case ref.JSONSchema != nil:
			_, loaded := entitiesMap[ref.JSONSchema.LayerID]
			if loaded {
				break
			}
			entity := jsonsch.Entity{
				LayerID:   ref.JSONSchema.LayerID,
				ValueType: variantType,
				Ref:       ref.JSONSchema.Ref,
			}
			entitiesMap[entity.LayerID] = entity
		}
		return nil
	}

	// Load all layers, construct entities
	for variantType, variant := range bundle.TypeNames {
		if err := processRef(variantType, variant.BundleSchemaRef, schemaEntities); err != nil {
			return nil, err
		}
		for _, ovl := range variant.Overlays {
			if err := processRef(variantType, ovl, ovlEntities); err != nil {
				return nil, err
			}
		}
	}

	// If there are entities, import them
	// First import schemas, then overlays
	importJson := func(input map[string]jsonsch.Entity, typeTerm string) error {
		importEntities := make([]jsonsch.Entity, 0, len(input))
		for _, entity := range input {
			importEntities = append(importEntities, entity)
		}
		if len(importEntities) > 0 {
			jlayers, err := bundle.importJSONSchema(ctx, typeTerm, relative, importEntities)
			if err != nil {
				return err
			}
			for k, v := range jlayers {
				if _, exists := layers[k]; exists {
					return fmt.Errorf("Multiple definitions for layer %s", k)
				}
				layers[k] = v
			}
		}
		return nil
	}

	if err := importJson(schemaEntities, ls.SchemaTerm); err != nil {
		return nil, err
	}
	if err := importJson(ovlEntities, ls.OverlayTerm); err != nil {
		return nil, err
	}

	resultBundle := ls.BundleByType{}
	for variantType, variant := range bundle.TypeNames {
		sch := layers[variant.GetLayerID()]
		ovl := make([]*ls.Layer, 0, len(variant.Overlays))
		for _, o := range variant.Overlays {
			ovl = append(ovl, layers[o.GetLayerID()])
		}
		_, err := resultBundle.Add(ctx, variantType, sch, ovl...)
		if err != nil {
			return nil, err
		}
	}
	ret := make(map[string]*ls.Layer)
	for variantId := range bundle.TypeNames {
		layer, _ := resultBundle.LoadSchema(variantId)
		if layer != nil {
			ret[variantId] = layer
		}
	}
	return ret, nil
}

func LoadBundle(ctx *ls.Context, file string) (ls.SchemaLoader, error) {
	var bundle Bundle
	if err := cmdutil.ReadJSON(file, &bundle); err != nil {
		return nil, err
	}
	dir := filepath.Dir(file)
	items, err := bundle.GetLayers(ctx, dir, func(data string) (*ls.Layer, error) {
		var input interface{}
		err := json.Unmarshal([]byte(data), &input)
		if err != nil {
			return nil, err
		}
		return ls.UnmarshalLayer(input, nil)
	})
	if err != nil {
		return nil, err
	}
	b := ls.BundleByType{}
	for k, v := range items {
		b.Add(ctx, k, v)
	}
	return &b, nil
}

// ReadLayers reads layer(s) from jsongraph, jsonld
func ReadLayers(input []byte, interner ls.Interner) ([]*ls.Layer, error) {
	var v interface{}
	err := json.Unmarshal(input, &v)
	if err != nil {
		return nil, err
	}
	// Input is JSON or JSON-LD
	// If input is []interface{}, it must be JSON-LD
	if _, arr := v.([]interface{}); arr {
		l, err := ls.UnmarshalLayer(v, interner)
		if err != nil {
			return nil, err
		}
		return []*ls.Layer{l}, nil
	}
	// If input has "nodes", it is a JSON graph
	if m, ok := v.(map[string]interface{}); ok {
		if _, exists := m["nodes"]; exists {
			target := ls.NewLayerGraph()
			if err := ls.NewJSONMarshaler(interner).Unmarshal(input, target); err != nil {
				return nil, err
			}
			layers := ls.LayersFromGraph(target)
			if len(layers) == 0 {
				return nil, fmt.Errorf("No layers in input")
			}
			return layers, nil
		}
	}
	// Try json-ld
	l, err := ls.UnmarshalLayer(v, interner)
	if err != nil {
		return nil, err
	}
	return []*ls.Layer{l}, nil
}

func getRelativeFileName(dir, fname string) string {
	if filepath.IsAbs(fname) {
		return fname
	}
	return filepath.Join(dir, filepath.Base(fname))
}
