package gl

import (
	"fmt"
)

// ValueArrayValue is a []Value on the stack
type ValueArrayValue []Value

var valueArraySelectors = map[string]func(ValueArrayValue) (Value, error){
	"length": func(v ValueArrayValue) (Value, error) {
		return ValueOf(len(v)), nil
	},
}

func (value ValueArrayValue) Selector(sel string) (Value, error) {
	selected := valueArraySelectors[sel]
	if selected != nil {
		return selected(value)
	}
	return nil, ErrUnknownSelector{Selector: sel}
}

func (value ValueArrayValue) Index(index Value) (Value, error) {
	i, err := index.AsInt()
	if err != nil {
		return nil, err
	}
	if i < 0 || i >= len(value) {
		return ValueOf(nil), nil
	}
	return ValueOf(value[i]), nil
}

func (value ValueArrayValue) Iterate(f func(Value) (Value, error)) (Value, error) {
	var ret Value
	for _, k := range value {
		v, err := f(k)
		if err != nil {
			return nil, err
		}
		if ret == nil {
			ret = v
		} else {
			acc, ok := ret.(Accumulator)
			if !ok {
				return nil, ErrCannotAccumulate
			}
			ret, err = acc.Add(v)
			if err != nil {
				return nil, err
			}
		}
	}
	return ret, nil
}

func (value ValueArrayValue) Add(v2 Value) (Value, error) {
	slice, ok := v2.(ValueArrayValue)
	if !ok {
		return nil, ErrIncompatibleValue
	}
	ret := ValueArrayValue(append(value, slice...))
	return ret, nil
}

func (value ValueArrayValue) AsBool() (bool, error) { return len(value) > 0, nil }
func (ValueArrayValue) AsInt() (int, error)         { return 0, ErrNotANumber }
func (ValueArrayValue) Call([]Value) (Value, error) { return nil, ErrNotCallable }
func (ValueArrayValue) Eq(Value) (bool, error)      { return false, ErrIncomparable }
func (value ValueArrayValue) AsString() (string, error) {
	s := make([]string, 0, len(value))
	for _, x := range value {
		str, err := x.AsString()
		if err != nil {
			return "", err
		}
		s = append(s, str)
	}
	return fmt.Sprint(s), nil
}
