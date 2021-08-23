package gl

import (
	"fmt"
)

// StringSliceValue is a []string on the stack
type StringSliceValue []string

var stringSliceSelectors = map[string]func(StringSliceValue) (Value, error){
	"length": func(v StringSliceValue) (Value, error) {
		return ValueOf(len(v)), nil
	},
	"has": func(v StringSliceValue) (Value, error) {
		return FunctionValue{
			MinArgs: 1,
			MaxArgs: 1,
			Name:    "has",
			Closure: func(ctx *Context, args []Value) (Value, error) {
				str, err := args[0].AsString()
				if err != nil {
					return nil, err
				}
				if v.has(str) {
					return TrueValue, nil
				}
				return FalseValue, nil
			},
		}, nil
	},
	"hasAny": func(v StringSliceValue) (Value, error) {
		return FunctionValue{
			MinArgs: 1,
			MaxArgs: -1,
			Name:    "hasAny",
			Closure: func(ctx *Context, args []Value) (Value, error) {
				str, err := argsToStrings(args)
				if err != nil {
					return nil, err
				}
				if v.hasAny(str...) {
					return TrueValue, nil
				}
				return FalseValue, nil
			},
		}, nil
	},
	"hasAll": func(v StringSliceValue) (Value, error) {
		return FunctionValue{
			MinArgs: 1,
			MaxArgs: -1,
			Name:    "hasAll",
			Closure: func(ctx *Context, args []Value) (Value, error) {
				str, err := argsToStrings(args)
				if err != nil {
					return nil, err
				}
				if v.hasAll(str...) {
					return TrueValue, nil
				}
				return FalseValue, nil
			},
		}, nil
	},
}

func argsToStrings(args []Value) (StringSliceValue, error) {
	ret := make([]string, 0, len(args))
	for _, arg := range args {
		switch t := arg.(type) {
		case StringValue:
			ret = append(ret, string(t))
		case StringSliceValue:
			ret = append(ret, []string(t)...)
		default:
			return nil, ErrNotAString
		}
	}
	return ret, nil
}

func (value StringSliceValue) has(s string) bool {
	for _, x := range value {
		if x == s {
			return true
		}
	}
	return false
}

func (value StringSliceValue) hasAny(strs ...string) bool {
	for _, s := range strs {
		for _, x := range value {
			if x == s {
				return true
			}
		}
	}
	return false
}

func (value StringSliceValue) hasAll(strs ...string) bool {
	for _, s := range strs {
		found := false
		for _, x := range value {
			if x == s {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
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

func (value StringSliceValue) AsBool() (bool, error)           { return len(value) > 0, nil }
func (StringSliceValue) AsInt() (int, error)                   { return 0, ErrNotANumber }
func (StringSliceValue) Call(*Context, []Value) (Value, error) { return nil, ErrNotCallable }
func (value StringSliceValue) AsString() (string, error)       { return fmt.Sprint(value), nil }
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
