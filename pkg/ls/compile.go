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
package ls

import (
	"fmt"
)

type Compiler struct {
	// Resolver resolves an ID and returns a strong reference
	Resolver func(string) (string, error)
	// Loader loads a layer using strong reference
	Loader func(string) (*Layer, error)

	compiledSchemas map[string]*Layer
}

// Compile compiles the schema by resolving all references and
// computing all compositions. Compilation process directly modifies
// the schema
func (compiler *Compiler) Compile(ref string) (*Layer, error) {
	if compiler.compiledSchemas == nil {
		compiler.compiledSchemas = make(map[string]*Layer)
	}
	id, err := compiler.Resolver(ref)
	if err != nil {
		return nil, err
	}
	ret := compiler.compiledSchemas[id]
	if ret != nil {
		return ret, nil
	}
	schema, err := compiler.Loader(id)
	if err != nil {
		return nil, err
	}
	if schema == nil {
		return nil, ErrNotFound(ref)
	}
	schema = schema.Clone()
	// Put the compiled schema here, so if there are loops, we can refer to the
	// same object
	compiler.compiledSchemas[id] = schema
	if err := compiler.resolveCompositions(schema.Root.GetAttributes(), schema); err != nil {
		return nil, err
	}
	if err := compiler.resolveReferences(schema.Root.GetAttributes()); err != nil {
		return nil, err
	}
	return schema, nil
}

func (compiler *Compiler) resolveCompositions(attributes *ObjectType, layer *Layer) error {
	for i := 0; i < attributes.Len(); i++ {
		attribute := attributes.Get(i)
		err := compiler.resolveAttributeCompositions(attribute, layer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (compiler *Compiler) resolveAttributeCompositions(attribute *Attribute, layer *Layer) error {
	switch a := attribute.Type.(type) {
	case *ObjectType:
		return compiler.resolveCompositions(a, layer)
	case *ArrayType:
		return compiler.resolveAttributeCompositions(a.Attribute, layer)
	case *PolymorphicType:
		for _, item := range attribute.GetPolymorphicOptions() {
			if err := compiler.resolveAttributeCompositions(item, layer); err != nil {
				return err
			}
		}
	case *CompositeType:
		if err := attribute.ComposeOptions(layer); err != nil {
			return err
		}
	}
	return nil
}

func (compiler *Compiler) resolveReferences(attributes *ObjectType) error {
	for i := 0; i < attributes.Len(); i++ {
		attribute := attributes.Get(i)
		err := compiler.resolveAttributeReferences(attribute)
		if err != nil {
			return err
		}
	}
	return nil
}

func (compiler *Compiler) resolveAttributeReferences(attribute *Attribute) error {
	switch a := attribute.Type.(type) {
	case *ReferenceType:
		sch, err := compiler.Compile(a.Reference)
		if err != nil {
			return fmt.Errorf("Error resolving ref %s: %w", a.Reference, err)
		}
		attribute.Type = sch.Root.GetAttributes()
		return nil
	case *ObjectType:
		return compiler.resolveReferences(a)
	case *ArrayType:
		return compiler.resolveAttributeReferences(a.Attribute)
	case *CompositeType:
		for _, item := range attribute.GetCompositionOptions() {
			if err := compiler.resolveAttributeReferences(item); err != nil {
				return err
			}
		}
	case *PolymorphicType:
		for _, item := range attribute.GetPolymorphicOptions() {
			if err := compiler.resolveAttributeReferences(item); err != nil {
				return err
			}
		}
	}
	return nil
}
