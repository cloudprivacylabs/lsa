package gl

import (
	"fmt"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/cloudprivacylabs/lsa/pkg/gl/parser"
)

//go:generate antlr4 -Dlanguage=Go gl.g4 -o parser

type errorListener struct {
	antlr.DefaultErrorListener
	err error
}

func (lst *errorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	if lst.err == nil {
		lst.err = ErrSyntax(fmt.Sprintf("line %d:%d %s ", line, column, msg))
	}
}

// Parse parses the given input string as a script and returns an evaluatable object
func Parse(input string) (Evaluatable, error) {
	lexer := parser.NewglLexer(antlr.NewInputStream(input))
	stream := antlr.NewCommonTokenStream(lexer, 0)
	p := parser.NewglParser(stream)
	errListener := errorListener{}
	p.AddErrorListener(&errListener)
	p.BuildParseTrees = true
	compiler := NewCompiler()
	antlr.ParseTreeWalkerDefault.Walk(compiler, p.Script())
	if err := compiler.Error(); err != nil {
		return nil, err
	}
	expr, ok := compiler.pop().(Evaluatable)
	if !ok {
		return nil, ErrNotExpression
	}
	return expr, nil
}

// EvaluateWith parses and evaluates a script with the given scope
func EvaluateWith(scope *Scope, input string) (Value, error) {
	e, err := Parse(input)
	if err != nil {
		return nil, err
	}
	return e.Evaluate(scope)
}

// Evaluate parses and evaluates a script with empty scope
func Evaluate(input string) (Value, error) {
	scope := NewScope()
	return EvaluateWith(scope, input)
}
