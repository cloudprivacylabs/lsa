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
	"bytes"
	"encoding/xml"
	"io"
	"strings"
)

type xmlElement struct {
	name       xml.Name
	attributes []xmlAttribute
	children   []interface{}
}

type xmlAttribute struct {
	name  xml.Name
	value string
}

type xmlText struct {
	text []byte
}

func (el xmlElement) findAttr(name xml.Name) (string, bool) {
	for _, x := range el.attributes {
		if MatchName(x.name, name) {
			return x.value, true
		}
	}
	return "", false
}

func decode(decoder *xml.Decoder) (*xmlElement, error) {
	wsFacet, _ := GetWhitespaceFacet("collapse")

	filterBOM := func(in []byte) []byte {
		if len(in) == 3 && in[0] == 0xEF && in[1] == 0xBB && in[2] == 0xBF {
			return []byte(" ")
		}
		return in
	}

	// At the top level there can be processing instructions,
	// directives, and the root element
	done := false
	rootSeen := false
	var rootNode *xmlElement
	var err error
	for !done {
		var tok xml.Token
		tok, err = decoder.Token()
		if err != nil {
			break
		}
		switch token := tok.(type) {
		case xml.CharData:
			data := token.Copy()
			if !rootSeen {
				data = filterBOM(data)
			}
			if strings.TrimSpace(string(data)) != "" {
				return nil, ErrExtraCharacters
			}

		case xml.StartElement:
			// This is the document root
			if rootSeen {
				return nil, ErrMultipleRoots
			}
			rootSeen = true
			rootNode, err = decodeElement(token, decoder, wsFacet)
			if err != nil {
				return nil, err
			}
			done = true

		case xml.Comment:

		case xml.Directive:

		case xml.ProcInst:

		default:
			return nil, ErrInvalidXML
		}
	}
	if err == io.EOF {
		err = nil
	}
	return rootNode, nil
}

func decodeElement(elToken xml.StartElement, decoder *xml.Decoder, wsFacet WhitespaceFacet) (*xmlElement, error) {
	element := xmlElement{
		name: elToken.Name,
	}
	for _, attribute := range elToken.Attr {
		if IsWhitespaceFacet(attribute.Name) {
			wf, err := GetWhitespaceFacet(attribute.Value)
			if err != nil {
				return nil, err
			}
			wsFacet = wf
		}
		attribute := xmlAttribute{
			name:  attribute.Name,
			value: attribute.Value,
		}
		element.attributes = append(element.attributes, attribute)
	}

	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			return nil, ErrElementNodeNotTerminated{elToken.Name}
		}
		if err != nil {
			return nil, err
		}

		switch token := tok.(type) {
		case xml.StartElement:
			el, err := decodeElement(token, decoder, wsFacet)
			if err != nil {
				return nil, err
			}
			element.children = append(element.children, el)

		case xml.EndElement:
			// Normalize text nodes
			w := 0
			var accumulatedText *bytes.Buffer
			for i := range element.children {
				if ch, chardata := element.children[i].(*xmlText); chardata {
					if accumulatedText == nil {
						accumulatedText = &bytes.Buffer{}
					}
					accumulatedText.Write(ch.text)
				} else {
					if accumulatedText != nil {
						element.children[w] = &xmlText{text: accumulatedText.Bytes()}
						w++
						accumulatedText = nil
					}
					element.children[w] = element.children[i]
					w++
				}
			}

			if accumulatedText != nil {
				element.children[w] = &xmlText{text: accumulatedText.Bytes()}
				w++
			}
			element.children = element.children[:w]

			// Apply wsfacet and create text nodes
			w = 0
			for i := range element.children {
				if chardata, isCharData := element.children[i].(*xmlText); isCharData {
					str := wsFacet.Filter(string(chardata.text))
					if len(str) != 0 {
						element.children[w] = element.children[i]
						element.children[w].(*xmlText).text = []byte(str)
						w++
					}
				} else {
					element.children[w] = element.children[i]
					w++
				}
			}
			element.children = element.children[:w]
			return &element, nil

		case xml.CharData:
			element.children = append(element.children, &xmlText{text: []byte(token.Copy())})

		case xml.ProcInst:
		case xml.Directive:
		case xml.Comment:
		}
	}
}
