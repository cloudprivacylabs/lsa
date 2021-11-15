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

// Interner interface is used to keep a string table to reduce memory footprint by eliminated repeated keys
type Interner interface {
	Intern(string) string
}

// InternSlice interns all elements of a slice
func InternSlice(interner Interner, slice []string) []string {
	out := make([]string, 0, len(slice))
	for _, x := range slice {
		out = append(out, interner.Intern(x))
	}
	return out
}

// StringInterner is used to intern strings so multiple identical
// copies of strings are minimized
type StringInterner struct {
	strings map[string]string
}

// Return a new interner
func NewInterner() StringInterner {
	return StringInterner{strings: make(map[string]string)}
}

// Intern a string and return the corresponding interned string
func (s StringInterner) Intern(key string) string {
	result, ok := s.strings[key]
	if !ok {
		result = key
		s.strings[key] = result
	}
	return result
}
