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

// EnterExpressionScript is called when production ExpressionScript is entered.
func (s *BaseglListener) EnterExpressionScript(ctx *ExpressionScriptContext) {}

// ExitExpressionScript is called when production ExpressionScript is exited.
func (s *BaseglListener) ExitExpressionScript(ctx *ExpressionScriptContext) {}

// EnterStatementListScript is called when production StatementListScript is entered.
func (s *BaseglListener) EnterStatementListScript(ctx *StatementListScriptContext) {}

// ExitStatementListScript is called when production StatementListScript is exited.
func (s *BaseglListener) ExitStatementListScript(ctx *StatementListScriptContext) {}

// EnterStatement is called when production statement is entered.
func (s *BaseglListener) EnterStatement(ctx *StatementContext) {}

// ExitStatement is called when production statement is exited.
func (s *BaseglListener) ExitStatement(ctx *StatementContext) {}

// EnterStatementList is called when production statementList is entered.
func (s *BaseglListener) EnterStatementList(ctx *StatementListContext) {}

// ExitStatementList is called when production statementList is exited.
func (s *BaseglListener) ExitStatementList(ctx *StatementListContext) {}

// EnterStatementBlock is called when production statementBlock is entered.
func (s *BaseglListener) EnterStatementBlock(ctx *StatementBlockContext) {}

// ExitStatementBlock is called when production statementBlock is exited.
func (s *BaseglListener) ExitStatementBlock(ctx *StatementBlockContext) {}

// EnterExpressionStatement is called when production expressionStatement is entered.
func (s *BaseglListener) EnterExpressionStatement(ctx *ExpressionStatementContext) {}

// ExitExpressionStatement is called when production expressionStatement is exited.
func (s *BaseglListener) ExitExpressionStatement(ctx *ExpressionStatementContext) {}

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

// EnterLiteralExpression is called when production LiteralExpression is entered.
func (s *BaseglListener) EnterLiteralExpression(ctx *LiteralExpressionContext) {}

// ExitLiteralExpression is called when production LiteralExpression is exited.
func (s *BaseglListener) ExitLiteralExpression(ctx *LiteralExpressionContext) {}

// EnterLogicalOrExpression is called when production LogicalOrExpression is entered.
func (s *BaseglListener) EnterLogicalOrExpression(ctx *LogicalOrExpressionContext) {}

// ExitLogicalOrExpression is called when production LogicalOrExpression is exited.
func (s *BaseglListener) ExitLogicalOrExpression(ctx *LogicalOrExpressionContext) {}

// EnterDefinitionExpression is called when production DefinitionExpression is entered.
func (s *BaseglListener) EnterDefinitionExpression(ctx *DefinitionExpressionContext) {}

// ExitDefinitionExpression is called when production DefinitionExpression is exited.
func (s *BaseglListener) ExitDefinitionExpression(ctx *DefinitionExpressionContext) {}

// EnterIndexExpression is called when production IndexExpression is entered.
func (s *BaseglListener) EnterIndexExpression(ctx *IndexExpressionContext) {}

// ExitIndexExpression is called when production IndexExpression is exited.
func (s *BaseglListener) ExitIndexExpression(ctx *IndexExpressionContext) {}

// EnterNotExpression is called when production NotExpression is entered.
func (s *BaseglListener) EnterNotExpression(ctx *NotExpressionContext) {}

// ExitNotExpression is called when production NotExpression is exited.
func (s *BaseglListener) ExitNotExpression(ctx *NotExpressionContext) {}

// EnterFunctionCallExpression is called when production FunctionCallExpression is entered.
func (s *BaseglListener) EnterFunctionCallExpression(ctx *FunctionCallExpressionContext) {}

// ExitFunctionCallExpression is called when production FunctionCallExpression is exited.
func (s *BaseglListener) ExitFunctionCallExpression(ctx *FunctionCallExpressionContext) {}

// EnterIdentifierExpression is called when production IdentifierExpression is entered.
func (s *BaseglListener) EnterIdentifierExpression(ctx *IdentifierExpressionContext) {}

// ExitIdentifierExpression is called when production IdentifierExpression is exited.
func (s *BaseglListener) ExitIdentifierExpression(ctx *IdentifierExpressionContext) {}

// EnterAssignmentExpression is called when production AssignmentExpression is entered.
func (s *BaseglListener) EnterAssignmentExpression(ctx *AssignmentExpressionContext) {}

// ExitAssignmentExpression is called when production AssignmentExpression is exited.
func (s *BaseglListener) ExitAssignmentExpression(ctx *AssignmentExpressionContext) {}

// EnterBlockClosureExpression is called when production BlockClosureExpression is entered.
func (s *BaseglListener) EnterBlockClosureExpression(ctx *BlockClosureExpressionContext) {}

// ExitBlockClosureExpression is called when production BlockClosureExpression is exited.
func (s *BaseglListener) ExitBlockClosureExpression(ctx *BlockClosureExpressionContext) {}

// EnterClosureExpression is called when production ClosureExpression is entered.
func (s *BaseglListener) EnterClosureExpression(ctx *ClosureExpressionContext) {}

// ExitClosureExpression is called when production ClosureExpression is exited.
func (s *BaseglListener) ExitClosureExpression(ctx *ClosureExpressionContext) {}

// EnterEqualityExpression is called when production EqualityExpression is entered.
func (s *BaseglListener) EnterEqualityExpression(ctx *EqualityExpressionContext) {}

// ExitEqualityExpression is called when production EqualityExpression is exited.
func (s *BaseglListener) ExitEqualityExpression(ctx *EqualityExpressionContext) {}

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
