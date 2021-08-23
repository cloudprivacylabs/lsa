package gl

import ()

type Expression interface {
	Evaluate(*Context) (Value, error)
}

type FuncAsExpression func(*Context) (Value, error)

func (f FuncAsExpression) Evaluate(ctx *Context) (Value, error) { return f(ctx) }

type LValueExpression string

func (l LValueExpression) Evaluate(*Context) (Value, error) { return LValue{Name: string(l)}, nil }

type NullLiteral struct{}

func (l NullLiteral) Evaluate(*Context) (Value, error) { return NullValue{}, nil }

type BoolLiteral bool

func (l BoolLiteral) Evaluate(*Context) (Value, error) { return BoolValue(l), nil }

type StringLiteral string

func (l StringLiteral) Evaluate(*Context) (Value, error) { return StringValue(l), nil }

type NumberLiteral string

func (l NumberLiteral) Evaluate(*Context) (Value, error) { return NumberValue(l), nil }

type IdentifierValue string

func (expr IdentifierValue) Evaluate(ctx *Context) (Value, error) {
	value := ctx.Get(string(expr))
	if value == nil {
		return nil, ErrUnknownIdentifier(expr)
	}
	return value, nil
}

type AssignmentExpression struct {
	LValue string
	RValue Expression
}

func (expr AssignmentExpression) Evaluate(ctx *Context) (Value, error) {
	v, err := expr.RValue.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	ctx.Set(expr.LValue, v)
	return v, nil
}

type LogicalAndExpression struct {
	Left  Expression
	Right Expression
}

func (expr LogicalAndExpression) Evaluate(ctx *Context) (Value, error) {
	v1, err := expr.Left.Evaluate(ctx)
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
	v2, err := expr.Right.Evaluate(ctx)
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

type LogicalOrExpression struct {
	Left  Expression
	Right Expression
}

func (expr LogicalOrExpression) Evaluate(ctx *Context) (Value, error) {
	v1, err := expr.Left.Evaluate(ctx)
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
	v2, err := expr.Right.Evaluate(ctx)
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

type SelectExpression struct {
	Base     Expression
	Selector string
}

func (expr SelectExpression) Evaluate(ctx *Context) (Value, error) {
	base, err := expr.Base.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	return base.Selector(expr.Selector)
}

type IndexExpression struct {
	Base  Expression
	Index Expression
}

func (expr IndexExpression) Evaluate(ctx *Context) (Value, error) {
	base, err := expr.Base.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	index, err := expr.Index.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	return base.Index(index)
}

type NotExpression struct {
	Base Expression
}

func (expr NotExpression) Evaluate(ctx *Context) (Value, error) {
	v, err := expr.Base.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	b, err := v.AsBool()
	if err != nil {
		return nil, err
	}
	return ValueOf(!b), nil
}

type EqualityExpression struct {
	Left  Expression
	Right Expression
}

func (expr EqualityExpression) Evaluate(ctx *Context) (Value, error) {
	v1, err := expr.Left.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	v2, err := expr.Right.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	b, err := v1.Eq(v2)
	if err != nil {
		return nil, err
	}
	return ValueOf(b), nil
}

type FunctionCallExpression struct {
	Function Expression
	Args     []Expression
}

func (expr FunctionCallExpression) Evaluate(ctx *Context) (Value, error) {
	f, err := expr.Function.Evaluate(ctx)
	if err != nil {
		return nil, err
	}
	argValues := make([]Value, 0, len(expr.Args))
	for _, x := range expr.Args {
		v, err := x.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		argValues = append(argValues, v)
	}
	return f.Call(ctx, argValues)
}

type ClosureExpression struct {
	Symbol string
	F      Expression
}

func (expr ClosureExpression) Evaluate(ctx *Context) (Value, error) {
	return Closure{Symbol: expr.Symbol, F: expr.F}, nil
}
