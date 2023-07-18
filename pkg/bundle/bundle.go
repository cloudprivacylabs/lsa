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

package bundle

import (
	"fmt"
	"io"

	"github.com/bserdar/jsonom"

	jsonsch "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/json/jsonschema"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// Bundle defines type names for variants so references can be resolved
type Bundle struct {
	Base         string                 `json:"base" yaml:"base"`
	Spreadsheets []SpreadsheetReference `json:"spreadsheets" yaml:"spreadsheets"`
	JSONSchemas  []JSONSchema           `json:"jsonSchemas" yaml:"jsonSchemas"`
	Variants     map[string]*Variant    `json:"variants" yaml:"variants"`

	// Layers, keyed by layer ID
	Layers map[string]*ls.Layer

	// Variant schemas, keyed by variant name
	variants map[string]*ls.Layer

	// Loaded JSON schemas, keyed by schema ID
	jsonSchemas map[string]jsonom.Node
}

type bundleContext struct {
	*ls.Context
	// Schema name -> layer map
	schemaNameMap map[string]*ls.Layer
	// Schema name -> schema map
	jsonSchemas map[string]jsonom.Node

	// Layer ID -> entity
	entitiesJSMap map[string]jsonsch.Entity
	// Layer ID -> entity
	overlaysJSMap map[string]jsonsch.Entity

	jsonLoader  func(*ls.Context, string) (io.ReadCloser, error)
	layerLoader func(*ls.Context, string) (*ls.Layer, error)
}

// Merge bundle into b
func (b *Bundle) Merge(bundle Bundle) {
	if b.Variants == nil {
		b.Variants = make(map[string]*Variant)
	}
	b.Spreadsheets = append(b.Spreadsheets, bundle.Spreadsheets...)
	b.JSONSchemas = append(b.JSONSchemas, bundle.JSONSchemas...)
	for typeName, variant := range bundle.Variants {
		existingVariant, ok := b.Variants[typeName]
		if !ok || existingVariant == nil {
			b.Variants[typeName] = variant
			continue
		}
		existingVariant.Merge(*variant)
	}
}

func (b *Bundle) getLayers() map[string]*ls.Layer {
	if b.Layers == nil {
		b.Layers = make(map[string]*ls.Layer)
	}
	return b.Layers
}

func (b *Bundle) addLayer(id string, layer *ls.Layer) error {
	m := b.getLayers()
	if _, exists := m[id]; exists {
		return ls.ErrDuplicate(id)
	}
	m[id] = layer
	return nil
}

func (b *Bundle) loadSpreadsheets(ctx *ls.Context, sheetLoader func(*ls.Context, string) ([][][]string, error)) error {
	for _, x := range b.Spreadsheets {
		m, err := x.Import(ctx, sheetLoader)
		if err != nil {
			return fmt.Errorf("While loading spreadsheet %s: %w", x.Name, err)
		}
		for k, v := range m {
			if err := b.addLayer(k, v); err != nil {
				return fmt.Errorf("In %s: %w", x.Name, err)
			}
		}
	}
	return nil
}

func (bundle *Bundle) loadJSONSchemas(ctx *ls.Context, jsonLoader func(*ls.Context, string) (io.ReadCloser, error)) error {
	if bundle.jsonSchemas == nil {
		bundle.jsonSchemas = make(map[string]jsonom.Node)
	}
	for _, sch := range bundle.JSONSchemas {
		root, err := sch.Load(ctx, jsonLoader)
		if err != nil {
			return fmt.Errorf("Cannot load JSON schema %s: %w", sch.Name, err)
		}
		if _, exists := bundle.jsonSchemas[sch.ID]; exists {
			return ls.ErrDuplicate(sch.ID)
		}
		bundle.jsonSchemas[sch.ID] = root
	}
	return nil
}

func (bundle *Bundle) loadVariants(ctx *ls.Context, layerLoader func(*ls.Context, string) (*ls.Layer, error)) error {
	// Load all jsonld variants
	layersByRef := make(map[string]*ls.Layer)
	loadBySchema := func(schemaRef, layerID string) (string, error) {
		// Did we already load this one?
		layer, exists := layersByRef[schemaRef]
		if exists {
			if len(layerID) != 0 && layerID != layer.GetID() {
				return "", fmt.Errorf("Layer %s has id %s, but defined with id %s", schemaRef, layer.GetID(), layerID)
			}
			return layer.GetID(), nil
		}
		layer, err := layerLoader(ctx, schemaRef)
		if err != nil {
			return "", err
		}
		if _, exists := bundle.getLayers()[layer.GetID()]; exists {
			return "", fmt.Errorf("Duplicate layer id %s in %s", layer.GetID(), schemaRef)
		}
		if len(layerID) != 0 && layerID != layer.GetID() {
			return "", fmt.Errorf("Layer %s has id %s, but defined with id %s", schemaRef, layer.GetID(), layerID)
		}
		layersByRef[schemaRef] = layer
		bundle.getLayers()[layer.GetID()] = layer
		return layer.GetID(), nil
	}
	for _, variant := range bundle.Variants {
		if len(variant.Schema) > 0 {
			// Load the schema
			layerID, err := loadBySchema(variant.Schema, variant.LayerID)
			if err != nil {
				return err
			}
			variant.LayerID = layerID
		}
		for ovl := range variant.Overlays {
			if len(variant.Overlays[ovl].Schema) > 0 {
				// Load the schema
				layerID, err := loadBySchema(variant.Overlays[ovl].Schema, variant.Overlays[ovl].LayerID)
				if err != nil {
					return err
				}
				variant.Overlays[ovl].LayerID = layerID
			}
		}
	}

	// Import json schemas
	for schemaName, _ := range bundle.jsonSchemas {
		// Find all variants using this JSON schema
		schemaEntities := make([]jsonsch.Entity, 0)
		overlayEntities := make([]jsonsch.Entity, 0)
		for variantName, variant := range bundle.Variants {
			if variant.JSONSchema != nil {
				if variant.JSONSchema.GetSchemaBase() == schemaName {
					variant.LayerID = variant.JSONSchema.LayerID
					entity := jsonsch.Entity{
						LayerID:       variant.JSONSchema.LayerID,
						ValueType:     variantName,
						Ref:           variant.JSONSchema.Ref,
						AttrNamespace: variant.JSONSchema.Namespace,
					}
					schemaEntities = append(schemaEntities, entity)
				}
			}
			for ovl := range variant.Overlays {
				if variant.Overlays[ovl].JSONSchema != nil {
					if variant.Overlays[ovl].JSONSchema.GetSchemaBase() == schemaName {
						variant.Overlays[ovl].LayerID = variant.Overlays[ovl].JSONSchema.LayerID
						entity := jsonsch.Entity{
							LayerID:       variant.Overlays[ovl].JSONSchema.LayerID,
							ValueType:     variantName,
							Ref:           variant.Overlays[ovl].JSONSchema.Ref,
							AttrNamespace: variant.Overlays[ovl].JSONSchema.Namespace,
						}
						overlayEntities = append(overlayEntities, entity)
					}
				}
			}
		}
		// Import the schema
		if len(schemaEntities) > 0 {
			if err := bundle.importJSONSchema(ctx, ls.SchemaTerm, schemaEntities); err != nil {
				return err
			}
		}
		if len(overlayEntities) > 0 {
			if err := bundle.importJSONSchema(ctx, ls.OverlayTerm, overlayEntities); err != nil {
				return err
			}
		}
	}
	return nil
}

func (bundle *Bundle) importJSONSchema(ctx *ls.Context, typeTerm string, importEntities []jsonsch.Entity) error {
	compiler := jsonschema.NewCompiler()
	compiler.LoadURL = func(s string) (io.ReadCloser, error) {
		sch, exists := bundle.jsonSchemas[s]
		if !exists {
			return nil, ls.ErrNotFound(s)
		}
		rd, wr := io.Pipe()
		go func() {
			sch.Encode(wr)
			wr.Close()
		}()
		return rd, nil
	}
	if len(importEntities) > 0 {
		compiled, err := jsonsch.CompileEntitiesWith(compiler, importEntities...)
		if err != nil {
			return err
		}
		g := ls.NewLayerGraph()
		layers, err := jsonsch.BuildEntityGraph(g, typeTerm, jsonsch.LinkRefsByValueType, compiled...)
		if err != nil {
			return err
		}
		for _, layer := range layers {
			if err := bundle.addLayer(layer.Layer.GetID(), layer.Layer); err != nil {
				return err
			}
		}
	}
	return nil
}

// Add a new variant to the bundle. The typeName must be unique. If
// empty, schema type will be used. If there are overlays, the variant
// will be built using the schema as the base, so caller must create a
// clone if necessary.
func (bundle *Bundle) Add(ctx *ls.Context, typeName string, schema *ls.Layer, overlays ...*ls.Layer) (*ls.Layer, error) {
	return bundle.add(ctx, typeName, schema, overlays...)
}

func (bundle *Bundle) add(ctx *ls.Context, typeName string, schema *ls.Layer, overlays ...*ls.Layer) (*ls.Layer, error) {
	if bundle.Layers == nil {
		bundle.Layers = make(map[string]*ls.Layer)
	}
	if bundle.variants == nil {
		bundle.variants = make(map[string]*ls.Layer)
	}
	if _, exists := bundle.variants[typeName]; exists {
		return nil, ls.ErrDuplicate(typeName)
	}
	output := schema
	for _, overlay := range overlays {
		if err := output.Compose(ctx, overlay); err != nil {
			return nil, err
		}
	}
	bundle.variants[typeName] = output
	return output, nil
}

func loadBundleChain(parent, base string, bundleLoader func(parentBundle, loadBundle string) (Bundle, error)) (Bundle, error) {
	b, err := bundleLoader(parent, base)
	if err != nil {
		return Bundle{}, fmt.Errorf("Cannot load %s: %w", base, err)
	}
	if len(b.Base) > 0 {
		newBase, err := loadBundleChain(base, b.Base, bundleLoader)
		if err != nil {
			return Bundle{}, err
		}
		b.Merge(newBase)
	}
	return b, nil
}

// LoadBundle loads a bundle and all the bundles it includes. The
// bundleLoader is called with the parent bundle name, and the next
// bundle to load. Parent bundle name can be empty, which means, that
// is the first bundle to load.
func LoadBundle(base string, bundleLoader func(parentBundle, loadBundle string) (Bundle, error)) (Bundle, error) {
	return loadBundleChain("", base, bundleLoader)
}

// NewBundleFromVariants creates a new bundle from the given variants
func NewBundleFromVariants(variantMap map[string]*ls.Layer) Bundle {
	ret := Bundle{
		Variants: make(map[string]*Variant),
		Layers:   make(map[string]*ls.Layer),
		variants: make(map[string]*ls.Layer),
	}
	for v, layer := range variantMap {
		ret.variants[v] = layer
		ret.Variants[v] = &Variant{
			SchemaRef: SchemaRef{
				layer:   layer,
				LayerID: layer.GetID(),
			},
		}
		ret.Layers[layer.GetID()] = layer
	}
	return ret
}

// Build collects all parts of a bundle and builds the layers
func (bundle *Bundle) Build(ctx *ls.Context, spreadsheetLoader func(*ls.Context, string) ([][][]string, error), jsonLoader func(*ls.Context, string) (io.ReadCloser, error), layerLoader func(*ls.Context, string) (*ls.Layer, error)) error {
	err := bundle.loadSpreadsheets(ctx, spreadsheetLoader)
	if err != nil {
		return err
	}
	err = bundle.loadJSONSchemas(ctx, jsonLoader)
	if err != nil {
		return err
	}

	err = bundle.loadVariants(ctx, layerLoader)
	return err
}

func (bundle *Bundle) GetCachedLayers() map[string]*ls.Layer {
	ret := make(map[string]*ls.Layer)
	for k, v := range bundle.getLayers() {
		ret[k] = v
	}
	return ret
}

func (bundle *Bundle) LoadSchema(variant string) (*ls.Layer, error) {
	if bundle.variants == nil {
		bundle.variants = make(map[string]*ls.Layer)
	}
	layer, exists := bundle.variants[variant]
	if exists {
		return layer, nil
	}
	v, exists := bundle.Variants[variant]
	if !exists {
		return nil, nil
	}
	base := bundle.getLayers()[v.LayerID]
	if base == nil {
		return nil, ls.ErrNotFound(fmt.Sprintf("Cannot find base layer for variant %s", variant))
	}

	ovl := make([]*ls.Layer, 0, len(v.Overlays))
	for _, o := range v.Overlays {
		overlay := bundle.getLayers()[o.LayerID]
		if overlay == nil {
			return nil, ls.ErrNotFound(fmt.Sprintf("Cannot find overlay %s for variant %s", o.LayerID, variant))
		}
		ovl = append(ovl, overlay)
	}
	if len(ovl) > 0 {
		base = base.Clone()
	}
	l, err := bundle.add(ls.DefaultContext(), variant, base, ovl...)
	if err != nil {
		return nil, fmt.Errorf("While composing variant: %s: %w", variant, err)
	}
	return l, nil
}

// GetLayer returns the layer for the given variant. Returns nil if
// not found. Panics if bundle is not initialized
func (bundle *Bundle) GetLayer(ctx *ls.Context, variant string) (*ls.Layer, error) {
	if bundle.variants == nil {
		bundle.variants = make(map[string]*ls.Layer)
	}
	layer, exists := bundle.variants[variant]
	if exists {
		return layer, nil
	}
	v, exists := bundle.Variants[variant]
	if !exists {
		return nil, nil
	}
	base := bundle.getLayers()[v.LayerID]
	if base == nil {
		return nil, ls.ErrNotFound(fmt.Sprintf("Cannot find base layer for variant %s", variant))
	}

	ovl := make([]*ls.Layer, 0, len(v.Overlays))
	for _, o := range v.Overlays {
		overlay := bundle.getLayers()[o.LayerID]
		if overlay == nil {
			return nil, ls.ErrNotFound(fmt.Sprintf("Cannot find overlay %s for variant %s", o.LayerID, variant))
		}
		ovl = append(ovl, overlay)
	}
	if len(ovl) > 0 {
		base = base.Clone()
	}
	l, err := bundle.add(ctx, variant, base, ovl...)
	if err != nil {
		return nil, fmt.Errorf("While composing variant: %s: %w", variant, err)
	}
	return l, nil
}
