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

package mem

import (
	"encoding/json"
	"errors"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// Repository is an in-memory schema repository. It keeps all parsed
// schemas and schema variants
type Repository struct {
	layers   map[string]*ls.Layer
	interner ls.Interner
}

// New returns a new empty repository
func New() *Repository {
	return &Repository{
		layers:   make(map[string]*ls.Layer),
		interner: ls.NewInterner(),
	}
}

// GetLayers returns layers in the repository
func (repo *Repository) GetLayers() []*ls.Layer {
	ret := make([]*ls.Layer, 0, len(repo.layers))
	for _, x := range repo.layers {
		ret = append(ret, x)
	}
	return ret
}

// AddLayer adds a new schema or overlay to the repo. If there is one
// with the same id, the new layer replaces the old one
func (repo *Repository) AddLayer(layer *ls.Layer) {
	repo.layers[layer.GetID()] = layer
}

// ParseAddObject parses the given layer or schema variant and adds
// it to the repository. Returns the parsed object
func (repo *Repository) ParseAddObject(in []byte) (interface{}, error) {
	var m interface{}
	err := json.Unmarshal(in, &m)
	if err != nil {
		return nil, err
	}
	return repo.ParseAddIntf(m)
}

// ParseAddIntf parses and adds the unmarshaled object
func (repo *Repository) ParseAddIntf(m interface{}) (interface{}, error) {
	layer, err1 := ls.UnmarshalLayer(m, repo.interner)
	if err1 != nil {
		return nil, err1
	}
	repo.AddLayer(layer)
	return layer, nil
}

// RemoveObject removes the object(s) with the given id
func (repo *Repository) RemoveObject(ID string) {
	delete(repo.layers, ID)
}

// GetSchema returns a schema with the given id
func (repo *Repository) GetSchema(id string) *ls.Layer {
	l := repo.layers[id]
	if l != nil && l.GetLayerType() == ls.SchemaTerm {
		return l
	}
	return nil
}

// GetOverlay returns an overlay with the given id
func (repo *Repository) GetOverlay(id string) *ls.Layer {
	l := repo.layers[id]
	if l != nil && l.GetLayerType() == ls.OverlayTerm {
		return l
	}
	return nil
}

// GetLayer returns a schema or an overlay with the given id
func (repo *Repository) GetLayer(id string) *ls.Layer {
	return repo.layers[id]
}

// GetComposedSchema returns a composed layer from the schema variant
func (repo *Repository) GetComposedSchema(context *ls.Context, id string) (*ls.Layer, error) {
	data, err := ls.MarshalLayer(repo.GetLayer(id))
	if err != nil {
		return nil, err
	}
	m := struct {
		Schema   string   `json:"schema"`
		Overlays []string `json:"overlays"`
	}{}
	b, err := json.Marshal(data)
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}
	var result *ls.Layer
	if len(m.Schema) > 0 {
		sch := repo.GetSchema(m.Schema)
		if sch == nil {
			return nil, ls.ErrNotFound(m.Schema)
		}
		result = sch.Clone()
	}
	for _, x := range m.Overlays {
		ovl := repo.GetLayer(x)
		if ovl == nil {
			return nil, ls.ErrNotFound(x)
		}
		if result == nil {
			result = ovl.Clone()
		} else {
			err := result.Compose(context, ovl)
			if err != nil {
				return nil, err
			}
		}
	}
	if result == nil {
		return nil, errors.New("empty schema")
	}
	return result, nil
}
