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
package rdf

// Interner is used to keep a single copy of ids
type Interner map[string]string

// GetStrings returns a copy of the strings in the interner
func (n Interner) GetStrings() []string {
	ret := make([]string, 0, len(n))
	for k := range n {
		ret = append(ret, k)
	}
	return ret
}

// Add a new string to interner and return an integer
func (n Interner) Add(s string) string {
	if v, exists := n[s]; exists {
		return v
	}
	n[s] = s
	return s
}
