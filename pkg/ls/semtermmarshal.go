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
	"reflect"

	"github.com/cloudprivacylabs/lpg/v2"
)

// TermMarshaler interface defines JSON and JSONLD unmarshalers for a
// custom schema extension
type TermMarshaler interface {
	// Unmarshal a flattened json-ld object.
	UnmarshalLd(target *Layer, key string, value any, node *LDNode, allNodes map[string]*LDNode, interner Interner) error
	// Marshal a property of a node as expanded json-ld
	MarshalLd(source *Layer, sourceNode *lpg.Node, key string) (any, error)
	UnmarshalJSON(target *Layer, key string, value any, node *lpg.Node, interner Interner) error
}

// GetTermMarshaler returns the custom marshaler for the term. If
// there is none, returns the default marshaler
func GetTermMarshaler(term string) TermMarshaler {
	ret, _ := GetTermMetadata(term).(TermMarshaler)
	if ret == nil {
		return DefaultTermMarshaler
	}
	return ret
}

type defaultTermMarshaler struct{}

var DefaultTermMarshaler defaultTermMarshaler

// UnmarshalLd for default marshaler tries to unmarshal the term as a property value
func (defaultTermMarshaler) UnmarshalLd(target *Layer, key string, value any, node *LDNode, allNodes map[string]*LDNode, interner Interner) error {
	// value must be an array
	arr, ok := value.([]any)
	if !ok {
		return nil
	}
	setValue := func(v string) {
		if key == AttributeIndexTerm.Name {
			node.GraphNode.SetProperty(key, AttributeIndexTerm.MustPropertyValue(v))
		} else {
			value, _ := node.GraphNode.GetProperty(key)
			if value == nil {
				node.GraphNode.SetProperty(key, NewPropertyValue(key, v))
			} else {
				pvalue, ok := value.(PropertyValue)
				if !ok {
					node.GraphNode.SetProperty(key, value)
				} else {
					node.GraphNode.SetProperty(key, NewPropertyValue(key, GenericListAppend(pvalue.Value(), v)))
				}
			}
		}
	}
	// If list, descend to its elements
	arr = LDDescendToListElements(arr)
	for _, element := range arr {
		m, ok := element.(map[string]any)
		if !ok {
			continue
		}
		// This is a value or an @id
		if len(m) == 1 {
			if v := m["@value"]; v != nil {
				for _, x := range LDGetValueArr(m) {
					setValue(fmt.Sprint(x))
				}
			} else if v := m["@id"]; v != nil {
				if id, ok := v.(string); ok {
					// Is this a link?
					referencedNode := allNodes[id]
					if referencedNode == nil {
						setValue(id)
					} else {
						target.Graph.NewEdge(node.GraphNode, referencedNode.GraphNode, key, nil)
					}
				}
			}
		}
	}
	return nil
}

// MarshalLd marshals an annotation term of a node as expanded json-ld
func (defaultTermMarshaler) MarshalLd(source *Layer, sourceNode *lpg.Node, key string) (any, error) {
	if key == NodeIDTerm.Name {
		return nil, nil
	}
	var k string
	if GetTerm(key).IsID {
		k = "@id"
	} else {
		k = "@value"
	}
	propv, ok := GetPropertyValue(sourceNode, key)
	if !ok {
		return nil, nil
	}
	val := reflect.ValueOf(propv.Value())
	if val.Type().Kind() == reflect.Array || val.Type().Kind() == reflect.Slice {
		arr := make([]any, 0, val.Len())
		for i := 0; i < val.Len(); i++ {
			arr = append(arr, map[string]any{k: val.Index(i).Interface()})
		}
		return arr, nil
	}
	return []any{map[string]any{k: propv.Value()}}, nil
}

// UnmarshalJSON for the default marshaler tries to unmarshal terms as property values
func (defaultTermMarshaler) UnmarshalJSON(target *Layer, key string, value any, node *lpg.Node, interner Interner) error {
	key = interner.Intern(key)
	node.SetProperty(key, NewPropertyValue(key, value))
	return nil
}
