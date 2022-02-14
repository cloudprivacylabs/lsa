package opencypher

import (
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

type ErrInvalidValueReferenceInPattern struct {
	Symbol string
}

func (e ErrInvalidValueReferenceInPattern) Error() string {
	return "Invalid value reference in pattern: " + e.Symbol
}

func (properties Properties) AsLiteral(ctx *EvalContext) ([]MapKeyValue, error) {
	if properties.Param != nil {
		param, err := ctx.GetParameter(string(*properties.Param))
		if err != nil {
			return nil, err
		}
		lit, ok := param.Value.(map[string]Value)
		if !ok {
			return nil, ErrPropertiesParameterExpected
		}
		kv := make([]MapKeyValue, 0, len(lit))
		for k, v := range lit {
			kv = append(kv, MapKeyValue{Key: k, Value: v})
		}
		return kv, nil
	}
	if properties.Map != nil {
		return properties.Map.KeyValues, nil
	}
	return nil, nil
}

type matchResultAccumulator struct {
	where   Expression
	evalCtx *EvalContext
	result  ResultSet
}

func (acc *matchResultAccumulator) StoreResult(path interface{}, symbols map[string]interface{}) {
	if acc.where != nil {
		// Evaluate the where clause
		// for k, v := range symbols {

		// }
	}
	// Record results in the context
	if node, ok := path.(graph.Node); ok {
		acc.result.AddPath(node, nil)
	}
	if edges, ok := path.([]graph.Edge); ok {
		acc.result.AddPath(nil, edges)
	}
	result := make(map[string]Value)
	for k, v := range symbols {
		result[k] = Value{Value: v}
	}
	acc.result.Append(result)
}

func (match Match) GetResults(ctx *EvalContext) (ResultSet, error) {
	if len(match.Pattern.Parts) > 1 {
		panic("Multiple match patterns are not implemented")
	}
	pattern, err := match.Pattern.Parts[0].getPattern(ctx)
	if err != nil {
		return ResultSet{}, err
	}

	newContext := ctx.SubContext()

	symbols := make(map[string]*graph.PatternSymbol)

	for symbol := range pattern.GetSymbolNames() {
		// If a symbol is in the context, then get its value. Otherwise, it is a local symbol. Add to context
		value, err := ctx.GetVar(symbol)
		if err != nil {
			continue
		}
		ps := &graph.PatternSymbol{}
		// A variable with the same name exists
		// Must be a Node, or []Edge
		switch val := value.Value.(type) {
		case graph.Node:
			ps.AddNode(val)
		case []graph.Edge:
			ps.AddPath(val)
		default:
			return ResultSet{}, ErrInvalidValueReferenceInPattern{Symbol: symbol}
		}
		symbols[symbol] = ps
	}

	resultAccumulator := matchResultAccumulator{
		where:   match.Where,
		evalCtx: newContext,
	}

	err = pattern.Run(ctx.graph, symbols, &resultAccumulator)
	if err != nil {
		return ResultSet{}, err
	}

	return resultAccumulator.result, nil
}

func (part PatternPart) getPattern(ctx *EvalContext) (graph.Pattern, error) {
	pattern := make([]graph.PatternItem, 0, len(part.Path)*2+1)
	np, err := part.Start.getPattern(ctx)
	if err != nil {
		return nil, err
	}
	pattern = append(pattern, np)
	for _, pathItem := range part.Path {
		pi, err := pathItem.Rel.getPattern(ctx)
		if err != nil {
			return nil, err
		}
		pattern = append(pattern, pi)

		pi, err = pathItem.Node.getPattern(ctx)
		if err != nil {
			return nil, err
		}
		pattern = append(pattern, pi)
	}
	return pattern, nil
}

func (np NodePattern) getPattern(ctx *EvalContext) (graph.PatternItem, error) {
	ret := graph.PatternItem{}
	if np.Var != nil {
		ret.Name = string(*np.Var)
	}
	ret.Labels = np.Labels.getPattern()
	var err error
	ret.Properties, err = np.Properties.getPattern(ctx)
	if err != nil {
		return graph.PatternItem{}, err
	}
	return ret, nil
}

func (rp RelationshipPattern) getPattern(ctx *EvalContext) (graph.PatternItem, error) {
	ret := graph.PatternItem{}
	if rp.Var != nil {
		ret.Name = string(*rp.Var)
	}
	ret.Labels = rp.RelTypes.getPattern()
	if rp.Range != nil {
		from, to, err := rp.Range.Evaluate(ctx)
		if err != nil {
			return graph.PatternItem{}, err
		}
		if from != nil {
			ret.Min = *from
		} else {
			ret.Min = 1
		}
		if to != nil {
			ret.Max = *to
		} else {
			ret.Max = 1
		}
	} else {
		ret.Min, ret.Max = 1, 1
	}
	if rp.Backwards {
		ret.Backwards = true
	}
	var err error
	ret.Properties, err = rp.Properties.getPattern(ctx)
	if err != nil {
		return graph.PatternItem{}, err
	}
	return ret, nil
}

func (rt *RelationshipTypes) getPattern() graph.StringSet {
	if rt == nil {
		return nil
	}
	if len(rt.Rel) == 0 {
		return nil
	}
	ret := graph.NewStringSet()
	for _, r := range rt.Rel {
		s := r.String()
		if len(s) > 0 {
			ret.Add(s)
		}
	}
	if len(ret) == 0 {
		return nil
	}
	return ret
}

func (nl *NodeLabels) getPattern() graph.StringSet {
	if nl == nil {
		return nil
	}
	ret := graph.NewStringSet()
	for _, l := range *nl {
		s := l.String()
		if len(s) > 0 {
			ret.Add(s)
		}
	}
	if len(ret) == 0 {
		return nil
	}
	return ret
}

func (p *Properties) getPattern(ctx *EvalContext) (map[string]interface{}, error) {
	if p == nil {
		return nil, nil
	}
	if p.Param != nil {
		value, err := ctx.GetParameter(string(*p.Param))
		if err != nil {
			return nil, err
		}
		m, ok := value.Value.(map[string]interface{})
		if !ok {
			return nil, ErrPropertiesParameterExpected
		}
		return m, nil
	}
	if p.Map != nil {
		value, err := p.Map.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		m, ok := value.Value.(map[string]interface{})
		if !ok {
			return nil, ErrPropertiesExpected
		}
		return m, nil
	}

	return nil, nil
}