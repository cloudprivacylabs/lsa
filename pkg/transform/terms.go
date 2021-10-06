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

	"github.com/cloudprivacylabs/lsa/pkg/gl"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

var ErrInvalidConditional = errors.New("Invalid conditional")
var ErrInvalidVariables = errors.New("Invalid vars section")

const RS = ls.LS + "reshape/"

// ReshapeTerms defines the terms used to specify reshaping layers
var ReshapeTerms = struct {
	// If given, the If term specifies a predicate that should be true to reshape the node
	If string
	// Vars defines a list of expressions that pull values from the
	// source graph and define them as variables
	Vars string
	// Source specifies the source value to be used to generate the target value
	Source string
	// IfEmpty determines whether to reshape the node even if it has no value
	IfEmpty string
	// JoinMethod determines how to join multiple values to generate a single value
	JoinMethod string
	// JoinDelimiter specifies the join delimiter if there are multiple values to be combined
	JoinDelimiter string
}{
	If: ls.NewTerm(RS+"if", false, false, ls.OverrideComposition, struct {
		ifSemantics
	}{}),
	Vars: ls.NewTerm(RS+"vars", false, true, ls.OverrideComposition, struct {
		varsSemantics
	}{}),
	Source: ls.NewTerm(RS+"source", false, false, ls.OverrideComposition, struct {
		sourceSemantics
	}{}),
	IfEmpty:       ls.NewTerm(RS+"ifEmpty", false, false, ls.OverrideComposition, nil),
	JoinMethod:    ls.NewTerm(RS+"joinMethod", false, false, ls.OverrideComposition, nil),
	JoinDelimiter: ls.NewTerm(RS+"joinDelimiter", false, false, ls.OverrideComposition, nil),
}

type ifSemantics struct{}
type varsSemantics struct{}
type sourceSemantics struct{}

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
	out := make([]gl.Evaluatable, 0, len(val))
	for _, x := range val {
		r, err := gl.Parse(x)
		if err != nil {
			return err
		}
		out = append(out, r)
	}
	target.GetCompiledDataMap()[term] = out
	return nil
}

// CompileTerm compiles the variable expressions
func (varsSemantics) CompileTerm(target ls.CompilablePropertyContainer, term string, value *ls.PropertyValue) error {
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
	out := make([]gl.Evaluatable, 0, len(val))
	for _, x := range val {
		r, err := gl.Parse(x)
		if err != nil {
			return err
		}
		out = append(out, r)
	}
	target.GetCompiledDataMap()[term] = out
	return nil
}

// CompileTerm compiles the source expression
func (sourceSemantics) CompileTerm(target ls.CompilablePropertyContainer, term string, value *ls.PropertyValue) error {
	if value == nil {
		return nil
	}
	if !value.IsString() {
		return ErrSourceMustBeString
	}
	e, err := gl.Parse(value.AsString())
	if err != nil {
		return err
	}
	target.GetCompiledDataMap()[term] = e
	return nil
}
