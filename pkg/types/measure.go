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
	"regexp"
	"strings"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher"
	"github.com/cloudprivacylabs/opencypher/graph"
)

// Measure is the data type that contains a value and a unit
type Measure struct {
	Value string `json:"value" yaml:"value"`
	Unit  string `json:"unit" yaml:"unit"`
}

func (m Measure) String() string {
	return strings.TrimSpace(m.Value + " " + m.Unit)
}

var unitRegexp = regexp.MustCompile(`^([+\-]?(?:(?:0|[1-9]\d*)(?:\.\d*)?|\.\d+)(?:\d[eE][+\-]?\d+)?)(?:(?:\s+(\S+.*))|([^\seE\d]+.*))$`)

// ParseMeasure parses a number and then a string for units
func ParseMeasure(in string) (Measure, error) {
	values := make([]string, 0, 2)
	for _, v := range unitRegexp.FindAllStringSubmatch("0.1e123K", -1) {
		for _, x := range v {
			x := strings.TrimSpace(x)
			if len(x) > 0 {
				values = append(values, x)
			}
		}
	}
	if len(values) != 2 {
		return Measure{}, ErrNotAMeasure{Value: in}
	}
	return Measure{
		Value: values[0],
		Unit:  values[1],
	}, nil
}

// When data elements are ingested, measures may appear in several forms:
//
//   * Node containing a string that has both the value and unit
//   * Node containing value and unit in properties
//   * One node with value, another with unit
//
// When data elements are captured this way, a Measure node can be
// contructed that has the normalized measure. This node has the
// Measure as its value type, and contains value and unit properties.
// A measure is a Value node, and contains the combined normalized
// unit as its node value. The measure node is a sibling of the value
// node.
//
// Example:
//
// Input: 5'4"
// Measure node:
//   measure/value: 64
//   measure/unit: [i_inch]
//   value: 64 [i_inch]
//

// MeasureTerm is used as a valuetype for a measure node
var MeasureTerm = ls.NewTerm(ls.LS, "Measure", false, false, ls.OverrideComposition, struct {
	measureParser
}{
	measureParser{},
}, "Measure")

// MeasureUnitTerm is a node property term giving measure unit
var MeasureUnitTerm = ls.NewTerm(ls.LS, "measure/unit", false, false, ls.OverrideComposition, nil)

// MeasureValueTerm is a node property term giving measure value.
var MeasureValueTerm = ls.NewTerm(ls.LS, "measure/value", false, false, ls.OverrideComposition, nil)

// MeasureUnitExpr gives the expression that returns the unit. The
// result can be a node or a value. The expression is evaluated with
// (valueNode) bound to the value node of the unit expr
var MeasureUnitExpr = ls.NewTerm(ls.LS, "measure/unitExpr", false, false, ls.OverrideComposition, ls.CompileOCSemantics{})

// MeasureValueExpr gives the expression that returns the measured value node. The result must be a node.
var MeasureValueNodeExpr = ls.NewTerm(ls.LS, "measure/valueNodeExpr", false, false, ls.OverrideComposition, ls.CompileOCSemantics{})

// MeasureUnitNode gives the schema node id containing the unit. This
// node must appear under the common parent with measure node
var MeasureUnitNode = ls.NewTerm(ls.LS, "measure/unitNode", false, false, ls.OverrideComposition, nil)

// MeasureValueNode gives the schema node id containing the value. This
// node must appear under the common parent with measure node
var MeasureValueNode = ls.NewTerm(ls.LS, "measure/valueNode", false, false, ls.OverrideComposition, nil)

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

type ErrMeasureProcessing struct {
	Msg string
	ID  string
}

func (e ErrMeasureProcessing) Error() string {
	return "Measure processing error for :" + e.ID + ":" + e.Msg
}

