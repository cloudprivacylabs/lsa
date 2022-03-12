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
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

const X_LS = "x-ls"

type ErrCyclicSchema struct {
	Loop []*jsonschema.Schema
}

func (e ErrCyclicSchema) Error() string {
	items := make([]string, 0)
	for _, x := range e.Loop {
		items = append(items, x.Location)
	}
	return "Cycle:" + strings.Join(items, " ")
}

// Entity defines a location in the schema as an entity
type Entity struct {
	// Reference to the  schema
	Ref string `json:"ref" bson:"ref" yaml:"ref"`
	// ID of the entity
	ID string `json:"id" bson:"id" yaml:"id"`
	// ID of the layer that will be generated
	LayerID string `json:"layerId" bson:"layerId" yaml:"layerId"`
}

// FindEntityByRef finds the entity by ref value
func FindEntityByRef(entities []Entity, ref string) *Entity {
	for i, x := range entities {
		if x.Ref == ref {
			return &entities[i]
		}
	}
	return nil
}

// FindEntityByID finds the entity by ID value
func FindEntityByID(entities []Entity, ID string) *Entity {
	for i, x := range entities {
		if x.ID == ID {
			return &entities[i]
		}
	}
	return nil
}

// CompiledEntity contains the JSON schema for the entity
type CompiledEntity struct {
	Entity
	Schema *jsonschema.Schema
}

// EntityLayer contains the layer for the entity
type EntityLayer struct {
	Entity CompiledEntity

	Layer *ls.Layer `json:"-"`
}

// CompileEntities compiles given entities
func CompileEntities(entities ...Entity) ([]CompiledEntity, error) {
	compiler := jsonschema.NewCompiler()
	return CompileEntitiesWith(compiler, entities...)
}

// The meta-schema for annotations
// mem:// is required for WASM
// compilation. Without that, JSON schema compiler tries to resolve
// relative dir and fails.
var annotationsMeta = jsonschema.MustCompileString("mem://annotations.json", `{}`)

type annotationsCompiler struct{}

type annotationExtSchema map[string]interface{}

func (annotationExtSchema) Validate(ctx jsonschema.ValidationContext, v interface{}) error {
	return nil
}

func (annotationsCompiler) Compile(ctx jsonschema.CompilerContext, m map[string]interface{}) (jsonschema.ExtSchema, error) {
	if ext, ok := m[X_LS]; ok {
		if extMap, ok := ext.(map[string]interface{}); ok {
			propertyMap := make(map[string]interface{})
			for k, v := range extMap {
				propertyMap[k] = v
			}
			return annotationExtSchema(propertyMap), nil
		} else if ext != nil {
			return nil, fmt.Errorf(X_LS + " is not an object")
		}
	}
	// nothing to compile, return nil
	return nil, nil
}

// CompileEntitiesWith compiles all entities as a single json schema unit using the given compiler
func CompileEntitiesWith(compiler *jsonschema.Compiler, entities ...Entity) ([]CompiledEntity, error) {
	ret := make([]CompiledEntity, 0, len(entities))
	compiler.ExtractAnnotations = true
	compiler.RegisterExtension(X_LS, annotationsMeta, annotationsCompiler{})
	for _, e := range entities {
		sch, err := compiler.Compile(e.Ref)
		if err != nil {
			return nil, fmt.Errorf("During %s: %w", e.ID, err)
		}
		ret = append(ret, CompiledEntity{Entity: e, Schema: sch})
	}
	return ret, nil
}

type importContext struct {
	entities      []CompiledEntity
	loop          []*jsonschema.Schema
	currentEntity *CompiledEntity
	interner      ls.Interner
}

func (ctx *importContext) checkLoopAndPush(sch *jsonschema.Schema) bool {
	for _, x := range ctx.loop {
		if sch == x {
			return true
		}
	}
	ctx.loop = append(ctx.loop, sch)
	return false
}

func (ctx *importContext) pop() {
	ctx.loop = ctx.loop[:len(ctx.loop)-1]
}

func (ctx *importContext) findEntity(sch *jsonschema.Schema) *CompiledEntity {
	for i := range ctx.entities {
		if ctx.entities[i].Schema == sch {
			return &ctx.entities[i]
		}
	}
	return nil
}

