package gl

import (
	"strconv"
)

// NumberValue is a numeric value on the stack
type NumberValue string

var numberSelectors = map[string]func(NumberValue) (Value, error){}

// AsInt returns the inv value, or error if value cannot be converted to int
func (v NumberValue) AsInt() (int, error) {
	return strconv.Atoi(string(v))
}

// AsBool returns true if value is nonzero
func (v NumberValue) AsBool() (bool, error) {
	i, err := v.AsInt()
	if err != nil {
		return false, err
	}
	return i == 0, nil
}

// Call returns ErrNotCallable
func (NumberValue) Call(*Scope, []Value) (Value, error) { return nil, ErrNotCallable }

// Index returns ErrNotIndexable
func (NumberValue) Index(Value) (Value, error) { return nil, ErrNotIndexable }

// Selector returns ErrUnknownSelector
func (NumberValue) Selector(sel string) (Value, error) { return nil, ErrUnknownSelector{Selector: sel} }

// AsString returns value as string
func (v NumberValue) AsString() (string, error) { return string(v), nil }

// Eq compares the string values of the numbers
func (v NumberValue) Eq(value Value) (bool, error) {
	s, err := value.AsString()
	if err != nil {
		return false, err
	}
	return string(v) == s, nil
}
