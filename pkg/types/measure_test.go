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
	"encoding/json"
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher"
	"github.com/cloudprivacylabs/opencypher/graph"
)

func TestMeasureValueNodesExpr(t *testing.T) {
	schema := `{
"@context": "../../schemas/ls.json",
"@type": "Schema",
"@id":"1",
"layer": {
  "@type": "Object",
  "@id":"root",
  "attributes": {
     "m1": {
       "@type": "Value",
       "attributeName": "m1"
     },
     "u1": {
       "@type": "Value",
       "attributeName": "u1"
     },
     "m": {
       "@type": "Value",
       "valueType": "Measure",
       "attributeName": "m",
       "https://lschema.org/measure/valueNodeExpr": "` +
		"match (v {`https://lschema.org/attributeName`:'m1'}) return v" + `",
       "https://lschema.org/measure/unitExpr": "` +
		"match (valueNode)<-[]-()-[]->(v {`https://lschema.org/attributeName`:'u1'}) return v" + `"
     }
  }
}
}`

	var layer *ls.Layer
	ctx := ls.DefaultContext()
	{
		var v interface{}
		err := json.Unmarshal([]byte(schema), &v)
		if err != nil {
			t.Error(err)
			return
		}
		layer, err = ls.UnmarshalLayer(v, nil)
		if err != nil {
			t.Error(err)
			return
		}
		compiler := ls.Compiler{}
		layer, err = compiler.CompileSchema(ctx, layer)
		if err != nil {
			t.Error(err)
			return
		}
	}

	root := layer.GetAttributeByID("root")
	m1 := layer.GetAttributeByID("m1")
	u1 := layer.GetAttributeByID("u1")
	m := layer.GetAttributeByID("m")
	bldr := ls.NewGraphBuilder(nil, ls.GraphBuilderOptions{EmbedSchemaNodes: true})
	_, rootNode, _ := bldr.ObjectAsNode(root, nil)
	bldr.ValueAsNode(m1, rootNode, "123")
	bldr.ValueAsNode(u1, rootNode, "unit")
	nodes, err := getMeasureValueNodes(ctx, bldr.GetGraph(), m)
	if err != nil {
		t.Error(err)
		return
	}
	if len(nodes) != 1 {
		t.Errorf("Expecting 1 nodes, got %d", len(nodes))
	}
	u, err := findUnit(nodes[0], m)
	if err != nil {
		t.Error(err)
		return
	}
	if u != "unit" {
		t.Errorf("Wrong unit: %s", u)
	}
	if err = BuildMeasureNodesForLayer(ctx, bldr, layer); err != nil {
		t.Error(err)
	}
	// There must be a measure node
	v, err := opencypher.ParseAndEvaluate("match (n:`https://lschema.org/Measure`) return n", opencypher.NewEvalContext(bldr.GetGraph()))
	if err != nil {
		t.Error(err)
		return
	}
	measureNode, _ := v.Get().(opencypher.ResultSet).Rows[0]["1"].Get().(graph.Node)
	if measureNode == nil {
		t.Error(err)
		return
	}
	x, _ := ls.GetNodeValue(measureNode)
	measure := x.(Measure)
	if measure.Value != "123" || measure.Unit != "unit" {
		t.Errorf("Wrong measure: %v", measure)
	}
}

func TestMeasureValueNodes(t *testing.T) {
	schema := `{
"@context": "../../schemas/ls.json",
"@type": "Schema",
"@id":"1",
"layer": {
  "@type": "Object",
  "@id":"root",
  "attributes": {
     "m1": {
       "@type": "Value",
       "attributeName": "m1"
     },
     "u1": {
       "@type": "Value",
       "attributeName": "u1"
     },
     "m": {
       "@type": "Value",
       "valueType": "Measure",
       "attributeName": "m",
       "https://lschema.org/measure/valueNode": "m1",
       "https://lschema.org/measure/unitNode": "u1"
     }
  }
}
}`

	var layer *ls.Layer
	ctx := ls.DefaultContext()
	{
		var v interface{}
		err := json.Unmarshal([]byte(schema), &v)
		if err != nil {
			t.Error(err)
			return
		}
		layer, err = ls.UnmarshalLayer(v, nil)
		if err != nil {
			t.Error(err)
			return
		}
		compiler := ls.Compiler{}
		layer, err = compiler.CompileSchema(ctx, layer)
		if err != nil {
			t.Error(err)
			return
		}
	}

	root := layer.GetAttributeByID("root")
	m1 := layer.GetAttributeByID("m1")
	u1 := layer.GetAttributeByID("u1")
	m := layer.GetAttributeByID("m")
	bldr := ls.NewGraphBuilder(nil, ls.GraphBuilderOptions{EmbedSchemaNodes: true})
	_, rootNode, _ := bldr.ObjectAsNode(root, nil)
	bldr.ValueAsNode(m1, rootNode, "123")
	bldr.ValueAsNode(u1, rootNode, "unit")
	nodes, err := getMeasureValueNodes(ctx, bldr.GetGraph(), m)
	if err != nil {
		t.Error(err)
		return
	}
	if len(nodes) != 1 {
		t.Errorf("Expecting 1 nodes, got %d", len(nodes))
	}
	u, err := findUnit(nodes[0], m)
	if err != nil {
		t.Error(err)
		return
	}
	if u != "unit" {
		t.Errorf("Wrong unit: %s", u)
	}
	if err = BuildMeasureNodesForLayer(ctx, bldr, layer); err != nil {
		t.Error(err)
	}
	// There must be a measure node
	v, err := opencypher.ParseAndEvaluate("match (n:`https://lschema.org/Measure`) return n", opencypher.NewEvalContext(bldr.GetGraph()))
	if err != nil {
		t.Error(err)
		return
	}
	measureNode, _ := v.Get().(opencypher.ResultSet).Rows[0]["1"].Get().(graph.Node)
	if measureNode == nil {
		t.Error(err)
		return
	}
	x, _ := ls.GetNodeValue(measureNode)
	measure := x.(Measure)
	if measure.Value != "123" || measure.Unit != "unit" {
		t.Errorf("Wrong measure: %v", measure)
	}
}
