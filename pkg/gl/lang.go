package gl

import (
	"fmt"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/cloudprivacylabs/lsa/pkg/gl/parser"
)

//go:generate antlr4 -Dlanguage=Go gl.g4 -o parser

// type Expression interface {
// 	//	parser.IExpressionContext
// }

type errorListener struct {
	antlr.DefaultErrorListener
	err error
}

func (lst *errorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	if lst.err == nil {
		lst.err = ErrSyntax(fmt.Sprintf("line %d:%d %s ", line, column, msg))
	}
}

// ParseExpression parses the given input string as an expression and returns an evaluatable expression object
func ParseExpression(input string) (Expression, error) {
	lexer := parser.NewglLexer(antlr.NewInputStream(input))
	stream := antlr.NewCommonTokenStream(lexer, 0)
	p := parser.NewglParser(stream)
	errListener := errorListener{}
	p.AddErrorListener(&errListener)
	p.BuildParseTrees = true
	compiler := NewCompiler()
	antlr.ParseTreeWalkerDefault.Walk(compiler, p.Expression())
	if err := compiler.Error(); err != nil {
		return nil, err
	}
	expr, ok := compiler.pop().(Expression)
	if !ok {
		return nil, ErrNotExpression
	}
	return expr, nil
}

// EvaluateExpression parses and evaluates an expression
func EvaluateExpression(context *Context, input string) (Value, error) {
	e, err := ParseExpression(input)
	if err != nil {
		return nil, err
	}
	return e.Evaluate(context)
}
