// Code generated from gl.g4 by ANTLR 4.9. DO NOT EDIT.

package parser // gl

import "github.com/antlr/antlr4/runtime/Go/antlr"

// glListener is a complete listener for a parse tree produced by glParser.
type glListener interface {
	antlr.ParseTreeListener

	// EnterExpressionScript is called when entering the ExpressionScript production.
	EnterExpressionScript(c *ExpressionScriptContext)

	// EnterStatementListScript is called when entering the StatementListScript production.
	EnterStatementListScript(c *StatementListScriptContext)

	// EnterStatement is called when entering the statement production.
	EnterStatement(c *StatementContext)

	// EnterStatementList is called when entering the statementList production.
	EnterStatementList(c *StatementListContext)

	// EnterStatementBlock is called when entering the statementBlock production.
	EnterStatementBlock(c *StatementBlockContext)

	// EnterExpressionStatement is called when entering the expressionStatement production.
	EnterExpressionStatement(c *ExpressionStatementContext)

	// EnterParenthesizedExpression is called when entering the ParenthesizedExpression production.
	EnterParenthesizedExpression(c *ParenthesizedExpressionContext)

	// EnterLogicalAndExpression is called when entering the LogicalAndExpression production.
	EnterLogicalAndExpression(c *LogicalAndExpressionContext)

	// EnterDotExpression is called when entering the DotExpression production.
	EnterDotExpression(c *DotExpressionContext)

	// EnterLiteralExpression is called when entering the LiteralExpression production.
	EnterLiteralExpression(c *LiteralExpressionContext)

	// EnterLogicalOrExpression is called when entering the LogicalOrExpression production.
	EnterLogicalOrExpression(c *LogicalOrExpressionContext)

	// EnterDefinitionExpression is called when entering the DefinitionExpression production.
	EnterDefinitionExpression(c *DefinitionExpressionContext)

	// EnterIndexExpression is called when entering the IndexExpression production.
	EnterIndexExpression(c *IndexExpressionContext)

	// EnterNotExpression is called when entering the NotExpression production.
	EnterNotExpression(c *NotExpressionContext)

	// EnterFunctionCallExpression is called when entering the FunctionCallExpression production.
	EnterFunctionCallExpression(c *FunctionCallExpressionContext)

	// EnterIdentifierExpression is called when entering the IdentifierExpression production.
	EnterIdentifierExpression(c *IdentifierExpressionContext)

	// EnterAssignmentExpression is called when entering the AssignmentExpression production.
	EnterAssignmentExpression(c *AssignmentExpressionContext)

	// EnterBlockClosureExpression is called when entering the BlockClosureExpression production.
	EnterBlockClosureExpression(c *BlockClosureExpressionContext)

	// EnterClosureExpression is called when entering the ClosureExpression production.
	EnterClosureExpression(c *ClosureExpressionContext)

	// EnterEqualityExpression is called when entering the EqualityExpression production.
	EnterEqualityExpression(c *EqualityExpressionContext)

	// EnterLvalue is called when entering the lvalue production.
	EnterLvalue(c *LvalueContext)

	// EnterArguments is called when entering the arguments production.
	EnterArguments(c *ArgumentsContext)

	// EnterArgumentList is called when entering the argumentList production.
	EnterArgumentList(c *ArgumentListContext)

	// EnterLiteral is called when entering the literal production.
	EnterLiteral(c *LiteralContext)

	// EnterNumericLiteral is called when entering the numericLiteral production.
	EnterNumericLiteral(c *NumericLiteralContext)

	// EnterIdentifierName is called when entering the identifierName production.
	EnterIdentifierName(c *IdentifierNameContext)

	// ExitExpressionScript is called when exiting the ExpressionScript production.
	ExitExpressionScript(c *ExpressionScriptContext)

	// ExitStatementListScript is called when exiting the StatementListScript production.
	ExitStatementListScript(c *StatementListScriptContext)

	// ExitStatement is called when exiting the statement production.
	ExitStatement(c *StatementContext)

	// ExitStatementList is called when exiting the statementList production.
	ExitStatementList(c *StatementListContext)

	// ExitStatementBlock is called when exiting the statementBlock production.
	ExitStatementBlock(c *StatementBlockContext)

	// ExitExpressionStatement is called when exiting the expressionStatement production.
	ExitExpressionStatement(c *ExpressionStatementContext)

	// ExitParenthesizedExpression is called when exiting the ParenthesizedExpression production.
	ExitParenthesizedExpression(c *ParenthesizedExpressionContext)

	// ExitLogicalAndExpression is called when exiting the LogicalAndExpression production.
	ExitLogicalAndExpression(c *LogicalAndExpressionContext)

	// ExitDotExpression is called when exiting the DotExpression production.
	ExitDotExpression(c *DotExpressionContext)

	// ExitLiteralExpression is called when exiting the LiteralExpression production.
	ExitLiteralExpression(c *LiteralExpressionContext)

	// ExitLogicalOrExpression is called when exiting the LogicalOrExpression production.
	ExitLogicalOrExpression(c *LogicalOrExpressionContext)

	// ExitDefinitionExpression is called when exiting the DefinitionExpression production.
	ExitDefinitionExpression(c *DefinitionExpressionContext)

	// ExitIndexExpression is called when exiting the IndexExpression production.
	ExitIndexExpression(c *IndexExpressionContext)

	// ExitNotExpression is called when exiting the NotExpression production.
	ExitNotExpression(c *NotExpressionContext)

	// ExitFunctionCallExpression is called when exiting the FunctionCallExpression production.
	ExitFunctionCallExpression(c *FunctionCallExpressionContext)

	// ExitIdentifierExpression is called when exiting the IdentifierExpression production.
	ExitIdentifierExpression(c *IdentifierExpressionContext)

	// ExitAssignmentExpression is called when exiting the AssignmentExpression production.
	ExitAssignmentExpression(c *AssignmentExpressionContext)

	// ExitBlockClosureExpression is called when exiting the BlockClosureExpression production.
	ExitBlockClosureExpression(c *BlockClosureExpressionContext)

	// ExitClosureExpression is called when exiting the ClosureExpression production.
	ExitClosureExpression(c *ClosureExpressionContext)

	// ExitEqualityExpression is called when exiting the EqualityExpression production.
	ExitEqualityExpression(c *EqualityExpressionContext)

	// ExitLvalue is called when exiting the lvalue production.
	ExitLvalue(c *LvalueContext)

	// ExitArguments is called when exiting the arguments production.
	ExitArguments(c *ArgumentsContext)

	// ExitArgumentList is called when exiting the argumentList production.
	ExitArgumentList(c *ArgumentListContext)

	// ExitLiteral is called when exiting the literal production.
	ExitLiteral(c *LiteralContext)

	// ExitNumericLiteral is called when exiting the numericLiteral production.
	ExitNumericLiteral(c *NumericLiteralContext)

	// ExitIdentifierName is called when exiting the identifierName production.
	ExitIdentifierName(c *IdentifierNameContext)
}
