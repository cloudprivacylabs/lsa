package gl

import (
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

// EdgeValue is zero or more edges
type EdgeValue struct {
	basicValue
	Edges map[graph.Edge]struct{}
}

// oneEdge is a convenience function that returns the edge if there is
// one, and that returns an error otherwise
func (e EdgeValue) oneEdge() (graph.Edge, error) {
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
	// edge.label
	//
	// The label of the edge, a string value
	"label": func(edge EdgeValue) (Value, error) {
		e, err := edge.oneEdge()
		if err != nil {
			return nil, err
		}
		return ValueOf(e.GetLabel()), nil
	},
	// edge.from
	//
	// The source node of the edge, a node value
	"from": func(edge EdgeValue) (Value, error) {
		e, err := edge.oneEdge()
		if err != nil {
			return nil, err
		}
		return ValueOf(e.GetFrom()), nil
	},
	// edge.to
	//
	// The target node of the edge, a node value
	"to": func(edge EdgeValue) (Value, error) {
		e, err := edge.oneEdge()
		if err != nil {
			return nil, err
		}
		return ValueOf(e.GetTo()), nil
	},
}

// Selector selects one of the selectors of the edge
func (e EdgeValue) Selector(sel string) (Value, error) {
	selected := edgeSelectors[sel]
	if selected != nil {
		return selected(e)
	}
	return e.basicValue.Selector(sel)
}

// Add returns the set union of to edge sets
func (e EdgeValue) Add(v2 Value) (Value, error) {
	edges, ok := v2.(EdgeValue)
	if !ok {
		return nil, ErrIncompatibleValue
	}
	ret := EdgeValue{Edges: map[graph.Edge]struct{}{}}
	for k := range e.Edges {
		ret.Edges[k] = struct{}{}
	}
	for k := range edges.Edges {
		ret.Edges[k] = struct{}{}
	}
	return ret, nil
}

// AsBool returns true if edge value is nonempty
func (e EdgeValue) AsBool() (bool, error) { return len(e.Edges) > 0, nil }

// AsString returns error
func (e EdgeValue) AsString() (string, error) { return "", ErrNotAString }

// Eq compares two edge sets
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
