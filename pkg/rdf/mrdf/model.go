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
package mrdf

import (
	"container/list"
	"fmt"

	"github.com/piprate/json-gold/ld"

	"github.com/cloudprivacylabs/lsa/pkg/rdf"
)

type ErrUnrecognizedNode struct {
	node interface{}
}

func (e ErrUnrecognizedNode) Error() string { return fmt.Sprintf("Unrecognized Node: %v", e.node) }

type node struct {
	rdf.Node
}

type triple struct {
	subject   string
	predicate string
	objectLit rdf.BasicLiteral
	objectID  string
}

func makeTriple(t rdf.Triple) triple {
	tr := triple{
		subject:   t.Subject.GetID(),
		predicate: t.Predicate.GetID(),
	}
	if lit, ok := t.Object.(rdf.Literal); ok {
		tr.objectLit = rdf.BasicLiteral{Value: lit.GetValue(), Type: lit.GetType(), Language: lit.GetLanguage()}
	}
	if id, ok := t.Object.(rdf.IDNode); ok {
		tr.objectID = id.GetID()
	}
	return tr
}

// G is the in-memory graph implementation
type G struct {
	// list of *node objects
	allNodes list.List

	// All id nodes that are not predicates
	idNodes map[string]*node
	triples map[triple][3]*node
}

func (g *G) newNode(n rdf.Node) *node {
	ret := &node{Node: n}
	g.allNodes.PushBack(ret)
	return ret
}

// NewGraph returns a new empty graph
func NewGraph() *G {
	return &G{idNodes: make(map[string]*node), triples: make(map[triple][3]*node)}
}

// AddTriple adds a triple to the graph. Returns true if triple was
// inserted. Returns false if triple already exists. If nodes are not in the graph, they are inserted
func (g *G) AddTriple(t rdf.Triple) bool {
	tr := makeTriple(t)
	if _, exists := g.triples[tr]; exists {
		return false
	}

	subjectNode := g.idNodes[tr.subject]
	if subjectNode == nil {
		subjectNode = g.newNode(rdf.ToBasicNode(t.Subject))
		g.idNodes[tr.subject] = subjectNode
	}

	predicateNode := g.newNode(rdf.ToBasicNode(t.Predicate))

	var objectNode *node
	if idnode, ok := t.Object.(rdf.IDNode); ok {
		objectNode = g.idNodes[idnode.GetID()]
		if objectNode == nil {
			objectNode = g.newNode(rdf.ToBasicNode(t.Object))
			g.idNodes[idnode.GetID()] = objectNode
		}
	} else {
		objectNode = g.newNode(rdf.ToBasicNode(t.Object))
	}
	g.triples[tr] = [3]*node{subjectNode, predicateNode, objectNode}
	return true
}

// RemoveTriple removes the given triple. Returns true if triple is removed
func (g *G) RemoveTriple(t rdf.Triple) bool {
	tr := makeTriple(t)
	trp, exists := g.triples[tr]
	if !exists {
		return false
	}

	delete(g.triples, tr)
	delete(g.idNodes, trp[0].Node.(rdf.IDNode).GetID())
	if id, ok := trp[2].Node.(rdf.IDNode); ok {
		delete(g.idNodes, id.GetID())
	}

	return true
}

// AddQuads adds all the quads from the jsonld library
func (g *G) AddQuads(quads []*ld.Quad) error {
	makeNode := func(in ld.Node) (rdf.Node, error) {
		switch t := in.(type) {
		case *ld.IRI:
			return rdf.BasicIRI(t.GetValue()), nil

		case *ld.BlankNode:
			return rdf.BasicBlankNode(t.GetValue()), nil

		case *ld.Literal:
			return rdf.BasicLiteral{Value: t.Value, Type: t.Datatype, Language: t.Language}, nil

		}
		return nil, ErrUnrecognizedNode{in}
	}
	for _, q := range quads {
		subject, err := makeNode(q.Subject)
		if err != nil {
			return err
		}
		pred, err := makeNode(q.Predicate)
		if err != nil {
			return err
		}
		obj, err := makeNode(q.Object)
		if err != nil {
			return err
		}
		g.AddTriple(rdf.Triple{Subject: subject.(rdf.IDNode), Predicate: pred.(rdf.IRI), Object: obj})
	}
	return nil
}

func (g *G) ToGraph() (nodes []rdf.GraphNode, edges [][2]string) {
	nm := make(map[*node]string)
	i := 0
	for el := g.allNodes.Front(); el != nil; el = el.Next() {
		node := el.Value.(*node)
		id := fmt.Sprintf("n_%d", i)
		nm[node] = id
		nodes = append(nodes, rdf.GraphNode{Node: node.Node, ID: id})
		i++
	}
	edg := make(map[[2]string]struct{})
	for _, tr := range g.triples {
		e := [2]string{nm[tr[0]], nm[tr[1]]}
		if _, ok := edg[e]; !ok {
			edges = append(edges, e)
			edg[e] = struct{}{}
		}
		e = [2]string{nm[tr[1]], nm[tr[2]]}
		if _, ok := edg[e]; !ok {
			edges = append(edges, e)
			edg[e] = struct{}{}
		}
	}
	return
}
