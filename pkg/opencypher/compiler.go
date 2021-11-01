package opencypher

import (
	//	"github.com/antlr/antlr4/runtime/Go/antlr"

	"github.com/cloudprivacylabs/lsa/pkg/opencypher/parser"
)

// Compiler is the Opencypher compiler
type Compiler struct {
	*parser.BaseCypherListener
	stack []interface{}
	err   error
}

// NewCompiler returns a new Opencypher compiler
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
		c.err = ErrInvalidExpression("Nothing on stack")
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

// EnterOC_Cypher is called when production oC_Cypher is entered.
func (s *Compiler) EnterOC_Cypher(ctx *parser.OC_CypherContext) {
}

// ExitOC_Cypher is called when production oC_Cypher is exited.
func (s *Compiler) ExitOC_Cypher(ctx *parser.OC_CypherContext) {
}

// EnterOC_Statement is called when production oC_Statement is entered.
func (s *Compiler) EnterOC_Statement(ctx *parser.OC_StatementContext) {}

// ExitOC_Statement is called when production oC_Statement is exited.
func (s *Compiler) ExitOC_Statement(ctx *parser.OC_StatementContext) {}

// EnterOC_Query is called when production oC_Query is entered.
func (s *Compiler) EnterOC_Query(ctx *parser.OC_QueryContext) {}

// ExitOC_Query is called when production oC_Query is exited.
func (s *Compiler) ExitOC_Query(ctx *parser.OC_QueryContext) {}

// EnterOC_RegularQuery is called when production oC_RegularQuery is entered.
func (s *Compiler) EnterOC_RegularQuery(ctx *parser.OC_RegularQueryContext) {}

// ExitOC_RegularQuery is called when production oC_RegularQuery is exited.
func (s *Compiler) ExitOC_RegularQuery(ctx *parser.OC_RegularQueryContext) {
	// The body contains a single query, or a list of unions of single query
}

// EnterOC_Union is called when production oC_Union is entered.
func (s *Compiler) EnterOC_Union(ctx *parser.OC_UnionContext) {}

// ExitOC_Union is called when production oC_Union is exited.
func (s *Compiler) ExitOC_Union(ctx *parser.OC_UnionContext) {
}

// EnterOC_SingleQuery is called when production oC_SingleQuery is entered.
func (s *Compiler) EnterOC_SingleQuery(ctx *parser.OC_SingleQueryContext) {}

// ExitOC_SingleQuery is called when production oC_SingleQuery is exited.
func (s *Compiler) ExitOC_SingleQuery(ctx *parser.OC_SingleQueryContext) {}

// EnterOC_SinglePartQuery is called when production oC_SinglePartQuery is entered.
func (s *Compiler) EnterOC_SinglePartQuery(ctx *parser.OC_SinglePartQueryContext) {}

// ExitOC_SinglePartQuery is called when production oC_SinglePartQuery is exited.
func (s *Compiler) ExitOC_SinglePartQuery(ctx *parser.OC_SinglePartQueryContext) {}

// EnterOC_MultiPartQuery is called when production oC_MultiPartQuery is entered.
func (s *Compiler) EnterOC_MultiPartQuery(ctx *parser.OC_MultiPartQueryContext) {}

// ExitOC_MultiPartQuery is called when production oC_MultiPartQuery is exited.
func (s *Compiler) ExitOC_MultiPartQuery(ctx *parser.OC_MultiPartQueryContext) {}

// EnterOC_UpdatingClause is called when production oC_UpdatingClause is entered.
func (s *Compiler) EnterOC_UpdatingClause(ctx *parser.OC_UpdatingClauseContext) {}

// ExitOC_UpdatingClause is called when production oC_UpdatingClause is exited.
func (s *Compiler) ExitOC_UpdatingClause(ctx *parser.OC_UpdatingClauseContext) {}

// EnterOC_ReadingClause is called when production oC_ReadingClause is entered.
func (s *Compiler) EnterOC_ReadingClause(ctx *parser.OC_ReadingClauseContext) {}

// ExitOC_ReadingClause is called when production oC_ReadingClause is exited.
func (s *Compiler) ExitOC_ReadingClause(ctx *parser.OC_ReadingClauseContext) {}

// EnterOC_Match is called when production oC_Match is entered.
func (s *Compiler) EnterOC_Match(ctx *parser.OC_MatchContext) {}

// ExitOC_Match is called when production oC_Match is exited.
func (s *Compiler) ExitOC_Match(ctx *parser.OC_MatchContext) {}

// EnterOC_Unwind is called when production oC_Unwind is entered.
func (s *Compiler) EnterOC_Unwind(ctx *parser.OC_UnwindContext) {}

// ExitOC_Unwind is called when production oC_Unwind is exited.
func (s *Compiler) ExitOC_Unwind(ctx *parser.OC_UnwindContext) {}

