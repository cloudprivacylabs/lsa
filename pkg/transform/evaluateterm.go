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

var EvaluateTerm = ls.NewTerm(TRANSFORM, "evaluate").SetComposition(ls.SetComposition).SetMetadata(EvaluateTermSemantics).SetTags(ls.SchemaElementTag).Register()

type evaluateTermSemantics struct{}

var EvaluateTermSemantics = evaluateTermSemantics{}

func (evaluateTermSemantics) Get(node ls.CompilablePropertyContainer) []string {
	return ls.AsPropertyValue(node.GetProperty(EvaluateTerm)).MustStringSlice()
}

func (evaluateTermSemantics) CompileTerm(ctx *ls.CompileContext, target ls.CompilablePropertyContainer, term string, value *ls.PropertyValue) error {
	if value == nil {
		return nil
	}
	expr := make([]opencypher.Evaluatable, 0)
	for _, str := range value.MustStringSlice() {
		e, err := ctx.CompileOpencypher(str)
		if err != nil {
			return err
		}
		expr = append(expr, e)
	}
	target.SetProperty("$compiled_"+EvaluateTerm, expr)
	return nil
}

// GetEvaluatables returns the contents of the compiled evaluate term
func (evaluateTermSemantics) GetEvaluatables(node ls.CompilablePropertyContainer) []opencypher.Evaluatable {
	v, _ := node.GetProperty("$compiled_" + EvaluateTerm)
	x, _ := v.([]opencypher.Evaluatable)
	return x
}
