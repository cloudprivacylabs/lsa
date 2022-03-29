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

// BundleByID stores layers by their ID
type BundleByID struct {
	variants map[string]*Layer
}

// Add a new variant to the bundle. The variant ID must be unique. If
// variant ID is empty, schema ID will be used. If there are overlays,
// the variant will be built using the schema as the base, so caller
// must create a clone if necessary.
func (b *BundleByID) Add(ctx *Context, variantID string, schema *Layer, overlays ...*Layer) (*Layer, error) {
	if b.variants == nil {
		b.variants = make(map[string]*Layer)
	}
	if len(variantID) == 0 {
		variantID = schema.GetID()
	}
	if len(variantID) == 0 {
		return nil, ErrInvalidInput{Msg: "Empty variant ID"}
	}
	if _, exists := b.variants[variantID]; exists {
		return nil, ErrDuplicate(variantID)
	}
	output := schema
	for _, overlay := range overlays {
		if err := output.Compose(ctx, overlay); err != nil {
			return nil, err
		}
	}
	b.variants[variantID] = output
	return output, nil
}

// LoadSchema is a schema loader function using schema IDs as reference
func (b *BundleByID) LoadSchema(ref string) (*Layer, error) {
	l := b.variants[ref]
	if l == nil {
		return nil, ErrNotFound(ref)
	}
	return l, nil
}

// BundleByType stores layers by their value type
type BundleByType struct {
	variants map[string]*Layer
}

// Add a new variant to the bundle. The schema type must be unique. If
// there are overlays, the variant will be built using the schema as
// the base, so caller must create a clone if necessary.
func (b *BundleByType) Add(ctx *Context, schema *Layer, overlays ...*Layer) (*Layer, error) {
	if b.variants == nil {
		b.variants = make(map[string]*Layer)
	}
	vtype := schema.GetValueType()
	if len(vtype) == 0 {
		return nil, ErrInvalidInput{Msg: "Schema has no type"}
	}
	if _, exists := b.variants[vtype]; exists {
		return nil, ErrDuplicate(vtype)
	}
	output := schema
	for _, overlay := range overlays {
		if err := output.Compose(ctx, overlay); err != nil {
			return nil, err
		}
	}
	b.variants[vtype] = output
	return output, nil
}

// SchemaLoader is a schema loader function using types as reference
func (b *BundleByType) LoadSchema(ref string) (*Layer, error) {
	l := b.variants[ref]
	if l == nil {
		return nil, ErrNotFound(ref)
	}
	return l, nil
}
