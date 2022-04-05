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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	jsonsch "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// type SchemaOverlays[T OverlayTypeConstraint] struct {
// 	Schema     string     `json:"schema"`
// 	JSONSchema JSONSchema `json:"jsonSchema"`
// 	Overlays   []T        `json:"overlays"`
// }

// type JSONSchema struct {
// 	Schema  string `json:"schema"`
// 	LayerId string `json:"layerId"`
// }

// type OverlayTypeConstraint interface {
// 	string | JSONSchema
// }

// --- >
// Bundle defines type names for variants so references can be resolved
type Bundle struct {
	ID        string                   `json:"id" bson:"id" yaml:"id" bson:"id"`
	CreatedAt time.Time                `json:"createdAt" bson:"createdAt" yaml:"createdAt"`
	UpdatedAt time.Time                `json:"updatedAt" bson:"updatedAt" yaml:"updatedAt"`
	Variants  map[string]BundleVariant `json:"variants" yaml:"variants"`
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

// ParseBundle parses a bundle from JSON or YAML input
func ParseBundle(text string, contentType string) (*Bundle, error) {
	var ret Bundle
	if err := cmdutil.ReadJSON(contentType, &ret); err != nil {
		return nil, err
	}
	return &ret, nil
}

func (bundle *Bundle) importJSONSchema(ctx context.Context, loader ls.SchemaLoader, importEntities []jsonsch.Entity) (map[string]*ls.Layer, error) {
	compiler := jsonschema.NewCompiler()
	compiler.LoadURL = func(s string) (io.ReadCloser, error) {
		obj, err := loader.LoadSchema(s)
		if err != nil {
			return nil, err
		}
		return ioutil.NopCloser(strings.NewReader(obj.GetID())), nil
	}
	// Import all JSON schemas into a graph
	jsonLayers := make(map[string]*ls.Layer)
	if len(importEntities) > 0 {
		compiled, err := jsonsch.CompileEntitiesWith(compiler, importEntities...)
		if err != nil {
			return nil, err
		}
		g := ls.NewLayerGraph()
		layers, err := jsonsch.BuildEntityGraph(g, ls.SchemaTerm, jsonsch.LinkRefsByValueType, compiled...)
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
func (bundle *Bundle) GetLayers(ctx context.Context, loader ls.SchemaLoader) (map[string]*ls.Layer, error) {
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
			b := ls.BundleByID{}
			layer, err := b.LoadSchema(ref.LayerID)
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
	for variantType, variant := range bundle.Variants {
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
	importJson := func(input map[string]jsonsch.Entity) error {
		importEntities := make([]jsonsch.Entity, 0, len(input))
		for _, entity := range input {
			importEntities = append(importEntities, entity)
		}
		if len(importEntities) > 0 {
			jlayers, err := bundle.importJSONSchema(ctx, loader, importEntities)
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

	if err := importJson(schemaEntities); err != nil {
		return nil, err
	}
	if err := importJson(ovlEntities); err != nil {
		return nil, err
	}

	lsctx := ls.NewContext(ctx)
	resultBundle := ls.BundleByType{}
	for variantType, variant := range bundle.Variants {
		sch := layers[variant.GetLayerID()]
		ovl := make([]*ls.Layer, 0, len(variant.Overlays))
		for _, o := range variant.Overlays {
			ovl = append(ovl, layers[o.GetLayerID()])
		}
		_, err := resultBundle.Add(lsctx, variantType, sch, ovl...)
		if err != nil {
			return nil, err
		}
	}
	ret := make(map[string]*ls.Layer)
	for variantId := range bundle.Variants {
		layer, _ := resultBundle.LoadSchema(variantId)
		if layer != nil {
			ret[variantId] = layer
		}
	}
	return ret, nil
}

// --->

// fn to process schemas, (go through bundles once)
// func ReadSchemaBundle[T OverlayTypeConstraint](ctx *ls.Context, file string) (map[string]*ls.Layer, error) {
// 	var ovl SchemaOverlays[T]
// 	if err := cmdutil.ReadJSON(file, &ovl); err != nil {
// 		return nil, err
// 	}
// 	layers, err := ovl.Load(ctx, filepath.Dir(file))
// 	if err != nil {
// 		return nil, err
// 	}
// 	b := ls.BundleByType{}
// 	bundle := make(map[string]*ls.Layer, 0)
// 	strLayers := make(map[string]*ls.Layer, 0)
// 	jsonschLayers := make(map[JSONSchema]*ls.Layer, 0)
// 	for idx, t := range ovl.Overlays {
// 		switch any(t).(type) {
// 		case string:
// 			strLayers[any(t).(string)] = layers[idx]
// 		case JSONSchema:
// 			jsonschLayers[any(t).(JSONSchema)] = layers[idx]
// 		}
// 	}
// 	for _, l := range layers {
// 		for k, v := range strLayers {
// 			c, err := b.Add(ctx, k, l, v)
// 			if err != nil {
// 				return nil, err
// 			}
// 			bundle[k] = c
// 		}
// 		for k, v := range jsonschLayers {
// 			c, err := b.Add(ctx, k.LayerId, l, v)
// 			if err != nil {
// 				return nil, err
// 			}
// 			bundle[k.LayerId] = c
// 		}
// 	}
// 	return bundle, nil
// }

func (bv *BundleVariant) Load(ctx *ls.Context, relativeDir string) ([]*ls.Layer, error) {
	loadFile := func(f string) ([]byte, error) {
		var fname string
		if filepath.IsAbs(f) {
			fname = f
		} else {
			fname = filepath.Join(relativeDir, f)
		}
		return cmdutil.ReadURL(fname)
	}
	ret := make([]*ls.Layer, 0)

	idx := 0
	for _, l := range append([]string{bv.LayerID}, bv.Overlays[idx].GetLayerID()) {
		data, err := loadFile(any(l).(string))
		if err != nil {
			return nil, fmt.Errorf("While loading %s: %w", l, err)
		}
		layers, err := ReadLayers(data, ctx.GetInterner())
		if err != nil {
			return nil, fmt.Errorf("While loading %s: %w", l, err)
		}
		if len(layers) > 1 {
			return nil, fmt.Errorf("Multiple layers in input %s: %s", relativeDir, l)
		}
		ret = append(ret, layers[0])
		idx++
	}
	return ret, nil
}

// type Bundle[T OverlayTypeConstraint] struct {
// 	// If types is nonempty, bundle is based on schema types
// 	Types map[string]SchemaOverlays[T] `json:"types"`
// 	// If variants is nonempty, bundle is based on variant IDs
// 	Variants map[string]SchemaOverlays[T] `json:"variants"`
// }

func LoadBundle(ctx *ls.Context, file string) (ls.SchemaLoader, error) {
	var bundle Bundle
	if err := cmdutil.ReadJSON(file, &bundle); err != nil {
		return nil, err
	}
	if len(bundle.ID) == 0 && len(bundle.Variants) == 0 {
		return nil, fmt.Errorf("%s: Empty bundle", file)
	}

	if len(bundle.Variants) != 0 {
		b := ls.BundleByType{}
		for typeName, layers := range bundle.Variants {
			items, err := layers.Load(ctx, filepath.Dir(file))
			if err != nil {
				return nil, err
			}
			_, err = b.Add(ctx, typeName, items[0], items[1:]...)
			if err != nil {
				return nil, err
			}
		}
		return &b, nil
	}

	b := ls.BundleByID{}
	for id, layers := range bundle.Variants {
		items, err := layers.Load(ctx, filepath.Dir(file))
		if err != nil {
			return nil, err
		}
		_, err = b.Add(ctx, id, items[0], items[1:]...)
		if err != nil {
			return nil, err
		}
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
