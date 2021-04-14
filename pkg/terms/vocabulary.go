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
package terms

// A Vocabulary is a map of Terms
type Vocabulary map[string]Term

// NewVocabulary creates a new vocabulary with the given terms
func NewVocabulary(terms ...Term) Vocabulary {
	ret := Vocabulary{}
	ret.Add(terms...)
	return ret
}

// Add terms to a vocabulary
func (v Vocabulary) Add(terms ...Term) {
	for _, t := range terms {
		v[t.GetTerm()] = t
	}
}

// AddVocabulary adds terms of the source vocabulary to v
func (v Vocabulary) AddVocabulary(src Vocabulary) {
	for k, x := range src {
		v[k] = x
	}
}
