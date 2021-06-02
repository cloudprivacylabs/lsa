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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudprivacylabs/lsa/pkg/layers"
	"github.com/piprate/json-gold/ld"
)

type ErrNotFound string

func (e ErrNotFound) Error() string { return "Not found: " + string(e) }

const indexFile = "index.ls"

// Repository implements a filesystem based schema repository under a
// given directory
type Repository struct {
	logger func(string, error)
	root   string
	index  []IndexEntry
	vocab  terms.Vocabulary
}

// New returns a new file repository under the given directory.
func New(root string, vocab terms.Vocabulary, errorLogger func(string, error)) *Repository {
	return &Repository{root: root, vocab: vocab, logger: errorLogger}
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
		if x.ID == id && x.Type == ls.TermSchemaManifestType {
			return x.unmarshaled.(*ls.SchemaManifest)
		}
	}
	return nil
}

func (repo *Repository) GetSchema(id string) *ls.Layer {
	for _, x := range repo.index {
		if x.ID == id && x.Type == ls.TermSchemaType {
			return x.unmarshaled.(*ls.Layer)
		}
	}
	return nil
}

func (repo *Repository) GetOverlay(id string) *ls.Layer {
	for _, x := range repo.index {
		if x.ID == id && x.Type == ls.TermOverlayType {
			return x.unmarshaled.(*ls.Layer)
		}
	}
	return nil
}

func (repo *Repository) GetLayer(id string) *ls.Layer {
	for _, x := range repo.index {
		if x.ID == id && (x.Type == ls.TermOverlayType || x.Type == ls.TermSchemaType) {
			return x.unmarshaled.(*ls.Layer)
		}
	}
	return nil
}

func (repo *Repository) GetSchemaManifestByObjectType(t string) *ls.SchemaManifest {
	for _, x := range repo.index {
		if x.hasType(t) && x.Type == ls.TermSchemaManifestType {
			return x.unmarshaled.(*ls.SchemaManifest)
		}
	}
	return nil
}

func (repo *Repository) GetComposedSchema(id string) (*ls.Layer, error) {
	for i, x := range repo.index {
		if x.ID == id && x.Type == ls.TermSchemaManifestType {
			return repo.compose(i)
		}
	}
	return nil, nil
}

func (repo *Repository) GetComposedSchemaByObjectType(t string) (*ls.Layer, error) {
	for i, x := range repo.index {
		if x.hasType(t) && x.Type == ls.TermSchemaManifestType {
			return repo.compose(i)
		}
	}
	return nil, nil
}

func (repo *Repository) compose(index int) (*ls.Layer, error) {
	if repo.index[index].composed != nil {
		return repo.index[index].composed, nil
	}
	layers := make([]*ls.Layer, 0)
	m := repo.index[index].unmarshaled.(*ls.SchemaManifest)
	if len(m.Schema) > 0 {
		sch := repo.GetSchema(m.Schema)
		if sch == nil {
			return nil, ErrNotFound(m.Schema)
		}
		layers = append(layers, sch)
	}
	for _, x := range m.Overlays {
		ovl := repo.GetLayer(x)
		if ovl == nil {
			return nil, ErrNotFound(x)
		}
		layers = append(layers, ovl)
	}
	result, err := ls.Compose(ls.ComposeOptions{}, repo.vocab, layers...)
	if err != nil {
		return nil, err
	}
	repo.index[index].composed = result
	return result, nil
}

type IndexEntry struct {
	Type       string      `json:"type"`
	ID         string      `json:"id"`
	TargetType []string    `json:"targetType,omitempty"`
	Payload    interface{} `json:"payload"`

	unmarshaled interface{}
	composed    *ls.Layer
}

func (i IndexEntry) hasType(t string) bool {
	for _, x := range i.TargetType {
		if x == t {
			return true
		}
	}
	return false
}

// Load loads the index under the directory. If buildIndexIfStale
// is true, build the index if necessary
func (repo *Repository) Load(buildIndexIfStale bool) error {
	if buildIndexIfStale {
		if repo.IsIndexStale() {
			if err := repo.UpdateIndex(); err != nil {
				return err
			}
		}
	}
	data, err := ioutil.ReadFile(filepath.Join(repo.root, indexFile))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &repo.index); err != nil {
		return err
	}
	for i, x := range repo.index {
		switch x.Type {
		case ls.TermSchemaManifestType:
			if repo.index[i].unmarshaled, err = ls.SchemaManifestFromLD(x.Payload); err != nil {
				return err
			}
		case ls.TermSchemaType, ls.TermOverlayType:
			if repo.index[i].unmarshaled, err = ls.LayerFromLD(x.Payload); err != nil {
				return err
			}
		}
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
	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() != indexFile {
			info, _ := entry.Info()
			if info != nil && info.ModTime().After(t) {
				return true
			}
		}
	}
	return false
}

// UpdateIndex builds and updates the index file
func (repo *Repository) UpdateIndex() error {
	index, err := repo.BuildIndex()
	if err != nil {
		return err
	}
	output := filepath.Join(repo.root, indexFile)
	f, err := os.Create(output)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	return enc.Encode(index)
}

// BuildIndex reads and parses all jsonld files and returns the index
// entries
func (repo *Repository) BuildIndex() ([]IndexEntry, error) {
	entries, err := os.ReadDir(repo.root)
	if err != nil {
		return nil, err
	}
	proc := ld.NewJsonLdProcessor()
	ret := make([]IndexEntry, 0)
	for _, entry := range entries {
		if !entry.IsDir() && entry.Name() != indexFile {
			fname := filepath.Join(repo.root, entry.Name())
			data, err := ioutil.ReadFile(fname)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", fname, err)
			}
			var obj interface{}
			if err := json.Unmarshal(data, &obj); err != nil {
				repo.logError(fname, err)
				continue
			}
			expanded, err := proc.Expand(obj, nil)
			if err != nil {
				repo.logError(fname, err)
				continue
			}
			obj, err = ParseRepositoryObject(expanded)
			if err != nil {
				repo.logError(fname, err)
				continue
			}
			if manifest, ok := obj.(*ls.SchemaManifest); ok {
				entry := IndexEntry{
					Type:       ls.TermSchemaManifestType,
					ID:         manifest.ID,
					TargetType: manifest.TargetType,
					Payload:    expanded,
				}
				ret = append(ret, entry)
			} else if layer, ok := obj.(*ls.Layer); ok {
				entry := IndexEntry{
					Type:       layer.Type,
					ID:         layer.ID,
					TargetType: layer.TargetType,
					Payload:    expanded,
				}
				ret = append(ret, entry)
			} else {
				repo.logError(fname, fmt.Errorf("Cannot read object"))
			}
		}
	}
	return ret, nil
}

func (repo *Repository) logError(fname string, err error) {
	if repo.logger != nil {
		repo.logger(fname, err)
	}
}
