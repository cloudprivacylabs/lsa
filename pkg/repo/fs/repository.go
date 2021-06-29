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
package fs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type ErrNotFound string

func (e ErrNotFound) Error() string { return "Not found: " + string(e) }

const indexFile = "index.ls"

// Repository implements a filesystem based schema repository under a
// given directory
type Repository struct {
	root  string
	index []IndexEntry
}

// New returns a new file repository under the given directory.
func New(root string) *Repository {
	return &Repository{root: root}
}

type IndexEntry struct {
	Type       string `json:"type"`
	ID         string `json:"id"`
	TargetType string `json:"targetType,omitempty"`
	File       string `json:"file"`
}

func (i IndexEntry) hasType(t string) bool {
	return i.TargetType == t
}

var ErrBadIndex = errors.New("Bad index file")
var ErrNoIndex = errors.New("No index file")

// Load loads the index under the directory.
func (repo *Repository) Load() error {
	data, err := ioutil.ReadFile(filepath.Join(repo.root, indexFile))
	if err != nil {
		return ErrNoIndex
	}
	if err := json.Unmarshal(data, &repo.index); err != nil {
		return ErrBadIndex
	}
	return nil
}

// IsIndexStale returns true if the index needs to be rebuilt
func (repo *Repository) IsIndexStale() bool {
	info, err := os.Stat(filepath.Join(repo.root, indexFile))
	if err == os.ErrNotExist || info == nil {
		return true
	}
	t := info.ModTime()
	entries, err := os.ReadDir(repo.root)
	if err != nil {
		return true
	}
	names := make(map[string]struct{})
	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() != indexFile {
			names[entry.Name()] = struct{}{}
			info, _ := entry.Info()
			if info != nil && info.ModTime().After(t) {
				return true
			}
			found := false
			for _, x := range repo.index {
				if x.File == entry.Name() {
					found = true
					break
				}
			}
			if !found {
				return true
			}
		}
	}
	// Any file deleted?
	for _, index := range repo.index {
		if _, exists := names[index.File]; !exists {
			return true
		}
	}
	return false
}

// UpdateIndex builds and updates the index file
func (repo *Repository) UpdateIndex() ([]string, error) {
	index, warnings, err := repo.BuildIndex()
	if err != nil {
		return warnings, err
	}
	output := filepath.Join(repo.root, indexFile)
	f, err := os.Create(output)
	if err != nil {
		return warnings, err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	return warnings, enc.Encode(index)
}

// BuildIndex reads and parses all jsonld files and returns the index
// entries
func (repo *Repository) BuildIndex() ([]IndexEntry, []string, error) {
	warnings := make([]string, 0)
	entries, err := os.ReadDir(repo.root)
	if err != nil {
		return nil, nil, err
	}
	ret := make([]IndexEntry, 0)
	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() != indexFile {
			fname := filepath.Join(repo.root, entry.Name())
			data, err := ioutil.ReadFile(fname)
			if err != nil {
				return nil, warnings, fmt.Errorf("%s: %w", fname, err)
			}
			var obj interface{}
			if err := json.Unmarshal(data, &obj); err != nil {
				warnings = append(warnings, fmt.Sprintf("Cannot load %s: %v", fname, err))
				continue
			}
			var typeName string
			if arr, ok := obj.([]interface{}); ok {
				if len(arr) == 1 {
					if m, ok := arr[0].(map[string]interface{}); ok {
						typeName = m["@type"].(string)
					}
				}
			} else if m, ok := obj.(map[string]interface{}); ok {
				typeName = m["@type"].(string)
			}
			switch typeName {
			case ls.SchemaTerm, ls.OverlayTerm, "Schema", "Overlay":
				layer, err := ls.UnmarshalLayer(obj)
				if err != nil {
					warnings = append(warnings, fmt.Sprintf("Cannot parse %s: %v", fname, err))
					continue
				}
				entry := IndexEntry{
					Type:       layer.GetLayerType(),
					ID:         layer.GetID(),
					TargetType: layer.GetTargetType(),
					File:       entry.Name(),
				}
				ret = append(ret, entry)
			case ls.SchemaManifestTerm, "SchemaManifest":
				manifest, err := ls.UnmarshalSchemaManifest(obj)
				if err != nil {
					warnings = append(warnings, fmt.Sprintf("Cannot parse %s: %v", fname, err))
					continue
				}
				entry := IndexEntry{
					Type:       ls.SchemaManifestTerm,
					ID:         manifest.ID,
					TargetType: manifest.TargetType,
					File:       entry.Name(),
				}
				ret = append(ret, entry)
			}
		}
	}
	return ret, warnings, nil
}

