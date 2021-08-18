package gl

import (
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// NodeValue is zero or mode nodes on the stack
type NodeValue struct {
	BasicValue
	Nodes map[ls.Node]struct{}
}

func NewNodeValue(nodes ...ls.Node) NodeValue {
	ret := NodeValue{Nodes: make(map[ls.Node]struct{})}
	for _, n := range nodes {
		ret.Nodes[n] = struct{}{}
	}
	return ret
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
	"length": func(node NodeValue) (Value, error) {
		return ValueOf(len(node.Nodes)), nil
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
	"firstReachable": func(node NodeValue) (Value, error) {
		// firstReachable(nodeClosure)
		// firstReachable(nodeClosure, edgeClosure)
		return FunctionValue{MinArgs: 1, MaxArgs: 2, Name: "firstReachable", Closure: func(ctx *Context, args []Value) (Value, error) {
			nodeClosure, ok := args[0].(Closure)
			if !ok {
				return nil, ErrNotAClosure
			}
			var edgeClosure Closure
			if len(args) == 2 {
				edgeClosure, ok = args[1].(Closure)
				if !ok {
					return nil, ErrNotAClosure
				}
			}
			var closureError error
			var found ls.Node
			for nd := range node.Nodes {
				ls.FirstReachable(nd, func(node ls.Node, _ []ls.Node) bool {
					b, err := AsBool(nodeClosure.Evaluate(ValueOf(node), ctx))
					if err != nil {
						closureError = err
						return true
					}
					if b {
						found = node
						return true
					}
					return false
				},
					func(edge ls.Edge, _ []ls.Node) bool {
						if edgeClosure.F == nil {
							return true
						}
						b, err := AsBool(edgeClosure.Evaluate(ValueOf(edge), ctx))
						if err != nil {
							closureError = err
							return false
						}
						if b {
							return false
						}
						return true
					})
				if closureError != nil {
					return nil, closureError
				}
				if found != nil {
					return NewNodeValue(found), nil
				}
			}
			return NewNodeValue(), nil
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
	ret := NewNodeValue()
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
	ret := NewNodeValue()
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
