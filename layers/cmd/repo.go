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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/piprate/json-gold/ld"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type SchemaRepository struct {
	sync.Mutex

	schemaIDs map[string]*ls.Schema
	objects   map[string]*ls.Schema
	layers    map[string]*ls.SchemaLayer

	resolvedSchemasID         map[string]*ls.SchemaLayer
	resolvedSchemasObjectType map[string]*ls.SchemaLayer
}

func (s *SchemaRepository) init() {
	s.Lock()
	defer s.Unlock()
	if s.schemaIDs == nil {
		s.schemaIDs = make(map[string]*ls.Schema)
	}
	if s.objects == nil {
		s.objects = make(map[string]*ls.Schema)
	}
	if s.layers == nil {
		s.layers = make(map[string]*ls.SchemaLayer)
	}
	if s.resolvedSchemasID == nil {
		s.resolvedSchemasID = make(map[string]*ls.SchemaLayer)
	}
	if s.resolvedSchemasObjectType == nil {
		s.resolvedSchemasObjectType = make(map[string]*ls.SchemaLayer)
	}
}

func parseSchemaObject(obj interface{}) (interface{}, error) {
	proc := ld.NewJsonLdProcessor()
	expanded, err := proc.Expand(obj, nil)
	if err != nil {
		return nil, err
	}
	l := ls.SchemaLayer{}
	s := ls.Schema{}
	err = l.UnmarshalExpanded(expanded)
	if err != nil {
		if err2 := s.UnmarshalExpanded(expanded); err2 != nil {
			return nil, fmt.Errorf("Unrecognized object type %v %v", err, err2)
		} else {
			return &s, nil
		}
	}
	return &l, nil
}

func (s *SchemaRepository) Add(obj interface{}) error {
	s.init()
	s.Lock()
	defer s.Unlock()
	sch, err := parseSchemaObject(obj)
	if err != nil {
		return err
	}
	switch k := sch.(type) {
	case *ls.Schema:
		s.schemaIDs[k.ID] = k
		s.objects[k.ObjectType] = k
	case *ls.SchemaLayer:
		s.layers[k.ID] = k
	}
	return nil
}

func (s *SchemaRepository) AddFile(file ...string) error {
	for _, x := range file {
		data, err := ioutil.ReadFile(x)
		if err != nil {
			return fmt.Errorf("%s: %w", x, err)
		}
		var obj interface{}
		if err := json.Unmarshal(data, &obj); err != nil {
			return fmt.Errorf("%s: %w", x, err)
		}
		if err := s.Add(obj); err != nil {
			return fmt.Errorf("%s: %w", x, err)
		}
	}
	return nil
}

func (s *SchemaRepository) LoadDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	wg := sync.WaitGroup{}
	for _, entry := range entries {
		if !entry.IsDir() {
			wg.Add(1)
			go func(name string) {
				defer wg.Done()
				if err := s.AddFile(filepath.Join(dir, name)); err != nil {
					log.Println(err)
				}
			}(entry.Name())
		}
	}
	wg.Wait()
	return nil
}

func (s *SchemaRepository) GetSchemaByID(id string) *ls.Schema {
	s.init()
	return s.schemaIDs[id]
}

func (s *SchemaRepository) GetSchemaByType(t string) *ls.Schema {
	s.init()
	return s.objects[t]
}

func (s *SchemaRepository) GetLayer(id string) *ls.SchemaLayer {
	s.init()
	return s.layers[id]
}

func (s *SchemaRepository) ComposeSchema(schema *ls.Schema) (*ls.SchemaLayer, error) {
	s.init()

	if schema.Composed != nil {
		return schema.Composed, nil
	}

	var composed *ls.SchemaLayer
	for _, ovl := range schema.Layers {
		overlay := s.GetLayer(ovl)
		if overlay == nil {
			return nil, fmt.Errorf("Cannot find layer %s", ovl)
		}
		if composed == nil {
			composed = overlay.Clone()
		} else if err := composed.Compose(ls.ComposeOptions{}, overlay); err != nil {
			return nil, err
		}
	}
	return composed, nil
}

