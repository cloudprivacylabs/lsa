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

package json

import (
	"encoding/json"
	"fmt"

	"github.com/cloudprivacylabs/lsa/pkg/gl"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// ExportOptions are used to produce the output from the document
type ExportOptions struct {
	// BuildNodeKeyFunc builds a node key from the node
	BuildNodeKeyFunc func(ls.Node) (string, bool, error)
}

// GetBuildNodeKeyBySchemaNodeFunc returns a function that gets the
// schema node and the doc node. If the doc node does not have a
// schema node, it is not exported. The function `f` should decide
// what key to use
func GetBuildNodeKeyBySchemaNodeFunc(f func(schemaNode, docNode ls.Node) (string, bool, error)) func(ls.Node) (string, bool, error) {
	return func(node ls.Node) (string, bool, error) {
		schemaNodes := node.OutWith(ls.InstanceOfTerm).Targets().All()
		if len(schemaNodes) != 1 {
			return "", false, nil
		}
		return f(schemaNodes[0].(ls.Node), node)
	}
}

// GetBuildNodeKeyExprFunc returns a function that builds node keys
// using an expression. The expression should be a closure getting a
// node argument
func GetBuildNodeKeyExprFunc(closure gl.Closure) func(ls.Node) (string, bool, error) {
	return func(node ls.Node) (string, bool, error) {
		value, err := closure.Evaluate(gl.ValueOf(node), gl.NewContext())
		if err != nil {
			return "", false, err
		}
		// Value must be a string
		str, err := value.AsString()
		if err != nil {
			return "", false, err
		}
		if len(str) == 0 {
			return "", false, nil
		}
		return str, true, nil
	}
}

func (options ExportOptions) BuildNodeKey(node ls.Node) (string, bool, error) {
	if options.BuildNodeKeyFunc != nil {
		return options.BuildNodeKeyFunc(node)
	}
	return DefaultBuildNodeKeyFunc(node)
}

// DefaultBuildNodeKeyFunc returns the attribute name term propertu
// from the node if it exists. If not, it return false
func DefaultBuildNodeKeyFunc(node ls.Node) (string, bool, error) {
	v := node.GetProperties()[ls.AttributeNameTerm]
	if v == nil {
		return "", false, nil
	}
	if v.IsString() {
		return v.AsString(), true, nil
	}
	return "", false, nil
}

type ErrInvalidBooleanValue struct {
	NodeID string
	Value  string
}

func (e ErrInvalidBooleanValue) Error() string {
	return fmt.Sprintf("Invalid boolean value at %s: %s", e.NodeID, e.Value)
}

// Export the document subtree to the target. The returned result is
// OM, which respects element ordering
func Export(node ls.Node, options ExportOptions) (OM, error) {
	return exportJSON(node, options, map[ls.Node]struct{}{})
}

type ErrValueExpected struct {
	NodeID string
}

func (e ErrValueExpected) Error() string {
	return fmt.Sprintf("Value expected at %s", e.NodeID)
}

func exportJSON(node ls.Node, options ExportOptions, seen map[ls.Node]struct{}) (OM, error) {
	// Loop protection
	if _, exists := seen[node]; exists {
		return nil, nil
	}
	seen[node] = struct{}{}

	nodeType := node.GetTypes()
	if !nodeType.Has(ls.DocumentNodeTerm) {
		// Not a document node
		return nil, nil
	}
	types := ls.CombineNodeTypes(ls.InstanceOf(node))
	switch {
	case types.Has(ls.AttributeTypes.Object):
		ret := Object{}
		gnodes := node.OutWith(ls.DataEdgeTerms.ObjectAttributes).Targets().All()
		nodes := make([]ls.Node, 0, len(gnodes))
		for _, node := range gnodes {
			nodes = append(nodes, node.(ls.Node))
		}
		ls.SortNodes(nodes)
		for _, nextNode := range nodes {
			key, ok, err := options.BuildNodeKey(nextNode)
			if err != nil {
				return nil, err
			}
			if ok {
				value, err := exportJSON(nextNode, options, seen)
				if err != nil {
					return nil, err
				}
				ret.Values = append(ret.Values, KeyValue{Key: key, Value: value})
			}
		}
		return ret, nil

	case types.Has(ls.AttributeTypes.Array):
		ret := Array{}
		gnodes := node.OutWith(ls.DataEdgeTerms.ArrayElements).Targets().All()
		nodes := make([]ls.Node, 0, len(gnodes))
		for _, node := range gnodes {
			nodes = append(nodes, node.(ls.Node))
		}
		ls.SortNodes(nodes)
		for _, nextNode := range nodes {
			value, err := exportJSON(nextNode, options, seen)
			if err != nil {
				return nil, err
			}
			ret.Elements = append(ret.Elements, value)
		}
		return ret, nil

	case types.Has(ls.AttributeTypes.Value):
		value := node.GetValue()
		if value == nil {
			return nil, nil
		}
		valueStr := value.(string)
		switch {
		case types.Has(BooleanTypeTerm):
			if valueStr == "true" {
				return Value{Value: true}, nil
			}
			if valueStr == "false" {
				return Value{Value: false}, nil
			}
			return nil, ErrInvalidBooleanValue{NodeID: node.GetID(), Value: valueStr}
		case types.Has(StringTypeTerm):
			return Value{Value: valueStr}, nil
		case types.Has(NumberTypeTerm), types.Has(IntegerTypeTerm):
			return Value{Value: json.Number(valueStr)}, nil
		case types.Has(ObjectTypeTerm), types.Has(ArrayTypeTerm):
			return nil, ErrValueExpected{NodeID: node.GetID()}
		default:
			return Value{Value: valueStr}, nil
		}
	}
	return nil, nil
}
