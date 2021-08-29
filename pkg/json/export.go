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
	"io"

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
		schemaNodes := node.GetAllOutgoingEdgesWithLabel(ls.InstanceOfTerm).Targets().All()
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

type Encodable interface {
	Encode(io.Writer) error
}

// ExportKeyValue is a JSON key-value pair
type ExportKeyValue struct {
	Key   string
	Value Encodable
}

// ExportValue is a JSON value
type ExportValue struct {
	Value json.RawMessage
}

// Encode a value
func (e ExportValue) Encode(w io.Writer) error {
	_, err := w.Write(e.Value)
	return err
}

// Encode a key-value pair
func (e ExportKeyValue) Encode(w io.Writer) error {
	data, err := json.Marshal(e.Key)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	if _, err := w.Write([]byte{':'}); err != nil {
		return err
	}
	return e.Value.Encode(w)
}

// ExportObject represents a JSON object
type ExportObject struct {
	Values []ExportKeyValue
}

// Encode a json object
func (e ExportObject) Encode(w io.Writer) error {
	if _, err := w.Write([]byte{'{'}); err != nil {
		return err
	}
	for i, v := range e.Values {
		if i > 0 {
			if _, err := w.Write([]byte{','}); err != nil {
				return err
			}
		}
		if err := v.Encode(w); err != nil {
			return err
		}
	}
	if _, err := w.Write([]byte{'}'}); err != nil {
		return err
	}
	return nil
}

// ExportArray represents a JSON array
type ExportArray struct {
	Elements []Encodable
}

func (e ExportArray) Encode(w io.Writer) error {
	if _, err := w.Write([]byte{'['}); err != nil {
		return err
	}
	for i, value := range e.Elements {
		if i > 0 {
			if _, err := w.Write([]byte{','}); err != nil {
				return err
			}
		}
		if err := value.Encode(w); err != nil {
			return err
		}
	}
	if _, err := w.Write([]byte{']'}); err != nil {
		return err
	}
	return nil
}

type ErrInvalidBooleanValue struct {
	NodeID string
	Value  string
}

func (e ErrInvalidBooleanValue) Error() string {
	return fmt.Sprintf("Invalid boolean value at %s: %s", e.NodeID, e.Value)
}

// Export the document subtree to the target. The returned result is
// Encodable, which respects element ordering
func Export(node ls.Node, options ExportOptions) (Encodable, error) {
	return exportJSON(node, options, map[ls.Node]struct{}{})
}

type ErrValueExpected struct {
	NodeID string
}

func (e ErrValueExpected) Error() string {
	return fmt.Sprintf("Value expected at %s", e.NodeID)
}

func exportJSON(node ls.Node, options ExportOptions, seen map[ls.Node]struct{}) (Encodable, error) {
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
		ret := ExportObject{}
		gnodes := node.GetAllOutgoingEdgesWithLabel(ls.DataEdgeTerms.ObjectAttributes).Targets().All()
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
				ret.Values = append(ret.Values, ExportKeyValue{Key: key, Value: value})
			}
		}
		return ret, nil

	case types.Has(ls.AttributeTypes.Array):
		ret := ExportArray{}
		gnodes := node.GetAllOutgoingEdgesWithLabel(ls.DataEdgeTerms.ArrayElements).Targets().All()
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
				return ExportValue{Value: []byte("true")}, nil
			}
			if valueStr == "false" {
				return ExportValue{Value: []byte("false")}, nil
			}
			return nil, ErrInvalidBooleanValue{NodeID: node.GetID(), Value: valueStr}
		case types.Has(StringTypeTerm):
			data, _ := json.Marshal(valueStr)
			return ExportValue{Value: data}, nil
		case types.Has(NumberTypeTerm), types.Has(IntegerTypeTerm):
			data, _ := json.Marshal(json.Number([]byte(valueStr)))
			return ExportValue{Value: data}, nil
		case types.Has(ObjectTypeTerm), types.Has(ArrayTypeTerm):
			return nil, ErrValueExpected{NodeID: node.GetID()}
		default:
			data, _ := json.Marshal(valueStr)
			return ExportValue{Value: data}, nil
		}
	}
	return nil, nil
}