func getMeasureValueNodes(ctx *ls.Context, g graph.Graph, measureSchemaNode graph.Node) ([]graph.Node, error) {
	valueNodes := make([]graph.Node, 0)
	evalCtx := opencypher.NewEvalContext(g)
	results, err := ls.CompileOCSemantics{}.Evaluate(measureSchemaNode, MeasureValueNodeExpr, evalCtx)
	if err != nil {
		return nil, ErrMeasureProcessing{
			ID:  ls.GetNodeID(measureSchemaNode),
			Msg: err.Error(),
		}
	}
	for _, rs := range results {
		for _, row := range rs.Rows {
			if len(row) == 0 {
				continue
			}
			if len(row) > 1 {
				return nil, ErrMeasureProcessing{
					ID:  ls.GetNodeID(measureSchemaNode),
					Msg: "Expression returns multiple columns",
				}
			}
			for _, v := range row {
				if v.Get() == nil {
					continue
				}
				node, ok := v.Get().(graph.Node)
				if !ok {
					return nil, ErrMeasureProcessing{
						ID:  ls.GetNodeID(measureSchemaNode),
						Msg: "Result is not a node",
					}
				}
				valueNodes = append(valueNodes, node)
			}
		}
	}
	if s := ls.AsPropertyValue(measureSchemaNode.GetProperty(MeasureValueNode)).AsString(); len(s) > 0 {
		valueNodes = append(valueNodes, ls.GetNodesInstanceOf(g, s)...)
	}
	return valueNodes, nil
}

// findUnit returns the unit of the value node based on the specification of the schema node
func findUnit(valueNode, measureSchemaNode graph.Node) (string, error) {
	if valueNode == nil {
		return "", nil
	}
	evalCtx := opencypher.NewEvalContext(valueNode.GetGraph())
	evalCtx.SetVar("valueNode", opencypher.ValueOf(valueNode))
	results, err := ls.CompileOCSemantics{}.Evaluate(measureSchemaNode, MeasureUnitExpr, evalCtx)
	if err != nil {
		return "", ErrMeasureProcessing{
			ID:  ls.GetNodeID(measureSchemaNode),
			Msg: fmt.Sprintf("Cannot get unit node: %s", err),
		}
	}
	// Select the first matching result
	for _, result := range results {
		if len(result.Rows) != 1 {
			return "", ErrMeasureProcessing{
				ID:  ls.GetNodeID(measureSchemaNode),
				Msg: "Multiple columns in unit expression",
			}
		}
		for _, v := range result.Rows[0] {
			// Must be string or node
			if v.Get() == nil {
				continue
			}
			node, ok := v.Get().(graph.Node)
			if ok {
				s, _ := ls.GetRawNodeValue(node)
				return s, nil
			}
			return fmt.Sprint(v.Get()), nil
		}
	}
	// If we are here, expressions did not match
	unitNodeID := ls.AsPropertyValue(measureSchemaNode.GetProperty(MeasureUnitNode)).AsString()
	if len(unitNodeID) == 0 {
		return "", nil
	}

	// Find the closest unit node starting from the value node
	found := make([]graph.Node, 0)
	addToFound := func(node graph.Node) bool {
		if ls.AsPropertyValue(node.GetProperty(ls.SchemaNodeIDTerm)).AsString() == unitNodeID {
			found = append(found, node)
			return true
		}
		return true
	}
	ls.IterateDescendants(valueNode, addToFound, ls.FollowEdgesInEntity, false)
	if len(found) == 0 {
		// Try parent
		sources := graph.SourceNodes(valueNode.GetEdges(graph.IncomingEdge))
		if len(sources) != 1 {
			// Cannot find unit
			return "", ErrMeasureProcessing{
				ID:  ls.GetNodeID(measureSchemaNode),
				Msg: "Cannot find unit node",
			}
		}
		ls.IterateDescendants(sources[0], addToFound, ls.FollowEdgesInEntity, false)
		if len(found) != 1 {
			return "", ErrMeasureProcessing{
				ID:  ls.GetNodeID(measureSchemaNode),
				Msg: "Cannot find unit node",
			}
		}
	}
	s, _ := ls.GetRawNodeValue(found[0])
	return s, nil
}

