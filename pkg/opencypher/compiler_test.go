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

func TestUnion(t *testing.T) {
	c := getParser(`match (n) return n union all match (n) return n`).OC_Cypher()

	oC_Cypher(c.(*parser.OC_CypherContext))
}

func TestExpr(t *testing.T) {
	c := getParser(`x  +  y+z`).OC_Expression()

	out := oC_Expression(c.(*parser.OC_ExpressionContext))
	t.Logf("%+v", out)
}
