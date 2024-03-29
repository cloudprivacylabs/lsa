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

package transform

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// TransformScript combines transformation related annotations into a
// single unit where they can be reached using schema node ids.
type TransformScript struct {
	// TargetSchemaNodes are keyed by schema node ID, and contains
	// transformation elements for each target schema node
	//
	// The schema node ID can be a suffix of the schema node path,
	// elements separated by space
	TargetSchemaNodes map[string]NodeTransformAnnotations `json:"reshapeNodes,omitempty" yaml:"reshapeNodes,omitempty"`

	// Keyed by the last part of the target schema node key
	compiledTargetSchemaNodes map[string][]*compiledAnnotations
}

type compiledAnnotations struct {
	path        []string
	annotations NodeTransformAnnotations
}

func (c compiledAnnotations) pathMatch(schemaPath []*lpg.Node) bool {
	if len(schemaPath) < len(c.path) {
		return false
	}
	s := schemaPath[len(schemaPath)-len(c.path):]
	for i := range s {
		if ls.GetNodeID(s[i]) != c.path[i] {
			return false
		}
	}
	return true
}

func (t *TransformScript) getMatching(schemaPath []*lpg.Node) *compiledAnnotations {
	last := ls.GetNodeID(schemaPath[len(schemaPath)-1])
	annotations := t.compiledTargetSchemaNodes[last]
	if len(annotations) == 0 {
		return nil
	}
	for _, ann := range annotations {
		if ann.pathMatch(schemaPath) {
			return ann
		}
	}
	return nil
}

func (t *TransformScript) Compile(ctx *ls.Context) error {
	if t == nil {
		return nil
	}
	cctx := &ls.CompileContext{}
	t.compiledTargetSchemaNodes = make(map[string][]*compiledAnnotations)
	for sid, ann := range t.TargetSchemaNodes {
		ctx.GetLogger().Debug(map[string]any{"script.compile.schemaNodeId": sid})

		fields := strings.Fields(sid)
		if len(fields) == 0 {
			return fmt.Errorf("Transformation script error: no elements in target schema path")
		}
		termNode := fields[len(fields)-1]
		val := &compiledAnnotations{
			path:        fields,
			annotations: NodeTransformAnnotations{},
		}
		for k, v := range ann {
			pv, ok := v.(ls.PropertyValue)
			if ok {
				if !ls.IsTermRegistered(k) {
					ctx.GetLogger().Info(map[string]any{"script.compile": "Unknown term", "term": k})
				}
				val.annotations[k] = v
				if err := ls.GetTermCompiler(k).CompileTerm(cctx, val.annotations, k, pv); err != nil {
					return err
				}
			}
		}
		t.compiledTargetSchemaNodes[termNode] = append(t.compiledTargetSchemaNodes[termNode], val)
	}

	return nil
}

// NodeTransformAnnotations contains a term, and one or more annotations
type NodeTransformAnnotations map[string]any

func (t *TransformScript) GetProperties(schemaPath []*lpg.Node) ls.CompilablePropertyContainer {
	if t == nil {
		return schemaPath[len(schemaPath)-1]
	}
	ann := t.getMatching(schemaPath)
	if ann == nil {
		return schemaPath[len(schemaPath)-1]
	}
	return ann.annotations
}

// GetSources returns the "source" or "sources" property
func (t *TransformScript) GetSources(schemaPath []*lpg.Node) []string {
	if t == nil {
		return nil
	}
	nd := t.GetProperties(schemaPath)
	if nd == nil {
		return nil
	}
	prop, ok := nd.GetProperty(SourcesTerm.Name)
	if ok {
		if s, ok := prop.([]any); ok {
			ret := make([]string, 0, len(s))
			for _, x := range s {
				ret = append(ret, x.(string))
			}
			return ret
		}
	}
	pv, ok := prop.(ls.PropertyValue)
	if ok {
		s := pv.AsStringSlice()
		if len(s) > 0 {
			return s
		}
	}
	prop, ok = nd.GetProperty(SourceTerm.Name)
	if ok {
		if s, ok := prop.(string); ok {
			return []string{s}
		}
	}
	pv, ok = prop.(ls.PropertyValue)
	if ok {
		if str, ok := pv.Value().(string); ok {
			return []string{str}
		}
	}
	return nil
}

func (t *TransformScript) Validate(targetSchema *ls.Layer) error {
	for key := range t.TargetSchemaNodes {
		for _, fld := range strings.Fields(key) {
			if targetSchema.GetAttributeByID(fld) == nil {
				return ls.ErrNotFound(fld)
			}
		}
	}
	return nil
}

func (nd NodeTransformAnnotations) GetProperty(key string) (any, bool) {
	v, ok := nd[key]
	return v, ok
}

func (nd NodeTransformAnnotations) SetProperty(key string, value any) {
	nd[key] = value
}

func (nd *NodeTransformAnnotations) setProperties(mv map[string]any) {
	set := func(k string, v any) {
		switch value := v.(type) {
		case string:
			(*nd)[k] = ls.NewPropertyValue(k, value)
		case []any:
			sl := make([]string, 0, len(value))
			for _, s := range value {
				sl = append(sl, s.(string))
			}
			(*nd)[k] = ls.NewPropertyValue(k, sl)
		}
	}
	*nd = make(map[string]any)
	for k, v := range mv {
		u, err := url.Parse(k)
		if err != nil || u.IsAbs() {
			set(k, v)
			continue
		}
		u.Scheme = "https"
		u.Path = "lschema.org/transform/" + u.Path
		set(u.String(), v)
	}
}

func (nd *NodeTransformAnnotations) UnmarshalJSON(in []byte) error {
	var mv map[string]any
	if err := json.Unmarshal(in, &mv); err != nil {
		return err
	}
	nd.setProperties(mv)
	return nil
}

func (nd *NodeTransformAnnotations) UnmarshalYAML(parse func(any) error) error {
	var mv map[string]any
	if err := parse(&mv); err != nil {
		return err
	}
	nd.setProperties(mv)
	return nil
}
