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
	"github.com/cloudprivacylabs/opencypher/graph"
)

var ValueExprTerm = ls.NewTerm(TRANSFORM, "valueExpr", false, false, ls.OverrideComposition, ValueExprTermSemantics)

type valueExprTermSemantics struct{}

var ValueExprTermSemantics = valueExprTermSemantics{}

func (valueExprTermSemantics) CompileTerm(target ls.CompilablePropertyContainer, term string, value *ls.PropertyValue) error {
	e, err := opencypher.Parse(value.AsString())
	if err != nil {
		return err
	}
	target.SetProperty("$compiled_"+ValueExprTerm, e)
	return nil
}

// GetEvaluatable returns the contents of the compiled valueExpr term
func (valueExprTermSemantics) GetEvaluatable(node graph.Node) opencypher.Evaluatable {
	v, _ := node.GetProperty("$compiled_" + ValueExprTerm)
	x, _ := v.(opencypher.Evaluatable)
	return x
}

func (valueExprTermSemantics) Evaluate(node graph.Node, ctx *opencypher.EvalContext) (bool, opencypher.Value, error) {
	ev := ValueExprTermSemantics.GetEvaluatable(node)
	if ev == nil {
		return false, opencypher.Value{}, nil
	}
	v, err := ev.Evaluate(ctx)
	return true, v, err
}
