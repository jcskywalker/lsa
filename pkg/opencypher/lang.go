package opencypher

import (
	"fmt"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	//	"github.com/cloudprivacylabs/lsa/pkg/gl/parser"
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
