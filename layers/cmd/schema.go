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

type SchemaOverlays struct {
	Schema   string   `json:"schema"`
	Overlays []string `json:"overlays"`
}

func (sch SchemaOverlays) Load(ctx *ls.Context, relativeDir string) ([]*ls.Layer, error) {
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

	for _, l := range append([]string{sch.Schema}, sch.Overlays...) {
		data, err := loadFile(l)
		if err != nil {
			return nil, err
		}
		layers, err := ReadLayers(data, ctx.GetInterner())
		if err != nil {
			return nil, err
		}
		if len(layers) > 1 {
			return nil, fmt.Errorf("Multiple layers in input %s", relativeDir)
		}
		ret = append(ret, layers[0])
	}
	return ret, nil
}

type Bundle struct {
	// If types is nonempty, bundle is based on schema types
	Types map[string]SchemaOverlays `json:"types"`
	// If variants is nonempty, bundle is based on variant IDs
	Variants map[string]SchemaOverlays `json:"variants"`
}

func LoadBundle(ctx *ls.Context, file string) (ls.SchemaLoader, error) {
	var bundle Bundle
	if err := cmdutil.ReadJSON(file, &bundle); err != nil {
		return nil, err
	}
	if len(bundle.Types) == 0 && len(bundle.Variants) == 0 {
		return nil, fmt.Errorf("%s: Empty bundle", file)
	}
	if len(bundle.Types) != 0 && len(bundle.Variants) != 0 {
		return nil, fmt.Errorf("%s: Bundle has both types and variants")
	}
	if len(bundle.Types) != 0 {
		b := ls.BundleByType{}
		for typeName, layers := range bundle.Types {
			items, err := layers.Load(ctx, filepath.Dir(file))
			if err != nil {
				return nil, err
			}
			layer, err := b.Add(ctx, items[0], items[1:]...)
			if err != nil {
				return nil, err
			}
			if layer.GetValueType() != typeName {
				return nil, fmt.Errorf("%s: Type %s has layers of type %s", file, typeName, layer.GetValueType())
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
	if err := json.Unmarshal(input, &v); err == nil {
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
	return nil, fmt.Errorf("Unrecognized input format")
}
