package gl

import (
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// EdgeValue is zero or more edge on the stack
type EdgeValue struct {
	BasicValue
	Edges map[ls.Edge]struct{}
}

func (e EdgeValue) oneEdge() (ls.Edge, error) {
	switch len(e.Edges) {
	case 0:
		return nil, ErrNoEdgesInResult
	case 1:
		for k := range e.Edges {
			return k, nil
		}
	}
	return nil, ErrMultipleEdgesInResult
}

var edgeSelectors = map[string]func(EdgeValue) (Value, error){
	"label": func(edge EdgeValue) (Value, error) {
		e, err := edge.oneEdge()
		if err != nil {
			return nil, err
		}
		return ValueOf(e.GetLabel()), nil
	},
	"from": func(edge EdgeValue) (Value, error) {
		e, err := edge.oneEdge()
		if err != nil {
			return nil, err
		}
		return ValueOf(e.GetFrom().(ls.Node)), nil
	},
	"to": func(edge EdgeValue) (Value, error) {
		e, err := edge.oneEdge()
		if err != nil {
			return nil, err
		}
		return ValueOf(e.GetTo().(ls.Node)), nil
	},
	"properties": func(edge EdgeValue) (Value, error) {
		e, err := edge.oneEdge()
		if err != nil {
			return nil, err
		}
		return ValueOf(e.GetProperties()), nil
	},
}

func (e EdgeValue) Selector(sel string) (Value, error) {
	selected := edgeSelectors[sel]
	if selected != nil {
		return selected(e)
	}
	return e.BasicValue.Selector(sel)
}

func (e EdgeValue) Iterate(f func(Value) (Value, error)) (Value, error) {
	var ret Value
	for edge := range e.Edges {
		v, err := f(ValueOf(edge))
		if err != nil {
			return nil, err
		}
		if ret == nil {
			ret = v
		} else {
			accumulator, ok := ret.(Accumulator)
			if !ok {
				return nil, ErrCannotAccumulate
			}
			ret, err = accumulator.Add(v)
			if err != nil {
				return nil, err
			}
		}
	}
	return ret, nil
}

func (e EdgeValue) Add(v2 Value) (Value, error) {
	edges, ok := v2.(EdgeValue)
	if !ok {
		return nil, ErrIncompatibleValue
	}
	ret := EdgeValue{Edges: map[ls.Edge]struct{}{}}
	for k := range e.Edges {
		ret.Edges[k] = struct{}{}
	}
	for k := range edges.Edges {
		ret.Edges[k] = struct{}{}
	}
	return ret, nil
}

func (e EdgeValue) AsBool() (bool, error) { return len(e.Edges) > 0, nil }

func (e EdgeValue) AsString() (string, error) { return "", ErrNotAString }

func (e EdgeValue) Eq(val Value) (bool, error) {
	ev, ok := val.(EdgeValue)
	if !ok {
		return false, ErrIncomparable
	}
	if len(ev.Edges) != len(e.Edges) {
		return false, nil
	}
	for k := range ev.Edges {
		if _, ok := e.Edges[k]; !ok {
			return false, nil
		}
	}
	return true, nil
}