// EnterOC_Merge is called when production oC_Merge is entered.
func (s *Compiler) EnterOC_Merge(ctx *parser.OC_MergeContext) {}

// ExitOC_Merge is called when production oC_Merge is exited.
func (s *Compiler) ExitOC_Merge(ctx *parser.OC_MergeContext) {}

// EnterOC_MergeAction is called when production oC_MergeAction is entered.
func (s *Compiler) EnterOC_MergeAction(ctx *parser.OC_MergeActionContext) {}

// ExitOC_MergeAction is called when production oC_MergeAction is exited.
func (s *Compiler) ExitOC_MergeAction(ctx *parser.OC_MergeActionContext) {}

// EnterOC_Create is called when production oC_Create is entered.
func (s *Compiler) EnterOC_Create(ctx *parser.OC_CreateContext) {}

// ExitOC_Create is called when production oC_Create is exited.
func (s *Compiler) ExitOC_Create(ctx *parser.OC_CreateContext) {}

// EnterOC_Set is called when production oC_Set is entered.
func (s *Compiler) EnterOC_Set(ctx *parser.OC_SetContext) {}

// ExitOC_Set is called when production oC_Set is exited.
func (s *Compiler) ExitOC_Set(ctx *parser.OC_SetContext) {}

// EnterOC_SetItem is called when production oC_SetItem is entered.
func (s *Compiler) EnterOC_SetItem(ctx *parser.OC_SetItemContext) {}

// ExitOC_SetItem is called when production oC_SetItem is exited.
func (s *Compiler) ExitOC_SetItem(ctx *parser.OC_SetItemContext) {}

// EnterOC_Delete is called when production oC_Delete is entered.
func (s *Compiler) EnterOC_Delete(ctx *parser.OC_DeleteContext) {}

// ExitOC_Delete is called when production oC_Delete is exited.
func (s *Compiler) ExitOC_Delete(ctx *parser.OC_DeleteContext) {}

// EnterOC_Remove is called when production oC_Remove is entered.
func (s *Compiler) EnterOC_Remove(ctx *parser.OC_RemoveContext) {}

// ExitOC_Remove is called when production oC_Remove is exited.
func (s *Compiler) ExitOC_Remove(ctx *parser.OC_RemoveContext) {}

// EnterOC_RemoveItem is called when production oC_RemoveItem is entered.
func (s *Compiler) EnterOC_RemoveItem(ctx *parser.OC_RemoveItemContext) {}

// ExitOC_RemoveItem is called when production oC_RemoveItem is exited.
func (s *Compiler) ExitOC_RemoveItem(ctx *parser.OC_RemoveItemContext) {}

// EnterOC_InQueryCall is called when production oC_InQueryCall is entered.
func (s *Compiler) EnterOC_InQueryCall(ctx *parser.OC_InQueryCallContext) {}

// ExitOC_InQueryCall is called when production oC_InQueryCall is exited.
func (s *Compiler) ExitOC_InQueryCall(ctx *parser.OC_InQueryCallContext) {}

// EnterOC_StandaloneCall is called when production oC_StandaloneCall is entered.
func (s *Compiler) EnterOC_StandaloneCall(ctx *parser.OC_StandaloneCallContext) {}

// ExitOC_StandaloneCall is called when production oC_StandaloneCall is exited.
func (s *Compiler) ExitOC_StandaloneCall(ctx *parser.OC_StandaloneCallContext) {}

// EnterOC_YieldItems is called when production oC_YieldItems is entered.
func (s *Compiler) EnterOC_YieldItems(ctx *parser.OC_YieldItemsContext) {}

// ExitOC_YieldItems is called when production oC_YieldItems is exited.
func (s *Compiler) ExitOC_YieldItems(ctx *parser.OC_YieldItemsContext) {}

// EnterOC_YieldItem is called when production oC_YieldItem is entered.
func (s *Compiler) EnterOC_YieldItem(ctx *parser.OC_YieldItemContext) {}

// ExitOC_YieldItem is called when production oC_YieldItem is exited.
func (s *Compiler) ExitOC_YieldItem(ctx *parser.OC_YieldItemContext) {}

// EnterOC_With is called when production oC_With is entered.
func (s *Compiler) EnterOC_With(ctx *parser.OC_WithContext) {}

// ExitOC_With is called when production oC_With is exited.
func (s *Compiler) ExitOC_With(ctx *parser.OC_WithContext) {}

// EnterOC_Return is called when production oC_Return is entered.
func (s *Compiler) EnterOC_Return(ctx *parser.OC_ReturnContext) {}

// ExitOC_Return is called when production oC_Return is exited.
func (s *Compiler) ExitOC_Return(ctx *parser.OC_ReturnContext) {}

