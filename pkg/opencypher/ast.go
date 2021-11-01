package opencypher

import (
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/parser"
)

type Evaluatable interface{}

func oC_Cypher(ctx *parser.OC_CypherContext) Evaluatable {
	return oC_Statement(ctx.OC_Statement().(*parser.OC_StatementContext))
}

func oC_Statement(ctx *parser.OC_StatementContext) Evaluatable {
	return oC_Query(ctx.OC_Query().(*parser.OC_QueryContext))
}

func oC_Query(ctx *parser.OC_QueryContext) Evaluatable {
	if x := ctx.OC_RegularQuery(); x != nil {
		return oC_RegularQuery(x.(*parser.OC_RegularQueryContext))
	}
	return oC_StandaloneCall(ctx.OC_StandaloneCall().(*parser.OC_StandaloneCallContext))
}

type RegularQuery struct {
	SingleQuery Evaluatable
	Unions      []Union
}

func oC_RegularQuery(ctx *parser.OC_RegularQueryContext) RegularQuery {
	ret := RegularQuery{
		SingleQuery: oC_SingleQuery(ctx.OC_SingleQuery().(*parser.OC_SingleQueryContext)),
	}
	for _, u := range ctx.AllOC_Union() {
		ret.Unions = append(ret.Unions, oC_Union(u.(*parser.OC_UnionContext)))
	}
	return ret
}

type Union struct {
	All         bool
	SingleQuery Evaluatable
}

func oC_Union(ctx *parser.OC_UnionContext) Union {
	return Union{
		All:         ctx.ALL() != nil,
		SingleQuery: oC_SingleQuery(ctx.OC_SingleQuery().(*parser.OC_SingleQueryContext)),
	}
}

func oC_SingleQuery(ctx *parser.OC_SingleQueryContext) Evaluatable {
	if x := ctx.OC_SinglePartQuery(); x != nil {
		return oC_SinglePartQuery(x.(*parser.OC_SinglePartQueryContext))
	}
	return oC_MultiPartQuery(ctx.OC_MultiPartQuery().(*parser.OC_MultiPartQueryContext))
}

type SinglePartQuery struct {
	Read   []ReadingClause
	Update []UpdatingClause
	Return *ReturnClause
}

func oC_SinglePartQuery(ctx *parser.OC_SinglePartQueryContext) SinglePartQuery {
	ret := SinglePartQuery{}
	for _, r := range ctx.AllOC_ReadingClause() {
		ret.Read = append(ret.Read, oC_ReadingClause(r.(*parser.OC_ReadingClauseContext)))
	}
	for _, u := range ctx.AllOC_UpdatingClause() {
		ret.Update = append(ret.Update, oC_UpdatingClause(u.(*parser.OC_UpdatingClauseContext)))
	}
	if x := ctx.OC_Return(); x != nil {
		ret.Return = oC_Return(x.(*parser.OC_ReturnContext))
	}
	return ret
}

type ReadingClause interface {
	Evaluatable
}

func oC_ReadingClause(ctx *parser.OC_ReadingClauseContext) ReadingClause {
	if match := ctx.OC_Match(); match != nil {
		return oC_Match(match.(*parser.OC_MatchContext))
	}
	if unwind := ctx.OC_Unwind(); unwind != nil {
		return oC_Unwind(unwind.(*parser.OC_UnwindContext))
	}
	return oC_InQueryCall(ctx.OC_InQueryCall().(*parser.OC_InQueryCallContext))
}

type UpdatingClause interface {
	Evaluatable
}

func oC_UpdatingClause(ctx *parser.OC_UpdatingClauseContext) UpdatingClause {
	if create := ctx.OC_Create(); create != nil {
		return oC_Create(create.(*parser.OC_CreateContext))
	}
	if merge := ctx.OC_Merge(); merge != nil {
		return oC_Merge(merge.(*parser.OC_MergeContext))
	}
	if del := ctx.OC_Delete(); del != nil {
		return oC_Delete(del.(*parser.OC_DeleteContext))
	}
	if set := ctx.OC_Set(); set != nil {
		return oC_Set(set.(*parser.OC_SetContext))
	}
	return oC_Remove(ctx.OC_Remove().(*parser.OC_RemoveContext))
}

