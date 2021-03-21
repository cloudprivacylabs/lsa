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
	"github.com/piprate/json-gold/ld"
)

type ComposeOptions struct {
	// While composing an object with layer1 and layer2, if layer2 has
	// attributes missing in layer1, then add those attributes to the
	// result. By default, the result will only have attributes of
	// layer1
	Union bool
}

// Compose schema layers. Directly modifies the target.
func (layer *SchemaLayer) Compose(options ComposeOptions, source *SchemaLayer) error {
	if err := layer.Attributes.Compose(options, &source.Attributes); err != nil {
		return err
	}
	return layer.Validate()
}

// Compose source into this attributes
func (attributes *Attributes) Compose(options ComposeOptions, source *Attributes) error {
	// Process target attributes and compose
	for i := 0; i < attributes.Len(); i++ {
		term1Attribute := attributes.Get(i)
		// Is there a matching source attribute? It has to be at this level
		term2Attribute := source.GetByID(term1Attribute.ID)
		if term2Attribute == nil {
			continue
		}
		// Descend the tree together
		if err := term1Attribute.Compose(options, term2Attribute); err != nil {
			return err
		}
	}
	// If source has any attributes that are not in attributes, and if
	// we are using a union composition, copy them over
	if options.Union {
		for i := 0; i < source.Len(); i++ {
			sourceAttribute := source.Get(i)
			if attributes.GetByID(sourceAttribute.ID) == nil {
				newAttribute := sourceAttribute.Clone(attributes)
				attributes.Add(newAttribute)
			}
		}
	}
	return nil
}

// Compose source into this attribute
func (attribute *Attribute) Compose(options ComposeOptions, source *Attribute) error {
	setType := func() {
		attribute.attributes = nil
		attribute.reference = ""
		attribute.arrayItems = nil
		attribute.allOf = nil
		attribute.oneOf = nil
		if source.attributes != nil {
			attribute.attributes = source.attributes.Clone(attribute)
		} else if len(source.reference) != 0 {
			attribute.reference = source.reference
		} else if source.arrayItems != nil {
			attribute.arrayItems = source.arrayItems.Clone(attribute)
		} else if source.allOf != nil {
			attribute.allOf = make([]*Attribute, len(source.allOf))
			for i, x := range source.allOf {
				attribute.allOf[i] = x.Clone(attribute)
			}
		} else if source.oneOf != nil {
			attribute.oneOf = make([]*Attribute, len(source.oneOf))
			for i, x := range source.oneOf {
				attribute.oneOf[i] = x.Clone(attribute)
			}
		}
	}
	compose := func(targetOptions, sourceOptions []*Attribute) ([]*Attribute, error) {
		for _, option := range sourceOptions {
			found := false
			for i := range targetOptions {
				if targetOptions[i].ID == option.ID {
					found = true
					if err := targetOptions[i].Compose(options, option); err != nil {
						return nil, err
					}
					break
				}
			}
			if !found {
				targetOptions = append(targetOptions, option.Clone(attribute))
			}
		}
		return targetOptions, nil
	}
	switch {
	case attribute.IsObject():
		if source.IsObject() {
			if err := attribute.GetAttributes().Compose(options, source.GetAttributes()); err != nil {
				return err
			}
		} else {
			setType()
		}
	case attribute.IsReference():
		if source.IsReference() {
			attribute.reference = source.reference
		} else {
			setType()
		}
	case attribute.IsArray():
		if source.IsArray() {
			if err := attribute.GetArrayItems().Compose(options, source.GetArrayItems()); err != nil {
				return err
			}
		} else {
			setType()
		}
	case attribute.IsComposition():
		if source.IsComposition() {
			var err error
			if attribute.allOf, err = compose(attribute.allOf, source.allOf); err != nil {
				return err
			}
		} else {
			setType()
		}
	case attribute.IsPolymorphic():
		if source.IsPolymorphic() {
			var err error
			if attribute.oneOf, err = compose(attribute.oneOf, source.oneOf); err != nil {
				return err
			}
		} else {
			setType()
		}
	}

	for term, v2 := range source.Values {
		v2Arr, _ := v2.([]interface{})
		if v2Arr == nil {
			continue
		}
		// If a term appears in both source and target, apply term-specific composition
		if v1, exists := attribute.Values[term]; exists {
			v1Arr, _ := v1.([]interface{})
			if v1Arr == nil {
				continue
			}
			result, err := composeTerm(options, term, v1, v2)
			if err != nil {
				return err
			}
			attribute.Values[term] = result
		} else {
			attribute.Values[term] = ld.CloneDocument(v2)
		}
	}
	return nil
}

func composeTerm(options ComposeOptions, term string, t1, t2 interface{}) (interface{}, error) {
	t := Terms[term]
	if t != nil {
		if t.Compose != nil {
			return t.Compose(options, t1, t2)
		}
	}
	return DefaultComposeTerm(options, t1, t2)
}

// DefaultComposeTerm is the default term composition algorithm
//
// If t2 is nil, returns copy of t1
// If t1 is nil, returns copy of t2
// If t1 and t2 are lists, append t2 to t1
// If t1 and t2 are sets, combine them
func DefaultComposeTerm(options ComposeOptions, t1, t2 interface{}) (interface{}, error) {
	if t2 == nil {
		if t1 == nil {
			return nil, nil
		}
		return ld.CloneDocument(t1), nil
	}
	if t1 == nil {
		return ld.CloneDocument(t2), nil
	}

	arr1, t1Arr := t1.([]interface{})
	arr2, t2Arr := t2.([]interface{})
	if !t1Arr || !t2Arr {
		return nil, ErrInvalidObject("Expanded node not array")
	}
	arr1List := len(arr1) == 1 && ld.IsList(arr1[0])
	arr2List := len(arr2) == 1 && ld.IsList(arr2[0])
	if arr1List && arr2List {
		l1 := arr1[0].(map[string]interface{})["@list"].([]interface{})
		l1 = append(l1, arr2[0].(map[string]interface{})["@list"].([]interface{})...)
		return []interface{}{map[string]interface{}{"@list": l1}}, nil
	}
	if (arr1List && !arr2List) || (!arr1List && arr2List) {
		return nil, ErrIncompatibleComposition{Msg: "Composing a list node with a non-list node"}
	}
	// Both nodes are sets
	return append(arr1, arr2...), nil
}

// OverrideComposeTerm overrides t1 with t2
func OverrideComposeTerm(options ComposeOptions, t1, t2 interface{}) (interface{}, error) {
	if t2 == nil {
		return ld.CloneDocument(t1), nil
	}
	return ld.CloneDocument(t2), nil
}
