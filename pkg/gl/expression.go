package gl

import ()

type Evaluatable interface {
	Evaluate(*Scope) (Value, error)
}

type lValueExpression string

func (l lValueExpression) Evaluate(*Scope) (Value, error) { return lValue{name: string(l)}, nil }

type nullLiteral struct{}

func (nullLiteral) Evaluate(*Scope) (Value, error) { return NullValue{}, nil }

type boolLiteral bool

func (l boolLiteral) Evaluate(*Scope) (Value, error) { return BoolValue(l), nil }

type stringLiteral string

func (l stringLiteral) Evaluate(*Scope) (Value, error) { return StringValue(l), nil }

type numberLiteral string

func (l numberLiteral) Evaluate(*Scope) (Value, error) { return NumberValue(l), nil }

type identifierValue string

func (expr identifierValue) Evaluate(scope *Scope) (Value, error) {
	value := scope.Get(string(expr))
	if value == nil {
		return nil, ErrUnknownIdentifier(expr)
	}
	return value, nil
}

type statementList struct {
	statements []Evaluatable
	newScope   bool
}

func (expr statementList) Evaluate(scope *Scope) (Value, error) {
	if expr.newScope {
		scope = scope.NewScope()
	}
	var ret Value
	var err error
	for _, stmt := range expr.statements {
		ret, err = stmt.Evaluate(scope)
		if err != nil {
			break
		}
	}
	return ret, err
}

type expressionStatement struct {
	value Evaluatable
}

func (expr expressionStatement) Evaluate(scope *Scope) (Value, error) {
	return expr.value.Evaluate(scope)
}

type assignmentExpression struct {
	lValue string
	rValue Evaluatable
}

func (expr assignmentExpression) Evaluate(scope *Scope) (Value, error) {
	v, err := expr.rValue.Evaluate(scope)
	if err != nil {
		return nil, err
	}
	scope.Set(expr.lValue, v)
	return v, nil
}

type logicalAndExpression struct {
	left  Evaluatable
	right Evaluatable
}

func (expr logicalAndExpression) Evaluate(scope *Scope) (Value, error) {
	v1, err := expr.left.Evaluate(scope)
	if err != nil {
		return nil, err
	}
	b1, err := v1.AsBool()
	if err != nil {
		return nil, err
	}
	if !b1 {
		return FalseValue, nil
	}
	v2, err := expr.right.Evaluate(scope)
	if err != nil {
		return nil, err
	}
	b2, err := v2.AsBool()
	if err != nil {
		return nil, err
	}
	if !b2 {
		return FalseValue, nil
	}
	return TrueValue, nil
}

type logicalOrExpression struct {
	left  Evaluatable
	right Evaluatable
}

func (expr logicalOrExpression) Evaluate(scope *Scope) (Value, error) {
	v1, err := expr.left.Evaluate(scope)
	if err != nil {
		return nil, err
	}
	b1, err := v1.AsBool()
	if err != nil {
		return nil, err
	}
	if b1 {
		return TrueValue, nil
	}
	v2, err := expr.right.Evaluate(scope)
	if err != nil {
		return nil, err
	}
	b2, err := v2.AsBool()
	if err != nil {
		return nil, err
	}
	if b2 {
		return TrueValue, nil
	}
	return FalseValue, nil
}

type selectExpression struct {
	base     Evaluatable
	selector string
}

func (expr selectExpression) Evaluate(scope *Scope) (Value, error) {
	base, err := expr.base.Evaluate(scope)
	if err != nil {
		return nil, err
	}
	return base.Selector(expr.selector)
}

type indexExpression struct {
	base  Evaluatable
	index Evaluatable
}

func (expr indexExpression) Evaluate(scope *Scope) (Value, error) {
	base, err := expr.base.Evaluate(scope)
	if err != nil {
		return nil, err
	}
	index, err := expr.index.Evaluate(scope)
	if err != nil {
		return nil, err
	}
	return base.Index(index)
}

type notExpression struct {
	base Evaluatable
}

func (expr notExpression) Evaluate(scope *Scope) (Value, error) {
	v, err := expr.base.Evaluate(scope)
	if err != nil {
		return nil, err
	}
	b, err := v.AsBool()
	if err != nil {
		return nil, err
	}
	return ValueOf(!b), nil
}

type equalityExpression struct {
	left  Evaluatable
	right Evaluatable
}

func (expr equalityExpression) Evaluate(scope *Scope) (Value, error) {
	v1, err := expr.left.Evaluate(scope)
	if err != nil {
		return nil, err
	}
	v2, err := expr.right.Evaluate(scope)
	if err != nil {
		return nil, err
	}
	b, err := v1.Eq(v2)
	if err != nil {
		return nil, err
	}
	return ValueOf(b), nil
}

type functionCallExpression struct {
	function Evaluatable
	args     []Evaluatable
}

func (expr functionCallExpression) Evaluate(scope *Scope) (Value, error) {
	f, err := expr.function.Evaluate(scope)
	if err != nil {
		return nil, err
	}
	argValues := make([]Value, 0, len(expr.args))
	for _, x := range expr.args {
		v, err := x.Evaluate(scope)
		if err != nil {
			return nil, err
		}
		argValues = append(argValues, v)
	}
	return f.Call(scope, argValues)
}

type closureExpression struct {
	symbol string
	f      Evaluatable
}

func (expr closureExpression) Evaluate(scope *Scope) (Value, error) {
	return Closure{Symbol: expr.symbol, F: expr.f}, nil
}
