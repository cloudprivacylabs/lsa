package gl

import (
	"errors"
	"fmt"
)

var ErrEvaluation = errors.New("Evaluation error")
var ErrNotCallable = errors.New("Not callable")
var ErrNotIndexable = errors.New("Not indexable")
var ErrNoEdgesInResult = errors.New("No edges in result")
var ErrMultipleEdgesInResult = errors.New("Multiple edges in result")
var ErrNoNodesInResult = errors.New("No nodes in result")
var ErrMultipleNodesInResult = errors.New("Multiple nodes in result")
var ErrNotANumber = errors.New("Not a number")
var ErrNotAString = errors.New("Not a string")
var ErrIncomparable = errors.New("Incomparable values")
var ErrNotLValue = errors.New("Not LValue")
var ErrCannotIterate = errors.New("Cannot iterate")
var ErrCannotAccumulate = errors.New("Cannot accumulate collection of values")
var ErrIncompatibleValue = errors.New("Incompatible value")

type ErrInvalidFunctionCall string

func (e ErrInvalidFunctionCall) Error() string {
	return fmt.Sprintf("Invalid function call: %s", string(e))
}

type ErrUnknownIdentifier string

func (e ErrUnknownIdentifier) Error() string {
	return fmt.Sprintf("Unknonwn identifier '%s'", string(e))
}

type ErrSyntax string

func (e ErrSyntax) Error() string {
	return fmt.Sprintf("Syntax error: %s", string(e))
}

type ErrUnknownSelector struct {
	Selector string
}

func (e ErrUnknownSelector) Error() string {
	return fmt.Sprintf("Unknown selector : %s", e.Selector)
}

type ErrInvalidArguments struct {
	Method string
	Msg    string
}

func (e ErrInvalidArguments) Error() string {
	return fmt.Sprintf("Invalid argument to '%s': %s", e.Method, e.Msg)
}
