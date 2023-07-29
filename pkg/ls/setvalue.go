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

package ls

import (
	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/opencypher"
)

var SetValueTerm = NewTerm(LS, "setValue").SetComposition(OverrideComposition).SetMetadata(setValueSemantics{}).SetTags(SchemaElementTag).Register()

type setValueSemantics struct{}

// CompileTerm will compile the expression in setValue term.
func (setValueSemantics) CompileTerm(ctx *CompileContext, node CompilablePropertyContainer, term string, value PropertyValue) error {
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
	node.SetProperty("$compiled_"+term, expr)
	return nil
}

// ProcessNodePostDocIngest will evaluate the opencypher expressions
// given in the term for the docNode and set the value of the docnode
// based on that
func (setValueSemantics) ProcessNodePostDocIngest(schemaRootNode, schemaNode *lpg.Node, term PropertyValue, docNode *lpg.Node) error {
	v, _ := docNode.GetProperty("$compiled_" + term.Sem().Name)
	exprs, _ := v.([]opencypher.Evaluatable)
	evalContext := NewEvalContext(docNode.GetGraph())
	evalContext.SetVar("this", opencypher.ValueOf(docNode))
	var lastResult any
	for _, expr := range exprs {
		result, err := expr.Evaluate(evalContext)
		if err != nil {
			return err
		}
		lastResult = result.Get()
		resultSet, ok := lastResult.(opencypher.ResultSet)
		if ok {
			if len(resultSet.Rows) == 1 {
				for k, v := range resultSet.Rows[0] {
					if opencypher.IsNamedResult(k) {
						evalContext.SetVar(k, v)
					}
				}
			}
		}
	}
	if rs, ok := lastResult.(opencypher.ResultSet); ok {
		if len(rs.Rows) == 1 {
			if len(rs.Rows[0]) == 1 {
				for _, v := range rs.Rows[0] {
					if err := SetNodeValue(docNode, v.Get()); err != nil {
						return err
					}
					rawValue, _ := GetRawNodeValue(docNode)
					SetEntityIDVectorElementFromNode(docNode, rawValue)
				}
			}
		}
	}
	return nil
}
