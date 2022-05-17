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
	Term string

	Namespace string
	LName     string
	Aliases   []string

	// If true, the term value is an @id (IRI). In JSON-LD, the values for
	// this term will be marshalled as @id
	IsID bool

	// If true, the term is a list. In JSON-LD, its elements will be
	// marshaled under @list
	IsList bool

	Composition CompositionType

	Metadata interface{}
}

// NewTerm registers a term with given semantics, and returns the term.
func NewTerm(ns, lname string, isID, isList bool, comp CompositionType, md interface{}, aliases ...string) string {
	t := TermSemantics{Term: ns + lname,
		Namespace:   ns,
		LName:       lname,
		Aliases:     aliases,
		IsID:        isID,
		IsList:      isList,
		Composition: comp,
		Metadata:    md,
	}
	RegisterTerm(t)
	return t.Term
}

func (t TermSemantics) Compose(target, src *PropertyValue) (*PropertyValue, error) {
	return t.Composition.Compose(target, src)
}

var registeredTerms = map[string]TermSemantics{}

// If a term is know, using this function avoids duplicate string
// copies
func knownTerm(s string) string {
	x, ok := registeredTerms[s]
	if ok {
		return x.Term
	}
	return s
}

func RegisterTerm(t TermSemantics) {
	reg := func(s string) {
		_, ok := registeredTerms[s]
		if ok {
			panic("Duplicate term :" + t.Term)
		}
		registeredTerms[s] = t
	}
	reg(t.Term)
	for _, alias := range t.Aliases {
		reg(alias)
	}
}

func GetTermInfo(term string) TermSemantics {
	t, ok := registeredTerms[term]
	if !ok {
		return TermSemantics{Term: term, Composition: SetComposition}
	}
	return t
}

// GetTermMetadata returns metadata about a term
func GetTermMetadata(term string) interface{} {
	t := GetTermInfo(term)
	return t.Metadata
}

func IsTermRegistered(term string) bool {
	_, ok := registeredTerms[term]
	return ok
}