// EnterOC_ProjectionBody is called when production oC_ProjectionBody is entered.
func (s *Compiler) EnterOC_ProjectionBody(ctx *parser.OC_ProjectionBodyContext) {}

// ExitOC_ProjectionBody is called when production oC_ProjectionBody is exited.
func (s *Compiler) ExitOC_ProjectionBody(ctx *parser.OC_ProjectionBodyContext) {}

// EnterOC_ProjectionItems is called when production oC_ProjectionItems is entered.
func (s *Compiler) EnterOC_ProjectionItems(ctx *parser.OC_ProjectionItemsContext) {}

// ExitOC_ProjectionItems is called when production oC_ProjectionItems is exited.
func (s *Compiler) ExitOC_ProjectionItems(ctx *parser.OC_ProjectionItemsContext) {}

// EnterOC_ProjectionItem is called when production oC_ProjectionItem is entered.
func (s *Compiler) EnterOC_ProjectionItem(ctx *parser.OC_ProjectionItemContext) {}

// ExitOC_ProjectionItem is called when production oC_ProjectionItem is exited.
func (s *Compiler) ExitOC_ProjectionItem(ctx *parser.OC_ProjectionItemContext) {}

// EnterOC_Order is called when production oC_Order is entered.
func (s *Compiler) EnterOC_Order(ctx *parser.OC_OrderContext) {}

// ExitOC_Order is called when production oC_Order is exited.
func (s *Compiler) ExitOC_Order(ctx *parser.OC_OrderContext) {}

// EnterOC_Skip is called when production oC_Skip is entered.
func (s *Compiler) EnterOC_Skip(ctx *parser.OC_SkipContext) {}

// ExitOC_Skip is called when production oC_Skip is exited.
func (s *Compiler) ExitOC_Skip(ctx *parser.OC_SkipContext) {}

// EnterOC_Limit is called when production oC_Limit is entered.
func (s *Compiler) EnterOC_Limit(ctx *parser.OC_LimitContext) {}

// ExitOC_Limit is called when production oC_Limit is exited.
func (s *Compiler) ExitOC_Limit(ctx *parser.OC_LimitContext) {}

// EnterOC_SortItem is called when production oC_SortItem is entered.
func (s *Compiler) EnterOC_SortItem(ctx *parser.OC_SortItemContext) {}

// ExitOC_SortItem is called when production oC_SortItem is exited.
func (s *Compiler) ExitOC_SortItem(ctx *parser.OC_SortItemContext) {}

// EnterOC_Where is called when production oC_Where is entered.
func (s *Compiler) EnterOC_Where(ctx *parser.OC_WhereContext) {}

// ExitOC_Where is called when production oC_Where is exited.
func (s *Compiler) ExitOC_Where(ctx *parser.OC_WhereContext) {}

// EnterOC_Pattern is called when production oC_Pattern is entered.
func (s *Compiler) EnterOC_Pattern(ctx *parser.OC_PatternContext) {}

// ExitOC_Pattern is called when production oC_Pattern is exited.
func (s *Compiler) ExitOC_Pattern(ctx *parser.OC_PatternContext) {}

// EnterOC_PatternPart is called when production oC_PatternPart is entered.
func (s *Compiler) EnterOC_PatternPart(ctx *parser.OC_PatternPartContext) {}

// ExitOC_PatternPart is called when production oC_PatternPart is exited.
func (s *Compiler) ExitOC_PatternPart(ctx *parser.OC_PatternPartContext) {}

// EnterOC_AnonymousPatternPart is called when production oC_AnonymousPatternPart is entered.
func (s *Compiler) EnterOC_AnonymousPatternPart(ctx *parser.OC_AnonymousPatternPartContext) {}

// ExitOC_AnonymousPatternPart is called when production oC_AnonymousPatternPart is exited.
func (s *Compiler) ExitOC_AnonymousPatternPart(ctx *parser.OC_AnonymousPatternPartContext) {}

// EnterOC_PatternElement is called when production oC_PatternElement is entered.
func (s *Compiler) EnterOC_PatternElement(ctx *parser.OC_PatternElementContext) {}

// ExitOC_PatternElement is called when production oC_PatternElement is exited.
func (s *Compiler) ExitOC_PatternElement(ctx *parser.OC_PatternElementContext) {}

// EnterOC_NodePattern is called when production oC_NodePattern is entered.
func (s *Compiler) EnterOC_NodePattern(ctx *parser.OC_NodePatternContext) {}

// ExitOC_NodePattern is called when production oC_NodePattern is exited.
func (s *Compiler) ExitOC_NodePattern(ctx *parser.OC_NodePatternContext) {}

// EnterOC_PatternElementChain is called when production oC_PatternElementChain is entered.
func (s *Compiler) EnterOC_PatternElementChain(ctx *parser.OC_PatternElementChainContext) {}

