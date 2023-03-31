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
	"errors"
	"fmt"

	"github.com/cloudprivacylabs/lpg/v2"
)

var ErrExtraCharacters = errors.New("Extra characters before document")
var ErrMultipleRoots = errors.New("Multiple roots")
var ErrInvalidXML = errors.New("Invalid XML")

type ErrIncorrectElement struct {
	Name xml.Name
}

func (err ErrIncorrectElement) Error() string {
	return fmt.Sprintf("Incorrect element: %v", err.Name)
}

const XSDNamespace = "http://www.w3.org/2001/XMLSchema"

type ErrAmbiguousSchemaAttribute struct {
	Attr xml.Name
}

func (e ErrAmbiguousSchemaAttribute) Error() string {
	if len(e.Attr.Space) > 0 {
		return fmt.Sprintf("Ambiguous schema attribute for %s:%s", e.Attr.Space, e.Attr.Local)
	}
	return fmt.Sprintf("Ambiguous schema attribute for %s", e.Attr.Local)
}

type ErrElementNodeNotTerminated struct {
	Name xml.Name
}

func (e ErrElementNodeNotTerminated) Error() string {
	if len(e.Name.Space) > 0 {
		return fmt.Sprintf("Element not terminated: %s:%s", e.Name.Space, e.Name.Local)
	}
	return fmt.Sprintf("Element not terminated: %s", e.Name.Local)
}

type ErrCannotHaveAttributes struct {
	Path string
}

func (e ErrCannotHaveAttributes) Error() string {
	return "Cannot have attributes:" + e.Path
}

// This finds the best matching schema attribute. If a schema
// attribute is an XML attribute, requireAttr must be
// true.
func findBestMatchingSchemaAttribute(name xml.Name, schemaAttrs []*lpg.Node, requireAttr bool) (*lpg.Node, error) {
	var matched *lpg.Node
	for _, attr := range schemaAttrs {
		_, attrTerm := attr.GetProperty(AttributeTerm)
		if requireAttr != attrTerm {
			continue
		}
		attrName := GetXMLName(attr)
		if MatchName(name, attrName) {
			if matched != nil {
				return nil, ErrAmbiguousSchemaAttribute{Attr: name}
			}
			matched = attr
		}
	}
	return matched, nil
}

func makeFullName(name xml.Name) string {
	if len(name.Space) == 0 {
		return name.Local
	}
	if name.Space[len(name.Space)-1] == '/' || name.Space[len(name.Space)-1] == '#' || name.Space[len(name.Space)-1] == ':' {
		return name.Space + name.Local
	}
	return name.Space + "/" + name.Local
}
