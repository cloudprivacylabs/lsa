package gl

import (
	"strconv"
)

// StringValue is a string on the stack
type StringValue string

var stringSelectors = map[string]func(StringValue) (Value, error){
	"length": func(v StringValue) (Value, error) {
		return ValueOf(len(v)), nil
	},
}

func (value StringValue) Selector(sel string) (Value, error) {
	selected := stringSelectors[sel]
	if selected != nil {
		return selected(value)
	}
	return nil, ErrUnknownSelector{Selector: sel}
}

func (StringValue) Call(*Context, []Value) (Value, error) { return nil, ErrNotCallable }
func (StringValue) Index(Value) (Value, error)            { return nil, ErrNotIndexable }

func (s StringValue) Iterate(f func(Value) (Value, error)) (Value, error) {
	return f(s)
}

func (value StringValue) AsBool() (bool, error) { return len(value) > 0, nil }

func (value StringValue) AsInt() (int, error) {
	return strconv.Atoi(string(value))
}

func (value StringValue) AsString() (string, error) { return string(value), nil }

func (value StringValue) Eq(v Value) (bool, error) {
	s, err := v.AsString()
	if err != nil {
		return false, err
	}
	return s == string(value), nil
}
