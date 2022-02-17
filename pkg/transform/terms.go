// Copyright 2021 Cloud Privacy Labs, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package transform

import (
	"errors"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher"
)

var ErrInvalidConditional = errors.New("Invalid conditional")
var ErrInvalidVariables = errors.New("Invalid vars section")

const RS = ls.LS + "reshape/"

// ReshapeTerms defines the terms used to specify reshaping layers
var ReshapeTerms = struct {
	// If given, the If term specifies a predicate that should be true to reshape the node
	If string
	// Export defines a list of symbols that will be exported from the
	// opencypher expressions run in this node
	Export string
	// Expressions specify one or more expression to evaluate
	Expressions string
	// ValueExpr specifies the query to be used to generate the target value
	ValueExpr string
	// IfEmpty determines whether to reshape the node even if it has no value
	IfEmpty string
	// JoinMethod determines how to join multiple values to generate a single value
	JoinMethod string
	// JoinDelimiter specifies the join delimiter if there are multiple values to be combined
	JoinDelimiter string
}{
	If: ls.NewTerm(RS, "if", false, false, ls.OverrideComposition, struct {
		ifSemantics
	}{}),
	Export: ls.NewTerm(RS, "export", false, true, ls.OverrideComposition, struct {
		exportSemantics
	}{}),
	Expressions: ls.NewTerm(RS, "expr", false, true, ls.OverrideComposition, struct {
		exprSemantics
	}{}),
	ValueExpr: ls.NewTerm(RS, "valueExpr", false, false, ls.OverrideComposition, struct {
		valueExprSemantics
	}{}),
	IfEmpty:       ls.NewTerm(RS, "ifEmpty", false, false, ls.OverrideComposition, nil),
	JoinMethod:    ls.NewTerm(RS, "joinMethod", false, false, ls.OverrideComposition, nil),
	JoinDelimiter: ls.NewTerm(RS, "joinDelimiter", false, false, ls.OverrideComposition, nil),
}

type ifSemantics struct{}
type exportSemantics struct{}
type valueExprSemantics struct{}
type exprSemantics struct{}

// CompileTerm compiles the if conditional
func (ifSemantics) CompileTerm(target ls.CompilablePropertyContainer, term string, value *ls.PropertyValue) error {
	if value == nil {
		return nil
	}
	var val []string
	if value.IsString() {
		val = []string{value.AsString()}
	} else if value.IsStringSlice() {
		val = value.AsStringSlice()
	} else {
		return ErrInvalidConditional
	}
	out := make([]opencypher.Evaluatable, 0, len(val))
	for _, x := range val {
		r, err := opencypher.Parse(x)
		if err != nil {
			return err
		}
		out = append(out, r)
	}
	target.SetProperty("$compiled_"+term, out)
	return nil
}

// CompileTerm compiles the export list
func (exportSemantics) CompileTerm(target ls.CompilablePropertyContainer, term string, value *ls.PropertyValue) error {
	if value == nil {
		return nil
	}
	var val []string
	if value.IsString() {
		val = []string{value.AsString()}
	} else if value.IsStringSlice() {
		val = value.AsStringSlice()
	} else {
		return ErrInvalidVariables
	}
	target.SetProperty("$compiled_"+term, val)
	return nil
}

func (exprSemantics) CompileTerm(target ls.CompilablePropertyContainer, term string, value *ls.PropertyValue) error {
	if value == nil {
		return nil
	}
	expr := make([]opencypher.Evaluatable, 0)
	if value.IsString() {
		e, err := opencypher.Parse(value.AsString())
		if err != nil {
			return err
		}
		expr = append(expr, e)
	} else if value.IsStringSlice() {
		for _, x := range value.AsStringSlice() {
			e, err := opencypher.Parse(x)
			if err != nil {
				return err
			}
			expr = append(expr, e)
		}
	}
	target.SetProperty("$compiled_"+term, expr)
	return nil
}

// CompileTerm compiles the value expressions
func (valueExprSemantics) CompileTerm(target ls.CompilablePropertyContainer, term string, value *ls.PropertyValue) error {
	if value == nil {
		return nil
	}
	if !value.IsString() {
		return ErrSourceMustBeString
	}
	e, err := opencypher.Parse(value.AsString())
	if err != nil {
		return err
	}
	target.SetProperty("$compiled_"+term, e)
	return nil
}