// BuildEntityGraph imports  JSON schemas or overlays
//
// A JSON schema may include many object definitions. This import
// algorithm creates a layer for each entity.
//
// typeTerm should be either ls.SchemaTerm or ls.OverlayTerm
func BuildEntityGraph(typeTerm string, entities ...CompiledEntity) ([]EntityLayer, error) {
	ctx := importContext{entities: entities, interner: ls.NewInterner()}
	ret := make([]EntityLayer, 0, len(ctx.entities))
	for i := range ctx.entities {
		ctx.currentEntity = &ctx.entities[i]

		s := schemaProperty{}
		ctx.loop = make([]*jsonschema.Schema, 0)
		if err := importSchema(&ctx, &s, ctx.currentEntity.Schema); err != nil {
			return nil, err
		}

		imported := EntityLayer{}
		imported.Entity = *ctx.currentEntity
		imported.Layer = ls.NewLayer()

		// Set the layer ID from the entity layer ID
		imported.Layer.SetID(ctx.currentEntity.LayerID)
		// Set the root node ID from the entity ID
		rootNode := imported.Layer.Graph.NewNode(nil, nil)
		ls.SetNodeID(rootNode, ctx.currentEntity.ID)
		// Set the target type of the layer to root node ID
		imported.Layer.SetTargetType(ctx.currentEntity.ID)
		imported.Layer.Graph.NewEdge(imported.Layer.GetLayerRootNode(), rootNode, ls.LayerRootTerm, nil)
		buildSchemaAttrs(ctx.currentEntity.ID, nil, s, imported.Layer, rootNode, ctx.interner)
		ret = append(ret, imported)
	}
	return ret, nil
}

func importSchema(ctx *importContext, target *schemaProperty, sch *jsonschema.Schema) error {
	if ctx.checkLoopAndPush(sch) {
		return ErrCyclicSchema{Loop: ctx.loop}
	}
	defer ctx.pop()

	switch {
	case sch.Ref != nil:
		ref := ctx.findEntity(sch.Ref)
		if ref != nil {
			target.reference = ref.ID
			return nil
		}
		return importSchema(ctx, target, sch.Ref)

	case len(sch.AllOf) > 0:
		for _, x := range sch.AllOf {
			prop := schemaProperty{}
			if err := importSchema(ctx, &prop, x); err != nil {
				return err
			}
			target.allOf = append(target.allOf, prop)
		}

	case len(sch.AnyOf) > 0:
		for _, x := range sch.AnyOf {
			prop := schemaProperty{}
			if err := importSchema(ctx, &prop, x); err != nil {
				return err
			}
			target.allOf = append(target.allOf, prop)
		}

	case len(sch.OneOf) > 0:
		for _, x := range sch.OneOf {
			prop := schemaProperty{}
			if err := importSchema(ctx, &prop, x); err != nil {
				return err
			}
			target.oneOf = append(target.oneOf, prop)
		}

	case len(sch.Properties) > 0:
		target.object = &objectSchema{properties: make(map[string]schemaProperty), required: sch.Required}
		for k, v := range sch.Properties {
			val := schemaProperty{key: k}
			err := importSchema(ctx, &val, v)
			if err != nil {
				return err
			}
			target.object.properties[k] = val
		}
		// TODO: patternProperties, etc
	case sch.Items != nil:
		target.array = &arraySchema{}
		if itemSchema, ok := sch.Items.(*jsonschema.Schema); ok {
			err := importSchema(ctx, &target.array.items, itemSchema)
			if err != nil {
				return err
			}
		} else {
			panic("Multiple item schemas not supported")
		}
	case sch.Items2020 != nil:
		target.array = &arraySchema{}
		err := importSchema(ctx, &target.array.items, sch.Items2020)
		if err != nil {
			return err
		}
	default:
		if len(sch.Types) > 0 {
			target.typ = sch.Types
		}
		target.format = sch.Format
		if len(sch.Enum) > 0 {
			target.enum = sch.Enum
		}
		if len(sch.Constant) > 0 {
			target.enum = sch.Constant
		}
		if sch.Pattern != nil {
			target.pattern = sch.Pattern.String()
		}
		if len(sch.Description) > 0 {
			target.description = sch.Description
		}
		if sch.Default != nil {
			s := fmt.Sprint(sch.Default)
			target.defaultValue = &s
		}
	}
	if ext, ok := sch.Extensions[X_LS]; ok {
		mext, _ := ext.(annotationExtSchema)
		if len(mext) > 0 {
			target.annotations = map[string]interface{}(mext)
		}
	}

	return nil
}