// ExitOC_PatternElementChain is called when production oC_PatternElementChain is exited.
func (s *Compiler) ExitOC_PatternElementChain(ctx *parser.OC_PatternElementChainContext) {}

// EnterOC_RelationshipPattern is called when production oC_RelationshipPattern is entered.
func (s *Compiler) EnterOC_RelationshipPattern(ctx *parser.OC_RelationshipPatternContext) {}

// ExitOC_RelationshipPattern is called when production oC_RelationshipPattern is exited.
func (s *Compiler) ExitOC_RelationshipPattern(ctx *parser.OC_RelationshipPatternContext) {}

// EnterOC_RelationshipDetail is called when production oC_RelationshipDetail is entered.
func (s *Compiler) EnterOC_RelationshipDetail(ctx *parser.OC_RelationshipDetailContext) {}

// ExitOC_RelationshipDetail is called when production oC_RelationshipDetail is exited.
func (s *Compiler) ExitOC_RelationshipDetail(ctx *parser.OC_RelationshipDetailContext) {}

// EnterOC_Properties is called when production oC_Properties is entered.
func (s *Compiler) EnterOC_Properties(ctx *parser.OC_PropertiesContext) {}

// ExitOC_Properties is called when production oC_Properties is exited.
func (s *Compiler) ExitOC_Properties(ctx *parser.OC_PropertiesContext) {}

// EnterOC_RelationshipTypes is called when production oC_RelationshipTypes is entered.
func (s *Compiler) EnterOC_RelationshipTypes(ctx *parser.OC_RelationshipTypesContext) {}

// ExitOC_RelationshipTypes is called when production oC_RelationshipTypes is exited.
func (s *Compiler) ExitOC_RelationshipTypes(ctx *parser.OC_RelationshipTypesContext) {}

// EnterOC_NodeLabels is called when production oC_NodeLabels is entered.
func (s *Compiler) EnterOC_NodeLabels(ctx *parser.OC_NodeLabelsContext) {}

// ExitOC_NodeLabels is called when production oC_NodeLabels is exited.
func (s *Compiler) ExitOC_NodeLabels(ctx *parser.OC_NodeLabelsContext) {}

// EnterOC_NodeLabel is called when production oC_NodeLabel is entered.
func (s *Compiler) EnterOC_NodeLabel(ctx *parser.OC_NodeLabelContext) {}

// ExitOC_NodeLabel is called when production oC_NodeLabel is exited.
func (s *Compiler) ExitOC_NodeLabel(ctx *parser.OC_NodeLabelContext) {}

// EnterOC_RangeLiteral is called when production oC_RangeLiteral is entered.
func (s *Compiler) EnterOC_RangeLiteral(ctx *parser.OC_RangeLiteralContext) {}

// ExitOC_RangeLiteral is called when production oC_RangeLiteral is exited.
func (s *Compiler) ExitOC_RangeLiteral(ctx *parser.OC_RangeLiteralContext) {}

// EnterOC_LabelName is called when production oC_LabelName is entered.
func (s *Compiler) EnterOC_LabelName(ctx *parser.OC_LabelNameContext) {}

// ExitOC_LabelName is called when production oC_LabelName is exited.
func (s *Compiler) ExitOC_LabelName(ctx *parser.OC_LabelNameContext) {}

// EnterOC_RelTypeName is called when production oC_RelTypeName is entered.
func (s *Compiler) EnterOC_RelTypeName(ctx *parser.OC_RelTypeNameContext) {}

// ExitOC_RelTypeName is called when production oC_RelTypeName is exited.
func (s *Compiler) ExitOC_RelTypeName(ctx *parser.OC_RelTypeNameContext) {}

// EnterOC_Expression is called when production oC_Expression is entered.
func (s *Compiler) EnterOC_Expression(ctx *parser.OC_ExpressionContext) {}

// ExitOC_Expression is called when production oC_Expression is exited.
func (s *Compiler) ExitOC_Expression(ctx *parser.OC_ExpressionContext) {}

// EnterOC_OrExpression is called when production oC_OrExpression is entered.
func (s *Compiler) EnterOC_OrExpression(ctx *parser.OC_OrExpressionContext) {}

// ExitOC_OrExpression is called when production oC_OrExpression is exited.
func (s *Compiler) ExitOC_OrExpression(ctx *parser.OC_OrExpressionContext) {}

// EnterOC_XorExpression is called when production oC_XorExpression is entered.
func (s *Compiler) EnterOC_XorExpression(ctx *parser.OC_XorExpressionContext) {}

// ExitOC_XorExpression is called when production oC_XorExpression is exited.
func (s *Compiler) ExitOC_XorExpression(ctx *parser.OC_XorExpressionContext) {}

// EnterOC_AndExpression is called when production oC_AndExpression is entered.
func (s *Compiler) EnterOC_AndExpression(ctx *parser.OC_AndExpressionContext) {}

