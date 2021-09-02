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

	"github.com/bserdar/digraph"
)

// Types is a set of strings for representing node types
type Types struct {
	slice []string
}

// Add adds new types. The result is the set-union of the existing
// types and the given types
func (types *Types) Add(t ...string) {
	for i := range t {
		x := knownTerm(t[i])
		if !types.Has(x) {
			types.slice = append(types.slice, x)
		}
	}
}

// AddTypes adds all the types into this one
func (types *Types) AddTypes(t ...*Types) {
	set := make(map[string]struct{})
	for _, k := range types.slice {
		set[k] = struct{}{}
	}
	for _, tt := range t {
		for _, k := range tt.slice {
			if _, exists := set[k]; !exists {
				types.slice = append(types.slice, k)
				set[k] = struct{}{}
			}
		}
	}
}

func (types Types) Slice() []string { return types.slice }
func (types Types) Len() int        { return len(types.slice) }

// Remove removes the given set of types from the node.
func (types *Types) Remove(t ...string) {
	types.slice = StringSetSubtract(types.slice, t)
}

// Set sets the types
func (types *Types) Set(t ...string) {
	types.slice = make([]string, 0, len(t))
	types.Add(t...)
}

// Has returns true if the set has the given type
func (types Types) Has(t string) bool {
	for _, x := range types.slice {
		if t == x {
			return true
		}
	}
	return false
}

func (types Types) String() string {
	return fmt.Sprint(types.slice)
}

// Node is the node type used for schema layer graphs
type Node interface {
	digraph.Node

	// Return the types of the node
	GetTypes() *Types

	// Return node ID
	GetID() string

	// Set node ID
	SetID(string)

	// Clone returns a new node that is a copy of this one, but the
	// returned node is not connected
	Clone() Node

	// Value of the document node, nil if the node is not a document node
	GetValue() interface{}

	SetValue(interface{})

	GetIndex() int
	SetIndex(int)

	// GetProperties returns the name/value pairs of the node. The
	// values are either string or []string. When cloned, the new node
	// receives a deep copy of the map
	GetProperties() map[string]*PropertyValue

	// Returns the compiled data map. This map is used to store
	// compilation information related to the node, and its contents are
	// unspecified. If the node is cloned with compiled data map, the
	// new node will get a shallow copy of the compiled data
	GetCompiledDataMap() map[interface{}]interface{}
}

// node is either an attribute node, document node, or an annotation
// node.  The attribute nodes have types Attribute plus the specific
// type of the attribute. Other nodes will have their own types
// marking them as literal or IRI, or something else. Annotations
// cannot have Attribute or one of the attribute types
type node struct {
	digraph.NodeHeader

	// The types of the schema node
	types Types

	// value for document nodes
	value interface{}

	// Properties associated with the node. These are assumed to be JSON-types
	properties map[string]*PropertyValue
	// These can be set during compilation. They are shallow-cloned
	compiled map[interface{}]interface{}
}

func (a *node) GetCompiledDataMap() map[interface{}]interface{} {
	if a.compiled == nil {
		a.compiled = make(map[interface{}]interface{})
	}
	return a.compiled
}

func (a *node) GetProperties() map[string]*PropertyValue {
	if a.properties == nil {
		a.properties = make(map[string]*PropertyValue)
	}
	return a.properties
}

func (a *node) GetValue() interface{} { return a.value }

func (a *node) SetValue(value interface{}) { a.value = value }

func (a *node) GetIndex() int {
	p := a.properties[AttributeIndexTerm]
	if p == nil || !p.IsString() {
		return 0
	}
	return p.AsInt()
}

func (a *node) SetIndex(index int) {
	a.properties[AttributeIndexTerm] = IntPropertyValue(index)
}

func IsDocumentNode(a Node) bool {
	return a.GetTypes().Has(DocumentNodeTerm)
}

// NewNode returns a new node with the given types
func NewNode(ID string, types ...string) Node {
	ret := node{}
	ret.types.Add(types...)
	ret.SetLabel(ID)
	return &ret
}

// NewNodes allocates n empty nodes
func NewNodes(n int) []Node {
	nodes := make([]node, n)
	ret := make([]Node, n)
	for i := range nodes {
		ret[i] = &nodes[i]
	}
	return ret
}

// GetID returns the node ID
func (a *node) GetID() string {
	l := a.GetLabel()
	if l == nil {
		return ""
	}
	return l.(string)
}

// SetID sets the node ID
func (a *node) SetID(ID string) {
	a.SetLabel(ID)
}

// GetTypes returns the types of the node
func (a *node) GetTypes() *Types {
	return &a.types
}

// Connect source node with the target node using an edge with the given label
func Connect(source, target Node, edgeLabel string) Edge {
	edge := NewEdge(edgeLabel)
	digraph.Connect(source, target, edge)
	return edge
}

// IsAttributeNode returns true if the node has Attribute type
func IsAttributeNode(a Node) bool {
	return a.GetTypes().Has(AttributeTypes.Attribute)
}

// Clone returns a copy of the node data. The returned node has the
// same label, types, and properties. The Compiled map is directly
// assigned to the new node
func (a *node) Clone() Node {
	ret := NewNode(a.GetID(), a.GetTypes().Slice()...).(*node)
	ret.value = a.value
	ret.properties = CopyPropertyMap(a.properties)
	ret.compiled = a.compiled
	return ret
}

// GetAttributeEdgeBetweenNodes returns the attribute edges between
// two nodes. If there are no direct edges, return nil
func GetLayerEdgeBetweenNodes(source, target Node) Edge {
	for edges := source.Out(); edges.HasNext(); {
		edge := edges.Next().(Edge)
		if IsAttributeTreeEdge(edge) && edge.GetTo() == target {
			return edge
		}
	}
	return nil
}

