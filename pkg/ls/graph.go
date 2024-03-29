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

	"github.com/cloudprivacylabs/lpg/v2"
)

// IsDocumentNode returns if the node has the DocumentNodeTerm as one of its labels
func IsDocumentNode(node *lpg.Node) bool {
	return node.GetLabels().Has(DocumentNodeTerm.Name)
}

// IsAttributeNode returns true if the node has Attribute type
func IsAttributeNode(node *lpg.Node) bool {
	return node.GetLabels().Has(AttributeNodeTerm.Name)
}

// GetNodeIndex returns the value of attribute index term as int
func GetNodeIndex(node *lpg.Node) int {
	v, _ := GetPropertyValueAs[int](node, AttributeIndexTerm.Name)
	return v
}

func SetNodeIndex(node *lpg.Node, index int) {
	node.SetProperty(AttributeIndexTerm.Name, AttributeIndexTerm.MustPropertyValue(index))
}

// GetNodeID returns the nodeid
func GetNodeID(node *lpg.Node) string {
	return NodeIDTerm.PropertyValue(node)
}

// SetNodeID sets the node ID
func SetNodeID(node *lpg.Node, ID string) {
	node.SetProperty(NodeIDTerm.Name, NewPropertyValue(NodeIDTerm.Name, ID))
}

// GetRawNodeValue returns the unprocessed node value
func GetRawNodeValue(node *lpg.Node) (string, bool) {
	return GetPropertyValueAs[string](node, NodeValueTerm.Name)
}

func RemoveRawNodeValue(node *lpg.Node) {
	node.RemoveProperty(NodeValueTerm.Name)
}

// SetRawNodeValue sets the unprocessed node value
func SetRawNodeValue(node *lpg.Node, value string) {
	node.SetProperty(NodeValueTerm.Name, NewPropertyValue(NodeValueTerm.Name, value))
}

// GetNodeValue returns the field value processed by the schema type
// information. The returned object is a Go native object based on the
// node type
func GetNodeValue(node *lpg.Node) (interface{}, error) {
	accessor, err := GetNodeValueAccessor(node)
	if err != nil {
		return nil, err
	}
	if nac, ok := accessor.(NodeValueAccessor); ok {
		return nac.GetNodeValue(node)
	}
	v, ok := GetRawNodeValue(node)
	if !ok {
		return nil, nil
	}
	if accessor == nil {
		return v, nil
	}
	return accessor.GetNativeValue(v, node)
}

// SetNodeValue sets the node value using the given native Go
// value. The value is expected to be interpreted by the node types
// and converted to string. If there are no value accessors specified
// for the node, the value will be fmt.Sprint(value)
func SetNodeValue(node *lpg.Node, value interface{}) error {
	accessor, err := GetNodeValueAccessor(node)
	if err != nil {
		return nil
	}
	if accessor == nil {
		if value == nil {
			node.RemoveProperty(NodeValueTerm.Name)
			return nil
		}
		SetRawNodeValue(node, fmt.Sprint(value))
		return nil
	}
	if nac, ok := accessor.(NodeValueAccessor); ok {
		return nac.SetNodeValue(value, node)
	}
	var oldValue interface{}
	if v, ok := GetRawNodeValue(node); ok {
		oldValue, err = accessor.GetNativeValue(v, node)
		if err != nil {
			return err
		}
	}
	svalue, err := accessor.FormatNativeValue(value, oldValue, node)
	if err != nil {
		return err
	}
	SetRawNodeValue(node, svalue)
	return nil
}

// GetNodeValueAccessor returns the value accessor for the node based
// on the node value type. If there is none, returns nil
func GetNodeValueAccessor(node *lpg.Node) (ValueAccessor, error) {
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
	typeFound := false
	p, _ := node.GetProperty(ValueTypeTerm.Name)
	if pv, ok := p.(PropertyValue); ok {
		typeFound = true
		for _, x := range pv.AsStringSlice() {
			if err := setAccessor(x); err != nil {
				return nil, err
			}
		}
	}
	if !typeFound {
		iedges := lpg.EdgeSlice(node.GetEdgesWithLabel(lpg.OutgoingEdge, InstanceOfTerm.Name))
		if len(iedges) == 1 {
			for t := range iedges[0].GetTo().GetLabels().M {
				if err := setAccessor(t); err != nil {
					return nil, err
				}
			}
		}
	}
	return accessor, nil
}

