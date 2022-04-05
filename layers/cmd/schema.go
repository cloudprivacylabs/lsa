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
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type SchemaOverlays[T OverlayTypeConstraint] struct {
	Schema     string     `json:"schema"`
	JSONSchema JSONSchema `json:"jsonSchema"`
	Overlays   []T        `json:"overlays"`
}

type JSONSchema struct {
	Schema  string `json:"schema"`
	LayerId string `json:"layerId"`
}

type OverlayTypeConstraint interface {
	string | JSONSchema
}

// fn to process schemas, (go through bundles once)
func ReadSchemaBundle[T OverlayTypeConstraint](ctx *ls.Context, file string) (map[string]*ls.Layer, error) {
	var ovl SchemaOverlays[T]
	if err := cmdutil.ReadJSON(file, &ovl); err != nil {
		return nil, err
	}
	layers, err := ovl.Load(ctx, filepath.Dir(file))
	if err != nil {
		return nil, err
	}
	b := ls.BundleByType{}
	bundle := make(map[string]*ls.Layer, 0)
	strLayers := make(map[string]*ls.Layer, 0)
	jsonschLayers := make(map[JSONSchema]*ls.Layer, 0)
	for idx, t := range ovl.Overlays {
		switch any(t).(type) {
		case string:
			strLayers[any(t).(string)] = layers[idx]
		case JSONSchema:
			jsonschLayers[any(t).(JSONSchema)] = layers[idx]
		}
	}
	for _, l := range layers {
		for k, v := range strLayers {
			c, err := b.Add(ctx, k, l, v)
			if err != nil {
				return nil, err
			}
			bundle[k] = c
		}
		for k, v := range jsonschLayers {
			c, err := b.Add(ctx, k.LayerId, l, v)
			if err != nil {
				return nil, err
			}
			bundle[k.LayerId] = c
		}
	}
	return bundle, nil
}

func (sch SchemaOverlays[T]) Load(ctx *ls.Context, relativeDir string) ([]*ls.Layer, error) {
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

	for _, l := range append([]T{any(sch.Schema).(T)}, sch.Overlays...) {
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
	}
	return ret, nil
}

type Bundle[T OverlayTypeConstraint] struct {
	// If types is nonempty, bundle is based on schema types
	Types map[string]SchemaOverlays[T] `json:"types"`
	// If variants is nonempty, bundle is based on variant IDs
	Variants map[string]SchemaOverlays[T] `json:"variants"`
}

func LoadBundle[T OverlayTypeConstraint](ctx *ls.Context, file string) (ls.SchemaLoader, error) {
	var bundle Bundle[T]
	if err := cmdutil.ReadJSON(file, &bundle); err != nil {
		return nil, err
	}
	if len(bundle.Types) == 0 && len(bundle.Variants) == 0 {
		return nil, fmt.Errorf("%s: Empty bundle", file)
	}
	if len(bundle.Types) != 0 && len(bundle.Variants) != 0 {
		return nil, fmt.Errorf("%s: Bundle has both types and variants", file)
	}
	if len(bundle.Types) != 0 {
		b := ls.BundleByType{}
		for typeName, layers := range bundle.Types {
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
