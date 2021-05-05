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

	"github.com/cloudprivacylabs/lsa/pkg/terms"
)

type ComposeOptions struct {
	// While composing an object with layer1 and layer2, if layer2 has
	// attributes missing in layer1, then add those attributes to the
	// result. By default, the result will only have attributes of
	// layer1
	Union bool
}

// Compose schema layers. Directly modifies the target. The source
// must be an overlay.
func (layer *Layer) Compose(options ComposeOptions, vocab terms.Vocabulary, source *Layer) error {
	if source.Type != TermOverlayType {
		return ErrInvalidComposition("Composition source is not an overlay")
	}
	if len(layer.TargetType) > 0 && len(source.TargetType) > 0 {
		compatible := false
		for _, t := range source.TargetType {
			found := false
			for _, x := range layer.TargetType {
				if x == t {
					found = true
					break
				}
			}
			if found {
				compatible = true
				break
			}
		}
		if !compatible {
			return ErrIncompatible{Target: layer.TargetType, Source: source.TargetType}
		}
	}
	err := ComposeTerms(vocab, layer.Root.Values, source.Root.Values)
	if err != nil {
		return err
	}
	// Process source attributes and compose
	source.Root.GetAttributes().Iterate(func(srcAttribute *Attribute) bool {
		// Is source in target?
		targetAttribute, exists := layer.Index[srcAttribute.ID]
		if !exists {
			// The attribute does not exist. If options = Union, we add this attribute to target
			if options.Union {
				if err = layer.addAttribute(srcAttribute); err != nil {
					return false
				}
			}
		} else {
			// Attribute exists in target. Do we have matching paths?
			if attributePathMatch(targetAttribute.GetPath(), srcAttribute.GetPath()) {
				err = targetAttribute.Compose(options, vocab, srcAttribute)
				if err != nil {
					return false
				}
			} else if options.Union {
				if err = layer.addAttribute(srcAttribute); err != nil {
					return false
				}
			}
		}
		return true
	})
	return err
}

func (layer *Layer) addAttribute(source *Attribute) error {
	// Add the attribute and all necessary parents
	srcParent := source.GetParent()
	var tgtParent *Attribute
	for {
		if srcParent == nil {
			tgtParent = layer.Root
			break
		} else {
			// Does parent exist in the layer? If so, attach to that parent
			target, exists := layer.Index[srcParent.ID]
			if exists {
				tgtParent = target
				break
			}
			srcParent = srcParent.GetParent()
		}
	}
	// Here, parent is non-nil
	// If there is a parent, then the parent is either an object or array
	newAttr := source.Clone(tgtParent)
	switch tgt := tgtParent.Type.(type) {
	case *ObjectType:
		if err := tgt.Add(newAttr, layer); err != nil {
			return err
		}
	case *PolymorphicType:
		tgt.Options = append(tgt.Options, newAttr)
	case *CompositeType:
		tgt.Options = append(tgt.Options, newAttr)
	default:
		return ErrInvalidComposition(source.ID)
	}
	return nil
}

// Compose source into this attributes
func (attribute *Attribute) Compose(options ComposeOptions, vocab terms.Vocabulary, source *Attribute) error {
	switch t := attribute.Type.(type) {
	case *ObjectType:
		if _, ok := source.Type.(*ObjectType); !ok {
			return ErrInvalidComposition(source.ID)
		}
	case *ReferenceType:
		if sref, ok := source.Type.(*ReferenceType); ok {
			attribute.Type = sref
		} else {
			return ErrInvalidComposition(source.ID)
		}
	case *ArrayType:
		if sarr, ok := source.Type.(*ArrayType); ok {
			if err := t.Compose(options, vocab, sarr.Attribute); err != nil {
				return err
			}
		} else {
			return ErrInvalidComposition(source.ID)
		}
	case *CompositeType:
		if _, ok := source.Type.(*CompositeType); !ok {
			return ErrInvalidComposition(source.ID)
		}
	case *PolymorphicType:
		if _, ok := source.Type.(*PolymorphicType); !ok {
			return ErrInvalidComposition(source.ID)
		}
	}
	return ComposeTerms(vocab, attribute.Values, source.Values)
}

// ComposeTerms composes the terms in target and source, and writes the result into target
func ComposeTerms(vocab terms.Vocabulary, target, source map[string]interface{}) error {
	for k, v := range source {
		targetTerm, exists := target[k]
		if !exists {
			target[k] = ld.CloneDocument(v)
		} else {
			term := vocab[k]
			if term == nil {
				value, err := DefaultComposeTerm(targetTerm, v)
				if err != nil {
					return err
				}
				target[k] = value
			} else if cmp, ok := term.(terms.Composable); ok {
				value, err := cmp.Compose(targetTerm, v)
				if err != nil {
					return err
				}
				target[k] = value
			} else {
				return ErrInvalidComposition(k)
			}
		}
	}
	return nil
}

// func composeTerm(options ComposeOptions, term string, t1, t2 interface{}) (interface{}, error) {
// 	t := Terms[term]
// 	if t != nil {
// 		if t.Compose != nil {
// 			return t.Compose(options, t1, t2)
// 		}
// 	}
// 	return DefaultComposeTerm(options, t1, t2)
// }

// DefaultComposeTerm is the default term composition algorithm
//
// If t2 is nil, returns copy of t1
// If t1 is nil, returns copy of t2
// If t1 and t2 are lists, append t2 to t1
// If t1 and t2 are sets, combine them
func DefaultComposeTerm(t1, t2 interface{}) (interface{}, error) {
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

// // OverrideComposeTerm overrides t1 with t2
// func OverrideComposeTerm(options ComposeOptions, t1, t2 interface{}) (interface{}, error) {
// 	if t2 == nil {
// 		return ld.CloneDocument(t1), nil
// 	}
// 	return ld.CloneDocument(t2), nil
// }

// Compose the layers to build another layer or schema. If the first
// layer is a schema, the result is a schema. Otherwise, the result is
// another overlay. A schema can only appear as the first element.
func Compose(options ComposeOptions, vocab terms.Vocabulary, layers ...*Layer) (*Layer, error) {
	var composed *Layer
	for i, layer := range layers {
		if i == 0 {
			composed = layer.Clone()
		} else {
			if layer.Type == TermSchemaType {
				return nil, ErrInvalidComposition("Schema must be the base element")
			}
			composed.Compose(options, vocab, layer)
		}
	}
	return composed, nil
}

// Check if the suffix attribute has matching ascendants
func attributePathMatch(path, suffix []*Attribute) bool {
	attributeMatch := func(a, b *Attribute) bool {
		return a.ID == b.ID && a.Type.GetType() == b.Type.GetType()
	}
	if len(suffix) > len(path) {
		return false
	}
	for i := range suffix {
		if !attributeMatch(path[len(path)-1-i], suffix[len(suffix)-1-i]) {
			return false
		}
	}
	return true
}
