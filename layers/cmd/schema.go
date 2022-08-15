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
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/bundle"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type bundlesSchemaLoader struct {
	bundles []*bundle.Bundle
	ctx     *ls.Context
}

func (b bundlesSchemaLoader) LoadSchema(ref string) (*ls.Layer, error) {
	var ret *ls.Layer
	for _, bnd := range b.bundles {
		l, err := bnd.GetLayer(b.ctx, ref)
		if err != nil {
			return nil, err
		}
		if l != nil {
			if ret != nil {
				return nil, fmt.Errorf("Duplicate definition of %s", ref)
			}
			ret = l
		}
	}
	if ret == nil {
		return nil, ls.ErrNotFound(ref)
	}
	return ret, nil
}

func recalculatePaths(bnd *bundle.Bundle, dir string) {
	if len(bnd.Base) > 0 {
		bnd.Base = getRelativeFileName(dir, bnd.Base)
	}
	for i := range bnd.Spreadsheets {
		bnd.Spreadsheets[i].Name = getRelativeFileName(dir, bnd.Spreadsheets[i].Name)
	}
	for i := range bnd.JSONSchemas {
		bnd.JSONSchemas[i].Name = getRelativeFileName(dir, bnd.JSONSchemas[i].Name)
		for j := range bnd.JSONSchemas[i].Overlays {
			bnd.JSONSchemas[i].Overlays[j] = getRelativeFileName(dir, bnd.JSONSchemas[i].Overlays[j])
		}
	}
	for _, v := range bnd.Variants {
		if len(v.SchemaRef.Schema) > 0 {
			v.SchemaRef.Schema = getRelativeFileName(dir, v.SchemaRef.Schema)
		}
		if v.SchemaRef.JSONSchema != nil {
			v.SchemaRef.JSONSchema.Ref = getRelativeFileName(dir, v.SchemaRef.JSONSchema.Ref)
		}
		for ovl := range v.Overlays {
			if len(v.Overlays[ovl].Schema) > 0 {
				v.Overlays[ovl].Schema = getRelativeFileName(dir, v.Overlays[ovl].Schema)
			}
			if v.Overlays[ovl].JSONSchema != nil {
				v.Overlays[ovl].JSONSchema.Ref = getRelativeFileName(dir, v.Overlays[ovl].JSONSchema.Ref)
			}
		}
	}
}

func LoadBundle(ctx *ls.Context, file []string) (ls.SchemaLoader, error) {
	ctx.GetLogger().Debug(map[string]interface{}{"loadBundle": file})
	bundles := make([]*bundle.Bundle, 0, len(file))
	for _, f := range file {
		b, err := bundle.LoadBundle(f, func(parentBundle, loadBundle string) (bundle.Bundle, error) {
			var bnd bundle.Bundle
			if err := cmdutil.ReadJSONOrYAML(f, &bnd); err != nil {
				return bnd, err
			}
			recalculatePaths(&bnd, filepath.Dir(loadBundle))
			return bnd, nil
		})
		if err != nil {
			return nil, fmt.Errorf("While reading %s: %w", f, err)
		}
		if err := b.Build(ctx, func(ctx *ls.Context, fname string) ([][][]string, error) {
			return cmdutil.ReadSheets(fname)
		}, func(ctx *ls.Context, fname string) (io.ReadCloser, error) {
			return os.Open(fname)
		}, func(ctx *ls.Context, fname string) (*ls.Layer, error) {
			data, err := os.ReadFile(fname)
			if err != nil {
				return nil, err
			}
			var v interface{}
			err = json.Unmarshal(data, &v)
			if err != nil {
				return nil, err
			}
			return ls.UnmarshalLayer(v, ctx.GetInterner())
		}); err != nil {
			return nil, err
		}
		bundles = append(bundles, &b)
	}
	return bundlesSchemaLoader{bundles: bundles, ctx: ctx}, nil
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
	// fname can be a URL
	u, err := url.Parse(fname)
	if err == nil {
		if len(u.Scheme) > 0 {
			return fname
		}
	}
	if filepath.IsAbs(fname) {
		return fname
	}
	return filepath.Join(dir, fname)
}
