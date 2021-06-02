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

// Recursively copy in
func copyIntf(in interface{}) interface{} {
	if arr, ok := in.([]interface{}); ok {
		ret := make([]interface{}, 0, len(arr))
		for _, x := range arr {
			ret = append(ret, copyIntf(x))
		}
		return ret
	}
	if m, ok := in.(map[string]interface{}); ok {
		ret := make(map[string]interface{}, len(m))
		for k, v := range m {
			ret[k] = copyIntf(v)
		}
		return ret
	}
	return in
}