// GetNodeFilteredValue returns the field value processed by the schema
// value filters, and then the node value filters
func GetNodeFilteredValue(node Node) interface{} {
	var schemaNode Node
	iedges := node.OutWith(InstanceOfTerm).All()
	if len(iedges) == 1 {
		schemaNode = iedges[0].GetTo().(Node)
	}
	return GetFilteredValue(schemaNode, node)
}

// GetFilteredValue filters the value through the schema properties
// and then through the node properties before returning
func GetFilteredValue(schemaNode, docNode Node) interface{} {
	value := docNode.GetValue()
	if schemaNode != nil {
		value = FilterValue(value, docNode, schemaNode.GetProperties())
	}
	return FilterValue(value, docNode, docNode.GetProperties())
}

// IsDocumentEdge returns true if the edge is a data edge term
func IsDocumentEdge(edge digraph.Edge) bool {
	switch edge.GetLabel() {
	case DataEdgeTerms.ObjectAttributes, DataEdgeTerms.ArrayElements:
		return true
	}
	return false
}

// SortNodes sorts nodes by their node index
func SortNodes(nodes []Node) {
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].GetIndex() < nodes[j].GetIndex() })
}

type EdgeFuncResult int

const (
	FollowEdgeResult EdgeFuncResult = iota
	SkipEdgeResult
	StopEdgeResult
)

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
func IterateDescendants(from Node, nodeFunc func(Node, []Node) bool, edgeFunc func(Edge, []Node) EdgeFuncResult, ordered bool) bool {
	return iterateDescendants(from, []Node{}, nodeFunc, edgeFunc, ordered, map[Node]struct{}{})
}

func iterateDescendants(root Node, path []Node, nodeFunc func(Node, []Node) bool, edgeFunc func(Edge, []Node) EdgeFuncResult, ordered bool, seen map[Node]struct{}) bool {
	if _, exists := seen[root]; exists {
		return true
	}
	seen[root] = struct{}{}

	path = append(path, root)

	if nodeFunc != nil && !nodeFunc(root, path) {
		return false
	}

	outgoing := root.Out()
	if ordered {
		outgoing = SortEdgesItr(outgoing)
	}

	for outgoing.HasNext() {
		edge := outgoing.Next().(Edge)
		follow := FollowEdgeResult
		if edgeFunc != nil {
			follow = edgeFunc(edge, path)
		}
		switch follow {
		case StopEdgeResult:
			return false
		case SkipEdgeResult:
		case FollowEdgeResult:
			next := edge.GetTo().(Node)
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
func FirstReachable(from Node, nodePredicate func(Node, []Node) bool, edgePredicate func(Edge, []Node) bool) (Node, []Node) {
	var (
		ret  Node
		path []Node
	)
	IterateDescendants(from, func(n Node, p []Node) bool {
		if nodePredicate(n, p) {
			ret = n
			path = p
			return false
		}
		return true
	},
		func(e Edge, p []Node) EdgeFuncResult {
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

// InstanceOf returns the transitive closure of all the nodes that are connect to this node via instanceOf term,
func InstanceOf(node Node) []Node {
	results := make(map[Node]struct{})
	IterateDescendants(node, func(n Node, p []Node) bool {
		results[n] = struct{}{}
		return true
	},
		func(e Edge, p []Node) EdgeFuncResult {
			if e.GetLabel() == InstanceOfTerm {
				return FollowEdgeResult
			}
			return SkipEdgeResult
		},
		false)
	ret := make([]Node, 0, len(results))
	for x := range results {
		ret = append(ret, x)
	}
	return ret
}

// CombineNodeTypes returns a combination of the types of all the given nodes
func CombineNodeTypes(nodes []Node) *Types {
	ret := Types{}
	t := make([]*Types, 0, len(nodes))
	for _, n := range nodes {
		t = append(t, n.GetTypes())
	}
	ret.AddTypes(t...)
	return &ret
}

// NodeSet is a set of nodes
type NodeSet struct {
	set   map[Node]struct{}
	nodes []Node
}

// NewNodeSet constructs a new nodeset from the given nodes
func NewNodeSet(node ...Node) NodeSet {
	ret := NodeSet{set: make(map[Node]struct{})}
	ret.Add(node...)
	return ret
}

// Add adds nodes to the set
func (n *NodeSet) Add(nodes ...Node) {
	for _, k := range nodes {
		if !n.Has(k) {
			n.set[k] = struct{}{}
			n.nodes = append(n.nodes, k)
		}
	}
}

// Has returns true if node is in the set
func (n NodeSet) Has(node Node) bool {
	_, ok := n.set[node]
	return ok
}

// Delete some nodes from the set
func (n *NodeSet) Delete(nodes ...Node) {
	for _, k := range nodes {
		delete(n.set, k)
	}
	w := 0
	for i := 0; i < len(n.nodes); i++ {
		if _, ok := n.set[n.nodes[i]]; ok {
			n.nodes[w] = n.nodes[i]
			w++
		}
	}
	n.nodes = n.nodes[:w]
}

// Slice returns the nodes in a nodeset as a slice
func (n NodeSet) Slice() []Node {
	return n.nodes
}

// Set returns the nodes as a map
func (n NodeSet) Map() map[Node]struct{} {
	return n.set
}

func (n NodeSet) Len() int { return len(n.nodes) }

// EqualSet returns if the two nodesets are equal without taking into account the node ordering
func (n NodeSet) EqualSet(n2 NodeSet) bool {
	if n.Len() != n2.Len() {
		return false
	}
	for k := range n.Map() {
		if !n2.Has(k) {
			return false
		}
	}
	return true
}
