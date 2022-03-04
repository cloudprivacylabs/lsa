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

// StringSetUnion returns s1 setunion s2
func StringSetUnion(s1, s2 []string) []string {
	output := make([]string, 0, len(s1)+len(s2))
	s := make(map[string]struct{})
	for _, x := range s1 {
		if _, ok := s[x]; !ok {
			s[x] = struct{}{}
			output = append(output, x)
		}
	}
	for _, x := range s2 {
		if _, ok := s[x]; !ok {
			s[x] = struct{}{}
			output = append(output, x)
		}
	}
	return output
}

// StringSetIntersection returns the common elements of s1 and s2
func StringSetIntersection(s1, s2 []string) []string {
	m := len(s1)
	if len(s2) > m {
		m = len(s2)
	}
	output := make([]string, 0, m)
	s := make(map[string]bool)
	for _, x := range s1 {
		s[x] = false
	}
	for _, x := range s2 {
		if v, ok := s[x]; ok {
			if !v {
				output = append(output, x)
				s[x] = true
			}
		}
	}
	return output
}

// StringSetSubtract returns all elements that are in s1 but not in s2
func StringSetSubtract(s1, s2 []string) []string {
	s := make(map[string]struct{})
	for _, x := range s2 {
		s[x] = struct{}{}
	}
	out := make([]string, 0, len(s1))
	for _, x := range s1 {
		if _, ok := s[x]; !ok {
			out = append(out, x)
		}
	}
	return out
}

// StringSetToSlice converts a string set to slice
func StringSetToSlice(str map[string]struct{}) []string {
	ret := make([]string, 0, len(str))
	for x := range str {
		ret = append(ret, x)
	}
	return ret
}
