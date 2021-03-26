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

	"github.com/lithammer/shortuuid/v3"
	"github.com/piprate/json-gold/ld"
)

type Node interface {
	GetValue() string
}

type IDNode interface {
	GetID() string
}

type ErrNodeExists string

func (e ErrNodeExists) Error() string { return fmt.Sprintf("Node exists: %s", string(e)) }

var ErrUnsupportedNode = fmt.Errorf("Unsupported node")

// An IRI is a generalization of a URI. It can be a subject, predicate, or object
type IRI struct {
	id string
}

// GetID returns the IRI value as the ID
func (node IRI) GetID() string    { return node.id }
func (node IRI) GetValue() string { return node.id }
func (node IRI) String() string   { return node.id }

// A BlankNode contains an ID. A BlankNode can be a subject or object.
type BlankNode struct {
	id string
}

// GetID returns the blank node ID
func (node BlankNode) GetID() string    { return node.id }
func (node BlankNode) GetValue() string { return node.id }
func (node BlankNode) String() string   { return node.id }

// A Literal has a value, type, and language. A Literal can only be an object
type Literal struct {
	Value    string
	Type     string
	Language string
}

func (node Literal) GetValue() string { return node.Value }
func (node Literal) String() string   { return node.Value }

// A Triple contains the subject (which is an IDNode), predicate, and object
type Triple struct {
	subject   Node
	predicate *IRI
	object    Node
}

func (t Triple) GetSubject() Node   { return t.subject }
func (t Triple) GetPredicate() *IRI { return t.predicate }
func (t Triple) GetObject() Node    { return t.object }

type Graph struct {
	// This keeps a single copy of all id strings in the memory
	ids Interner

	// A map of all nodes in the graph
	nodes map[Node]struct{}
	// This map keeps all triples. It is also used to keep a unique set
	// of triples
	triples map[Triple]*Triple

	// Map of all nodes with an ID (IRIs and BlankNodes)
	nodeIDs map[string]Node
	// Triple indexes based on subject
	subjects tripleIndex
	// Triple indexes based on object
	objects tripleIndex
	// Triple indexes based on predicate
	predicates tripleIndex
}

type tripleIndex map[Node]map[*Triple]struct{}

func (index tripleIndex) add(n Node, t *Triple) {
	m := index[n]
	if m == nil {
		m = make(map[*Triple]struct{})
		index[n] = m
	}
	m[t] = struct{}{}
}

func (index tripleIndex) remove(n Node, t *Triple) {
	m := index[n]
	if m != nil {
		delete(m, t)
	}
}

// NewGraph returns an empty graph
func NewGraph() *Graph {
	return &Graph{
		ids:        make(Interner),
		nodes:      make(map[Node]struct{}),
		triples:    make(map[Triple]*Triple),
		nodeIDs:    make(map[string]Node),
		subjects:   make(tripleIndex),
		objects:    make(tripleIndex),
		predicates: make(tripleIndex),
	}
}

// GetNodeByID returns a node by its ID. The returned node, if not
// nil, is either a *IRI or a *BlonkNode
func (g *Graph) GetNodeByID(id string) Node {
	return g.nodeIDs[id]
}

// NewIRI creates a new IRI node and returns it. The IRI must not
// exist before, otherwise ErrNodeExists error is returned.
func (g *Graph) NewIRI(value string) (*IRI, error) {
	if _, exists := g.nodeIDs[value]; exists {
		return nil, ErrNodeExists(value)
	}
	value = g.ids.Add(value)
	iri := &IRI{id: value}
	g.nodeIDs[value] = iri
	g.nodes[iri] = struct{}{}
	return iri, nil
}

// GetIRI returns the IRI node with the given value, if one
// exists. Returns nil if no such node exists.
func (g *Graph) GetIRI(value string) *IRI {
	n, _ := g.nodeIDs[value].(*IRI)
	return n
}

// NewBlankNode creates a new blank node. If id is empty, also
// generates a new blank node id
func (g *Graph) NewBlankNode(id string) (*BlankNode, error) {
	if len(id) == 0 {
		id = shortuuid.New()
	}
	if _, exists := g.nodeIDs[id]; exists {
		return nil, ErrNodeExists(id)
	}
	id = g.ids.Add(id)
	node := &BlankNode{id: id}
	g.nodeIDs[id] = node
	g.nodes[node] = struct{}{}
	return node, nil
}

