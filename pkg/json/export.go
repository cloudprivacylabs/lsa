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
	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// ExportOptions are used to produce the output from the document
type ExportOptions struct {
	// BuildNodeKeyFunc builds a node key from the node
	BuildNodeKeyFunc func(*lpg.Node) (string, bool, error)

	// If ExportTypeProperty is set, exports "@type" properties that
	// have non-LS related types
	ExportTypeProperty bool
}

// GetBuildNodeKeyBySchemaNodeFunc returns a function that gets the
// schema node and the doc node. If the doc node does not have a
// schema node, it is not exported. The function `f` should decide
// what key to use
func GetBuildNodeKeyBySchemaNodeFunc(f func(schemaNode, docNode *lpg.Node) (string, bool, error)) func(*lpg.Node) (string, bool, error) {
	return func(node *lpg.Node) (string, bool, error) {
		schemaNodes := lpg.TargetNodes(node.GetEdgesWithLabel(lpg.OutgoingEdge, ls.InstanceOfTerm.Name))
		if len(schemaNodes) != 1 {
			return "", false, nil
		}
		return f(schemaNodes[0], node)
	}
}

func (options ExportOptions) BuildNodeKey(node *lpg.Node) (string, bool, error) {
	if options.BuildNodeKeyFunc != nil {
		return options.BuildNodeKeyFunc(node)
	}
	return DefaultBuildNodeKeyFunc(node)
}

// DefaultBuildNodeKeyFunc returns the attribute name term property
// from the node if it exists. If not, it looks at the attributeName
// of the node reached by instanceOf edge. If none found it return false
func DefaultBuildNodeKeyFunc(node *lpg.Node) (string, bool, error) {
	nv, _ := node.GetProperty(ls.AttributeNameTerm.Name)
	v, ok := nv.(ls.PropertyValue)
	sl := ls.StringSliceType{}
	if ok {
		s, _ := sl.Coerce(v.Value())
		if slice, _ := s.([]string); len(slice) > 0 {
			return slice[0], true, nil
		}
	}
	found := false
	foundLabel := ""
	for _, inst := range append(ls.InstanceOf(node), node) {
		nv, _ := inst.GetProperty(ls.AttributeNameTerm.Name)
		v, ok := nv.(ls.PropertyValue)
		if ok {
			if found {
				return "", false, nil
			}
			found = true
			s, _ := sl.Coerce(v)
			if slice, _ := s.([]string); len(slice) > 0 {
				foundLabel = slice[0]
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
func Export(node *lpg.Node, options ExportOptions) (jsonom.Node, error) {
	return exportJSON(node, options, map[*lpg.Node]struct{}{})
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

func exportJSON(node *lpg.Node, options ExportOptions, seen map[*lpg.Node]struct{}) (jsonom.Node, error) {
	// Loop protection
	if _, exists := seen[node]; exists {
		return nil, nil
	}
	seen[node] = struct{}{}

	nodeType := node.GetLabels()
	if !nodeType.Has(ls.DocumentNodeTerm.Name) {
		// Not a document node
		return nil, nil
	}
	types := node.GetLabels()

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
	case types.Has(ls.AttributeTypeObject.Name):
		ret := jsonom.NewObject()
		if t := getTypes(); t != nil {
			ret.Set("@type", t)
		}
		gnodes := lpg.TargetNodes(node.GetEdgesWithLabel(lpg.OutgoingEdge, ls.HasTerm.Name))
		nodes := make([]*lpg.Node, 0, len(gnodes))
		for _, node := range gnodes {
			nodes = append(nodes, node)
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

	case types.Has(ls.AttributeTypeArray.Name):
		ret := jsonom.NewArray()
		gnodes := lpg.TargetNodes(node.GetEdgesWithLabel(lpg.OutgoingEdge, ls.HasTerm.Name))
		nodes := make([]*lpg.Node, 0, len(gnodes))
		for _, node := range gnodes {
			nodes = append(nodes, node)
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

	case types.Has(ls.AttributeTypeValue.Name):
		nativeValue, err := ls.GetNodeValue(node)
		if err != nil {
			return nil, err
		}
		if nativeValue == nil {
			return jsonom.NewValue(nil), nil
		}
		switch nativeValue.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool, string, json.Number:
			return jsonom.NewValue(nativeValue), nil
		}
		raw, ok := ls.GetRawNodeValue(node)
		if !ok {
			return nil, nil
		}
		return jsonom.NewValue(raw), nil
	}
	return nil, nil
}
