package gl

import ()

// BoolValue is a boolean value on the stack
type BoolValue bool

var TrueValue = BoolValue(true)
var FalseValue = BoolValue(false)

var boolSelectors = map[string]func(BoolValue) (Value, error){}

func (v BoolValue) AsInt() (int, error) {
	if v == true {
		return 1, nil
	}
	return 0, nil
}

func (BoolValue) Call(*Context, []Value) (Value, error) { return nil, ErrNotCallable }
func (BoolValue) Index(Value) (Value, error)            { return nil, ErrNotIndexable }
func (BoolValue) Selector(sel string) (Value, error)    { return nil, ErrUnknownSelector{Selector: sel} }
func (b BoolValue) Eq(v Value) (bool, error) {
	x, err := v.AsBool()
	if err != nil {
		return false, err
	}
	return bool(b) == x, nil
}

func (b BoolValue) Iterate(f func(Value) (Value, error)) (Value, error) {
	return f(b)
}

func (v BoolValue) AsBool() (bool, error) { return bool(v), nil }
func (v BoolValue) AsString() (string, error) {
	if v == true {
		return "true", nil
	}
	return "false", nil
}
