package gl

import ()

// BoolValue is a boolean value on the stack
type BoolValue bool

// Boolean literals
var (
	TrueValue  = BoolValue(true)
	FalseValue = BoolValue(false)
)

// AsInt returns 0 or 1
func (v BoolValue) AsInt() (int, error) {
	if v == true {
		return 1, nil
	}
	return 0, nil
}

// Call returns ErrNotCallable
func (BoolValue) Call(*Scope, []Value) (Value, error) { return nil, ErrNotCallable }

// Index returns ErrNotIndexable
func (BoolValue) Index(Value) (Value, error) { return nil, ErrNotIndexable }

// Selector returns ErrUnknownSelector
func (BoolValue) Selector(sel string) (Value, error) { return nil, ErrUnknownSelector{Selector: sel} }

// Eq returns true of the value.AsBool has the same value as this
func (b BoolValue) Eq(v Value) (bool, error) {
	x, err := v.AsBool()
	if err != nil {
		return false, err
	}
	return bool(b) == x, nil
}

// AsBool returns the value
func (v BoolValue) AsBool() (bool, error) { return bool(v), nil }

// AsString returns "true" or "false"
func (v BoolValue) AsString() (string, error) {
	if v == true {
		return "true", nil
	}
	return "false", nil
}
