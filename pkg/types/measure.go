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

package types

import (
	"fmt"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher"
	"github.com/cloudprivacylabs/opencypher/graph"
)

// Measure is the data type that contains a value and a unit
type Measure struct {
	Value string `json:"value" yaml:"value"`
	Unit  string `json:"unit" yaml:"unit"`
}

// MeasureTerm is used as a valueType.
//
// A measure should have a value and a unit. There are several ways
// the values and units are specified.
//
// A node may specify value and unit in its node value.
//
// Node:
//  Measure
//  value: value and unit
//
//
// A node may specify value and unit separately in its properties:
//
// Node:
//  Measure
//  value: 123
//  measureUnit: <unit>
//
// A node may be point to other nodes containing value or unit.
//
//  Node:                                  Node:
//   Measure                              value: <unit>
//   value: 123                            schemaNodeId: A
//   measureUnitNode: A
//
//  Node:                     Node:               Node:
//   Measure                  value: <unit>       value: <value>
//   measureUnitNode: A       schemaNodeId: A     schemaNodeId: B
//   measureValueNode: B
//
// A node may refer to other nodes using pattern expressions
//
//  Node:
//    Measure
//    value: 123
//    measureUnitPath: (this)<-[]-()-[]->(target : schemaNodeId:B)
//
//              Node:
//              value: <unit>
//              schemaNodeId: B
//
var MeasureTerm = ls.NewTerm(ls.LS, "Measure", false, false, ls.OverrideComposition, struct {
	measureParser
}{
	measureParser{},
})

// MeasureUnitTerm is a node property term giving measure unit
var MeasureUnitTerm = ls.NewTerm(ls.LS, "measure/unit", false, false, ls.OverrideComposition, nil)

// MeasureUnitNodeTerm gives the schema node ID of the unit node
var MeasureUnitNodeTerm = ls.NewTerm(ls.LS, "measure/unitNode", false, false, ls.OverrideComposition, nil)

// MeasureValueNodeTerm gives the schema node ID of the value node
var MeasureValueNodeTerm = ls.NewTerm(ls.LS, "measure/valueNode", false, false, ls.OverrideComposition, nil)

// MeasureUnitPathTerm gives the path to the unit node
var MeasureUnitPathTerm = ls.NewTerm(ls.LS, "measure/unitPath", false, false, ls.OverrideComposition, nil)

// MeasureValuePathTerm gives the path to the  value node
var MeasureValuePathTerm = ls.NewTerm(ls.LS, "measure/valuePath", false, false, ls.OverrideComposition, nil)

type ErrMultipleNodesMatch struct {
	Src string
}

func (e ErrMultipleNodesMatch) Error() string {
	return "Multiple nodes match: " + e.Src
}

type ErrNotAMeasure struct {
	Value string
}

func (e ErrNotAMeasure) Error() string {
	return "Not a Measure:" + e.Value
}

// Find node by schema Id without crossing entity boundaries
func findMeasureNodeBySchemaNode(node graph.Node, schemaNodeID string) (graph.Node, error) {
	var found graph.Node
	if !ls.WalkNodesInEntity(node, func(n graph.Node) bool {
		if ls.AsPropertyValue(n.GetProperty(ls.SchemaNodeIDTerm)).AsString() == schemaNodeID {
			if found != nil {
				return false
			}
			found = n
		}
		return true
	}) {
		return nil, ErrMultipleNodesMatch{schemaNodeID}
	}
	return found, nil
}

func findMeasureNodeByPath(node graph.Node, path string) (graph.Node, error) {
	p, err := opencypher.ParsePatternExpr(path)
	if err != nil {
		return nil, err
	}
	nodes, err := p.FindRelative(node)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, nil
	}
	if len(nodes) > 1 {
		return nil, ErrMultipleNodesMatch{path}
	}
	return nodes[0], nil
}

