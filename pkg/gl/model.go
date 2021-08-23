package gl

import (
	"fmt"

	"github.com/bserdar/digraph"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// Value represents a value on the evaluation stack
type Value interface {
	// Selector selects a field/method from the underlying value.
	Selector(string) (Value, error)
	// Index selects an value from an indexable value
	Index(Value) (Value, error)
	// Call a method/function with arguments
	Call(*Context, []Value) (Value, error)

	// Eq checks equivalence
	Eq(Value) (bool, error)

	Iterate(func(Value) (Value, error)) (Value, error)

	// Return value as an integer, or return error if it cannot be represented as an integer
	AsInt() (int, error)
	// Return value as a boolean. Returns false for empty string, empty collection, etc.
	AsBool() (bool, error)
	AsString() (string, error)
}

type Accumulator interface {
	Add(Value) (Value, error)
}

type BasicValue struct{}

func (BasicValue) Selector(sel string) (Value, error) {
	return nil, ErrUnknownSelector{Selector: sel}
}

func (BasicValue) Call(*Context, []Value) (Value, error) {
	return nil, ErrNotCallable
}

func (BasicValue) Index(i Value) (Value, error) {
	return nil, ErrNotIndexable
}

func (BasicValue) Iterate(func(Value) (Value, error)) (Value, error) { return nil, ErrCannotIterate }

func (BasicValue) AsInt() (int, error) {
	return 0, ErrNotANumber
}

func (BasicValue) AsBool() (bool, error) {
	return false, nil
}

func (BasicValue) AsString() (string, error) {
	return "", ErrNotAString
}

func (BasicValue) Eq(Value) (bool, error) { return false, ErrIncomparable }

type NullValue struct {
	BasicValue
}

func (NullValue) AsString() (string, error) { return "null", nil }

func (NullValue) Eq(v Value) (bool, error) {
	if v == nil {
		return true, nil
	}
	if _, ok := v.(NullValue); ok {
		return true, nil
	}
	return false, nil
}

type Closure struct {
	BasicValue
	Symbol string
	F      Expression
}

func (c Closure) Evaluate(arg Value, ctx *Context) (Value, error) {
	newContext := ctx.NewNestedContext()
	if len(c.Symbol) > 0 {
		newContext.Set(c.Symbol, arg)
	}
	return c.F.Evaluate(newContext)
}

type LValue struct {
	BasicValue
	Name string
}

type FunctionValue struct {
	BasicValue
	MinArgs int
	MaxArgs int
	Name    string
	Closure func(*Context, []Value) (Value, error)
}

func (FunctionValue) Eq(Value) (bool, error) { return false, nil }

func (f FunctionValue) Call(ctx *Context, args []Value) (Value, error) {
	if len(args) < f.MinArgs {
		return nil, ErrInvalidFunctionCall(fmt.Sprintf("'%s' needs at least %d args but got %d", f.Name, f.MinArgs, len(args)))
	}
	if f.MaxArgs >= 0 && len(args) > f.MaxArgs {
		return nil, ErrInvalidFunctionCall(fmt.Sprintf("'%s' needs at most %d args but got %d", f.Name, f.MaxArgs, len(args)))
	}
	return f.Closure(ctx, args)
}

func ValueOf(value interface{}) Value {
	if value == nil {
		return NullValue{}
	}
	if v, ok := value.(Value); ok {
		return v
	}
	switch t := value.(type) {
	case bool:
		return BoolValue(t)
	case string:
		return StringValue(t)
	case []string:
		return StringSliceValue(t)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return NumberValue(fmt.Sprint(t))
	case ls.Node:
		return NodeValue{Nodes: ls.NewNodeSet(t)}
	case []ls.Node:
		return NodeValue{Nodes: ls.NewNodeSet(t...)}
	case ls.Edge:
		return EdgeValue{Edges: map[ls.Edge]struct{}{t: struct{}{}}}
	case []ls.Edge:
		ret := EdgeValue{Edges: make(map[ls.Edge]struct{}, len(t))}
		for _, x := range t {
			ret.Edges[x] = struct{}{}
		}
		return ret
	case digraph.NodeIterator:
		ret := NodeValue{Nodes: ls.NewNodeSet()}
		for t.HasNext() {
			ret.Nodes.Add(t.Next().(ls.Node))
		}
		return ret
	case digraph.EdgeIterator:
		ret := EdgeValue{Edges: make(map[ls.Edge]struct{})}
		for t.HasNext() {
			ret.Edges[t.Next().(ls.Edge)] = struct{}{}
		}
		return ret
	case *digraph.Graph:
		return ValueOf(t.GetAllNodes())
	case []Value:
		return ValueArrayValue(t)
	case map[string]*ls.PropertyValue:
		return PropertiesValue{Properties: t}
	}
	panic("Unrepresentable value")
}

// AsBool is a convenience function that will return an error if the
// input contains an error, or if the value cannot be converted to a
// bool value. Use it as:
//
//   b, err:=AsBool(expression.Evaluate(ctx))
//
func AsBool(v Value, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	b, err := v.AsBool()
	if err != nil {
		return false, err
	}
	return b, nil
}
