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
	"strings"

	"github.com/bserdar/jsonom"
	"github.com/cloudprivacylabs/lsa/pkg/gl"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

// ExportOptions are used to produce the output from the document
type ExportOptions struct {
	// BuildNodeKeyFunc builds a node key from the node
	BuildNodeKeyFunc func(graph.Node) (string, bool, error)

	// If ExportTypeProperty is set, exports "@type" properties that
	// have non-LS related types
	ExportTypeProperty bool
}

// GetBuildNodeKeyBySchemaNodeFunc returns a function that gets the
// schema node and the doc node. If the doc node does not have a
// schema node, it is not exported. The function `f` should decide
// what key to use
func GetBuildNodeKeyBySchemaNodeFunc(f func(schemaNode, docNode graph.Node) (string, bool, error)) func(graph.Node) (string, bool, error) {
	return func(node graph.Node) (string, bool, error) {
		schemaNodes := graph.TargetNodes(node.GetEdgesWithLabel(graph.OutgoingEdge, ls.InstanceOfTerm))
		if len(schemaNodes) != 1 {
			return "", false, nil
		}
		return f(schemaNodes[0].(graph.Node), node)
	}
}

// GetBuildNodeKeyExprFunc returns a function that builds node keys
// using an expression. The expression should be a closure getting a
// node argument
func GetBuildNodeKeyExprFunc(closure gl.Closure) func(graph.Node) (string, bool, error) {
	return func(node graph.Node) (string, bool, error) {
		value, err := closure.Evaluate(gl.ValueOf(node), gl.NewScope())
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

func (options ExportOptions) BuildNodeKey(node graph.Node) (string, bool, error) {
	if options.BuildNodeKeyFunc != nil {
		return options.BuildNodeKeyFunc(node)
	}
	return DefaultBuildNodeKeyFunc(node)
}

// DefaultBuildNodeKeyFunc returns the attribute name term property
// from the node if it exists. If not, it looks at the attributeName
// of the node reached by instanceOf edge. If none found it return false
func DefaultBuildNodeKeyFunc(node graph.Node) (string, bool, error) {
	v := ls.AsPropertyValue(node.GetProperty(ls.AttributeNameTerm))
	if v != nil {
		if v.IsString() {
			return v.AsString(), true, nil
		}
		if v.IsStringSlice() {
			return v.AsStringSlice()[0], true, nil
		}
	}
	found := false
	foundLabel := ""
	for _, inst := range ls.InstanceOf(node) {
		v := ls.AsPropertyValue(inst.GetProperty(ls.AttributeNameTerm))
		if v != nil {
			if found {
				return "", false, nil
			}
			found = true
			if v.IsString() {
				foundLabel = v.AsString()
			} else if v.IsStringSlice() {
				foundLabel = v.AsStringSlice()[0]
			}
		}
	}
	if found {
		return foundLabel, true, nil
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
func Export(node graph.Node, options ExportOptions) (jsonom.Node, error) {
	return exportJSON(node, options, map[graph.Node]struct{}{})
}

type ErrValueExpected struct {
	NodeID string
}

func (e ErrValueExpected) Error() string {
	return fmt.Sprintf("Value expected at %s", e.NodeID)
}

func filterTypes(types []string) []string {
	ret := make([]string, 0, len(types))
	for _, x := range types {
		if !strings.HasPrefix(x, ls.LS) {
			ret = append(ret, x)
		}
	}
	return ret
}

func exportJSON(node graph.Node, options ExportOptions, seen map[graph.Node]struct{}) (jsonom.Node, error) {
	// Loop protection
	if _, exists := seen[node]; exists {
		return nil, nil
	}
	seen[node] = struct{}{}

	nodeType := node.GetLabels()
	if !nodeType.Has(ls.DocumentNodeTerm) {
		// Not a document node
		return nil, nil
	}
	types := ls.CombineNodeTypes(ls.InstanceOf(node))

	getTypes := func() jsonom.Node {
		if !options.ExportTypeProperty {
			return nil
		}
		nodeTypes := filterTypes(types.Slice())
		if len(nodeTypes) == 0 {
			return nil
		}
		arr := jsonom.NewArray()
		for _, x := range nodeTypes {
			arr.Append(jsonom.StringValue(x))
		}
		return arr
	}

	switch {
	case types.Has(ls.AttributeTypeObject):
		ret := jsonom.NewObject()
		if t := getTypes(); t != nil {
			ret.Set("@type", t)
		}
		gnodes := graph.TargetNodes(node.GetEdgesWithLabel(graph.OutgoingEdge, ls.HasTerm))
		nodes := make([]graph.Node, 0, len(gnodes))
		for _, node := range gnodes {
			nodes = append(nodes, node.(graph.Node))
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
				ret.Set(key, value)
			}
		}
		return ret, nil

	case types.Has(ls.AttributeTypeArray):
		ret := jsonom.NewArray()
		gnodes := graph.TargetNodes(node.GetEdgesWithLabel(graph.OutgoingEdge, ls.HasTerm))
		nodes := make([]graph.Node, 0, len(gnodes))
		for _, node := range gnodes {
			nodes = append(nodes, node.(graph.Node))
		}
		ls.SortNodes(nodes)
		for _, nextNode := range nodes {
			value, err := exportJSON(nextNode, options, seen)
			if err != nil {
				return nil, err
			}
			ret.Append(value)
		}
		return ret, nil

	case types.Has(ls.AttributeTypeValue):
		value := ls.GetRawNodeValue(node)
		if value == nil {
			return nil, nil
		}
		valueStr := value.(string)
		switch {
		case types.Has(BooleanTypeTerm):
			if valueStr == "true" {
				return jsonom.BoolValue(true), nil
			}
			if valueStr == "false" {
				return jsonom.BoolValue(false), nil
			}
			return nil, ErrInvalidBooleanValue{NodeID: ls.GetNodeID(node), Value: valueStr}
		case types.Has(StringTypeTerm):
			return jsonom.StringValue(valueStr), nil
		case types.Has(NumberTypeTerm), types.Has(IntegerTypeTerm):
			return jsonom.NewValue(json.Number(valueStr)), nil
		case types.Has(ObjectTypeTerm), types.Has(ArrayTypeTerm):
			return nil, ErrValueExpected{NodeID: ls.GetNodeID(node)}
		default:
			return jsonom.NewValue(valueStr), nil
		}
	}
	return nil, nil
}
