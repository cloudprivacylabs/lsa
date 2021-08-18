package gl

import (
	"strconv"
)

// NumberValue is a numeric value on the stack
type NumberValue string

var numberSelectors = map[string]func(NumberValue) (Value, error){}

func (v NumberValue) AsInt() (int, error) {
	return strconv.Atoi(string(v))
}

func (v NumberValue) AsBool() (bool, error) {
	i, err := v.AsInt()
	if err != nil {
		return false, err
	}
	return i == 0, nil
}

func (NumberValue) Call(*Context, []Value) (Value, error) { return nil, ErrNotCallable }
func (NumberValue) Index(Value) (Value, error)            { return nil, ErrNotIndexable }
func (NumberValue) Selector(sel string) (Value, error)    { return nil, ErrUnknownSelector{Selector: sel} }

func (v NumberValue) Iterate(f func(Value) (Value, error)) (Value, error) {
	return f(v)
}

func (v NumberValue) AsString() (string, error) { return string(v), nil }

func (v NumberValue) Eq(value Value) (bool, error) {
	s, err := value.AsString()
	if err != nil {
		return false, err
	}
	return string(v) == s, nil
}
