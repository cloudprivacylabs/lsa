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

	"github.com/cloudprivacylabs/opencypher/graph"
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

// // Ingester converts an XML object model into a graph using a schema
// type Ingester struct {
// 	ls.Ingester

// 	Interner ls.Interner

// 	IncludeProcessingInstructions bool
// 	IncludeDirectives             bool
// }

// // IngestStream ingests an XML from the input stream
// func IngestStream(context *ls.Context, ingester *Ingester, baseID string, input io.Reader) (graph.Node, error) {
// 	decoder := xml.NewDecoder(input)
// 	return ingester.IngestDocument(context, baseID, decoder)
// }

// type xmlElement struct {
// 	name       xml.Name
// 	attributes []xmlAttribute
// 	children   []interface{}
// }

// type xmlAttribute struct {
// 	name  xml.Name
// 	value string
// }

// type xmlText struct {
// 	text []byte
// }

// func (ingester *Ingester) decode(decoder *xml.Decoder) (*xmlElement, error) {
// 	wsFacet, _ := GetWhitespaceFacet("collapse")

// 	filterBOM := func(in []byte) []byte {
// 		if len(in) == 3 && in[0] == 0xEF && in[1] == 0xBB && in[2] == 0xBF {
// 			return []byte(" ")
// 		}
// 		return in
// 	}

// 	// At the top level there can be processing instructions,
// 	// directives, and the root element
// 	done := false
// 	rootSeen := false
// 	var rootNode *xmlElement
// 	var err error
// 	for !done {
// 		var tok xml.Token
// 		tok, err = decoder.Token()
// 		if err != nil {
// 			break
// 		}
// 		switch token := tok.(type) {
// 		case xml.CharData:
// 			data := token.Copy()
// 			if !rootSeen {
// 				data = filterBOM(data)
// 			}
// 			if strings.TrimSpace(string(data)) != "" {
// 				return nil, ErrExtraCharacters
// 			}

// 		case xml.StartElement:
// 			// This is the document root
// 			if rootSeen {
// 				return nil, ErrMultipleRoots
// 			}
// 			rootSeen = true
// 			rootNode, err = ingester.decodeElement(token, decoder, wsFacet)
// 			if err != nil {
// 				return nil, err
// 			}
// 			done = true

// 		case xml.Comment:

// 		case xml.Directive:

// 		case xml.ProcInst:

// 		default:
// 			return nil, ErrInvalidXML
// 		}
// 	}
// 	if err == io.EOF {
// 		err = nil
// 	}
// 	return rootNode, nil
// }

// func (ingester *Ingester) decodeElement(elToken xml.StartElement, decoder *xml.Decoder, wsFacet WhitespaceFacet) (*xmlElement, error) {
// 	element := xmlElement{
// 		name: elToken.Name,
// 	}
// 	for _, attribute := range elToken.Attr {
// 		if IsWhitespaceFacet(attribute.Name) {
// 			wf, err := GetWhitespaceFacet(attribute.Value)
// 			if err != nil {
// 				return nil, err
// 			}
// 			wsFacet = wf
// 		}
// 		attribute := xmlAttribute{
// 			name:  attribute.Name,
// 			value: attribute.Value,
// 		}
// 		element.attributes = append(element.attributes, attribute)
// 	}

// 	for {
// 		tok, err := decoder.Token()
// 		if err == io.EOF {
// 			return nil, ErrElementNodeNotTerminated{elToken.Name}
// 		}
// 		if err != nil {
// 			return nil, err
// 		}

// 		switch token := tok.(type) {
// 		case xml.StartElement:
// 			el, err := ingester.decodeElement(token, decoder, wsFacet)
// 			if err != nil {
// 				return nil, err
// 			}
// 			element.children = append(element.children, el)

// 		case xml.EndElement:
// 			// Normalize text nodes
// 			w := 0
// 			var accumulatedText *bytes.Buffer
// 			for i := range element.children {
// 				if ch, chardata := element.children[i].(*xmlText); chardata {
// 					if accumulatedText == nil {
// 						accumulatedText = &bytes.Buffer{}
// 					}
// 					accumulatedText.Write(ch.text)
// 				} else {
// 					if accumulatedText != nil {
// 						element.children[w] = &xmlText{text: accumulatedText.Bytes()}
// 						w++
// 						accumulatedText = nil
// 					}
// 					element.children[w] = element.children[i]
// 					w++
// 				}
// 			}