// IsDocumentEdge returns true if the edge is not an attribute link term
func IsDocumentEdge(edge *lpg.Edge) bool {
	return !IsAttributeTreeEdge(edge)
}

// SortNodes sorts nodes by their node index
func SortNodes(nodes []*lpg.Node) {
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
func SkipEdgesToNodeWithType(typ string) func(*lpg.Edge) EdgeFuncResult {
	return func(edge *lpg.Edge) EdgeFuncResult {
		if edge.GetTo().GetLabels().Has(typ) {
			return SkipEdgeResult
		}
		return FollowEdgeResult
	}
}

// FollowEdgesToNodeWithType returns a function that only follows edges that go
// to a node with the given type
func FollowEdgesToNodeWithType(typ string) func(*lpg.Edge) EdgeFuncResult {
	return func(edge *lpg.Edge) EdgeFuncResult {
		if edge.GetTo().GetLabels().Has(typ) {
			return FollowEdgeResult
		}
		return SkipEdgeResult
	}
}

// FollowEdgesInEntity follows only the document edges that do not cross entity boundaries
func FollowEdgesInEntity(edge *lpg.Edge) EdgeFuncResult {
	if _, ok := GetNodeOrSchemaProperty(edge.GetTo(), EntitySchemaTerm.Name); ok {
		return SkipEdgeResult
	}
	return FollowEdgeResult
}

// IsNodeEntityRoot checks if node is an entity root
func IsNodeEntityRoot(node *lpg.Node) bool {
	_, ok := GetNodeOrSchemaProperty(node, EntitySchemaTerm.Name)
	return ok
}

// SkipSchemaNodes can be used in IterateDescendants edge func
// to skip all edges that go to a schema node
var SkipSchemaNodes = SkipEdgesToNodeWithType(AttributeNodeTerm.Name)

// SkipDocumentNodes can be used in IterateDescendants edge func
// to skip all edges that go to a document node
var SkipDocumentNodes = SkipEdgesToNodeWithType(DocumentNodeTerm.Name)

// OnlyDocumentNodes can be used in IterateDescendants edge func to
// follow edges that reach to document nodes
var OnlyDocumentNodes = FollowEdgesToNodeWithType(DocumentNodeTerm.Name)

// OnlyAttributeNodes can be used in IterateDescendants edge func to
// follow edges that reach to Attribute nodes
var OnlyAttributeNodes = FollowEdgesToNodeWithType(AttributeNodeTerm.Name)

// IterateDescendants iterates the descendants of the node based on
// the results of nodeFunc and edgeFunc.
//
// For each visited node, if nodeFunc is not nil, nodeFunc is called
// with the node and the path to the node. If nodeFunc returns false,
// processing stops.
//
// For each outgoing edge, if edgeFunc is not nil, edgeFunc is called
// with the edge and the path to the source node. If edgeFunc returns
// FollowEdgeResult, the edge is followed. If edgeFunc returned
// DontFollowEdgeResult, edge is skipped. If edgeFunc returns
// StopEdgeResult, iteration stops.
func IterateDescendants(from *lpg.Node, nodeFunc func(*lpg.Node) bool, edgeFunc func(*lpg.Edge) EdgeFuncResult, ordered bool) bool {
	return iterateDescendants(from, func(node *lpg.Node, _ []*lpg.Node) bool {
		return nodeFunc(node)
	}, edgeFunc, ordered, make([]*lpg.Node, 0, 16), map[*lpg.Node]struct{}{}, false)
}

func IterateDescendantsp(from *lpg.Node, nodeFunc func(*lpg.Node, []*lpg.Node) bool, edgeFunc func(*lpg.Edge) EdgeFuncResult, ordered bool, visitSeenNodes bool) bool {
	return iterateDescendants(from, nodeFunc, edgeFunc, ordered, make([]*lpg.Node, 0, 16), map[*lpg.Node]struct{}{}, visitSeenNodes)
}

func iterateDescendants(root *lpg.Node, nodeFunc func(*lpg.Node, []*lpg.Node) bool, edgeFunc func(*lpg.Edge) EdgeFuncResult, ordered bool, path []*lpg.Node, seen map[*lpg.Node]struct{}, visitSeenNodes bool) bool {
	_, seenNode := seen[root]
	if seenNode && !visitSeenNodes {
		return true
	}
	if !seenNode {
		seen[root] = struct{}{}
	}

	path = append(path, root)
	if nodeFunc != nil && !nodeFunc(root, path) {
		return false
	}
	if seenNode {
		return true
	}

	outgoing := root.GetEdges(lpg.OutgoingEdge)
	if ordered {
		outgoing = SortEdgesItr(outgoing)
	}

	for outgoing.Next() {
		edge := outgoing.Edge()
		follow := FollowEdgeResult
		if edgeFunc != nil {
			follow = edgeFunc(edge)
		}
		switch follow {
		case StopEdgeResult:
			return false
		case SkipEdgeResult:
		case FollowEdgeResult:
			next := edge.GetTo()
			if !iterateDescendants(next, nodeFunc, edgeFunc, ordered, path, seen, visitSeenNodes) {
				return false
			}
		}
	}
	return true
}

// IterateAncestors iterates the ancestors of the node, calling
// nodeFunc for each node, and edgeFunc for each edge. If nodeFunc
// returns false, stops iteration and returns. The behavior after
// calling edgefunc depends on the return value. The edgeFunc may
// skip the edge, follow it, or stop processing.
func IterateAncestors(root *lpg.Node, nodeFunc func(*lpg.Node) bool, edgeFunc func(*lpg.Edge) EdgeFuncResult) bool {
	seen := make(map[*lpg.Node]struct{})
	var f func(*lpg.Node) bool
	f = func(node *lpg.Node) bool {
		if _, exists := seen[node]; exists {
			return true
		}
		seen[node] = struct{}{}
		if nodeFunc != nil && !nodeFunc(node) {
			return false
		}
		for incoming := node.GetEdges(lpg.IncomingEdge); incoming.Next(); {
			edge := incoming.Edge()
			follow := FollowEdgeResult
			if edgeFunc != nil {
				follow = edgeFunc(edge)
			}
			switch follow {
			case StopEdgeResult:
				return false
			case SkipEdgeResult:
			case FollowEdgeResult:
				next := edge.GetFrom()
				if !f(next) {
					return false
				}
			}
		}
		return true
	}
	return f(root)
}

// InstanceOf returns the nodes that are connect to this node via
// instanceOf term,
func InstanceOf(node *lpg.Node) []*lpg.Node {
	return lpg.NextNodesWith(node, InstanceOfTerm.Name)
}

// CombineNodeTypes returns a combination of the types of all the given nodes
func CombineNodeTypes(nodes []*lpg.Node) lpg.StringSet {
	ret := lpg.NewStringSet()
	for _, n := range nodes {
		for x := range n.GetLabels().M {
			ret.Add(x)
		}
	}
	return ret
}

// DocumentNodesUnder returns all document nodes under the given node(s)
func DocumentNodesUnder(node ...*lpg.Node) []*lpg.Node {
	set := make(map[*lpg.Node]struct{})
	for _, x := range node {
		IterateDescendants(x, func(n *lpg.Node) bool {
			if IsDocumentNode(n) {
				set[n] = struct{}{}
			}
			return true
		}, func(e *lpg.Edge) EdgeFuncResult {
			if IsDocumentNode(e.GetTo()) {
				return FollowEdgeResult
			}
			return SkipEdgeResult
		}, false)
	}
	ret := make([]*lpg.Node, 0, len(set))
	for x := range set {
		ret = append(ret, x)
	}
	return ret
}

// GetNodeOrSchemaProperty gets the node property with the key from
// the node, or from the schema nodes it is attached to
func GetNodeOrSchemaProperty(node *lpg.Node, key string) (PropertyValue, bool) {
	prop, ok := GetPropertyValue(node, key)
	if ok {
		return prop, true
	}
	for _, n := range InstanceOf(node) {
		prop, ok = GetPropertyValue(n, key)
		if ok {
			return prop, true
		}
	}
	return PropertyValue{}, false
}

// GetNodeSchemaNodeID returns the schema node ID of a document node. Returns empty string if not found.
func GetNodeSchemaNodeID(documentNode *lpg.Node) string {
	p, ok := GetNodeOrSchemaProperty(documentNode, SchemaNodeIDTerm.Name)
	if !ok {
		return ""
	}
	if s, ok := p.Value().(string); ok {
		return s
	}
	return ""
}

// IsAttributeTreeEdge returns true if the edge is an edge between two
// attribute nodes
func IsAttributeTreeEdge(edge *lpg.Edge) bool {
	if edge == nil {
		return false
	}
	l := edge.GetLabel()
	return l == ObjectAttributesTerm.Name ||
		l == ObjectAttributeListTerm.Name ||
		l == ArrayItemsTerm.Name ||
		l == AllOfTerm.Name ||
		l == OneOfTerm.Name
}

// SortEdges sorts edges by their target node index
func SortEdges(edges []*lpg.Edge) []*lpg.Edge {
	sort.Slice(edges, func(i, j int) bool {
		return GetNodeIndex(edges[i].GetTo()) < GetNodeIndex(edges[j].GetTo())
	})
	return edges
}

type sortedEdgeIterator struct {
	edges   []*lpg.Edge
	current *lpg.Edge
}

func (n *sortedEdgeIterator) Next() bool {
	if len(n.edges) == 0 {
		return false
	}
	n.current = n.edges[0]
	n.edges = n.edges[1:]
	return true
}

func (n *sortedEdgeIterator) Value() interface{} {
	return n.current
}

func (n *sortedEdgeIterator) Edge() *lpg.Edge {
	return n.current
}

func (n *sortedEdgeIterator) MaxSize() int {
	return len(n.edges)
}

// SortEdgesItr sorts the edges by index
func SortEdgesItr(edges lpg.EdgeIterator) lpg.EdgeIterator {
	e := make([]*lpg.Edge, 0)
	for edges.Next() {
		e = append(e, edges.Edge())
	}
	SortEdges(e)
	return &sortedEdgeIterator{edges: e}
}

// CloneNode clones the sourcenode in targetgraph
func CloneNode(sourceNode *lpg.Node, targetGraph *lpg.Graph) *lpg.Node {
	return lpg.CopyNode(sourceNode, targetGraph, func(key string, value any) any {
		return value
	})
}

func CloneEdge(fromInTarget, toInTarget *lpg.Node, sourceEdge *lpg.Edge, targetGraph *lpg.Graph) *lpg.Edge {
	return lpg.CloneEdge(fromInTarget, toInTarget, sourceEdge, targetGraph, func(key string, value any) any {
		return value
	})
}

func FindNodeByID(g *lpg.Graph, ID string) []*lpg.Node {
	ret := make([]*lpg.Node, 0)
	lpg.ForEachNode(g, func(node *lpg.Node) bool {
		if GetNodeID(node) == ID {
			ret = append(ret, node)
		}
		return true
	})
	return ret
}

// FindChildInstanceOf returns the childnodes of the parent that are
// instance of the given attribute id
func FindChildInstanceOf(parent *lpg.Node, childAttrID string) []*lpg.Node {
	ret := make([]*lpg.Node, 0)
	for edges := parent.GetEdges(lpg.OutgoingEdge); edges.Next(); {
		edge := edges.Edge()
		child := edge.GetTo()
		if !child.GetLabels().Has(DocumentNodeTerm.Name) {
			continue
		}
		if childAttrID == SchemaNodeIDTerm.PropertyValue(child) {
			ret = append(ret, child)
		}
	}
	return ret
}

// FindFirstChildIf returns the first child node for which the predicate returns true
func FindFirstChildIf(parent *lpg.Node, pred func(*lpg.Node) bool) *lpg.Node {
	for edges := parent.GetEdges(lpg.OutgoingEdge); edges.Next(); {
		edge := edges.Edge()
		child := edge.GetTo()
		if pred(child) {
			return child
		}
	}
	return nil
}

// GetEntityRootNode returns the entity root node containing this node
func GetEntityRootNode(aNode *lpg.Node) *lpg.Node {
	trc := aNode
	for {
		if IsNodeEntityRoot(trc) {
			return trc
		}

		nNodes := 0
		for edges := trc.GetEdges(lpg.IncomingEdge); edges.Next(); {
			edge := edges.Edge()
			nextNode := edge.GetFrom()
			if nextNode == edge.GetTo() {
				continue
			}
			if !nextNode.GetLabels().Has(DocumentNodeTerm.Name) {
				continue
			}
			nNodes++
			if nNodes > 1 {
				// Cannot find root
				return nil
			}
			trc = nextNode
		}
		if nNodes == 0 {
			return nil
		}
	}
}

// WalkNodesInEntity walks through all the nodes without crossing
// entity boundaries. It calls the function f for each node. The
// entity root containing the given node is also traversed.
func WalkNodesInEntity(aNode *lpg.Node, f func(*lpg.Node) bool) bool {
	root := GetEntityRootNode(aNode)
	if root == nil {
		return true
	}
	return IterateDescendants(root, func(node *lpg.Node) bool {
		return f(node)
	}, FollowEdgesInEntity, false)
}