func oC_MultiPartQuery(ctx *parser.OC_MultiPartQueryContext) Evaluatable {
	return nil
}

type ReturnClause struct {
}

func oC_Return(ctx *parser.OC_ReturnContext) *ReturnClause {
	return nil
}

type Match struct {
	Optional bool
	Pattern  Evaluatable
	Where    Expression
}

func oC_Match(ctx *parser.OC_MatchContext) Match {
	ret := Match{
		Optional: ctx.OPTIONAL() != nil,
		Pattern:  oC_Pattern(ctx.OC_Pattern().(*parser.OC_PatternContext)),
	}
	if w := ctx.OC_Where(); w != nil {
		ret.Where = oC_Where(w.(*parser.OC_WhereContext))
	}
	return ret
}

func oC_Where(ctx *parser.OC_WhereContext) Expression {
	return oC_Expression(ctx.OC_Expression().(*parser.OC_ExpressionContext))
}

type Expression interface {
	Evaluatable
}

func oC_Expression(ctx *parser.OC_ExpressionContext) Expression {
	return oC_OrExpression(ctx.OC_OrExpression().(*parser.OC_OrExpressionContext))
}

type OrExpression struct {
	Parts []Evaluatable
}

func oC_OrExpression(ctx *parser.OC_OrExpressionContext) Expression {
	ret := OrExpression{}
	for _, x := range ctx.AllOC_XorExpression() {
		ret.Parts = append(ret.Parts, oC_XorExpression(x.(*parser.OC_XorExpressionContext)))
	}
	return ret
}

type XorExpression struct {
	Parts []Evaluatable
}

func oC_XorExpression(ctx *parser.OC_XorExpressionContext) Expression {
	ret := XorExpression{}
	for _, x := range ctx.AllOC_AndExpression() {
		ret.Parts = append(ret.Parts, oC_AndExpression(x.(*parser.OC_AndExpressionContext)))
	}
	return ret
}

type AndExpression struct {
	Parts []Evaluatable
}

func oC_AndExpression(ctx *parser.OC_AndExpressionContext) Expression {
	ret := AndExpression{}
	for _, x := range ctx.AllOC_NotExpression() {
		ret.Parts = append(ret.Parts, oC_NotExpression(x.(*parser.OC_NotExpressionContext)))
	}
	return ret
}

type NotExpression struct {
	Part Evaluatable
}

func oC_NotExpression(ctx *parser.OC_NotExpressionContext) Expression {
	if len(ctx.AllNOT())%2 == 1 {
		return NotExpression{
			Part: oC_ComparisonExpression(ctx.OC_ComparisonExpression().(*parser.OC_ComparisonExpressionContext)),
		}
	}
	return oC_ComparisonExpression(ctx.OC_ComparisonExpression().(*parser.OC_ComparisonExpressionContext))
}

type ComparisonExpression struct {
	First  Expression
	Second []PartialComparisonExpression
}

func oC_ComparisonExpression(ctx *parser.OC_ComparisonExpression) Expression {
	ret := ComparisonExpression{
		First: oC_AddOrSubtractExpression(ctx.OC_AddOrSubtractExpression().(*parser.OC_AddOrSubtractExpressionContext)),
	}
	for _, x := range ctx.AllOC_PartialComparisonExpression() {
		ret.Second = append(ret.Second, oC_PartialComparisonExpression(x.(*parser.OC_PartialComparisonExpression)))
	}
	return ret
}

func oC_Pattern(ctx *parser.OC_PatternContext) Evaluatable {
	return nil
}

func oC_Unwind(ctx *parser.OC_UnwindContext) Evaluatable {
	return nil
}

func oC_InQueryCall(ctx *parser.OC_InQueryCallContext) Evaluatable {
	return nil
}

func oC_Create(ctx *parser.OC_CreateContext) Evaluatable {
	return nil
}

func oC_Merge(ctx *parser.OC_MergeContext) Evaluatable {
	return nil
}

func oC_Delete(ctx *parser.OC_DeleteContext) Evaluatable {
	return nil
}

func oC_Set(ctx *parser.OC_SetContext) Evaluatable {
	return nil
}

func oC_Remove(ctx *parser.OC_RemoveContext) Evaluatable {
	return nil
}

func oC_StandaloneCall(ctx *parser.OC_StandaloneCallContext) Evaluatable {
	return nil
}