// ExitOC_AndExpression is called when production oC_AndExpression is exited.
func (s *Compiler) ExitOC_AndExpression(ctx *parser.OC_AndExpressionContext) {}

// EnterOC_NotExpression is called when production oC_NotExpression is entered.
func (s *Compiler) EnterOC_NotExpression(ctx *parser.OC_NotExpressionContext) {}

// ExitOC_NotExpression is called when production oC_NotExpression is exited.
func (s *Compiler) ExitOC_NotExpression(ctx *parser.OC_NotExpressionContext) {}

// EnterOC_ComparisonExpression is called when production oC_ComparisonExpression is entered.
func (s *Compiler) EnterOC_ComparisonExpression(ctx *parser.OC_ComparisonExpressionContext) {}

// ExitOC_ComparisonExpression is called when production oC_ComparisonExpression is exited.
func (s *Compiler) ExitOC_ComparisonExpression(ctx *parser.OC_ComparisonExpressionContext) {}

// EnterOC_AddOrSubtractExpression is called when production oC_AddOrSubtractExpression is entered.
func (s *Compiler) EnterOC_AddOrSubtractExpression(ctx *parser.OC_AddOrSubtractExpressionContext) {
}

// ExitOC_AddOrSubtractExpression is called when production oC_AddOrSubtractExpression is exited.
func (s *Compiler) ExitOC_AddOrSubtractExpression(ctx *parser.OC_AddOrSubtractExpressionContext) {}

// EnterOC_MultiplyDivideModuloExpression is called when production oC_MultiplyDivideModuloExpression is entered.
func (s *Compiler) EnterOC_MultiplyDivideModuloExpression(ctx *parser.OC_MultiplyDivideModuloExpressionContext) {
}

// ExitOC_MultiplyDivideModuloExpression is called when production oC_MultiplyDivideModuloExpression is exited.
func (s *Compiler) ExitOC_MultiplyDivideModuloExpression(ctx *parser.OC_MultiplyDivideModuloExpressionContext) {
}

// EnterOC_PowerOfExpression is called when production oC_PowerOfExpression is entered.
func (s *Compiler) EnterOC_PowerOfExpression(ctx *parser.OC_PowerOfExpressionContext) {}

// ExitOC_PowerOfExpression is called when production oC_PowerOfExpression is exited.
func (s *Compiler) ExitOC_PowerOfExpression(ctx *parser.OC_PowerOfExpressionContext) {}

// EnterOC_UnaryAddOrSubtractExpression is called when production oC_UnaryAddOrSubtractExpression is entered.
func (s *Compiler) EnterOC_UnaryAddOrSubtractExpression(ctx *parser.OC_UnaryAddOrSubtractExpressionContext) {
}

// ExitOC_UnaryAddOrSubtractExpression is called when production oC_UnaryAddOrSubtractExpression is exited.
func (s *Compiler) ExitOC_UnaryAddOrSubtractExpression(ctx *parser.OC_UnaryAddOrSubtractExpressionContext) {
}

// EnterOC_StringListNullOperatorExpression is called when production oC_StringListNullOperatorExpression is entered.
func (s *Compiler) EnterOC_StringListNullOperatorExpression(ctx *parser.OC_StringListNullOperatorExpressionContext) {
}

// ExitOC_StringListNullOperatorExpression is called when production oC_StringListNullOperatorExpression is exited.
func (s *Compiler) ExitOC_StringListNullOperatorExpression(ctx *parser.OC_StringListNullOperatorExpressionContext) {
}

// EnterOC_ListOperatorExpression is called when production oC_ListOperatorExpression is entered.
func (s *Compiler) EnterOC_ListOperatorExpression(ctx *parser.OC_ListOperatorExpressionContext) {}

// ExitOC_ListOperatorExpression is called when production oC_ListOperatorExpression is exited.
func (s *Compiler) ExitOC_ListOperatorExpression(ctx *parser.OC_ListOperatorExpressionContext) {}

// EnterOC_StringOperatorExpression is called when production oC_StringOperatorExpression is entered.
func (s *Compiler) EnterOC_StringOperatorExpression(ctx *parser.OC_StringOperatorExpressionContext) {
}

// ExitOC_StringOperatorExpression is called when production oC_StringOperatorExpression is exited.
func (s *Compiler) ExitOC_StringOperatorExpression(ctx *parser.OC_StringOperatorExpressionContext) {
}

// EnterOC_NullOperatorExpression is called when production oC_NullOperatorExpression is entered.
func (s *Compiler) EnterOC_NullOperatorExpression(ctx *parser.OC_NullOperatorExpressionContext) {}

// ExitOC_NullOperatorExpression is called when production oC_NullOperatorExpression is exited.
func (s *Compiler) ExitOC_NullOperatorExpression(ctx *parser.OC_NullOperatorExpressionContext) {}

