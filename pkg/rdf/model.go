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

import (
	"github.com/lithammer/shortuuid/v3"
	"github.com/piprate/json-gold/ld"
)

// RDF constants, from github.com/piprate/json-gold/
const (
	RDFSyntaxNS string = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
	RDFSchemaNS string = "http://www.w3.org/2000/01/rdf-schema#"
	XSDNS       string = "http://www.w3.org/2001/XMLSchema#"

	XSDAnyType string = XSDNS + "anyType"
	XSDBoolean string = XSDNS + "boolean"
	XSDDouble  string = XSDNS + "double"
	XSDInteger string = XSDNS + "integer"
	XSDFloat   string = XSDNS + "float"
	XSDDecimal string = XSDNS + "decimal"
	XSDAnyURI  string = XSDNS + "anyURI"
	XSDString  string = XSDNS + "string"

	RDFType         string = RDFSyntaxNS + "type"
	RDFFirst        string = RDFSyntaxNS + "first"
	RDFRest         string = RDFSyntaxNS + "rest"
	RDFNil          string = RDFSyntaxNS + "nil"
	RDFPlainLiteral string = RDFSyntaxNS + "PlainLiteral"
	RDFXMLLiteral   string = RDFSyntaxNS + "XMLLiteral"
	RDFJSONLiteral  string = RDFSyntaxNS + "JSON"
	RDFObject       string = RDFSyntaxNS + "object"
	RDFLangString   string = RDFSyntaxNS + "langString"
	RDFList         string = RDFSyntaxNS + "List"
)

type ErrNodeExists string

func (e ErrNodeExists) Error() string { return "Node exists: " + string(e) }

// A Node can be an IRI, BlankNode, or Literal
type Node interface {
	GetValue() string
}

// An IDNode is either a BlankNode or IRI
type IDNode interface {
	Node
	GetID() string
}

// A Liteal node containing a value. Literal implements the Node interface
type Literal interface {
	// The value of the literal
	GetValue() string
	// Type of the literal
	GetType() string
	// The language of the value
	GetLanguage() string
}

// BasicLiteral is the basic implementation of literal.
type BasicLiteral struct {
	Value    string
	Type     string
	Language string
}

func (b BasicLiteral) GetValue() string    { return b.Value }
func (b BasicLiteral) GetType() string     { return b.Type }
func (b BasicLiteral) GetLanguage() string { return b.Language }

// NewStringLiteral returns a new string literal with no language
func NewStringLiteral(value string) BasicLiteral {
	return BasicLiteral{Value: value, Type: XSDString}
}

// A BlankNode is an IDNode and Node.
type BlankNode interface {
	GetID() string
	GetValue() string
}

type BasicBlankNode string

func (b BasicBlankNode) GetID() string    { return string(b) }
func (b BasicBlankNode) GetValue() string { return string(b) }

// NewBasicBlankNode generates a blank node using a uuid
func NewBasicBlankNode() BasicBlankNode {
	return BasicBlankNode("_b:" + shortuuid.New())
}

// An IRI is an IDNode and Node
type IRI interface {
	GetID() string
	GetIRI() string
	GetValue() string
}

type BasicIRI string

func (b BasicIRI) GetID() string    { return string(b) }
func (b BasicIRI) GetIRI() string   { return string(b) }
func (b BasicIRI) GetValue() string { return string(b) }

// ToBasicNode returns a BasicIRI, BasicBlankNode, or BasicLiteral
func ToBasicNode(in Node) Node {
	if i, ok := in.(IRI); ok {
		return BasicIRI(i.GetID())
	}
	if i, ok := in.(BlankNode); ok {
		return BasicBlankNode(i.GetID())
	}
	lit := in.(Literal)
	return BasicLiteral{Value: lit.GetValue(), Type: lit.GetType(), Language: lit.GetLanguage()}
}

type Triple struct {
	// A subject is a BlankNode or IRI
	Subject IDNode
	// A pedicate is an IRI
	Predicate IRI
	// An object can be a Literal, BlankNode, or IRI
	Object Node
}

func (t Triple) ToNQuad() *ld.Quad {
	ret := &ld.Quad{}
	if iri, ok := t.Subject.(IRI); ok {
		ret.Subject = ld.NewIRI(iri.GetIRI())
	} else {
		ret.Subject = ld.NewBlankNode(t.Subject.(BlankNode).GetID())
	}
	ret.Predicate = ld.NewIRI(t.Predicate.(IRI).GetIRI())
	if iri, ok := t.Object.(IRI); ok {
		ret.Object = ld.NewIRI(iri.GetIRI())
	} else if b, ok := t.Object.(BlankNode); ok {
		ret.Object = ld.NewBlankNode(b.GetID())
	} else {
		l := t.Object.(Literal)
		ret.Object = ld.NewLiteral(l.GetValue(), l.GetType(), l.GetLanguage())
	}
	return ret
}

func ToRDFDataset(t []Triple) *ld.RDFDataset {
	ret := ld.NewRDFDataset()
	triples := make([]*ld.Quad, 0, len(t))
	for _, x := range t {
		triples = append(triples, x.ToNQuad())
	}
	ret.Graphs["@default"] = triples
	return ret
}

type Graph interface {

	// AddTriple adds a triple to the graph. Returns true if triple was
	// inserted. Returns false if triple already exists
	AddTriple(Triple) bool
	Add(IDNode, IRI, Node) bool

	// Remove the given triple. Returns true if triple is removed
	RemoveTriple(Triple) bool
	Remove(IDNode, IRI, Node) bool

	GetTriples() []Triple
}