func (s *SchemaRepository) GetComposedSchemaByID(id string) (*ls.Schema, *ls.SchemaLayer, error) {
	s.init()
	sch := s.GetSchemaByID(id)
	if sch == nil {
		return nil, nil, fmt.Errorf("Cannot find schema %s", id)
	}
	i, err := s.ComposeSchema(sch)
	return sch, i, err
}

func (s *SchemaRepository) GetComposedSchemaByType(t string) (*ls.Schema, *ls.SchemaLayer, error) {
	s.init()
	sch := s.GetSchemaByType(t)
	if sch == nil {
		return nil, nil, fmt.Errorf("Cannot find schema for %s", t)
	}
	i, err := s.ComposeSchema(sch)
	return sch, i, err
}

func (s *SchemaRepository) resolveReferencesAttributes(attributes *ls.Attributes) error {
	for i := 0; i < attributes.Len(); i++ {
		attribute := attributes.Get(i)
		err := s.resolveReferencesAttribute(attribute)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SchemaRepository) resolveCompositionsAttributes(attributes *ls.Attributes) error {
	for i := 0; i < attributes.Len(); i++ {
		attribute := attributes.Get(i)
		err := s.resolveCompositionsAttribute(attribute)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SchemaRepository) resolveReferencesAttribute(attribute *ls.Attribute) error {

	if attribute.IsReference() {
		sch, err := s.ResolveSchemaForObjectType(attribute.GetReference())
		if err != nil {
			return fmt.Errorf("Error resolving ref %s", err)
		}
		attribute.MakeObject(&sch.Attributes)
		return nil
	}
	if attribute.IsObject() {
		return s.resolveReferencesAttributes(attribute.GetAttributes())
	}
	if attribute.IsArray() {
		return s.resolveReferencesAttribute(attribute.GetArrayItems())
	}
	if attribute.IsComposition() {
		for _, item := range attribute.GetCompositionOptions() {
			if err := s.resolveReferencesAttribute(item); err != nil {
				return err
			}
		}
	}
	if attribute.IsPolymorphic() {
		for _, item := range attribute.GetPolymorphicOptions() {
			if err := s.resolveReferencesAttribute(item); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *SchemaRepository) resolveCompositionsAttribute(attribute *ls.Attribute) error {
	if attribute.IsObject() {
		return s.resolveCompositionsAttributes(attribute.GetAttributes())
	}
	if attribute.IsArray() {
		return s.resolveCompositionsAttribute(attribute.GetArrayItems())
	}
	if attribute.IsPolymorphic() {
		for _, item := range attribute.GetPolymorphicOptions() {
			if err := s.resolveCompositionsAttribute(item); err != nil {
				return err
			}
		}
	}
	if attribute.IsComposition() {
		if err := attribute.ComposeOptions(); err != nil {
			return err
		}
	}
	return nil
}

func (s *SchemaRepository) ResolveSchemaForObjectType(objType string) (*ls.SchemaLayer, error) {
	s.init()
	obj := s.resolvedSchemasObjectType[objType]
	if obj != nil {
		return obj, nil
	}

	sch, layer, err := s.GetComposedSchemaByType(objType)
	if err != nil {
		return nil, err
	}

	s.resolvedSchemasID[sch.ID] = layer
	s.resolvedSchemasObjectType[sch.ObjectType] = layer
	err = s.resolveCompositionsAttributes(&layer.Attributes)
	if err != nil {
		return nil, err
	}
	err = s.resolveReferencesAttributes(&layer.Attributes)
	if err != nil {
		return nil, err
	}
	return layer, nil
}

func (s *SchemaRepository) ResolveSchemaForID(ID string) (*ls.SchemaLayer, error) {
	s.init()
	obj := s.resolvedSchemasID[ID]
	if obj != nil {
		return obj, nil
	}

	sch, layer, err := s.GetComposedSchemaByID(ID)
	if err != nil {
		return nil, err
	}

	s.resolvedSchemasID[sch.ID] = layer
	s.resolvedSchemasObjectType[sch.ObjectType] = layer
	err = s.resolveCompositionsAttributes(&layer.Attributes)
	if err != nil {
		return nil, err
	}
	err = s.resolveReferencesAttributes(&layer.Attributes)
	if err != nil {
		return nil, err
	}
	return layer, nil
}
