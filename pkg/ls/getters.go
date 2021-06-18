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

// GetKeyValue returns the value of the key in the node. The node must
// be a map
func GetKeyValue(key string, node interface{}) (interface{}, bool) {
	var m map[string]interface{}
	arr, ok := node.([]interface{})
	if ok {
		if len(arr) == 1 {
			m, _ = arr[0].(map[string]interface{})
		}
	} else {
		m, _ = node.(map[string]interface{})
	}
	if m == nil {
		return "", false
	}
	v, ok := m[key]
	return v, ok
}

// GetStringValue returns a string value from the node with the
// key. The node must be a map
func GetStringValue(key string, node interface{}) string {
	v, _ := GetKeyValue(key, node)
	if v == nil {
		return ""
	}
	return v.(string)
}

// GetNodeID returns the node @id. The argument must be a map
func GetNodeID(node interface{}) string {
	return GetStringValue("@id", node)
}

// GetNodeTypes returns the node @type. The argument must be a map
func GetNodeTypes(node interface{}) []string {
	m, ok := node.(map[string]interface{})
	if !ok {
		return nil
	}
	arr, ok := m["@type"].([]interface{})
	if ok {
		ret := make([]string, 0, len(arr))
		for _, x := range arr {
			s, _ := x.(string)
			if len(s) > 0 {
				ret = append(ret, s)
			}
		}
		return ret
	}
	return nil
}

// GetNodeValue returns the node @value. The argument must be a map
func GetNodeValue(node interface{}) interface{} {
	v, _ := GetKeyValue("@value", node)
	return v
}

// GetListElements returns the elements of a @list node. The input can
// be a [{"@list":elements}] or {@list:elements}. If the input cannot
// be interpreted as a list, returns nil
func GetListElements(node interface{}) []interface{} {
	var m map[string]interface{}
	if arr, ok := node.([]interface{}); ok {
		if len(arr) == 1 {
			m, _ = arr[0].(map[string]interface{})
		}
	}
	if m == nil {
		m, _ = node.(map[string]interface{})
	}
	if len(m) == 0 {
		return []interface{}{}
	}
	lst, ok := m["@list"]
	if !ok {
		return nil
	}
	elements, ok := lst.([]interface{})
	if !ok {
		return nil
	}
	return elements
}

// GetFlattenedNodeIndex returns a map of node objects from a flattened graph
func GetFlattenedNodeIndex(nodes []interface{}) (map[string]interface{}, error) {
	ret := make(map[string]interface{}, len(nodes))
	for _, node := range nodes {
		id := GetNodeID(node)
		if len(id) == 0 {
			return nil, ErrAttributeWithoutID
		}
		_, exists := ret[id]
		if exists {
			return nil, ErrDuplicateAttributeID(id)
		}
		ret[id] = node
	}
	return ret, nil
}

// If in is a @list, returns its elements
func DescendToListElements(in []interface{}) []interface{} {
	if len(in) == 1 {
		if m, ok := in[0].(map[string]interface{}); ok {
			if l, ok := m["@list"]; ok {
				if a, ok := l.([]interface{}); ok {
					return a
				}
			}
		}
	}
	return in
}