// LoadAndCompose loads the layer or schema manifest with the given
// ID. If the loaded object is a schema manifest, computes the
// composite schema and returns it.
func (repo *Repository) LoadAndCompose(id string) (*ls.Layer, error) {
	layer := repo.GetLayer(id)
	if layer != nil {
		return layer, nil
	}
	return repo.GetComposedSchema(id)
}

func (repo *Repository) GetSchemaManifest(id string) *ls.SchemaManifest {
	for _, x := range repo.index {
		if x.ID == id && x.Type == ls.SchemaManifestTerm {
			return repo.mustLoadSchemaManifest(x.File)
		}
	}
	return nil
}

func (repo *Repository) GetSchema(id string) *ls.Layer {
	for _, x := range repo.index {
		if x.ID == id && x.Type == ls.SchemaTerm {
			return repo.loadLayer(x.File)
		}
	}
	return nil
}

func (repo *Repository) GetOverlay(id string) *ls.Layer {
	for _, x := range repo.index {
		if x.ID == id && x.Type == ls.OverlayTerm {
			return repo.loadLayer(x.File)
		}
	}
	return nil
}

func (repo *Repository) GetLayer(id string) *ls.Layer {
	for _, x := range repo.index {
		if x.ID == id && (x.Type == ls.OverlayTerm || x.Type == ls.SchemaTerm) {
			return repo.loadLayer(x.File)
		}
	}
	return nil
}

func (repo *Repository) GetSchemaManifestByObjectType(t string) *ls.SchemaManifest {
	for _, x := range repo.index {
		if x.hasType(t) && x.Type == ls.SchemaManifestTerm {
			return repo.mustLoadSchemaManifest(x.File)
		}
	}
	return nil
}

func (repo *Repository) readJson(file string) (interface{}, error) {
	data, err := ioutil.ReadFile(filepath.Join(repo.root, file))
	if err != nil {
		return nil, err
	}
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return v, nil
}

func (repo *Repository) mustLoadSchemaManifest(file string) *ls.SchemaManifest {
	ret, err := repo.loadSchemaManifest(file)
	if err != nil {
		panic("Cannot load manifest:" + err.Error())
	}
	return ret
}

func (repo *Repository) loadSchemaManifest(file string) (*ls.SchemaManifest, error) {
	data, err := repo.readJson(file)
	if err != nil {
		return nil, err
	}
	ret, err := ls.UnmarshalSchemaManifest(data)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (repo *Repository) loadLayer(file string) *ls.Layer {
	data, err := repo.readJson(file)
	if err != nil {
		panic("Cannot read " + file)
	}
	ret, err := ls.UnmarshalLayer(data)
	if err != nil {
		panic("Cannot parse layer: " + err.Error())
	}
	return ret
}

func (repo *Repository) GetComposedSchema(id string) (*ls.Layer, error) {
	for _, x := range repo.index {
		if x.ID == id && x.Type == ls.SchemaManifestTerm {
			return repo.compose(x)
		}
	}
	return nil, nil
}

func (repo *Repository) GetComposedSchemaByObjectType(t string) (*ls.Layer, error) {
	for _, x := range repo.index {
		if x.hasType(t) && x.Type == ls.SchemaManifestTerm {
			return repo.compose(x)
		}
	}
	return nil, nil
}

func (repo *Repository) compose(index IndexEntry) (*ls.Layer, error) {
	m, err := repo.loadSchemaManifest(index.File)
	if err != nil {
		return nil, err
	}
	var result *ls.Layer
	if len(m.Schema) > 0 {
		sch := repo.GetSchema(m.Schema)
		if sch == nil {
			return nil, ErrNotFound(m.Schema)
		}
		result = sch
	}
	for _, x := range m.Overlays {
		ovl := repo.GetLayer(x)
		if ovl == nil {
			return nil, ErrNotFound(x)
		}
		if result == nil {
			result = ovl
		} else {
			err := result.Compose(ovl)
			if err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}
