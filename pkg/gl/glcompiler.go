package gl

import (
	"errors"
	"strconv"

	"github.com/antlr/antlr4/runtime/Go/antlr"

	"github.com/cloudprivacylabs/lsa/pkg/gl/parser"
)

var ErrInvalidExpression = errors.New("Invalid expression")
var ErrNotExpression = errors.New("Not an expression")

type Compiler struct {
	*parser.BaseglListener
	stack []interface{}
	err   error
}

func NewCompiler() *Compiler {
	return &Compiler{
		stack: make([]interface{}, 0, 64),
	}
}

func (c *Compiler) Error() error {
	return c.err
}

func (c *Compiler) push(val interface{}) {
	if c.err != nil {
		return
	}
	c.stack = append(c.stack, val)
}

func (c *Compiler) pop() interface{} {
	if c.err != nil {
		return nil
	}
	if len(c.stack) == 0 {
		c.err = ErrInvalidExpression
		return nil
	}
	ret := c.stack[len(c.stack)-1]
	c.stack = c.stack[:len(c.stack)-1]
	return ret
}

func (c *Compiler) setError(err error) {
	if c.err == nil && err != nil {
		c.err = err
	}
}

func (compiler *Compiler) ExitAssignmentExpression(c *parser.AssignmentExpressionContext) {
	val, ok := compiler.pop().(Expression)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	lv, ok := compiler.pop().(LValueExpression)
	if !ok {
		compiler.setError(ErrNotLValue)
	}
	compiler.push(AssignmentExpression{LValue: string(lv), RValue: val})
}

func (compiler *Compiler) ExitLvalue(c *parser.LvalueContext) {
	compiler.push(LValueExpression(c.GetText()))
}

func (compiler *Compiler) ExitLiteralExpression(c *parser.LiteralExpressionContext) {
	text := c.GetText()
	if text == "null" {
		compiler.push(NullLiteral{})
	} else if text == "true" || text == "false" {
		compiler.push(BoolLiteral(text == "true"))
	} else if len(text) > 0 && text[0] == '"' {
		q, err := strconv.Unquote(text)
		compiler.setError(err)
		compiler.push(StringLiteral(q))
	} else {
		compiler.push(NumberLiteral(text))
	}
}

func (compiler *Compiler) ExitIdentifierExpression(c *parser.IdentifierExpressionContext) {
	compiler.push(IdentifierValue(c.GetText()))
}

func (compiler *Compiler) ExitLogicalAndExpression(c *parser.LogicalAndExpressionContext) {
	v1, ok := compiler.pop().(Expression)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	v2, ok := compiler.pop().(Expression)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	compiler.push(LogicalAndExpression{Right: v1, Left: v2})
}

func (compiler *Compiler) ExitDotExpression(c *parser.DotExpressionContext) {
	selector := c.GetStop().GetText()
	base, ok := compiler.pop().(Expression)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	compiler.push(SelectExpression{Base: base, Selector: selector})
}

func (compiler *Compiler) ExitLogicalOrExpression(c *parser.LogicalOrExpressionContext) {
	v1, ok := compiler.pop().(Expression)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	v2, ok := compiler.pop().(Expression)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	compiler.push(LogicalOrExpression{Right: v1, Left: v2})
}

func (compiler *Compiler) ExitIndexExpression(c *parser.IndexExpressionContext) {
	index, ok := compiler.pop().(Expression)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	base, ok := compiler.pop().(Expression)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	compiler.push(IndexExpression{Base: base, Index: index})
}

func (compiler *Compiler) ExitNotExpression(c *parser.NotExpressionContext) {
	val, ok := compiler.pop().(Expression)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	compiler.push(NotExpression{Base: val})
}

func (compiler *Compiler) ExitEqualityExpression(c *parser.EqualityExpressionContext) {
	v1, ok := compiler.pop().(Expression)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	v2, ok := compiler.pop().(Expression)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	eq := Expression(EqualityExpression{Left: v2, Right: v1})
	if c.GetChild(1).(antlr.TerminalNode).GetSymbol().GetText() == "!=" {
		eq = NotExpression{Base: eq}
	}
	compiler.push(eq)
}

func (compiler *Compiler) ExitFunctionCallExpression(c *parser.FunctionCallExpressionContext) {
	a := compiler.pop()
	args, ok := a.([]Expression)
	if !ok {
		compiler.setError(ErrEvaluation)
		return
	}
	f, ok := compiler.pop().(Expression)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	compiler.push(FunctionCallExpression{Function: f, Args: args})
}

func (compiler *Compiler) ExitArguments(c *parser.ArgumentsContext) {
	argumentList := c.GetChild(1) // Argument list
	ch := argumentList.GetChildCount()
	ch = (ch + 1) / 2 // Subtract the commas
	if ch < 0 {
		return
	}
	args := make([]Expression, ch)
	for i := ch - 1; i >= 0; i-- {
		var ok bool
		args[i], ok = compiler.pop().(Expression)
		if !ok {
			compiler.setError(ErrNotExpression)
		}
	}
	compiler.push(args)
}

func (compiler *Compiler) ExitClosure(c *parser.ClosureContext) {
	expr, ok := compiler.pop().(Expression)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	identifier := c.GetStart().GetText()
	compiler.push(Closure{Symbol: identifier, Closure: expr})
}

func (compiler *Compiler) ExitSearchExpression(c *parser.SearchExpressionContext) {
	closure, ok := compiler.pop().(Closure)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	expr, ok := compiler.pop().(Expression)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	compiler.push(SearchExpression{Base: expr, Closure: closure})
}