// GetNodeMeasureValue tries to load the measure and unit from the given Measure node
func GetNodeMeasureValue(node graph.Node) (*Measure, error) {
	if ls.AsPropertyValue(node.GetProperty(ls.ValueTypeTerm)).AsString() != MeasureTerm {
		return nil, nil
	}
	var err error
	ret := &Measure{}

	getNodeValue := func(found graph.Node, err error) (string, error) {
		if err != nil {
			return "", err
		}
		if found == nil {
			return "", nil
		}
		raw, ok := ls.GetRawNodeValue(found)
		if !ok {
			return "", nil
		}
		return raw, nil
	}

	// find value
	if p, ok := node.GetProperty(MeasureValueNodeTerm); ok {
		ret.Value, err = getNodeValue(findMeasureNodeBySchemaNode(node, ls.AsPropertyValue(p, ok).AsString()))
		if err != nil {
			return nil, err
		}
	} else if p, ok := node.GetProperty(MeasureValuePathTerm); ok {
		ret.Value, err = getNodeValue(findMeasureNodeByPath(node, ls.AsPropertyValue(p, ok).AsString()))
		if err != nil {
			return nil, err
		}
	} else if raw, ok := ls.GetRawNodeValue(node); ok {
		ret.Value = raw
	} else {
		return nil, nil
	}

	// find unit
	if u, ok := node.GetProperty(MeasureUnitTerm); ok {
		ret.Unit = ls.AsPropertyValue(u, ok).AsString()
	} else if u, ok := node.GetProperty(MeasureUnitNodeTerm); ok {
		ret.Unit, err = getNodeValue(findMeasureNodeBySchemaNode(node, ls.AsPropertyValue(u, ok).AsString()))
		if err != nil {
			return nil, err
		}
	} else if u, ok := node.GetProperty(MeasureUnitPathTerm); ok {
		ret.Unit, err = getNodeValue(findMeasureNodeByPath(node, ls.AsPropertyValue(u, ok).AsString()))
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

// SetNodeMeasureValue tries to set the measure and unit values based on the given Measure value
func SetNodeMeasureValue(node graph.Node, value Measure) error {
	if p, ok := node.GetProperty(MeasureValueNodeTerm); ok {
		targetNode, err := findMeasureNodeBySchemaNode(node, ls.AsPropertyValue(p, ok).AsString())
		if err != nil {
			return err
		}
		ls.SetRawNodeValue(targetNode, value.Value)
	} else if p, ok := node.GetProperty(MeasureValuePathTerm); ok {
		targetNode, err := findMeasureNodeByPath(node, ls.AsPropertyValue(p, ok).AsString())
		if err != nil {
			return err
		}
		ls.SetRawNodeValue(targetNode, value.Value)
	} else {
		ls.SetRawNodeValue(node, value.Value)
	}

	// find unit
	if _, ok := node.GetProperty(MeasureUnitTerm); ok {
		node.SetProperty(MeasureUnitTerm, value.Unit)
	} else if u, ok := node.GetProperty(MeasureUnitNodeTerm); ok {
		targetNode, err := findMeasureNodeBySchemaNode(node, ls.AsPropertyValue(u, ok).AsString())
		if err != nil {
			return err
		}
		ls.SetRawNodeValue(targetNode, value.Unit)
	} else if u, ok := node.GetProperty(MeasureUnitPathTerm); ok {
		targetNode, err := findMeasureNodeByPath(node, ls.AsPropertyValue(u, ok).AsString())
		if err != nil {
			return err
		}
		ls.SetRawNodeValue(targetNode, value.Unit)
	}
	return nil
}

type measureParser struct{}

func (measureParser) GetNodeValue(node graph.Node) (interface{}, error) {
	m, err := GetNodeMeasureValue(node)
	if m == nil {
		return nil, err
	}
	return *m, err
}

func (measureParser) SetNodeValue(node graph.Node, value interface{}) error {
	if value == nil {
		return SetNodeMeasureValue(node, Measure{})
	}
	switch t := value.(type) {
	case Measure:
		return SetNodeMeasureValue(node, t)
	case string:
	}
	return ErrNotAMeasure{Value: fmt.Sprintf("%+v %T", value, value)}
}
