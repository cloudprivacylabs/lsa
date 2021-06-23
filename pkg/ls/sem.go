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

// TermSemantics is used to describe how a term operates within the
// layered schemas framework.
type TermSemantics struct {
	// The term
	Term string `json:"term" yaml:"term"`

	// If true, the term value is an @id (IRI). In JSON-LD, the values for
	// this term will be marshalled as @id
	IsID bool `json:"isId" yaml:"isId"`

	// If true, the term is a list. In JSON-LD, its elements will be
	// marshaled under @list
	IsList bool `json:"isList" yaml:"isList"`

	Composition CompositionType `json:"composition,omitempty" yaml:"composition,omitempty"`

	Metadata interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// NewTerm registers a term with given semantics, and returns the term.
func NewTerm(term string, isID, isList bool, comp CompositionType, md interface{}) string {
	t := TermSemantics{Term: term,
		IsID:        isID,
		IsList:      isList,
		Composition: comp,
		Metadata:    md,
	}
	RegisterTerm(t)
	return term
}

func (t TermSemantics) Compose(target, src interface{}) (interface{}, error) {
	return t.Composition.Compose(target, src)
}
