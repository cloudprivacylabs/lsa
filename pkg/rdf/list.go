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

// ListBuilder builds RDF lists. The first node is a blank
// node. Subsequent nodes are added to that blank node. Triples list
// can be added, the list builder keeps the index of the last node in
// the list. Do not insert or shuffle nodes.
type ListBuilder struct {
	Triples []Triple
	last    int
}

// AddNode adds a new node to the list
func (l *ListBuilder) AddNode(node Node) {
	newBlankNode := NewBasicBlankNode()
	if len(l.Triples) != 0 {
		l.Triples[l.last].Object = newBlankNode
	}
	l.Triples = append(l.Triples, Triple{Subject: newBlankNode, Predicate: BasicIRI(RDFFirst), Object: node})
	l.Triples = append(l.Triples, Triple{Subject: newBlankNode, Predicate: BasicIRI(RDFRest), Object: BasicIRI(RDFNil)})
	l.last = len(l.Triples) - 1
}
