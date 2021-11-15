package opencypher

import (
	"testing"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/parser"
)

func getParser(input string) *parser.CypherParser {
	lexer := parser.NewCypherLexer(antlr.NewInputStream(input))
	stream := antlr.NewCommonTokenStream(lexer, 0)
	p := parser.NewCypherParser(stream)
	p.BuildParseTrees = true
	return p
}

func TestExpr(t *testing.T) {
	c := getParser(`5  +  7+1`).OC_Expression()
	out := oC_Expression(c.(*parser.OC_ExpressionContext))
	result, err := out.Evaluate(NewEvalContext())
	if err != nil {
		t.Error(err)
	}
	if result.Value != 13 {
		t.Errorf("Wrong result: %+v %T", result, result.Value)
	}
}
