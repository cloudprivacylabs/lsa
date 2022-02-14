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
	"fmt"
	"sort"

	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

// IsDocumentNode returns if the node has the DocumentNodeTerm as one of its labels
func IsDocumentNode(node graph.Node) bool {
	return node.GetLabels().Has(DocumentNodeTerm)
}

// IsAttributeNode returns true if the node has Attribute type
func IsAttributeNode(node graph.Node) bool {
	return node.GetLabels().Has(AttributeNodeTerm)
}

// GetNodeIndex returns the value of attribute index term as int
func GetNodeIndex(node graph.Node) int {
	p := AsPropertyValue(node.GetProperty(AttributeIndexTerm))
	if p == nil || !p.IsString() {
		return 0
	}
	return p.AsInt()
}

func SetNodeIndex(node graph.Node, index int) {
	node.SetProperty(AttributeIndexTerm, IntPropertyValue(index))
}

// GetNodeID returns the nodeid
func GetNodeID(node graph.Node) string {
	v, _ := node.GetProperty(NodeIDTerm)
	s, _ := v.(string)
	return s
}

// SetNodeID sets the node ID
func SetNodeID(node graph.Node, ID string) {
	node.SetProperty(NodeIDTerm, ID)
}

// GetRawNodeValue returns the unprocessed node value
func GetRawNodeValue(node graph.Node) interface{} {
	val, _ := node.GetProperty(NodeValueTerm)
	return val
}

// SetRawNodeValue sets the unprocessed node value
func SetRawNodeValue(node graph.Node, value interface{}) {
	node.SetProperty(NodeValueTerm, value)
}

// GetNodeValue returns the field value processed by the schema type
// information. The returned object is a Go native object based on the
// node type
func GetNodeValue(node graph.Node) (interface{}, error) {
	accessor, err := GetNodeValueAccessor(node)
	if err != nil {
		return nil, err
	}
	if accessor == nil {
		return GetRawNodeValue(node), nil
	}
	return accessor.GetNodeValue(node)
}

// SetNodeValue sets the node value using the given native Go
// value. The value is expected to be interpreted by the node types
// and converted to string. If there are no value accessors specified
// for the node, the value will be fmt.Sprint(value)
func SetNodeValue(node graph.Node, value interface{}) error {
	accessor, err := GetNodeValueAccessor(node)
	if err != nil {
		return nil
	}
	if accessor == nil {
		if value == nil {
			node.RemoveProperty(NodeValueTerm)
			return nil
		}
		SetRawNodeValue(node, fmt.Sprint(value))
		return nil
	}
	return accessor.SetNodeValue(node, value)
}

// GetNodeValueAccessor returns the value accessor for the node based
// on the node type. If there is none, returns nil
func GetNodeValueAccessor(node graph.Node) (ValueAccessor, error) {
	var (
		accessor ValueAccessor
		typeName string
	)

	setAccessor := func(term string) error {
		a := GetValueAccessor(term)
		if a != nil {
			if accessor != nil && typeName != term {
				return ErrInconsistentTypes{TypeNames: []string{typeName, term}}
			}
			accessor = a
			typeName = term
		}
		return nil
	}
	iedges := graph.EdgeSlice(node.GetEdgesWithLabel(graph.OutgoingEdge, InstanceOfTerm))
	if len(iedges) == 1 {
		for _, t := range iedges[0].GetTo().GetLabels().Slice() {
			if err := setAccessor(t); err != nil {
				return nil, err
			}
		}
	}
	for _, t := range node.GetLabels().Slice() {
		if err := setAccessor(t); err != nil {
			return nil, err
		}
	}
	return accessor, nil
}

// IsDocumentEdge returns true if the edge is not an attribute link term
func IsDocumentEdge(edge graph.Edge) bool {
	return !IsAttributeTreeEdge(edge)
}

// SortNodes sorts nodes by their node index
func SortNodes(nodes []graph.Node) {
	sort.Slice(nodes, func(i, j int) bool {
		return GetNodeIndex(nodes[i]) < GetNodeIndex(nodes[j])
	})
}

type EdgeFuncResult int

const (
	FollowEdgeResult EdgeFuncResult = iota
	SkipEdgeResult
	StopEdgeResult
)

// SkipEdgesToNodeWithType returns a function that skips edges that go
// to a node with the given type
func SkipEdgesToNodeWithType(typ string) func(graph.Edge, []graph.Node) EdgeFuncResult {
	return func(edge graph.Edge, _ []graph.Node) EdgeFuncResult {
		if edge.GetTo().GetLabels().Has(typ) {
			return SkipEdgeResult
		}
		return FollowEdgeResult
	}
}

