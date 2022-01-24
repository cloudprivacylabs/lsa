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
package xml

import (
	"encoding/xml"
	"strings"
)

type ErrInvalidWhitespaceFacet struct {
	Name string
}

func (e ErrInvalidWhitespaceFacet) Error() string { return "Invalid whitespace facet: " + e.Name }

// WhitespaceFacet is the only pre-lexical facet in XSD spec, that is,
// whitespace normalization must be done before validation
type WhitespaceFacet interface {
	Filter(string) string
}

// IsWhitespaceFacet returns if the given xml name is the whitespace facet
func IsWhitespaceFacet(name xml.Name) bool {
	return (len(name.Space) == 0 && name.Local == "whiteSpace") ||
		(name.Space == XSDNamespace && name.Local == "whiteSpace")
}

// GetWhitespaceFacet returns the facet based on facetName being "preserve", "replace", or "collapse"
func GetWhitespaceFacet(facetName string) (WhitespaceFacet, error) {
	switch facetName {
	case "preserve":
		return PreserveWhitespaceFacet{}, nil
	case "replace":
		return ReplaceWhitespaceFacet{}, nil
	case "collapse":
		return CollapseWhitespaceFacet{}, nil
	}
	return nil, ErrInvalidWhitespaceFacet{Name: facetName}
}

// PreserveWhitespaceFacet does not do any string normalization
type PreserveWhitespaceFacet struct{}

func (PreserveWhitespaceFacet) Filter(input string) string {
	return input
}

func replaceRune(in rune) rune {
	if in == '\n' || in == '\t' || in == '\r' {
		return ' '
	}
	return in
}

// ReplaceWhitespaceFacet replaces all occurrences of #x9 (tab), #xA (line feed) and #xD
// (carriage return) are replaced with #x20 (space).
type ReplaceWhitespaceFacet struct{}

func (ReplaceWhitespaceFacet) Filter(input string) string {
	return strings.Map(replaceRune, input)
}

// CollapseWhitespaceFacet works as follows: Subsequent to the replacements specified
// above under replace, contiguous sequences of #x20s are collapsed to
// a single #x20, and initial and/or final #x20s are deleted.
type CollapseWhitespaceFacet struct{}

func (CollapseWhitespaceFacet) Filter(input string) string {
	input = strings.Map(replaceRune, input)
	out := make([]rune, 0, len(input))
	lastSp := false
	for _, x := range input {
		if x == ' ' {
			lastSp = true
		} else {
			if lastSp {
				if len(out) > 0 {
					out = append(out, ' ')
				}
			}
			out = append(out, x)
			lastSp = false
		}
	}
	return string(out)
}
