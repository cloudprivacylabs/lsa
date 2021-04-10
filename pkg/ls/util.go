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

// GetNodeType returns the node @type. The argument must be a map
func GetNodeType(node interface{}) string {
	m, ok := node.(map[string]interface{})
	if !ok {
		return ""
	}
	arr, ok := m["@type"].([]interface{})
	if ok {
		if len(arr) == 1 {
			return arr[0].(string)
		}
	}
	return ""
}

// GetNodeValue returns the node @value. The argument must be a map
func GetNodeValue(node interface{}) interface{} {
	v, _ := GetKeyValue("@value", node)
	return v
}

// GetListElements returns the elements of a @list node. The input can
// be a [{"@list":elements}] or {@list:elements}
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
	if m == nil {
		return nil
	}
	elements, _ := m["@list"].([]interface{})
	return elements
}