// BuildMeasureNode uses the measureSchemaNode to locate measure node
// instances in the graph, and creates/updates measure nodes in the
// graph
func BuildMeasureNodes(ctx *ls.Context, builder ls.GraphBuilder, measureSchemaNode graph.Node) error {
	ctx.GetLogger().Debug(map[string]interface{}{
		"mth":   "buildMeasureNodes",
		"stage": "start",
	})
	valueNodes, err := getMeasureValueNodes(ctx, builder.GetGraph(), measureSchemaNode)
	if err != nil {
		return err
	}
	if len(valueNodes) == 0 {
		return nil
	}
	ctx.GetLogger().Debug(map[string]interface{}{
		"mth":        "buildMeasureNodes",
		"valueNodes": len(valueNodes),
	})
	valuesUnits := make(map[graph.Node]string)
	for _, node := range valueNodes {
		unit, err := findUnit(node, measureSchemaNode)
		if err != nil {
			return err
		}
		valuesUnits[node] = unit
	}
	ctx.GetLogger().Debug(map[string]interface{}{
		"mth":   "buildMeasureNodes",
		"units": valuesUnits,
	})

	// Build/update measure nodes
	for value, unit := range valuesUnits {
		sources := graph.SourceNodes(value.GetEdges(graph.IncomingEdge))
		if len(sources) != 1 {
			// Cannot find parent node
			return ErrMeasureProcessing{
				ID:  ls.GetNodeID(measureSchemaNode),
				Msg: "Cannot find parent node",
			}
		}
		var measureNode graph.Node
		// Is there a measure node already?
		measureNodes := ls.FindChildInstanceOf(sources[0], ls.GetNodeID(measureSchemaNode))
		if len(measureNodes) == 0 {
			// Create one
			_, measureNode, err = builder.ValueAsNode(measureSchemaNode, sources[0], "")
			if err != nil {
				return ErrMeasureProcessing{
					ID:  ls.GetNodeID(measureSchemaNode),
					Msg: fmt.Sprintf("Cannot create measure node: %s", err.Error()),
				}
			}
			measureNode.SetLabels(measureNode.GetLabels().Add(MeasureTerm))
		} else {
			measureNode = sources[0]
		}
		v, _ := ls.GetRawNodeValue(value)
		measureNode.SetProperty(MeasureValueTerm, ls.StringPropertyValue(v))
		measureNode.SetProperty(MeasureUnitTerm, ls.StringPropertyValue(unit))
		ls.SetRawNodeValue(measureNode, Measure{Value: v, Unit: unit}.String())
	}

	return nil
}

// BuildMeasureNodesForLayer builds all the measure nodes for the layer
func BuildMeasureNodesForLayer(ctx *ls.Context, bldr ls.GraphBuilder, layer *ls.Layer) error {
	var err error
	layer.ForEachAttribute(func(node graph.Node, _ []graph.Node) bool {
		err = BuildMeasureNodes(ctx, bldr, node)
		if err != nil {
			return false
		}
		return true
	})
	return err
}

type measureParser struct{}

func (measureParser) GetNativeValue(value string, node graph.Node) (interface{}, error) {
	return ParseMeasure(value)
}

func (measureParser) FormatNativeValue(newValue, oldValue interface{}, node graph.Node) (string, error) {
	if newValue == nil {
		return "", nil
	}
	switch t := newValue.(type) {
	case Measure:
		return t.String(), nil
	case string:
		m, err := ParseMeasure(t)
		if err != nil {
			return "", err
		}
		return m.String(), nil
	}
	return "", ErrNotAMeasure{Value: fmt.Sprintf("%+v %T", newValue, newValue)}
}

func (measureParser) GetNodeValue(node graph.Node) (interface{}, error) {
	ret := Measure{
		Value: ls.AsPropertyValue(node.GetProperty(MeasureValueTerm)).AsString(),
		Unit:  ls.AsPropertyValue(node.GetProperty(MeasureUnitTerm)).AsString(),
	}
	return ret, nil
}

func (measureParser) SetNodeValue(value interface{}, node graph.Node) error {
	if value == nil {
		node.RemoveProperty(MeasureValueTerm)
		node.RemoveProperty(MeasureUnitTerm)
		ls.RemoveRawNodeValue(node)
		return nil
	}
	switch t := value.(type) {
	case Measure:
		node.SetProperty(MeasureValueTerm, ls.StringPropertyValue(t.Value))
		node.SetProperty(MeasureUnitTerm, ls.StringPropertyValue(t.Unit))
		ls.SetRawNodeValue(node, t.String())
		return nil
	}
	return ErrNotAMeasure{Value: fmt.Sprintf("%+v %T", value, value)}
}
