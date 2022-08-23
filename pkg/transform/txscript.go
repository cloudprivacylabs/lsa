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

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
)

// TransformScript combines transformation related annotations into a
// single unit where they can be reached using schema node ids.
type TransformScript struct {
	// TargetSchemaNodes are keyed by schema node ID, and contains
	// transformation elements for each target schema node
	TargetSchemaNodes map[string]NodeTransformAnnotations `json:"reshapeNodes,omitempty" yaml:"reshapeNodes,omitempty"`

	// Map specifies source schema nodes that map to one or more target schema nodes
	Map []NodeMapping `json:"map,omitempty" yaml:"map,omitempty"`

	nodeMappingsByTarget map[string][]*NodeMapping
}

type NodeMapping struct {
	SourceNodeID  string   `json:"source" yaml:"source"`
	SourceNodeIDs []string `json:"sources" yaml:"sources"`
	TargetNodeID  string   `json:"target" yaml:"target"`
	TargetNodeIDs []string `json:"targets" yaml:"targets"`
}

func (m NodeMapping) GetSources() []string {
	if len(m.SourceNodeID) > 0 {
		return []string{m.SourceNodeID}
	}
	return m.SourceNodeIDs
}

func getMappingSources(s []*NodeMapping) []string {
	ret := make([]string, 0)
	for _, x := range s {
		ret = append(ret, x.GetSources()...)
	}
	return ret
}

func (t *TransformScript) Compile(ctx *ls.Context) error {
	if t == nil {
		return nil
	}
	for sid, ann := range t.TargetSchemaNodes {
		ctx.GetLogger().Debug(map[string]interface{}{"script.compile.schemaNodeId": sid})
		for k, v := range ann {
			pv, ok := v.(*ls.PropertyValue)
			if ok {
				if !ls.IsTermRegistered(k) {
					ctx.GetLogger().Info(map[string]interface{}{"script.compile": "Unknown term", "term": k})
				}
				if err := ls.GetTermCompiler(k).CompileTerm(ann, k, pv); err != nil {
					return err
				}
			}
		}
	}

	t.nodeMappingsByTarget = make(map[string][]*NodeMapping)
	for i, m := range t.Map {
		if len(m.SourceNodeID) > 0 && len(m.SourceNodeIDs) > 0 {
			return fmt.Errorf("Both source and sources in mapping")
		}
		if len(m.TargetNodeID) > 0 && len(m.TargetNodeIDs) > 0 {
			return fmt.Errorf("Both target and targets in mapping")
		}
		if len(m.TargetNodeID) > 0 {
			t.nodeMappingsByTarget[m.TargetNodeID] = append(t.nodeMappingsByTarget[m.TargetNodeID], &t.Map[i])
		}
		for _, x := range m.TargetNodeIDs {
			t.nodeMappingsByTarget[x] = append(t.nodeMappingsByTarget[x], &t.Map[i])
		}
	}
	return nil
}

func (t *TransformScript) GetMappingsByTarget(target string) []*NodeMapping {
	if t == nil || t.nodeMappingsByTarget == nil {
		return nil
	}
	return t.nodeMappingsByTarget[target]
}

func (t *TransformScript) GetProperties(schemaNode graph.Node) ls.CompilablePropertyContainer {
	if t == nil {
		return schemaNode
	}
	id := ls.GetNodeID(schemaNode)
	tn, ok := t.TargetSchemaNodes[id]
	if !ok {
		return schemaNode
	}
	return tn
}

func (nd NodeTransformAnnotations) GetProperty(key string) (interface{}, bool) {
	v, ok := nd[key]
	return v, ok
}

func (nd NodeTransformAnnotations) SetProperty(key string, value interface{}) {
	nd[key] = value
}

type SourceAnnotations struct {
}

// NodeTransformAnnotations contains a term, and one or more annotations
type NodeTransformAnnotations map[string]interface{}

func (nd *NodeTransformAnnotations) setProperties(mv map[string]interface{}) {
	*nd = make(map[string]interface{})
	for k, v := range mv {
		u, err := url.Parse(k)
		if err != nil || u.IsAbs() {
			switch value := v.(type) {
			case string:
				(*nd)[k] = ls.StringPropertyValue(k, value)
			case []interface{}:
				sl := make([]string, 0, len(value))
				for _, s := range value {
					sl = append(sl, s.(string))
				}
				(*nd)[k] = ls.StringSlicePropertyValue(k, sl)
			}
			continue
		}
		u.Scheme = "https"
		u.Path = "lschema.org/transform/" + u.Path
		(*nd)[u.String()] = v
	}
}

func (nd *NodeTransformAnnotations) UnmarshalJSON(in []byte) error {
	var mv map[string]interface{}
	if err := json.Unmarshal(in, &mv); err != nil {
		return err
	}
	nd.setProperties(mv)
	return nil
}

func (nd *NodeTransformAnnotations) UnmarshalYAML(parse func(interface{}) error) error {
	var mv map[string]interface{}
	if err := parse(&mv); err != nil {
		return err
	}
	nd.setProperties(mv)
	return nil
}
