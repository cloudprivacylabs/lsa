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
	"encoding/json"

	"github.com/bserdar/digraph"
)

type Expression interface {
	EvaluateExpression(*digraph.Graph) (interface{}, error)
}

type NullExpression struct{}

func (NullExpression) EvaluateExpression(*digraph.Graph) (interface{}, error) { return nil, nil }

// SelectNodesExpression selects graph nodes based on a predicate
type SelectNodesExpression struct {
	Predicate NodePredicate `json:"$selectNodes"`
}

// EvaluateExpression evaluates the predicate for all nodes, and returns []digraph.Node
func (expr SelectNodesExpression) EvaluateExpression(g *digraph.Graph) (interface{}, error) {
	ret := make([]digraph.Node, 0)
	for nodes := g.AllNodes(); nodes.HasNext(); {
		node := nodes.Next().(Node)
		result, err := expr.Predicate.EvaluateNode(node)
		if err != nil {
			return nil, err
		}
		if result {
			ret = append(ret, node)
		}
	}
	return ret, nil
}

func unmarshalSelectNodesExpression(in json.RawMessage) (Expression, error) {
	np, err := UnmarshalNodePredicate(in)
	if err != nil {
		return nil, err
	}
	return SelectNodesExpression{Predicate: np}, nil
}

var expressionFactories = map[string]func(json.RawMessage) (Expression, error){
	"$selectNodes": unmarshalSelectNodesExpression,
}

func UnmarshalExpression(in []byte) (Expression, error) {
	var input map[string]json.RawMessage
	if err := json.Unmarshal(in, &input); err != nil {
		return nil, err
	}
	if len(input) == 0 {
		return NullExpression{}, nil
	}
	if len(input) > 1 {
		return nil, ErrInvalidExpression
	}
	for k, v := range input {
		factory, ok := expressionFactories[k]
		if !ok {
			return nil, ErrUnknownOperator{k}
		}
		value, err := factory(v)
		if err != nil {
			return nil, err
		}
		return value, nil
	}
	return nil, nil
}
