package gl

import (
	"errors"
	"fmt"
)

// Script compilation/evaluation errors
var (
	ErrEvaluation               = errors.New("Evaluation error")
	ErrNotCallable              = errors.New("Not callable")
	ErrNotIndexable             = errors.New("Not indexable")
	ErrNoEdgesInResult          = errors.New("No edges in result")
	ErrMultipleEdgesInResult    = errors.New("Multiple edges in result")
	ErrNoNodesInResult          = errors.New("No nodes in result")
	ErrMultipleNodesInResult    = errors.New("Multiple nodes in result")
	ErrNotANumber               = errors.New("Not a number")
	ErrNotAString               = errors.New("Not a string")
	ErrIncomparable             = errors.New("Incomparable values")
	ErrNotLValue                = errors.New("Not LValue")
	ErrCannotIterate            = errors.New("Cannot iterate")
	ErrCannotAccumulate         = errors.New("Cannot accumulate collection of values")
	ErrIncompatibleValue        = errors.New("Incompatible value")
	ErrNotAClosure              = errors.New("Not a closure")
	ErrClosureOrBooleanExpected = errors.New("Closure or a boolean value expected")
	ErrInvalidArgumentType      = errors.New("Invalid argument type")
	ErrInvalidStatementList     = errors.New("Invalid statement list")
	ErrInvalidExpression        = errors.New("Invalid expression")
	ErrNotExpression            = errors.New("Not an expression")
)

// ErrInvalidFunctionCall is returned if the function call cannot be completed
type ErrInvalidFunctionCall string

func (e ErrInvalidFunctionCall) Error() string {
	return fmt.Sprintf("Invalid function call: %s", string(e))
}

// ErrUnknownIdentifier is returned if an identifier is not in scope
type ErrUnknownIdentifier string

func (e ErrUnknownIdentifier) Error() string {
	return fmt.Sprintf("Unknonwn identifier '%s'", string(e))
}

// ErrSyntax is returned for generic syntax errors
type ErrSyntax string

func (e ErrSyntax) Error() string {
	return fmt.Sprintf("Syntax error: %s", string(e))
}

// ErrUnknownSelector is returned when a dot-expression uses an
// unknown selector for the type
type ErrUnknownSelector struct {
	Selector string
}

func (e ErrUnknownSelector) Error() string {
	return fmt.Sprintf("Unknown selector : %s", e.Selector)
}

// ErrInvalidArgument is returned when a method is called with an
// incompatible argument
type ErrInvalidArguments struct {
	Method string
	Msg    string
}

func (e ErrInvalidArguments) Error() string {
	return fmt.Sprintf("Invalid argument to '%s': %s", e.Method, e.Msg)
}
