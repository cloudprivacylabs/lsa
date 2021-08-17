package gl

import (
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// NodeValue is zero or mode nodes on the stack
type NodeValue struct {
	BasicValue
	Nodes map[ls.Node]struct{}
}

func (n NodeValue) oneNode() (ls.Node, error) {
	switch len(n.Nodes) {
	case 0:
		return nil, ErrNoNodesInResult
	case 1:
		for k := range n.Nodes {
			return k, nil
		}
	}
	return nil, ErrMultipleNodesInResult
}

var nodeSelectors = map[string]func(NodeValue) (Value, error){
	"id": func(node NodeValue) (Value, error) {
		n, err := node.oneNode()
		if err != nil {
			return nil, err
		}
		return ValueOf(n.GetID()), nil
	},
	"type": func(node NodeValue) (Value, error) {
		n, err := node.oneNode()
		if err != nil {
			return nil, err
		}
		return ValueOf(n.GetTypes()), nil
	},
	"value": func(node NodeValue) (Value, error) {
		n, err := node.oneNode()
		if err != nil {
			return nil, err
		}
		return ValueOf(n.GetValue()), nil
	},
	"properties": func(node NodeValue) (Value, error) {
		n, err := node.oneNode()
		if err != nil {
			return nil, err
		}
		return ValueOf(n.GetProperties()), nil
	},
	"descendants": func(node NodeValue) (Value, error) {
		return FunctionValue{MinArgs: 0, MaxArgs: -1, Name: "descendants", Closure: func(args []Value) (Value, error) {
			result := make(map[ls.Node]struct{})
			labels := make([]string, 0, len(args))
			for _, arg := range args {
				v, err := arg.AsString()
				if err != nil {
					return nil, err
				}
				labels = append(labels, v)
			}
			for nd := range node.Nodes {
				nodes := nd.GetDescendants(labels...)
				for k := range nodes {
					result[k] = struct{}{}
				}
			}
			return NodeValue{Nodes: result}, nil
		}}, nil
	},
}

func (v NodeValue) Selector(sel string) (Value, error) {
	selected := nodeSelectors[sel]
	if selected != nil {
		return selected(v)
	}
	return v.BasicValue.Selector(sel)
}

func (v NodeValue) Iterate(f func(Value) (Value, error)) (Value, error) {
	ret := NodeValue{Nodes: make(map[ls.Node]struct{})}
	for node := range v.Nodes {
		n, err := f(ValueOf(node))
		if err != nil {
			return nil, err
		}
		x, err := ret.Add(n)
		if err != nil {
			return nil, err
		}
		ret = x.(NodeValue)
	}
	return ret, nil
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
	ret := NodeValue{Nodes: map[ls.Node]struct{}{}}
	for k := range v.Nodes {
		ret.Nodes[k] = struct{}{}
	}
	for k := range nodes.Nodes {
		ret.Nodes[k] = struct{}{}
	}
	return ret, nil
}

func (v NodeValue) AsBool() (bool, error) { return len(v.Nodes) > 0, nil }

func (v NodeValue) AsString() (string, error) { return "", ErrNotAString }

func (v NodeValue) Eq(val Value) (bool, error) {
	nv, ok := val.(NodeValue)
	if !ok {
		return false, ErrIncomparable
	}
	if len(nv.Nodes) != len(v.Nodes) {
		return false, nil
	}
	for k := range nv.Nodes {
		if _, ok := v.Nodes[k]; !ok {
			return false, nil
		}
	}
	return true, nil
}