// FollowEdgesToNodeWithType returns a function that only follows edges that go
// to a node with the given type
func FollowEdgesToNodeWithType(typ string) func(graph.Edge, []graph.Node) EdgeFuncResult {
	return func(edge graph.Edge, _ []graph.Node) EdgeFuncResult {
		if edge.GetTo().GetLabels().Has(typ) {
			return FollowEdgeResult
		}
		return SkipEdgeResult
	}
}

// SkipSchemaNodes can be used in IterateDescendants edge func
// to skip all edges that go to a schema node
var SkipSchemaNodes = SkipEdgesToNodeWithType(AttributeNodeTerm)

// SkipDocumentNodes can be used in IterateDescendants edge func
// to skip all edges that go to a document node
var SkipDocumentNodes = SkipEdgesToNodeWithType(DocumentNodeTerm)

// OnlyDocumentNodes can be used in IterateDescendants edge func to
// follow edges that reach to document nodes
var OnlyDocumentNodes = FollowEdgesToNodeWithType(DocumentNodeTerm)

// IterateDescendants iterates the descendants of the node based on
// the results of nodeFunc and edgeFunc.
//
// For each visited node, if nodeFunc is not nil, nodeFunc is called
// with the node and the path to the node. If nodeFunc returns false,
// processing stops.
//
// For each outgoing edge, if edgeFunc is not nil, edgeFunc is called
// with the edge and the path to the source node. If edgeFunc returns
// FollowEdgeResult, the edge is followed. If edgeFunc returnd
// DontFollowEdgeResult, edge is skipped. If edgeFunc returns
// StopEdgeResult, iteration stops.
func IterateDescendants(from graph.Node, nodeFunc func(graph.Node, []graph.Node) bool, edgeFunc func(graph.Edge, []graph.Node) EdgeFuncResult, ordered bool) bool {
	return iterateDescendants(from, []graph.Node{}, nodeFunc, edgeFunc, ordered, map[graph.Node]struct{}{})
}

func iterateDescendants(root graph.Node, path []graph.Node, nodeFunc func(graph.Node, []graph.Node) bool, edgeFunc func(graph.Edge, []graph.Node) EdgeFuncResult, ordered bool, seen map[graph.Node]struct{}) bool {
	if _, exists := seen[root]; exists {
		return true
	}
	seen[root] = struct{}{}

	path = append(path, root)

	if nodeFunc != nil && !nodeFunc(root, path) {
		return false
	}

	outgoing := root.GetEdges(graph.OutgoingEdge)
	if ordered {
		outgoing = SortEdgesItr(outgoing)
	}

	for outgoing.Next() {
		edge := outgoing.Edge()
		follow := FollowEdgeResult
		if edgeFunc != nil {
			follow = edgeFunc(edge, path)
		}
		switch follow {
		case StopEdgeResult:
			return false
		case SkipEdgeResult:
		case FollowEdgeResult:
			next := edge.GetTo()
			if !iterateDescendants(next, path, nodeFunc, edgeFunc, ordered, seen) {
				return false
			}
		}
	}
	return true

}

// FirstReachable returns the first reachable node for which
// nodePredicate returns true, using only the edges for which
// edgePredicate returns true.
func FirstReachable(from graph.Node, nodePredicate func(graph.Node, []graph.Node) bool, edgePredicate func(graph.Edge, []graph.Node) bool) (graph.Node, []graph.Node) {
	var (
		ret  graph.Node
		path []graph.Node
	)
	IterateDescendants(from, func(n graph.Node, p []graph.Node) bool {
		if nodePredicate(n, p) {
			ret = n
			path = p
			return false
		}
		return true
	},
		func(e graph.Edge, p []graph.Node) EdgeFuncResult {
			if edgePredicate == nil {
				return FollowEdgeResult
			}
			if edgePredicate(e, p) {
				return FollowEdgeResult
			}
			return SkipEdgeResult
		},
		true)
	return ret, path
}

// // InstanceOfID returns the IDs of the schema nodes this node is an instance of
// func InstanceOfID(node Node) []string {
// 	out := make(map[string]struct{})
// 	ForEachInstanceOf(node, func(n Node) bool {
// 		v, has := n.GetProperties()[InstanceOfTerm]
// 		if has {
// 			if v.IsString() {
// 				out[v.AsString()] = struct{}{}
// 			} else if v.IsStringSlice() {
// 				for _, x := range v.AsStringSlice() {
// 					out[x] = struct{}{}
// 				}
// 			}
// 		}
// 		if IsAttributeNode(n) {
// 			out[n.GetID()] = struct{}{}
// 		}
// 		return true
// 	})
// 	ret := make([]string, 0, len(out))
// 	for x := range out {
// 		ret = append(ret, x)
// 	}
// 	return ret
// }

