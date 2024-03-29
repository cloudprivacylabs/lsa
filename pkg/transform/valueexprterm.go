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
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher"
)

// ValueExprTerm defines one or more opencypher expressions that
// defines the value of the node. The first one that returns nonempty
// resultset will be evaluated
var ValueExprTerm = ls.RegisterStringSliceTerm(ls.NewTerm(TRANSFORM, "valueExpr").SetComposition(ls.OverrideComposition).SetMetadata(ValueExprTermSemantics).SetTags(ls.SchemaElementTag))
var ValueExprFirstTerm = ls.RegisterStringSliceTerm(ls.NewTerm(TRANSFORM, "valueExpr.first").SetComposition(ls.OverrideComposition).SetMetadata(ValueExprTermSemantics).SetTags(ls.SchemaElementTag))
var ValueExprAllTerm = ls.RegisterStringSliceTerm(ls.NewTerm(TRANSFORM, "valueExpr.all").SetComposition(ls.OverrideComposition).SetMetadata(ValueExprTermSemantics).SetTags(ls.SchemaElementTag))

type valueExprTermSemantics struct{}

var ValueExprTermSemantics = valueExprTermSemantics{}

func (valueExprTermSemantics) Get(node ls.CompilablePropertyContainer) []string {
	s := ValueExprAllTerm.PropertyValue(node)
	if len(s) > 0 {
		return s
	}
	s = ValueExprFirstTerm.PropertyValue(node)
	if len(s) > 0 {
		return s
	}
	s = ValueExprTerm.PropertyValue(node)
	return s
}

func (valueExprTermSemantics) CompileTerm(ctx *ls.CompileContext, target ls.CompilablePropertyContainer, term string, value ls.PropertyValue) error {
	if value.Value() == nil {
		return nil
	}
	expr := make([]opencypher.Evaluatable, 0)
	for _, str := range value.AsStringSlice() {
		e, err := ctx.CompileOpencypher(str)
		if err != nil {
			return err
		}
		expr = append(expr, e)
	}
	target.SetProperty("$compiled_"+term, expr)
	return nil
}

// GetEvaluatables returns the contents of the compiled valueExpr terms
func (valueExprTermSemantics) GetEvaluatables(term string, node ls.CompilablePropertyContainer) []opencypher.Evaluatable {
	v, _ := node.GetProperty("$compiled_" + term)
	x, _ := v.([]opencypher.Evaluatable)
	return x
}