// 			if accumulatedText != nil {
// 				element.children[w] = &xmlText{text: accumulatedText.Bytes()}
// 				w++
// 			}
// 			element.children = element.children[:w]

// 			// Apply wsfacet and create text nodes
// 			w = 0
// 			for i := range element.children {
// 				if chardata, isCharData := element.children[i].(*xmlText); isCharData {
// 					str := wsFacet.Filter(string(chardata.text))
// 					if len(str) != 0 {
// 						element.children[w] = element.children[i]
// 						element.children[w].(*xmlText).text = []byte(str)
// 						w++
// 					}
// 				} else {
// 					element.children[w] = element.children[i]
// 					w++
// 				}
// 			}
// 			element.children = element.children[:w]
// 			return &element, nil

// 		case xml.CharData:
// 			element.children = append(element.children, &xmlText{text: []byte(token.Copy())})

// 		case xml.ProcInst:
// 		case xml.Directive:
// 		case xml.Comment:
// 		}
// 	}
// }

// IngestDocument ingests an XML using the schema. The output will have all input
// nodes associated with schema nodes.
//
// BaseID is the ID of the root object. All other attribute names are
// generated by appending the attribute path to baseID
// func (ingester *Ingester) IngestDocument(context *ls.Context, baseID string, decoder *xml.Decoder) (graph.Node, error) {
// 	if ingester.Interner == nil {
// 		ingester.Interner = ls.NewInterner()
// 	}

// 	root, err := ingester.decode(decoder)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// Root matches schema root
// 	ictx := ingester.Start(context, baseID)
// 	ictx.SourcePath.AppendString(root.name.Local)
// 	rootNode, err := ingester.element(ictx, root)
// 	ingester.Finish(ictx, rootNode)
// 	return rootNode, err
// }

// func (ingester *Ingester) element(ctx ls.IngestionContext, element *xmlElement) (graph.Node, error) {
// 	schemaNode := ctx.GetSchemaNode()
// 	// If schemaNode is nil and we are only ingesting known nodes, ignore this node
// 	if schemaNode == nil && ingester.OnlySchemaAttributes {
// 		return nil, nil
// 	}
// 	validate := func(node graph.Node, err error) (graph.Node, error) {
// 		if err != nil {
// 			return nil, err
// 		}
// 		if err := ingester.Validate(ctx, node); err != nil {
// 			return nil, err
// 		}
// 		return node, nil
// 	}

// 	// What are we ingesting? If there is a schema, it dictates the type
// 	if schemaNode != nil {
// 		switch {
// 		case schemaNode.HasLabel(ls.AttributeTypeValue):
// 			return validate(ingester.ingestAsValue(ctx, element, schemaNode))
// 		case schemaNode.HasLabel(ls.AttributeTypeObject):
// 			return validate(ingester.ingestAsObject(ctx, element, schemaNode))
// 		case schemaNode.HasLabel(ls.AttributeTypeArray):
// 			return validate(ingester.ingestAsArray(ctx, element, schemaNode))
// 		case schemaNode.HasLabel(ls.AttributeTypePolymorphic):
// 			return ingester.ingestPolymorphic(ctx, element, schemaNode)
// 		}
// 		return nil, ls.ErrInvalidSchema(fmt.Sprintf("Cannot determine attribute type for %s", ls.GetNodeID(schemaNode)))
// 	}
// 	// No schemas. If this is an element with a single text node, ingest as value
// 	if len(element.children) == 1 {
// 		if _, text := element.children[0].(*xmlText); text {
// 			return ingester.ingestAsValue(ctx, element, nil)
// 		}
// 	}
// 	return ingester.ingestAsObject(ctx, element, nil)
// }