// EnterOC_PropertyOrLabelsExpression is called when production oC_PropertyOrLabelsExpression is entered.
func (s *Compiler) EnterOC_PropertyOrLabelsExpression(ctx *parser.OC_PropertyOrLabelsExpressionContext) {
}

// ExitOC_PropertyOrLabelsExpression is called when production oC_PropertyOrLabelsExpression is exited.
func (s *Compiler) ExitOC_PropertyOrLabelsExpression(ctx *parser.OC_PropertyOrLabelsExpressionContext) {
}

// EnterOC_Atom is called when production oC_Atom is entered.
func (s *Compiler) EnterOC_Atom(ctx *parser.OC_AtomContext) {}

// ExitOC_Atom is called when production oC_Atom is exited.
func (s *Compiler) ExitOC_Atom(ctx *parser.OC_AtomContext) {}

// EnterOC_Literal is called when production oC_Literal is entered.
func (s *Compiler) EnterOC_Literal(ctx *parser.OC_LiteralContext) {}

// ExitOC_Literal is called when production oC_Literal is exited.
func (s *Compiler) ExitOC_Literal(ctx *parser.OC_LiteralContext) {}

// EnterOC_BooleanLiteral is called when production oC_BooleanLiteral is entered.
func (s *Compiler) EnterOC_BooleanLiteral(ctx *parser.OC_BooleanLiteralContext) {}

// ExitOC_BooleanLiteral is called when production oC_BooleanLiteral is exited.
func (s *Compiler) ExitOC_BooleanLiteral(ctx *parser.OC_BooleanLiteralContext) {}

// EnterOC_ListLiteral is called when production oC_ListLiteral is entered.
func (s *Compiler) EnterOC_ListLiteral(ctx *parser.OC_ListLiteralContext) {}

// ExitOC_ListLiteral is called when production oC_ListLiteral is exited.
func (s *Compiler) ExitOC_ListLiteral(ctx *parser.OC_ListLiteralContext) {}

// EnterOC_PartialComparisonExpression is called when production oC_PartialComparisonExpression is entered.
func (s *Compiler) EnterOC_PartialComparisonExpression(ctx *parser.OC_PartialComparisonExpressionContext) {
}

// ExitOC_PartialComparisonExpression is called when production oC_PartialComparisonExpression is exited.
func (s *Compiler) ExitOC_PartialComparisonExpression(ctx *parser.OC_PartialComparisonExpressionContext) {
}

// EnterOC_ParenthesizedExpression is called when production oC_ParenthesizedExpression is entered.
func (s *Compiler) EnterOC_ParenthesizedExpression(ctx *parser.OC_ParenthesizedExpressionContext) {
}

// ExitOC_ParenthesizedExpression is called when production oC_ParenthesizedExpression is exited.
func (s *Compiler) ExitOC_ParenthesizedExpression(ctx *parser.OC_ParenthesizedExpressionContext) {}

// EnterOC_RelationshipsPattern is called when production oC_RelationshipsPattern is entered.
func (s *Compiler) EnterOC_RelationshipsPattern(ctx *parser.OC_RelationshipsPatternContext) {}

// ExitOC_RelationshipsPattern is called when production oC_RelationshipsPattern is exited.
func (s *Compiler) ExitOC_RelationshipsPattern(ctx *parser.OC_RelationshipsPatternContext) {}

// EnterOC_FilterExpression is called when production oC_FilterExpression is entered.
func (s *Compiler) EnterOC_FilterExpression(ctx *parser.OC_FilterExpressionContext) {}

// ExitOC_FilterExpression is called when production oC_FilterExpression is exited.
func (s *Compiler) ExitOC_FilterExpression(ctx *parser.OC_FilterExpressionContext) {}

// EnterOC_IdInColl is called when production oC_IdInColl is entered.
func (s *Compiler) EnterOC_IdInColl(ctx *parser.OC_IdInCollContext) {}

// ExitOC_IdInColl is called when production oC_IdInColl is exited.
func (s *Compiler) ExitOC_IdInColl(ctx *parser.OC_IdInCollContext) {}

// EnterOC_FunctionInvocation is called when production oC_FunctionInvocation is entered.
func (s *Compiler) EnterOC_FunctionInvocation(ctx *parser.OC_FunctionInvocationContext) {}

// ExitOC_FunctionInvocation is called when production oC_FunctionInvocation is exited.
func (s *Compiler) ExitOC_FunctionInvocation(ctx *parser.OC_FunctionInvocationContext) {}

// EnterOC_FunctionName is called when production oC_FunctionName is entered.
func (s *Compiler) EnterOC_FunctionName(ctx *parser.OC_FunctionNameContext) {}

