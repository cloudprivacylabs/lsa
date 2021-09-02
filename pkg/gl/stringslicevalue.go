package gl

import (
	"fmt"
)

// StringSliceValue is a []string on the stack
type StringSliceValue []string

var stringSliceSelectors = map[string]func(StringSliceValue) (Value, error){
	// slice.length
	//
	// Returns the number of elements in the string slice
	"length": func(v StringSliceValue) (Value, error) {
		return ValueOf(len(v)), nil
	},

	// slice.has(v)
	//
	// Returns if the slice has the string representation of the value
	"has": func(v StringSliceValue) (Value, error) {
		return FunctionValue{
			MinArgs: 1,
			MaxArgs: 1,
			Name:    "has",
			Closure: func(scope *Scope, args []Value) (Value, error) {
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

	// slice.hasAny(v,...)
	//
	// Returns if the slice has any one of the string representations of the values
	"hasAny": func(v StringSliceValue) (Value, error) {
		return FunctionValue{
			MinArgs: 1,
			MaxArgs: -1,
			Name:    "hasAny",
			Closure: func(scope *Scope, args []Value) (Value, error) {
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

	// slice.hasAll(v,...)
	//
	// Returns if the slice has all the string representations of the values
	"hasAll": func(v StringSliceValue) (Value, error) {
		return FunctionValue{
			MinArgs: 1,
			MaxArgs: -1,
			Name:    "hasAll",
			Closure: func(scope *Scope, args []Value) (Value, error) {
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

func (value StringSliceValue) Add(v2 Value) (Value, error) {
	slice, ok := v2.(StringSliceValue)
	if !ok {
		return nil, ErrIncompatibleValue
	}
	ret := StringSliceValue(append(value, slice...))
	return ret, nil
}

func (value StringSliceValue) AsBool() (bool, error)         { return len(value) > 0, nil }
func (StringSliceValue) AsInt() (int, error)                 { return 0, ErrNotANumber }
func (StringSliceValue) Call(*Scope, []Value) (Value, error) { return nil, ErrNotCallable }
func (value StringSliceValue) AsString() (string, error)     { return fmt.Sprint(value), nil }
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
