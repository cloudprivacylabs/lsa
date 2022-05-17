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

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
)

// TransformScript combines transformation related annotations into a
// single unit where they can be reached using schema node ids.
type TransformScript struct {
	// TargetSchemaNodes are keyed by schema node ID, and contains
	// transformation elements for each target schema node
	TargetSchemaNodes map[string]NodeTransformAnnotations `json:"targetSchemaNodes,omitempty" yaml:"targetSchemaNodes,ommitempty"`

	// Map specifies source schema nodes that map to one or more target schema nodes
	Map []NodeMapping `json:"map,omitempty" yaml:"map,omitempty"`
}

type NodeMapping struct {
	SourceNodeID string `json:"source" yaml:"source"`
	TargetNodeID string `json:"target" yaml:"target"`
}

func (t *TransformScript) Compile() error {
	if t == nil {
		return nil
	}
	for _, ann := range t.TargetSchemaNodes {
		for k, v := range ann {
			pv, ok := v.(*ls.PropertyValue)
			if ok {
				if err := ls.GetTermCompiler(k).CompileTerm(ann, k, pv); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (t *TransformScript) GetMappingByTarget(target string) *NodeMapping {
	if t == nil {
		return nil
	}
	for _, m := range t.Map {
		if m.TargetNodeID == target {
			m := m
			return &m
		}
	}
	return nil
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

func (nd *NodeTransformAnnotations) UnmarshalJSON(in []byte) error {
	var mv map[string]*ls.PropertyValue
	if err := json.Unmarshal(in, &mv); err != nil {
		return err
	}
	*nd = make(map[string]interface{})
	for k, v := range mv {
		(*nd)[k] = v
	}
	return nil
}

func (nd *NodeTransformAnnotations) UnmarshalYAML(parse func(interface{}) error) error {
	var mv map[string]*ls.PropertyValue
	if err := parse(&mv); err != nil {
		return err
	}
	*nd = make(map[string]interface{})
	for k, v := range mv {
		(*nd)[k] = v
	}
	return nil
}