// ExitOC_FunctionName is called when production oC_FunctionName is exited.
func (s *Compiler) ExitOC_FunctionName(ctx *parser.OC_FunctionNameContext) {}

// EnterOC_ExplicitProcedureInvocation is called when production oC_ExplicitProcedureInvocation is entered.
func (s *Compiler) EnterOC_ExplicitProcedureInvocation(ctx *parser.OC_ExplicitProcedureInvocationContext) {
}

// ExitOC_ExplicitProcedureInvocation is called when production oC_ExplicitProcedureInvocation is exited.
func (s *Compiler) ExitOC_ExplicitProcedureInvocation(ctx *parser.OC_ExplicitProcedureInvocationContext) {
}

// EnterOC_ImplicitProcedureInvocation is called when production oC_ImplicitProcedureInvocation is entered.
func (s *Compiler) EnterOC_ImplicitProcedureInvocation(ctx *parser.OC_ImplicitProcedureInvocationContext) {
}

// ExitOC_ImplicitProcedureInvocation is called when production oC_ImplicitProcedureInvocation is exited.
func (s *Compiler) ExitOC_ImplicitProcedureInvocation(ctx *parser.OC_ImplicitProcedureInvocationContext) {
}

// EnterOC_ProcedureResultField is called when production oC_ProcedureResultField is entered.
func (s *Compiler) EnterOC_ProcedureResultField(ctx *parser.OC_ProcedureResultFieldContext) {}

// ExitOC_ProcedureResultField is called when production oC_ProcedureResultField is exited.
func (s *Compiler) ExitOC_ProcedureResultField(ctx *parser.OC_ProcedureResultFieldContext) {}

// EnterOC_ProcedureName is called when production oC_ProcedureName is entered.
func (s *Compiler) EnterOC_ProcedureName(ctx *parser.OC_ProcedureNameContext) {}

// ExitOC_ProcedureName is called when production oC_ProcedureName is exited.
func (s *Compiler) ExitOC_ProcedureName(ctx *parser.OC_ProcedureNameContext) {}

// EnterOC_Namespace is called when production oC_Namespace is entered.
func (s *Compiler) EnterOC_Namespace(ctx *parser.OC_NamespaceContext) {}

// ExitOC_Namespace is called when production oC_Namespace is exited.
func (s *Compiler) ExitOC_Namespace(ctx *parser.OC_NamespaceContext) {}

// EnterOC_ListComprehension is called when production oC_ListComprehension is entered.
func (s *Compiler) EnterOC_ListComprehension(ctx *parser.OC_ListComprehensionContext) {}

// ExitOC_ListComprehension is called when production oC_ListComprehension is exited.
func (s *Compiler) ExitOC_ListComprehension(ctx *parser.OC_ListComprehensionContext) {}

// EnterOC_PatternComprehension is called when production oC_PatternComprehension is entered.
func (s *Compiler) EnterOC_PatternComprehension(ctx *parser.OC_PatternComprehensionContext) {}

// ExitOC_PatternComprehension is called when production oC_PatternComprehension is exited.
func (s *Compiler) ExitOC_PatternComprehension(ctx *parser.OC_PatternComprehensionContext) {}

// EnterOC_PropertyLookup is called when production oC_PropertyLookup is entered.
func (s *Compiler) EnterOC_PropertyLookup(ctx *parser.OC_PropertyLookupContext) {}

// ExitOC_PropertyLookup is called when production oC_PropertyLookup is exited.
func (s *Compiler) ExitOC_PropertyLookup(ctx *parser.OC_PropertyLookupContext) {}

// EnterOC_CaseExpression is called when production oC_CaseExpression is entered.
func (s *Compiler) EnterOC_CaseExpression(ctx *parser.OC_CaseExpressionContext) {}

// ExitOC_CaseExpression is called when production oC_CaseExpression is exited.
func (s *Compiler) ExitOC_CaseExpression(ctx *parser.OC_CaseExpressionContext) {}

// EnterOC_CaseAlternatives is called when production oC_CaseAlternatives is entered.
func (s *Compiler) EnterOC_CaseAlternatives(ctx *parser.OC_CaseAlternativesContext) {}

// ExitOC_CaseAlternatives is called when production oC_CaseAlternatives is exited.
func (s *Compiler) ExitOC_CaseAlternatives(ctx *parser.OC_CaseAlternativesContext) {}

// EnterOC_Variable is called when production oC_Variable is entered.
func (s *Compiler) EnterOC_Variable(ctx *parser.OC_VariableContext) {}

// ExitOC_Variable is called when production oC_Variable is exited.
func (s *Compiler) ExitOC_Variable(ctx *parser.OC_VariableContext) {}

// EnterOC_NumberLiteral is called when production oC_NumberLiteral is entered.
func (s *Compiler) EnterOC_NumberLiteral(ctx *parser.OC_NumberLiteralContext) {}

