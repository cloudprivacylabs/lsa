package gl

import (
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// NodeValue is zero or mode nodes on the stack
type NodeValue struct {
	basicValue
	Nodes ls.NodeSet
}

func NewNodeValue(nodes ...ls.Node) NodeValue {
	return NodeValue{Nodes: ls.NewNodeSet(nodes...)}
}

func (n NodeValue) oneNode() (ls.Node, error) {
	switch n.Nodes.Len() {
	case 0:
		return nil, ErrNoNodesInResult
	case 1:
		return n.Nodes.Slice()[0], nil
	}
	return nil, ErrMultipleNodesInResult
}

var nodeSelectors = map[string]func(NodeValue) (Value, error){
	"id":             oneNode(func(node ls.Node) (Value, error) { return StringValue(node.GetID()), nil }),
	"length":         func(node NodeValue) (Value, error) { return ValueOf(node.Nodes.Len()), nil },
	"type":           oneNode(func(node ls.Node) (Value, error) { return StringSliceValue(node.GetTypes().Slice()), nil }),
	"value":          oneNode(func(node ls.Node) (Value, error) { return ValueOf(node.GetValue()), nil }),
	"properties":     oneNode(func(node ls.Node) (Value, error) { return PropertiesValue{Properties: node.GetProperties()}, nil }),
	"firstReachable": nodeFirstReachableFunc,
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
