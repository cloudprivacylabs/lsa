// Code generated from gl.g4 by ANTLR 4.9. DO NOT EDIT.

package parser // gl

import "github.com/antlr/antlr4/runtime/Go/antlr"

// BaseglListener is a complete listener for a parse tree produced by glParser.
type BaseglListener struct{}

var _ glListener = &BaseglListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseglListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseglListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseglListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseglListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterParenthesizedExpression is called when production ParenthesizedExpression is entered.
func (s *BaseglListener) EnterParenthesizedExpression(ctx *ParenthesizedExpressionContext) {}

// ExitParenthesizedExpression is called when production ParenthesizedExpression is exited.
func (s *BaseglListener) ExitParenthesizedExpression(ctx *ParenthesizedExpressionContext) {}

// EnterLogicalAndExpression is called when production LogicalAndExpression is entered.
func (s *BaseglListener) EnterLogicalAndExpression(ctx *LogicalAndExpressionContext) {}

// ExitLogicalAndExpression is called when production LogicalAndExpression is exited.
func (s *BaseglListener) ExitLogicalAndExpression(ctx *LogicalAndExpressionContext) {}

// EnterDotExpression is called when production DotExpression is entered.
func (s *BaseglListener) EnterDotExpression(ctx *DotExpressionContext) {}

// ExitDotExpression is called when production DotExpression is exited.
func (s *BaseglListener) ExitDotExpression(ctx *DotExpressionContext) {}

// EnterAssignmentExpression is called when production AssignmentExpression is entered.
func (s *BaseglListener) EnterAssignmentExpression(ctx *AssignmentExpressionContext) {}

// ExitAssignmentExpression is called when production AssignmentExpression is exited.
func (s *BaseglListener) ExitAssignmentExpression(ctx *AssignmentExpressionContext) {}

// EnterLiteralExpression is called when production LiteralExpression is entered.
func (s *BaseglListener) EnterLiteralExpression(ctx *LiteralExpressionContext) {}

// ExitLiteralExpression is called when production LiteralExpression is exited.
func (s *BaseglListener) ExitLiteralExpression(ctx *LiteralExpressionContext) {}

// EnterLogicalOrExpression is called when production LogicalOrExpression is entered.
func (s *BaseglListener) EnterLogicalOrExpression(ctx *LogicalOrExpressionContext) {}

// ExitLogicalOrExpression is called when production LogicalOrExpression is exited.
func (s *BaseglListener) ExitLogicalOrExpression(ctx *LogicalOrExpressionContext) {}

// EnterSearchExpression is called when production SearchExpression is entered.
func (s *BaseglListener) EnterSearchExpression(ctx *SearchExpressionContext) {}

// ExitSearchExpression is called when production SearchExpression is exited.
func (s *BaseglListener) ExitSearchExpression(ctx *SearchExpressionContext) {}

// EnterIndexExpression is called when production IndexExpression is entered.
func (s *BaseglListener) EnterIndexExpression(ctx *IndexExpressionContext) {}

// ExitIndexExpression is called when production IndexExpression is exited.
func (s *BaseglListener) ExitIndexExpression(ctx *IndexExpressionContext) {}

// EnterNotExpression is called when production NotExpression is entered.
func (s *BaseglListener) EnterNotExpression(ctx *NotExpressionContext) {}

// ExitNotExpression is called when production NotExpression is exited.
func (s *BaseglListener) ExitNotExpression(ctx *NotExpressionContext) {}

// EnterEqualityExpression is called when production EqualityExpression is entered.
func (s *BaseglListener) EnterEqualityExpression(ctx *EqualityExpressionContext) {}

// ExitEqualityExpression is called when production EqualityExpression is exited.
func (s *BaseglListener) ExitEqualityExpression(ctx *EqualityExpressionContext) {}

// EnterFunctionCallExpression is called when production FunctionCallExpression is entered.
func (s *BaseglListener) EnterFunctionCallExpression(ctx *FunctionCallExpressionContext) {}

// ExitFunctionCallExpression is called when production FunctionCallExpression is exited.
func (s *BaseglListener) ExitFunctionCallExpression(ctx *FunctionCallExpressionContext) {}

// EnterIdentifierExpression is called when production IdentifierExpression is entered.
func (s *BaseglListener) EnterIdentifierExpression(ctx *IdentifierExpressionContext) {}

// ExitIdentifierExpression is called when production IdentifierExpression is exited.
func (s *BaseglListener) ExitIdentifierExpression(ctx *IdentifierExpressionContext) {}

// EnterClosure is called when production closure is entered.
func (s *BaseglListener) EnterClosure(ctx *ClosureContext) {}

// ExitClosure is called when production closure is exited.
func (s *BaseglListener) ExitClosure(ctx *ClosureContext) {}

// EnterLvalue is called when production lvalue is entered.
func (s *BaseglListener) EnterLvalue(ctx *LvalueContext) {}

// ExitLvalue is called when production lvalue is exited.
func (s *BaseglListener) ExitLvalue(ctx *LvalueContext) {}

// EnterArguments is called when production arguments is entered.
func (s *BaseglListener) EnterArguments(ctx *ArgumentsContext) {}

// ExitArguments is called when production arguments is exited.
func (s *BaseglListener) ExitArguments(ctx *ArgumentsContext) {}

// EnterArgumentList is called when production argumentList is entered.
func (s *BaseglListener) EnterArgumentList(ctx *ArgumentListContext) {}

// ExitArgumentList is called when production argumentList is exited.
func (s *BaseglListener) ExitArgumentList(ctx *ArgumentListContext) {}

// EnterLiteral is called when production literal is entered.
func (s *BaseglListener) EnterLiteral(ctx *LiteralContext) {}

// ExitLiteral is called when production literal is exited.
func (s *BaseglListener) ExitLiteral(ctx *LiteralContext) {}

// EnterNumericLiteral is called when production numericLiteral is entered.
func (s *BaseglListener) EnterNumericLiteral(ctx *NumericLiteralContext) {}

// ExitNumericLiteral is called when production numericLiteral is exited.
func (s *BaseglListener) ExitNumericLiteral(ctx *NumericLiteralContext) {}

// EnterIdentifierName is called when production identifierName is entered.
func (s *BaseglListener) EnterIdentifierName(ctx *IdentifierNameContext) {}

// ExitIdentifierName is called when production identifierName is exited.
func (s *BaseglListener) ExitIdentifierName(ctx *IdentifierNameContext) {}