// ExitOC_NumberLiteral is called when production oC_NumberLiteral is exited.
func (s *Compiler) ExitOC_NumberLiteral(ctx *parser.OC_NumberLiteralContext) {}

// EnterOC_MapLiteral is called when production oC_MapLiteral is entered.
func (s *Compiler) EnterOC_MapLiteral(ctx *parser.OC_MapLiteralContext) {}

// ExitOC_MapLiteral is called when production oC_MapLiteral is exited.
func (s *Compiler) ExitOC_MapLiteral(ctx *parser.OC_MapLiteralContext) {}

// EnterOC_Parameter is called when production oC_Parameter is entered.
func (s *Compiler) EnterOC_Parameter(ctx *parser.OC_ParameterContext) {}

// ExitOC_Parameter is called when production oC_Parameter is exited.
func (s *Compiler) ExitOC_Parameter(ctx *parser.OC_ParameterContext) {}

// EnterOC_PropertyExpression is called when production oC_PropertyExpression is entered.
func (s *Compiler) EnterOC_PropertyExpression(ctx *parser.OC_PropertyExpressionContext) {}

// ExitOC_PropertyExpression is called when production oC_PropertyExpression is exited.
func (s *Compiler) ExitOC_PropertyExpression(ctx *parser.OC_PropertyExpressionContext) {}

// EnterOC_PropertyKeyName is called when production oC_PropertyKeyName is entered.
func (s *Compiler) EnterOC_PropertyKeyName(ctx *parser.OC_PropertyKeyNameContext) {}

// ExitOC_PropertyKeyName is called when production oC_PropertyKeyName is exited.
func (s *Compiler) ExitOC_PropertyKeyName(ctx *parser.OC_PropertyKeyNameContext) {}

// EnterOC_IntegerLiteral is called when production oC_IntegerLiteral is entered.
func (s *Compiler) EnterOC_IntegerLiteral(ctx *parser.OC_IntegerLiteralContext) {}

// ExitOC_IntegerLiteral is called when production oC_IntegerLiteral is exited.
func (s *Compiler) ExitOC_IntegerLiteral(ctx *parser.OC_IntegerLiteralContext) {}

// EnterOC_DoubleLiteral is called when production oC_DoubleLiteral is entered.
func (s *Compiler) EnterOC_DoubleLiteral(ctx *parser.OC_DoubleLiteralContext) {}

// ExitOC_DoubleLiteral is called when production oC_DoubleLiteral is exited.
func (s *Compiler) ExitOC_DoubleLiteral(ctx *parser.OC_DoubleLiteralContext) {}

// EnterOC_SchemaName is called when production oC_SchemaName is entered.
func (s *Compiler) EnterOC_SchemaName(ctx *parser.OC_SchemaNameContext) {}

// ExitOC_SchemaName is called when production oC_SchemaName is exited.
func (s *Compiler) ExitOC_SchemaName(ctx *parser.OC_SchemaNameContext) {}

// EnterOC_ReservedWord is called when production oC_ReservedWord is entered.
func (s *Compiler) EnterOC_ReservedWord(ctx *parser.OC_ReservedWordContext) {}

// ExitOC_ReservedWord is called when production oC_ReservedWord is exited.
func (s *Compiler) ExitOC_ReservedWord(ctx *parser.OC_ReservedWordContext) {}

// EnterOC_SymbolicName is called when production oC_SymbolicName is entered.
func (s *Compiler) EnterOC_SymbolicName(ctx *parser.OC_SymbolicNameContext) {}

// ExitOC_SymbolicName is called when production oC_SymbolicName is exited.
func (s *Compiler) ExitOC_SymbolicName(ctx *parser.OC_SymbolicNameContext) {}

// EnterOC_LeftArrowHead is called when production oC_LeftArrowHead is entered.
func (s *Compiler) EnterOC_LeftArrowHead(ctx *parser.OC_LeftArrowHeadContext) {}

// ExitOC_LeftArrowHead is called when production oC_LeftArrowHead is exited.
func (s *Compiler) ExitOC_LeftArrowHead(ctx *parser.OC_LeftArrowHeadContext) {}

// EnterOC_RightArrowHead is called when production oC_RightArrowHead is entered.
func (s *Compiler) EnterOC_RightArrowHead(ctx *parser.OC_RightArrowHeadContext) {}

// ExitOC_RightArrowHead is called when production oC_RightArrowHead is exited.
func (s *Compiler) ExitOC_RightArrowHead(ctx *parser.OC_RightArrowHeadContext) {}

// EnterOC_Dash is called when production oC_Dash is entered.
func (s *Compiler) EnterOC_Dash(ctx *parser.OC_DashContext) {}

// ExitOC_Dash is called when production oC_Dash is exited.
func (s *Compiler) ExitOC_Dash(ctx *parser.OC_DashContext) {}
