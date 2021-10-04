package gl

import (
	"strconv"

	"github.com/antlr/antlr4/runtime/Go/antlr"

	"github.com/cloudprivacylabs/lsa/pkg/gl/parser"
)

// Compiler is the graph language script compiler
type Compiler struct {
	*parser.BaseglListener
	stack []interface{}
	err   error
}

// NewCompiler returns a new gl compiler
func NewCompiler() *Compiler {
	return &Compiler{
		stack: make([]interface{}, 0, 64),
	}
}

// Error returns the first detected error
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

type statementBlockMarker struct {
	newScope bool
}

func (compiler *Compiler) EnterStatementList(c *parser.StatementListContext) {
	compiler.push(statementBlockMarker{})
}

func (compiler *Compiler) EnterStatementBlock(c *parser.StatementBlockContext) {
	compiler.push(statementBlockMarker{newScope: true})
}

func (compiler *Compiler) exitStatements() {
	// Pop until we hit the statement list
	statements := make([]Evaluatable, 0)
	for {
		v := compiler.pop()
		if s, ok := v.(statementBlockMarker); ok {
			result := statementList{newScope: s.newScope}
			for i := len(statements) - 1; i >= 0; i-- {
				result.statements = append(result.statements, statements[i])
			}
			compiler.push(result)
			return
		}
		ev, ok := v.(Evaluatable)
		if !ok {
			compiler.setError(ErrNotExpression)
			return
		}
		statements = append(statements, ev)
	}
}

func (compiler *Compiler) ExitStatementList(c *parser.StatementListContext) {
	compiler.exitStatements()
}

func (compiler *Compiler) ExitStatementBlock(c *parser.StatementBlockContext) {
	compiler.exitStatements()
}

func (compiler *Compiler) ExitExpressionStatement(c *parser.ExpressionStatementContext) {
	v, ok := compiler.pop().(Evaluatable)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	compiler.push(expressionStatement{value: v})
}

func (compiler *Compiler) ExitAssignmentExpression(c *parser.AssignmentExpressionContext) {
	val, ok := compiler.pop().(Evaluatable)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	lv, ok := compiler.pop().(lValueExpression)
	if !ok {
		compiler.setError(ErrNotLValue)
	}
	compiler.push(assignmentExpression{lValue: string(lv), rValue: val})
}

func (compiler *Compiler) ExitDefinitionExpression(c *parser.DefinitionExpressionContext) {
	val, ok := compiler.pop().(Evaluatable)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	lv, ok := compiler.pop().(lValueExpression)
	if !ok {
		compiler.setError(ErrNotLValue)
	}
	compiler.push(definitionExpression{lValue: string(lv), rValue: val})
}

func (compiler *Compiler) ExitLvalue(c *parser.LvalueContext) {
	compiler.push(lValueExpression(c.GetText()))
}

func (compiler *Compiler) ExitLiteralExpression(c *parser.LiteralExpressionContext) {
	text := c.GetText()
	if text == "null" {
		compiler.push(nullLiteral{})
	} else if text == "true" || text == "false" {
		compiler.push(boolLiteral(text == "true"))
	} else if len(text) > 0 && text[0] == '"' || text[0] == '\'' {
		ntext := "\"" + text[1:len(text)-1] + "\""
		q, err := strconv.Unquote(ntext)
		compiler.setError(err)
		compiler.push(stringLiteral(q))
	} else {
		compiler.push(numberLiteral(text))
	}
}

func (compiler *Compiler) ExitIdentifierExpression(c *parser.IdentifierExpressionContext) {
	compiler.push(identifierValue(c.GetText()))
}

func (compiler *Compiler) ExitLogicalAndExpression(c *parser.LogicalAndExpressionContext) {
	v1, ok := compiler.pop().(Evaluatable)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	v2, ok := compiler.pop().(Evaluatable)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	compiler.push(logicalAndExpression{right: v1, left: v2})
}

func (compiler *Compiler) ExitDotExpression(c *parser.DotExpressionContext) {
	selector := c.GetStop().GetText()
	base, ok := compiler.pop().(Evaluatable)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	compiler.push(selectExpression{base: base, selector: selector})
}

func (compiler *Compiler) ExitLogicalOrExpression(c *parser.LogicalOrExpressionContext) {
	v1, ok := compiler.pop().(Evaluatable)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	v2, ok := compiler.pop().(Evaluatable)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	compiler.push(logicalOrExpression{right: v1, left: v2})
}

func (compiler *Compiler) ExitIndexExpression(c *parser.IndexExpressionContext) {
	index, ok := compiler.pop().(Evaluatable)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	base, ok := compiler.pop().(Evaluatable)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	compiler.push(indexExpression{base: base, index: index})
}

func (compiler *Compiler) ExitNotExpression(c *parser.NotExpressionContext) {
	val, ok := compiler.pop().(Evaluatable)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	compiler.push(notExpression{base: val})
}

func (compiler *Compiler) ExitEqualityExpression(c *parser.EqualityExpressionContext) {
	v1, ok := compiler.pop().(Evaluatable)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	v2, ok := compiler.pop().(Evaluatable)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	eq := Evaluatable(equalityExpression{left: v2, right: v1})
	if c.GetChild(1).(antlr.TerminalNode).GetSymbol().GetText() == "!=" {
		eq = notExpression{base: eq}
	}
	compiler.push(eq)
}

func (compiler *Compiler) ExitFunctionCallExpression(c *parser.FunctionCallExpressionContext) {
	a := compiler.pop()
	args, ok := a.([]Evaluatable)
	if !ok {
		compiler.setError(ErrEvaluation)
		return
	}
	f, ok := compiler.pop().(Evaluatable)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	compiler.push(functionCallExpression{function: f, args: args})
}

func (compiler *Compiler) ExitArguments(c *parser.ArgumentsContext) {
	argumentList := c.GetChild(1) // Argument list
	ch := argumentList.GetChildCount()
	ch = (ch + 1) / 2 // Subtract the commas
	if ch < 0 {
		return
	}
	args := make([]Evaluatable, ch)
	for i := ch - 1; i >= 0; i-- {
		var ok bool
		args[i], ok = compiler.pop().(Evaluatable)
		if !ok {
			compiler.setError(ErrNotExpression)
		}
	}
	compiler.push(args)
}

func (compiler *Compiler) exitClosureExpression(identifier string) {
	expr, ok := compiler.pop().(Evaluatable)
	if !ok {
		compiler.setError(ErrNotExpression)
	}
	compiler.push(closureExpression{symbol: identifier, f: expr})
}

func (compiler *Compiler) ExitClosureExpression(c *parser.ClosureExpressionContext) {
	compiler.exitClosureExpression(c.GetStart().GetText())
}

func (compiler *Compiler) ExitBlockClosureExpression(c *parser.BlockClosureExpressionContext) {
	compiler.exitClosureExpression(c.GetStart().GetText())
}
