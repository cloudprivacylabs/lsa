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
	"encoding/json"
	"fmt"

	"github.com/bserdar/digraph"
)

// TermMarshaler interface defines JSON and JSONLD unmarshalers for a
// custom schema extension
type TermMarshaler interface {
	// Unmarshal a flattened json-ld object.
	UnmarshalLd(target *Layer, key string, value interface{}, node *LDNode, allNodes map[string]*LDNode, interner Interner) error
	// Marshal a property of a node as expanded json-ld
	MarshalLd(source *Layer, sourceNode Node, key string) (interface{}, error)
	UnmarshalJSON(target *Layer, key string, value interface{}, node Node, interner Interner) error
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
func (defaultTermMarshaler) UnmarshalLd(target *Layer, key string, value interface{}, node *LDNode, allNodes map[string]*LDNode, interner Interner) error {
	// value must be an array
	arr, ok := value.([]interface{})
	if !ok {
		return nil
	}
	setValue := func(v string) {
		if key == AttributeIndexTerm {
			node.GraphNode.GetProperties()[key] = StringPropertyValue(v)
		} else {
			value := node.GraphNode.GetProperties()[key]
			if value == nil {
				node.GraphNode.GetProperties()[key] = StringPropertyValue(v)
			} else if value.IsStringSlice() {
				node.GraphNode.GetProperties()[key] = StringSlicePropertyValue(append(value.AsStringSlice(), v))
			} else {
				node.GraphNode.GetProperties()[key] = StringSlicePropertyValue([]string{value.AsString(), v})
			}
		}
	}
	// If list, descend to its elements
	arr = DescendToListElements(arr)
	for _, element := range arr {
		m, ok := element.(map[string]interface{})
		if !ok {
			continue
		}
		// This is a value or an @id
		if len(m) == 1 {
			if v := m["@value"]; v != nil {
				setValue(fmt.Sprint(v))
			} else if v := m["@id"]; v != nil {
				if id, ok := v.(string); ok {
					// Is this a link?
					referencedNode := allNodes[id]
					if referencedNode == nil {
						setValue(id)
					} else {
						digraph.Connect(node.GraphNode, referencedNode.GraphNode, NewEdge(key))
					}
				}
			}
		}
	}
	return nil
}

// MarshalLd marshals an annotation term of a node as expanded json-ld
func (defaultTermMarshaler) MarshalLd(source *Layer, sourceNode Node, key string) (interface{}, error) {
	var k string
	if GetTermInfo(key).IsID {
		k = "@id"
	} else {
		k = "@value"
	}
	v := sourceNode.GetProperties()[key]
	if v.IsString() {
		return []interface{}{map[string]interface{}{k: v.AsString()}}, nil
	} else if v.IsStringSlice() {
		arr := make([]interface{}, 0)
		for _, elem := range v.AsStringSlice() {
			arr = append(arr, map[string]interface{}{k: elem})
		}
		return arr, nil
	}
	return nil, nil
}

// UnmarshalJSON for the default marshaler tries to unmarshal terms as property values
func (defaultTermMarshaler) UnmarshalJSON(target *Layer, key string, value interface{}, node Node, interner Interner) error {
	key = interner.Intern(key)
	switch val := value.(type) {
	case string, json.Number, float64, bool:
		node.GetProperties()[key] = StringPropertyValue(fmt.Sprint(val))
	case []interface{}:
		arr := make([]string, 0, len(val))
		for _, v := range val {
			switch v.(type) {
			case string, json.Number, float64, bool:
				arr = append(arr, fmt.Sprint(v))
			default:
				return fmt.Errorf("Invalid value: %s=%v", key, value)
			}
		}
		node.GetProperties()[key] = StringSlicePropertyValue(arr)
	default:
		return fmt.Errorf("Invalid  value: %s=%v", key, value)
	}
	return nil
}