// GetBlankNode returns the blank node with the given ID. If the blank
// node does not exist, returns nil
func (g *Graph) GetBlankNode(id string) *BlankNode {
	n, _ := g.nodeIDs[id].(*BlankNode)
	return n
}

// NewLiteralNode creates a new literal
func (g *Graph) NewLiteralNode(value, typ, lang string) *Literal {
	lit := &Literal{Value: value, Type: typ, Language: lang}
	g.nodes[lit] = struct{}{}
	return lit
}

// AddTriple adds the given triple to the graph. If the nodes are not
// in the graph, they are added
func (g *Graph) AddTriple(subject, predicate, object Node) *Triple {
	if _, ok := subject.(IDNode); !ok {
		panic("Subject is not an IDNode")
	}
	if _, ok := predicate.(*IRI); !ok {
		panic("Predicate is not an IRI")
	}
	if object == nil {
		panic("Null object")
	}
	subject = g.ensureNode(subject)
	predicate = g.ensureNode(predicate)
	object = g.ensureNode(object)
	return g.addTriple(subject, predicate, object)
}

func (g *Graph) ensureNode(node Node) Node {
	_, exists := g.nodes[node]
	if exists {
		return node
	}
	if idNode, ok := node.(IDNode); ok {
		id := idNode.GetID()
		found := g.nodeIDs[id]
		if found != nil {
			return found
		}
		if iri, ok := node.(*IRI); ok {
			node, _ := g.NewIRI(id)
			return node
		}
		node, _ := g.NewBlankNode(id)
		return node
	}
	// Literal node
	l := node.(*Literal)
	return g.NewLiteralNode(l.Value, l.Type, l.Language)
}

// Adds a triple if the triple does not already exist
func (g *Graph) addTriple(subject, predicate, object Node) *Triple {
	t := Triple{subject: subject,
		predicate: predicate.(*IRI),
		object:    object}

	if existing := g.triples[t]; existing != nil {
		return existing
	}
	g.triples[t] = &t
	g.subjects.add(t.subject, &t)
	g.objects.add(t.object, &t)
	g.predicates.add(t.predicate, &t)
	return &t
}

// GetTriple looks up a triple in the graph
func (g *Graph) GetTriple(subject, predicate, object Node) *Triple {
	return g.triples[Triple{subject: subject,
		predicate: predicate.(*IRI),
		object:    object}]
}

// Remove removes the triple if it is in the graph
func (g *Graph) Remove(t *Triple) {
	t = g.triples[*t]
	if t == nil {
		return
	}
	delete(g.triples, *t)
	g.subjects.remove(t.subject, t)
	g.objects.remove(t.object, t)
	g.predicates.remove(t.predicate, t)
}

// RemoveTriple removes a triple if it exists in the graph
func (g *Graph) RemoveTriple(subject, predicate, object Node) {
	g.Remove(&Triple{subject: subject, predicate: predicate.(*IRI), object: object})
}

// AllTriples returns all triples in the graph
func (g *Graph) AllTriples() []*Triple {
	ret := make([]*Triple, 0, len(g.triples))
	for _, x := range g.triples {
		ret = append(ret, x)
	}
	return ret
}

// AddQuads adds the given quads to the graph. It may return duplicate
// node error
func (g *Graph) AddQuads(quads []*ld.Quad) error {
	for _, q := range quads {
		if err := g.AddQuad(q); err != nil {
			return err
		}
	}
	return nil
}

// AddQuad adds the triple of the given quad to the graph.
func (g *Graph) AddQuad(quad *ld.Quad) error {
	addLdNode := func(node ld.Node) (Node, error) {
		switch k := node.(type) {
		case *ld.IRI:
			if iri := g.GetIRI(k.Value); iri != nil {
				return iri, nil
			}		if node==nil {
			panic("Null node")

			iri, err := g.NewIRI(k.Value)
			if err != nil {
				return nil, err
			}
			return iri, nil

		case *ld.BlankNode:
			if b := g.GetBlankNode(k.Value); b != nil {
				return b, nil
			}
			b, err := g.NewBlankNode(k.Value)
			if err != nil {
				return nil, err
			}
			return b, nil

		case *ld.Literal:
			return g.NewLiteralNode(k.Value, k.DataType, k.Language)
		}
		return nil, ErrUnsupportedNode
	}
}