// // ForEachInstanceOf traverses the transitive closure of all nodes
// // connected to the given nodes by instanceOf, and calls f until f
// // returns false or all nodes are traversed
// func ForEachInstanceOf(node Node, f func(Node) bool) {
// 	IterateDescendants(node, func(n Node, p []Node) bool {
// 		return f(n)
// 	},
// 		func(e Edge, p []Node) EdgeFuncResult {
// 			if e.GetLabel() == InstanceOfTerm {
// 				return FollowEdgeResult
// 			}
// 			return SkipEdgeResult
// 		},
// 		false)
// }

// InstanceOf returns the nodes that are connect to this node via
// instanceOf term,
func InstanceOf(node graph.Node) []graph.Node {
	return graph.NextNodesWith(node, InstanceOfTerm)
}

// CombineNodeTypes returns a combination of the types of all the given nodes
func CombineNodeTypes(nodes []graph.Node) graph.StringSet {
	ret := graph.NewStringSet()
	for _, n := range nodes {
		for x := range n.GetLabels() {
			ret.Add(x)
		}
	}
	return ret
}

// // DocumentNodesUnder returns all document nodes under the given node(s)
// func DocumentNodesUnder(node ...Node) []Node {
// 	input := make([]digraph.Node, 0, len(node))
// 	for _, x := range node {
// 		input = append(input, x)
// 	}
// 	itr := digraph.NewNodeWalkIterator(input...).Select(func(n digraph.Node) bool {
// 		lsnode, ok := n.(Node)
// 		if !ok {
// 			return false
// 		}
// 		if !lsnode.GetTypes().Has(DocumentNodeTerm) {
// 			return false
// 		}
// 		return true
// 	})
// 	all := itr.All()
// 	ret := make([]Node, 0, len(all))
// 	for _, x := range all {
// 		ret = append(ret, x.(Node))
// 	}
// 	return ret
// }

// GetNodeOrSchemaProperty gets the node property with the key from
// the node, or from the schema nodes it is attached to
func GetNodeOrSchemaProperty(node graph.Node, key string) (*PropertyValue, bool) {
	prop, _ := node.GetProperty(key)
	if pd, ok := prop.(*PropertyValue); ok {
		return pd, true
	}
	for _, n := range InstanceOf(node) {
		prop, _ = n.GetProperty(key)
		if pd, ok := prop.(*PropertyValue); ok {
			return pd, true
		}
	}
	return nil, false
}

// IsAttributeTreeEdge returns true if the edge is an edge between two
// attribute nodes
func IsAttributeTreeEdge(edge graph.Edge) bool {
	if edge == nil {
		return false
	}
	l := edge.GetLabel()
	return l == ObjectAttributesTerm ||
		l == ObjectAttributeListTerm ||
		l == ArrayItemsTerm ||
		l == AllOfTerm ||
		l == OneOfTerm
}

// SortEdges sorts edges by their target node index
func SortEdges(edges []graph.Edge) []graph.Edge {
	sort.Slice(edges, func(i, j int) bool {
		return GetNodeIndex(edges[i].GetTo().(graph.Node)) < GetNodeIndex(edges[j].GetTo().(graph.Node))
	})
	return edges
}

// SortEdgesItr sorts the edges by index
func SortEdgesItr(edges graph.EdgeIterator) graph.EdgeIterator {
	e := make([]graph.Edge, 0)
	for edges.Next() {
		e = append(e, edges.Edge())
	}
	SortEdges(e)
	return graph.NewEdgeIterator(e...)
}

// CloneNode clones the sourcenode in targetgraph
func CloneNode(sourceNode graph.Node, targetGraph graph.Graph) graph.Node {
	properties := make(map[string]interface{})
	sourceNode.ForEachProperty(func(key string, value interface{}) bool {
		if p, ok := value.(*PropertyValue); ok {
			properties[key] = p.Clone()
		} else {
			properties[key] = value
		}
		return true
	})
	newNode := targetGraph.NewNode(sourceNode.GetLabels().Slice(), properties)
	return newNode
}

func CloneEdge(fromInTarget, toInTarget graph.Node, sourceEdge graph.Edge, targetGraph graph.Graph) graph.Edge {
	properties := make(map[string]interface{})
	sourceEdge.ForEachProperty(func(key string, value interface{}) bool {
		if p, ok := value.(*PropertyValue); ok {
			properties[key] = p.Clone()
		} else {
			properties[key] = value
		}
		return true
	})
	return targetGraph.NewEdge(fromInTarget, toInTarget, sourceEdge.GetLabel(), properties)
}