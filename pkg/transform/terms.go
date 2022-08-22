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

// MapPropertyTerm defines the name of the property in the source
// graph nodes that contain the mapped schema node id. The contents of
// the nodes under the mapContext that have prop:schemaNodeId will be
// assigned to the current node
var MapPropertyTerm = ls.NewTerm(TRANSFORM, "mapProperty", false, false, ls.OverrideComposition, nil)

// MapContextTerm gives an opencypher expression that results in a
// node. That node will be used as the context for the map operations
// under that node
var MapContextTerm = ls.NewTerm(TRANSFORM, "mapContext", false, false, ls.OverrideComposition, MapContextSemantics)

var MapContextSemantics = mapContextSemantics{}

type mapContextSemantics struct{}

func (mapContextSemantics) CompileTerm(target ls.CompilablePropertyContainer, term string, value *ls.PropertyValue) error {
	e, err := opencypher.Parse(value.AsString())
	if err != nil {
		return err
	}
	target.SetProperty("$compiled_"+MapContextTerm, e)
	return nil
}

// GetEvaluatable returns the contents of the compiled mapContext term
func (mapContextSemantics) GetEvaluatable(node ls.CompilablePropertyContainer) opencypher.Evaluatable {
	v, _ := node.GetProperty("$compiled_" + MapContextTerm)
	x, _ := v.(opencypher.Evaluatable)
	return x
}

func (mapContextSemantics) Evaluate(node ls.CompilablePropertyContainer, ctx *opencypher.EvalContext) (bool, opencypher.Value, error) {
	ev := MapContextSemantics.GetEvaluatable(node)
	if ev == nil {
		return false, opencypher.RValue{}, nil
	}
	v, err := ev.Evaluate(ctx)
	return true, v, err
}
