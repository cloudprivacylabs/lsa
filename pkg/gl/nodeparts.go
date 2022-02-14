package gl

import (
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

func oneNode(f func(node graph.Node) (Value, error)) func(NodeValue) (Value, error) {
	return func(n NodeValue) (Value, error) {
		on, err := n.oneNode()
		if err != nil {
			return nil, err
		}
		return f(on)
	}
}

var trueClosure = Closure{F: boolLiteral(true)}

// closureOrBool will return a predicate closure, or if the value is a bool, a closure that returns that bool
func closureOrBool(v Value) (Closure, error) {
	cl, ok := v.(Closure)
	if !ok {
		b, err := v.AsBool()
		if err != nil {
			return cl, ErrClosureOrBooleanExpected
		}
		return Closure{F: boolLiteral(b)}, nil
	}
	return cl, nil
}

// firstReachable(nodeClosure|predicate)
// firstReachable(nodeClosure|predicate,edgeClosure|predicate)
func nodeFirstReachableFunc(node NodeValue) (Value, error) {
	return FunctionValue{
		MinArgs: 1,
		MaxArgs: 2,
		Name:    "firstReachable",
		Closure: func(scope *Scope, args []Value) (Value, error) {
			return firstReachableNodeImpl(scope, node, args)
		},
	}, nil
}

// firstDoc(nodeClosure|predicate)
func nodeFirstReachableDocNodeFunc(node NodeValue) (Value, error) {
	return FunctionValue{
		MinArgs: 1,
		MaxArgs: 1,
		Name:    "firstDoc",
		Closure: func(scope *Scope, args []Value) (Value, error) {
			nodeClosure, err := closureOrBool(args[0])
			if err != nil {
				return nil, err
			}
			var closureError error
			var found graph.Node
			for _, nd := range node.Nodes.Slice() {
				ls.FirstReachable(nd, func(node graph.Node, _ []graph.Node) bool {
					if !node.GetLabels().Has(ls.DocumentNodeTerm) {
						return true
					}
					b, err := AsBool(nodeClosure.Evaluate(ValueOf(node), scope))
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
					func(edge graph.Edge, _ []graph.Node) bool {
						if edge.GetLabel() == ls.InstanceOfTerm {
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
		},
	}, nil
}

func firstReachableNodeImpl(scope *Scope, root NodeValue, args []Value) (Value, error) {
	nodeClosure, err := closureOrBool(args[0])
	if err != nil {
		return nil, err
	}
	edgeClosure := trueClosure
	if len(args) == 2 {
		edgeClosure, err = closureOrBool(args[1])
		if err != nil {
			return nil, err
		}
	}
	var closureError error
	var found graph.Node
	for _, nd := range root.Nodes.Slice() {
		ls.FirstReachable(nd, func(node graph.Node, _ []graph.Node) bool {
			b, err := AsBool(nodeClosure.Evaluate(ValueOf(node), scope))
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
			func(edge graph.Edge, _ []graph.Node) bool {
				b, err := AsBool(edgeClosure.Evaluate(ValueOf(edge), scope))
				if err != nil {
					closureError = err
					return false
				}
				if b {
					return true
				}
				return false
			})
		if closureError != nil {
			return nil, closureError
		}
		if found != nil {
			return NewNodeValue(found), nil
		}
	}
	return NewNodeValue(), nil
}

func nodeInstanceOfFunc(node NodeValue) (Value, error) {
	return FunctionValue{
		MinArgs: 1,
		MaxArgs: 1,
		Name:    "instanceOf",
		Closure: func(scope *Scope, args []Value) (Value, error) {
			result := NewNodeValue()
			id, err := args[0].AsString()
			if err != nil {
				return nil, err
			}
			for n := range node.Nodes.Map() {
				for _, instanceOfNode := range ls.InstanceOf(n) {
					if ls.GetNodeID(instanceOfNode) == id {
						result.Nodes.Add(n)
						break
					}
				}
			}
			return result, nil
		},
	}, nil
}

func nodeWalk(node NodeValue) (Value, error) {
	return FunctionValue{
		MinArgs: 0,
		MaxArgs: -1,
		Name:    "walk",
		Closure: func(scope *Scope, args []Value) (Value, error) {
			result := NewNodeValue()
			walk := NewWalk()
			var edgePredicate func(graph.Edge) bool
			var nodePredicate func(graph.Node) bool
			var err error
			for i, arg := range args {
				arg := arg
				if err != nil {
					return nil, err
				}
				if i%2 == 0 {
					// Edge predicate
					var edgeClosure Closure
					cl, ok := arg.(Closure)
					if ok {
						edgeClosure = cl
					} else {
						s, ok := arg.(StringValue)
						if ok {
							edgeClosure = Closure{
								Symbol: "e",
								F: equalityExpression{
									right: selectExpression{
										base:     identifierValue("e"),
										selector: "label",
									},
									left: stringLiteral(s),
								},
							}
						} else {
							b, err := arg.AsBool()
							if err != nil {
								return cl, ErrInvalidArgumentType
							}
							edgeClosure = Closure{F: boolLiteral(b)}
						}
					}
					edgePredicate = func(e graph.Edge) bool {
						var b bool
						b, err = AsBool(edgeClosure.Evaluate(ValueOf(e), scope))
						return b
					}
				} else {
					// Node predicate
					var nodeClosure Closure
					cl, ok := arg.(Closure)
					if ok {
						nodeClosure = cl
					} else {
						s, ok := arg.(StringValue)
						if ok {
							nodeClosure = Closure{
								Symbol: "n",
								F: equalityExpression{
									right: selectExpression{
										base:     identifierValue("n"),
										selector: "id",
									},
									left: stringLiteral(s),
								},
							}
						} else {
							b, err := arg.AsBool()
							if err != nil {
								return cl, ErrInvalidArgumentType
							}
							nodeClosure = Closure{F: boolLiteral(b)}
						}
					}
					nodePredicate = func(n graph.Node) bool {
						var b bool
						b, err = AsBool(nodeClosure.Evaluate(ValueOf(n), scope))
						return b
					}
					walk.Step(edgePredicate, nodePredicate)
					edgePredicate = nil
					nodePredicate = nil
				}
			}
			if edgePredicate != nil && nodePredicate == nil {
				walk.Step(edgePredicate, AnyNodePredicate)
			}

			result.Nodes.Add(walk.Walk(node.Nodes.Slice())...)
			return result, nil
		},
	}, nil
}
