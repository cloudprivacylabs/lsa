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

	// EnterStatements is called when entering the Statements production.
	EnterStatements(c *StatementsContext)

	// EnterStatementBlock is called when entering the StatementBlock production.
	EnterStatementBlock(c *StatementBlockContext)

	// EnterExpressionStatement is called when entering the ExpressionStatement production.
	EnterExpressionStatement(c *ExpressionStatementContext)

	// EnterParenthesizedExpression is called when entering the ParenthesizedExpression production.
	EnterParenthesizedExpression(c *ParenthesizedExpressionContext)

	// EnterLogicalAndExpression is called when entering the LogicalAndExpression production.
	EnterLogicalAndExpression(c *LogicalAndExpressionContext)

	// EnterDotExpression is called when entering the DotExpression production.
	EnterDotExpression(c *DotExpressionContext)

	// EnterAssignmentExpression is called when entering the AssignmentExpression production.
	EnterAssignmentExpression(c *AssignmentExpressionContext)

	// EnterLiteralExpression is called when entering the LiteralExpression production.
	EnterLiteralExpression(c *LiteralExpressionContext)

	// EnterLogicalOrExpression is called when entering the LogicalOrExpression production.
	EnterLogicalOrExpression(c *LogicalOrExpressionContext)

	// EnterIndexExpression is called when entering the IndexExpression production.
	EnterIndexExpression(c *IndexExpressionContext)

	// EnterNotExpression is called when entering the NotExpression production.
	EnterNotExpression(c *NotExpressionContext)

	// EnterClosureExpression is called when entering the ClosureExpression production.
	EnterClosureExpression(c *ClosureExpressionContext)

	// EnterEqualityExpression is called when entering the EqualityExpression production.
	EnterEqualityExpression(c *EqualityExpressionContext)

	// EnterFunctionCallExpression is called when entering the FunctionCallExpression production.
	EnterFunctionCallExpression(c *FunctionCallExpressionContext)

	// EnterIdentifierExpression is called when entering the IdentifierExpression production.
	EnterIdentifierExpression(c *IdentifierExpressionContext)

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

	// ExitStatements is called when exiting the Statements production.
	ExitStatements(c *StatementsContext)

	// ExitStatementBlock is called when exiting the StatementBlock production.
	ExitStatementBlock(c *StatementBlockContext)

	// ExitExpressionStatement is called when exiting the ExpressionStatement production.
	ExitExpressionStatement(c *ExpressionStatementContext)

	// ExitParenthesizedExpression is called when exiting the ParenthesizedExpression production.
	ExitParenthesizedExpression(c *ParenthesizedExpressionContext)

	// ExitLogicalAndExpression is called when exiting the LogicalAndExpression production.
	ExitLogicalAndExpression(c *LogicalAndExpressionContext)

	// ExitDotExpression is called when exiting the DotExpression production.
	ExitDotExpression(c *DotExpressionContext)

	// ExitAssignmentExpression is called when exiting the AssignmentExpression production.
	ExitAssignmentExpression(c *AssignmentExpressionContext)

	// ExitLiteralExpression is called when exiting the LiteralExpression production.
	ExitLiteralExpression(c *LiteralExpressionContext)

	// ExitLogicalOrExpression is called when exiting the LogicalOrExpression production.
	ExitLogicalOrExpression(c *LogicalOrExpressionContext)

	// ExitIndexExpression is called when exiting the IndexExpression production.
	ExitIndexExpression(c *IndexExpressionContext)

	// ExitNotExpression is called when exiting the NotExpression production.
	ExitNotExpression(c *NotExpressionContext)

	// ExitClosureExpression is called when exiting the ClosureExpression production.
	ExitClosureExpression(c *ClosureExpressionContext)

	// ExitEqualityExpression is called when exiting the EqualityExpression production.
	ExitEqualityExpression(c *EqualityExpressionContext)

	// ExitFunctionCallExpression is called when exiting the FunctionCallExpression production.
	ExitFunctionCallExpression(c *FunctionCallExpressionContext)

	// ExitIdentifierExpression is called when exiting the IdentifierExpression production.
	ExitIdentifierExpression(c *IdentifierExpressionContext)

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
