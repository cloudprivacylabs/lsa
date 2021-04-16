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
	"fmt"
	"io"
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

// A BlankNode is an IDNode and Node.
type BlankNode interface {
	GetID() string
	GetValue() string
}

type BasicBlankNode string

func (b BasicBlankNode) GetID() string    { return string(b) }
func (b BasicBlankNode) GetValue() string { return string(b) }

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

type Graph interface {

	// AddTriple adds a triple to the graph. Returns true if triple was
	// inserted. Returns false if triple already exists
	AddTriple(Triple) bool

	// Remove the given triple. Returns true if triple is removed
	RemoveTriple(Triple) bool
}

type GraphNode struct {
	Node
	ID string
}

func ToDOT(graphName string, nodes []GraphNode, edges [][2]string, out io.Writer) error {
	if _, err := fmt.Fprintf(out, "digraph %s {\n", graphName); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "rankdir=\"LR\";\n"); err != nil {
		return err
	}
	for _, node := range nodes {
		if _, err := fmt.Fprintf(out, "  %s [label=\"%s\"];\n", node.ID, node.GetValue()); err != nil {
			return err
		}
	}
	for _, e := range edges {
		if _, err := fmt.Fprintf(out, "  %s -> %s;\n", e[0], e[1]); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(out, "}\n"); err != nil {
		return err
	}
	return nil
}
