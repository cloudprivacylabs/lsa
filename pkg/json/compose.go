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

package json

import (
	"fmt"
	"io"

	"github.com/bserdar/jsonom"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// ComposeSchema composes a JSON schema with zero or more
// overlays. The return is a JSON schema. While composing, mathcing
// objects are merged, matching arrays and values are overridden by
// the overlays. Overlays are composed in the order given.
func ComposeSchema(ctx *ls.Context, name string, overlays []string, loader func(*ls.Context, string) (io.ReadCloser, error)) (jsonom.Node, error) {
	ctx.GetLogger().Debug(map[string]interface{}{"jsonSchema.load": name, "stage": "begin"})
	stream, err := loader(ctx, name)
	if err != nil {
		ctx.GetLogger().Debug(map[string]interface{}{"jsonSchema.load": name, "stage": "cannot load", "err": err})
		return nil, fmt.Errorf("Cannot load %s: %w", name, err)
	}
	root, err := jsonom.UnmarshalReader(stream, ctx.GetInterner())
	stream.Close()
	if err != nil {
		return nil, fmt.Errorf("While reading %s: %w", name, err)
	}
	ctx.GetLogger().Debug(map[string]interface{}{"jsonSchema.load": name, "stage": "loaded"})
	for _, overlay := range overlays {
		ovlStream, err := loader(ctx, overlay)
		if err != nil {
			ctx.GetLogger().Debug(map[string]interface{}{"jsonSchema.load": name, "stage": "compose", "overlay": overlay, "err": err})
			return nil, fmt.Errorf("Cannot load %s: %w", overlay, err)
		}
		ctx.GetLogger().Debug(map[string]interface{}{"jsonSchema.load": name, "stage": "compose", "overlay": overlay})
		ovlNode, err := jsonom.UnmarshalReader(ovlStream, ctx.GetInterner())
		ovlStream.Close()
		if err != nil {
			ctx.GetLogger().Debug(map[string]interface{}{"jsonSchema.load": name, "stage": "compose", "overlay": overlay, "err": err})
			return nil, fmt.Errorf("While reading %s: %w", overlay, err)
		}
		root, err = composeSchema(root, ovlNode)
		if err != nil {
			return nil, fmt.Errorf("While composing %s: %w", overlay, err)
		}
	}
	return root, nil
}

func composeSchema(root, ovl jsonom.Node) (jsonom.Node, error) {
	if root == nil {
		return ovl, nil
	}
	if ovl == nil {
		return root, nil
	}
	switch ovlNode := ovl.(type) {
	case *jsonom.Object:
		rootObj, ok := root.(*jsonom.Object)
		if ok {
			return composeObjects(rootObj, ovlNode)
		}
	case *jsonom.Array:
		rootArr, ok := root.(*jsonom.Array)
		if ok {
			return composeArrays(rootArr, ovlNode)
		}
	case *jsonom.Value:
		rootValue, ok := root.(*jsonom.Value)
		if ok {
			return composeValues(rootValue, ovlNode)
		}
	}
	return ovl, nil
}

func composeObjects(root, ovl *jsonom.Object) (*jsonom.Object, error) {
	if ovl.Len() == 0 {
		return root, nil
	}
	for i := 0; i < ovl.Len(); i++ {
		k := ovl.N(i)
		rootKV := root.Get(k.Key())
		if rootKV == nil {
			root.AddOrSet(k)
			continue
		}
		node, err := composeSchema(rootKV.Value(), k.Value())
		if err != nil {
			return nil, err
		}
		root.Set(k.Key(), node)
	}
	return root, nil
}

func composeArrays(root, ovl *jsonom.Array) (*jsonom.Array, error) {
	if ovl == nil {
		return root, nil
	}
	return ovl, nil
}

func composeValues(root, ovl *jsonom.Value) (*jsonom.Value, error) {
	if ovl == nil {
		return root, nil
	}
	return ovl, nil
}
