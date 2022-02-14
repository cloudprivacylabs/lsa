package opencypher

import (
	"fmt"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/parser"
)

//go:generate antlr4 -Dlanguage=Go Cypher.g4 -o parser

type errorListener struct {
	antlr.DefaultErrorListener
	err error
}

type ErrSyntax string
type ErrInvalidExpression string

func (e ErrSyntax) Error() string            { return "Syntax error: " + string(e) }
func (e ErrInvalidExpression) Error() string { return "Invalid expression: " + string(e) }

func (lst *errorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	if lst.err == nil {
		lst.err = ErrSyntax(fmt.Sprintf("line %d:%d %s ", line, column, msg))
	}
}

// GetParser returns a parser that will parse the input string
func GetParser(input string) *parser.CypherParser {
	lexer := parser.NewCypherLexer(antlr.NewInputStream(input))
	stream := antlr.NewCommonTokenStream(lexer, 0)
	p := parser.NewCypherParser(stream)
	p.BuildParseTrees = true
	return p
}

// GetEvaluatable returns an evaluatable object
func Parse(input string) (Evaluatable, error) {
	pr := GetParser(input)
	errListener := errorListener{}
	pr.AddErrorListener(&errListener)
	c := pr.OC_Cypher()
	if errListener.err != nil {
		return nil, errListener.err
	}
	out := oC_Cypher(c.(*parser.OC_CypherContext))
	return out, nil
}
