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

import (
	"strings"

	"github.com/bserdar/digraph"
)

// EvaluatePathExpression evaluates the path expression expr from the
// node. The path expression contains a sequence of edge labels and
// property predicates of the form:
//
//   edgeLabel property=value  edgeLabel property=value ...
func EvaluatePathExpression(nodes []Node, expr ...string) []Node {
	if len(nodes) == 0 {
		return nil
	}
	current := map[Node]struct{}{}
	for _, x := range nodes {
		current[x] = struct{}{}
	}
	edgeLabel := true
	for _, curExpr := range expr {
		nextNodes := make(map[Node]struct{})
		if edgeLabel {
			// Advance current to the next nodes
			for n := range current {
				var edges digraph.Edges
				if len(curExpr) == 0 {
					edges = n.GetAllOutgoingEdges()
				} else {
					edges = n.GetAllOutgoingEdgesWithLabel(curExpr)
				}
				for edges.HasNext() {
					nextNodes[edges.Next().GetTo().(Node)] = struct{}{}
				}
			}
		} else {
			// Select nodes matching predicate
			type idSupport interface{ GetID() string }
			type valueSupport interface{ GetValue() interface{} }
			type propSupport interface {
				GetProperties() map[string]*PropertyValue
			}
			filter := func(f func(digraph.Node) bool) {
				for x := range current {
					if f(x) {
						nextNodes[x] = struct{}{}
					}
				}
			}
			pieces := strings.Split(curExpr, "=")
			switch len(pieces) {
			case 1:
				switch pieces[0] {
				case "":
					nextNodes = current
				case "@id":
					filter(func(n digraph.Node) bool {
						_, ok := n.(idSupport)
						return ok
					})
				case "@value":
					filter(func(n digraph.Node) bool {
						_, ok := n.(valueSupport)
						return ok
					})
				default:
					filter(func(n digraph.Node) bool {
						p, ok := n.(propSupport)
						if ok {
							_, ok = p.GetProperties()[pieces[0]]
						}
						return ok
					})
				}
			case 2:
				switch pieces[0] {
				case "@id":
					filter(func(n digraph.Node) bool {
						s, ok := n.(idSupport)
						return ok && s.GetID() == pieces[1]
					})
				case "@value":
					filter(func(n digraph.Node) bool {
						s, ok := n.(valueSupport)
						return ok && s.GetValue() == pieces[1]
					})
				default:
					filter(func(n digraph.Node) bool {
						p, ok := n.(propSupport)
						if !ok {
							return false
						}
						s, ok := p.GetProperties()[pieces[0]]
						if !ok {
							return false
						}
						return s.Has(pieces[1])
					})
				}
			}
		}
		current = nextNodes
		edgeLabel = !edgeLabel
	}
	ret := make([]Node, 0, len(current))
	for x := range current {
		ret = append(ret, x)
	}
	return ret
}
