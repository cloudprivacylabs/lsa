package gl

import (
	"fmt"
)

// StringSliceValue is a []string on the stack
type StringSliceValue []string

var StringHas = FunctionValue{MinArgs: 1, MaxArgs: 1, Name: "string.has"}

var stringSliceSelectors = map[string]func(StringSliceValue) (Value, error){
	"length": func(v StringSliceValue) (Value, error) {
		return ValueOf(len(v)), nil
	},
	"has": func(v StringSliceValue) (Value, error) {
		ret := StringHas
		ret.Closure = func(args []Value) (Value, error) {
			str, err := args[0].AsString()
			if err != nil {
				return nil, err
			}
			for _, x := range v {
				if x == str {
					return TrueValue, nil
				}
			}
			return FalseValue, nil
		}
		return ret, nil
	},
}

func (value StringSliceValue) Selector(sel string) (Value, error) {
	selected := stringSliceSelectors[sel]
	if selected != nil {
		return selected(value)
	}
	return nil, ErrUnknownSelector{Selector: sel}
}

func (value StringSliceValue) Index(index Value) (Value, error) {
	i, err := index.AsInt()
	if err != nil {
		return nil, err
	}
	if i < 0 || i >= len(value) {
		return ValueOf(nil), nil
	}
	return ValueOf(value[i]), nil
}

func (value StringSliceValue) Iterate(f func(Value) (Value, error)) (Value, error) {
	var ret Value
	for _, k := range value {
		v, err := f(ValueOf(k))
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

func (value StringSliceValue) Add(v2 Value) (Value, error) {
	slice, ok := v2.(StringSliceValue)
	if !ok {
		return nil, ErrIncompatibleValue
	}
	ret := StringSliceValue(append(value, slice...))
	return ret, nil
}

func (value StringSliceValue) AsBool() (bool, error)     { return len(value) > 0, nil }
func (StringSliceValue) AsInt() (int, error)             { return 0, ErrNotANumber }
func (StringSliceValue) Call([]Value) (Value, error)     { return nil, ErrNotCallable }
func (value StringSliceValue) AsString() (string, error) { return fmt.Sprint(value), nil }
func (value StringSliceValue) Eq(v Value) (bool, error) {
	sl, ok := v.(StringSliceValue)
	if !ok {
		return false, ErrIncomparable
	}
	if len(sl) != len(value) {
		return false, nil
	}
	for i, x := range value {
		if sl[i] != x {
			return false, nil
		}
	}
	return true, nil
}
