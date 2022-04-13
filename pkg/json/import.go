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
	"net/url"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
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

// Entity defines an entity as a layered schema with respect to a JSON schema
type Entity struct {
	// Reference to the schema. This includes the full reference to the
	// JSON schema, and the reference in that schema that defines the
	// entity. For example, this can be
	// https://somenamespace/myschema.json#/definitions/Object
	Ref string `json:"ref" bson:"ref" yaml:"ref"`
	// ID of the layer that will be generated
	LayerID string `json:"layerId,omitempty" bson:"layerId,omitempty" yaml:"layerId,omitempty"`
	// The ID of the root node. If empty, ValueType is used for the root node id
	RootNodeID string `json:"rootNodeId,omitempty" bson:"rootNodeId,omitempty" yaml:"rootNodeId,omitempty"`
	// ValueType is the value type of the schema, that is, the entity type defined with this schema
	ValueType string `json:"valueType" bson:"valueType" yaml:"valueType"`
}

func (e Entity) GetLayerRoot() string {
	if len(e.RootNodeID) == 0 {
		return e.ValueType
	}
	return e.RootNodeID
}

// LinkRefsBy is an enumeration that specifies how the links for the
// imported schema should be written
type LinkRefsBy int

const (
	// Remote references are schema references (entity.Ref)
	LinkRefsBySchemaRef LinkRefsBy = iota
	// Remote references are layer ids (entity.LayerID)
	LinkRefsByLayerID
	// Remote references are value types (entity.ValueType)
	LinkRefsByValueType
)

// FindEntityByRef finds the entity by ref value
func FindEntityByRef(entities []Entity, ref string) *Entity {
	for i, x := range entities {
		if x.Ref == ref {
			return &entities[i]
		}
	}
	return nil
}

// FindEntityByValueType finds the entity by ValueType value
func FindEntityByValueType(entities []Entity, valueType string) *Entity {
	for i, x := range entities {
		if x.ValueType == valueType {
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
			return nil, fmt.Errorf("During %s: %w", e.Ref, err)
		}
		ret = append(ret, CompiledEntity{Entity: e, Schema: sch})
	}
	return ret, nil
}

type importContext struct {
	entities []CompiledEntity

	currentEntity *CompiledEntity
	interner      ls.Interner
	// Maps schemas to their corresponding objects
	schMap map[string]*schemaProperty
}

func (ctx *importContext) newProp(sch *jsonschema.Schema, prop *schemaProperty) {
	ctx.schMap[sch.Location] = prop
}

func (ctx *importContext) findProp(sch *jsonschema.Schema) *schemaProperty {
	return ctx.schMap[sch.Location]
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
// algorithm creates a layer for each entity in the given target graph.
//
// typeTerm should be either ls.SchemaTerm or ls.OverlayTerm
func BuildEntityGraph(targetGraph graph.Graph, typeTerm string, linkRefsBy LinkRefsBy, entities ...CompiledEntity) ([]EntityLayer, error) {
	ctx := importContext{entities: entities, interner: ls.NewInterner(), schMap: make(map[string]*schemaProperty)}
	ret := make([]EntityLayer, 0, len(ctx.entities))
	for i := range ctx.entities {
		ctx.currentEntity = &ctx.entities[i]

		s, err := importSchema(&ctx, ctx.currentEntity.Schema)
		if err != nil {
			return nil, err
		}

		imported := EntityLayer{}
		imported.Entity = *ctx.currentEntity
		imported.Layer = ls.NewLayerInGraph(targetGraph)
		imported.Layer.SetLayerType(typeTerm)

		// Set the layer ID from the entity layer ID
		imported.Layer.SetID(ctx.currentEntity.LayerID)
		// Set the root node ID from the entity ID
		rootNode := imported.Layer.Graph.NewNode(nil, nil)
		//ls.SetNodeID(rootNode, ctx.currentEntity.ID)
		// Set the value type of the layer to root node ID
		imported.Layer.SetValueType(ctx.currentEntity.ValueType)
		imported.Layer.Graph.NewEdge(imported.Layer.GetLayerRootNode(), rootNode, ls.LayerRootTerm, nil)
		importer := schemaImporter{
			entityId: ctx.currentEntity.GetLayerRoot(),
			layer:    imported.Layer,
			interner: ctx.interner,
			linkRefs: linkRefsBy,
		}
		if err := importer.setNodeProperties(s, rootNode); err != nil {
			return nil, err
		}
		if err := importer.buildChildAttrs(s, rootNode); err != nil {
			return nil, err
		}
		ret = append(ret, imported)
	}
	return ret, nil
}

func importSchema(ctx *importContext, sch *jsonschema.Schema) (*schemaProperty, error) {
	if p := ctx.findProp(sch); p != nil {
		return p, nil
	}
	// Schema node ID is the layer ID + schema location
	target := &schemaProperty{
		ID: sch.Location,
	}
	{
		u, err := url.Parse(sch.Location)
		if err == nil {
			target.ID = u.Fragment
		}
	}

	ctx.newProp(sch, target)
	if sch.Ref != nil {
		ref := ctx.findEntity(sch.Ref)
		if ref != nil {
			target.reference = ref
			return target, nil
		}
		p, err := importSchema(ctx, sch.Ref)
		if err != nil {
			return nil, err
		}
		target.localReference = p
		return target, nil
	}

	ctx.newProp(sch, target)
	switch {
	case len(sch.AllOf) > 0:
		for _, x := range sch.AllOf {
			prop, err := importSchema(ctx, x)
			if err != nil {
				return nil, err
			}
			target.allOf = append(target.allOf, prop)
		}

	case len(sch.AnyOf) > 0:
		for _, x := range sch.AnyOf {
			prop, err := importSchema(ctx, x)
			if err != nil {
				return nil, err
			}
			target.allOf = append(target.allOf, prop)
		}

	case len(sch.OneOf) > 0:
		for _, x := range sch.OneOf {
			prop, err := importSchema(ctx, x)
			if err != nil {
				return nil, err
			}
			target.oneOf = append(target.oneOf, prop)
		}

	case len(sch.Properties) > 0:
		target.object = &objectSchema{properties: make(map[string]*schemaProperty), required: sch.Required}
		for k, v := range sch.Properties {
			val, err := importSchema(ctx, v)
			if err != nil {
				return nil, err
			}
			val.key = k
			target.object.properties[k] = val
		}
		// TODO: patternProperties, etc
	case sch.Items != nil:
		target.array = &arraySchema{}
		var err error
		if itemSchema, ok := sch.Items.(*jsonschema.Schema); ok {
			target.array.items, err = importSchema(ctx, itemSchema)
			if err != nil {
				return nil, err
			}
		} else {
			panic("Multiple item schemas not supported")
		}
	case sch.Items2020 != nil:
		target.array = &arraySchema{}
		var err error
		target.array.items, err = importSchema(ctx, sch.Items2020)
		if err != nil {
			return nil, err
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

	return target, nil
}
