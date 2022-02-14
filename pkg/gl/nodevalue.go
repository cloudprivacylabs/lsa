package gl

import (
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

// NodeValue is zero or mode nodes on the stack
type NodeValue struct {
	basicValue
	Nodes NodeSet
}

func NewNodeValue(nodes ...graph.Node) NodeValue {
	return NodeValue{Nodes: NewNodeSet(nodes...)}
}

func (n NodeValue) oneNode() (graph.Node, error) {
	switch n.Nodes.Len() {
	case 0:
		return nil, ErrNoNodesInResult
	case 1:
		return n.Nodes.Slice()[0], nil
	}
	return nil, ErrMultipleNodesInResult
}

var nodeSelectors = map[string]func(NodeValue) (Value, error){
	"id":             oneNode(func(node graph.Node) (Value, error) { return StringValue(ls.GetNodeID(node)), nil }),
	"length":         func(node NodeValue) (Value, error) { return ValueOf(node.Nodes.Len()), nil },
	"type":           oneNode(func(node graph.Node) (Value, error) { return StringSliceValue(node.GetLabels().Slice()), nil }),
	"value":          oneNode(func(node graph.Node) (Value, error) { v, _ := ls.GetNodeValue(node); return ValueOf(v), nil }),
	"firstReachable": nodeFirstReachableFunc,
	"first":          nodeFirstReachableFunc,
	"firstDoc":       nodeFirstReachableDocNodeFunc,
	"instanceOf":     nodeInstanceOfFunc,
	"walk":           nodeWalk,
}

func (v NodeValue) Selector(sel string) (Value, error) {
	selected := nodeSelectors[sel]
	if selected != nil {
		return selected(v)
	}
	return v.basicValue.Selector(sel)
}

func (v NodeValue) Add(v2 Value) (Value, error) {
	if v2 == nil {
		return v, nil
	}
	if _, ok := v2.(NullValue); ok {
		return v, nil
	}
	nodes, ok := v2.(NodeValue)
	if !ok {
		return nil, ErrIncompatibleValue
	}
	ret := NewNodeValue()
	ret.Nodes.Add(v.Nodes.Slice()...)
	ret.Nodes.Add(nodes.Nodes.Slice()...)
	return ret, nil
}

func (v NodeValue) AsBool() (bool, error) { return v.Nodes.Len() > 0, nil }

func (v NodeValue) AsString() (string, error) { return "", ErrNotAString }

func (v NodeValue) Eq(val Value) (bool, error) {
	nv, ok := val.(NodeValue)
	if !ok {
		return false, ErrIncomparable
	}
	return v.Nodes.EqualSet(nv.Nodes), nil
}