// This finds the best matching schema attribute. If a schema
// attribute is an XML attribute, requireAttr must be
// true.
func findBestMatchingSchemaAttribute(name xml.Name, schemaAttrs []graph.Node, requireAttr bool) (graph.Node, error) {
	var matched graph.Node
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

// func (ingester *Ingester) ingestAsValue(ctx ls.IngestionContext, element *xmlElement, schemaNode graph.Node) (graph.Node, error) {
// 	// element has at most one text node
// 	var value string
// 	if len(element.children) > 1 {
// 		return nil, ls.ErrSchemaValidation{Msg: "Cannot ingest element as a value because it has multiple child nodes", Path: ctx.SourcePath.Copy()}
// 	}
// 	if len(element.children) == 1 {
// 		t, ok := element.children[0].(*xmlText)
// 		if !ok {
// 			return nil, ls.ErrSchemaValidation{Msg: "Cannot ingest element as a value because it has elements as children", Path: ctx.SourcePath.Copy()}
// 		}
// 		value = string(t.text)
// 	}
// 	ingestedAs, _, node, err := ingester.Value(ctx.New(element.name.Local, schemaNode), value)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if node != nil && ingestedAs != ls.IngestAsProperty {
// 		node.SetProperty(LocalNameTerm, ls.StringPropertyValue(element.name.Local))
// 		if len(element.name.Space) > 0 {
// 			node.SetProperty(NamespaceTerm, ls.StringPropertyValue(element.name.Space))
// 		}
// 	}
// 	// Process attributes
// 	for _, attribute := range element.attributes {
// 		if !ingester.IngestEmptyValues && len(attribute.value) == 0 {
// 			continue
// 		}
// 		// If the value is ingested as property, we cannot have attributes
// 		if ingestedAs == ls.IngestAsProperty {
// 			return nil, ErrCannotHaveAttributes{Path: ctx.SourcePath.String()}
// 		}
// 		if node != nil {
// 			node.SetProperty(makeFullName(attribute.name), ls.StringPropertyValue(attribute.value))
// 		}
// 	}

// 	return node, err
// }

// func (ingester *Ingester) ingestAsObject(ctx ls.IngestionContext, element *xmlElement, schemaNode graph.Node) (graph.Node, error) {
// 	// Make sure tag matches
// 	if schemaNode != nil {
// 		schName := GetXMLName(schemaNode)
// 		if len(schName.Local) > 0 {
// 			if !MatchName(element.name, schName) {
// 				return nil, ErrIncorrectElement{element.name}
// 			}
// 		}
// 	}
// 	// Get all the possible child nodes from the schema. If the
// 	// schemaNode is nil, the returned schemaNodes will be empty
// 	childSchemaNodes := ls.GetObjectAttributeNodes(schemaNode)
// 	_, _, node, err := ingester.Object(ctx.New(element.name.Local, schemaNode))
// 	if err != nil {
// 		return nil, err
// 	}
// 	node.SetProperty(LocalNameTerm, ls.StringPropertyValue(element.name.Local))
// 	if len(element.name.Space) > 0 {
// 		node.SetProperty(NamespaceTerm, ls.StringPropertyValue(element.name.Space))
// 	}
// 	newCtx := ctx.NewLevel(node)

// 	for index, attribute := range element.attributes {
// 		if !ingester.IngestEmptyValues && len(attribute.value) == 0 {
// 			continue
// 		}
// 		var attrSchema graph.Node
// 		if schemaNode != nil {
// 			attrSchema, err = findBestMatchingSchemaAttribute(attribute.name, childSchemaNodes, true)
// 			if err != nil {
// 				return nil, err
// 			}
// 		}

// 		_, _, attrNode, err := ingester.Value(newCtx.New(fmt.Sprintf("attr-%d", index), attrSchema), attribute.value)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if attrNode != nil {
// 			if len(attribute.name.Space) > 0 {
// 				attrNode.SetProperty(NamespaceTerm, ls.StringPropertyValue(ingester.Interner.Intern(attribute.name.Space)))
// 			}
// 			attrNode.SetProperty(LocalNameTerm, ls.StringPropertyValue(ingester.Interner.Intern(attribute.name.Local)))
// 		}
// 	}

// 	for index, child := range element.children {
// 		var newNode graph.Node
// 		switch childNode := child.(type) {
// 		case *xmlElement:
// 			childSchema, err := findBestMatchingSchemaAttribute(childNode.name, childSchemaNodes, false)
// 			if err != nil {
// 				return nil, err
// 			}
// 			// If nothing was found, check if there are polymorphic nodes
// 			if childSchema == nil {
// 				for _, childSchemaNode := range childSchemaNodes {
// 					if !childSchemaNode.HasLabel(ls.AttributeTypePolymorphic) {
// 						continue
// 					}
// 					// A polymorphic node
// 					newNode, err = ingester.element(newCtx.New(childNode.name.Local, childSchemaNode), childNode)
// 					if err != nil {
// 						return nil, err
// 					}
// 				}
// 			}
// 			if newNode == nil {
// 				newNode, err = ingester.element(newCtx.New(childNode.name.Local, childSchema), childNode)
// 				if err != nil {
// 					return nil, err
// 				}
// 			}
// 		case *xmlText:
// 			_, _, newNode, err = ingester.Value(newCtx.New("", nil), string(childNode.text))
// 			if err != nil {
// 				return nil, err
// 			}
// 		}
// 		if newNode != nil {
// 			newNode.SetProperty(ls.AttributeIndexTerm, ls.IntPropertyValue(index))
// 		}
// 	}
// 	return node, nil
// }

// func (ingester *Ingester) ingestAsArray(ctx ls.IngestionContext, element *xmlElement, schemaNode graph.Node) (graph.Node, error) {
// 	elementNode := ls.GetArrayElementNode(schemaNode)
// 	_, _, node, err := ingester.Array(ctx.New(element.name.Local, schemaNode))
// 	if err != nil {
// 		return nil, err
// 	}
// 	node.SetProperty(LocalNameTerm, ls.StringPropertyValue(element.name.Local))
// 	if len(element.name.Space) > 0 {
// 		node.SetProperty(NamespaceTerm, ls.StringPropertyValue(element.name.Space))
// 	}
// 	newCtx := ctx.NewLevel(node)

// 	for index, attribute := range element.attributes {
// 		if !ingester.IngestEmptyValues && len(attribute.value) == 0 {
// 			continue
// 		}
// 		_, _, _, err := ingester.Value(newCtx.New(fmt.Sprintf("attr-%d", index), nil), attribute.value)
// 		if err != nil {
// 			return nil, err
// 		}
// 	}

// 	for index, child := range element.children {
// 		var newNode graph.Node
// 		switch childNode := child.(type) {
// 		case *xmlElement:
// 			newNode, err = ingester.element(newCtx.New(childNode.name.Local, elementNode), childNode)
// 			if err != nil {
// 				return nil, err
// 			}
// 		case *xmlText:
// 			_, _, newNode, err = ingester.Value(newCtx.New("", elementNode), string(childNode.text))
// 			if err != nil {
// 				return nil, err
// 			}
// 		}
// 		if newNode != nil {
// 			newNode.SetProperty(ls.AttributeIndexTerm, ls.IntPropertyValue(index))
// 		}
// 	}
// 	return node, nil
// }

// func (ingester *Ingester) ingestPolymorphic(ctx ls.IngestionContext, element *xmlElement, schemaNode graph.Node) (graph.Node, error) {
// 	f := func(ing *ls.Ingester, ctx ls.IngestionContext) (graph.Node, error) {
// 		newIngester := *ingester
// 		newIngester.Ingester = *ing
// 		node, err := newIngester.element(ctx, element)
// 		if err != nil {
// 			return nil, err
// 		}
// 		if node == nil {
// 			return nil, fmt.Errorf("Polymorhic option does not produce output")
// 		}
// 		return node, nil
// 	}
// 	return ingester.Polymorphic(ctx, f, f)
// }

func makeFullName(name xml.Name) string {
	if len(name.Space) == 0 {
		return name.Local
	}
	if name.Space[len(name.Space)-1] == '/' || name.Space[len(name.Space)-1] == '#' || name.Space[len(name.Space)-1] == ':' {
		return name.Space + name.Local
	}
	return name.Space + "/" + name.Local
}
