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
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
)

func TestMeasureValueUnitInNode(t *testing.T) {
	g := graph.NewOCGraph()
	n1 := g.NewNode([]string{ls.DocumentNodeTerm}, map[string]interface{}{
		ls.EntitySchemaTerm: ls.StringPropertyValue("sch"),
	})
	n2 := g.NewNode([]string{ls.DocumentNodeTerm}, map[string]interface{}{
		ls.ValueTypeTerm: ls.StringPropertyValue(MeasureTerm),
		ls.NodeValueTerm: ls.StringPropertyValue("123 cm"),
	})
	n3 := g.NewNode([]string{ls.DocumentNodeTerm}, map[string]interface{}{
		ls.ValueTypeTerm: ls.StringPropertyValue(MeasureTerm),
		ls.NodeValueTerm: ls.StringPropertyValue("123"),
		MeasureUnitTerm:  ls.StringPropertyValue("cm"),
	})

	g.NewEdge(n1, n2, ls.HasTerm, nil)
	g.NewEdge(n1, n3, ls.HasTerm, nil)

	m, err := GetNodeMeasureValue(n2)
	if err != nil {
		t.Error(err)
		return
	}
	if m.Value != "123 cm" || m.Unit != "" {
		t.Errorf("Wrong measure: %+v", m)
	}
	m, err = GetNodeMeasureValue(n3)
	if err != nil {
		t.Error(err)
		return
	}
	if m.Value != "123" || m.Unit != "cm" {
		t.Errorf("Wrong measure: %+v", m)
	}
}

func TestMeasureValueInNodeUnitInSchemaNode(t *testing.T) {
	g := graph.NewOCGraph()
	n1 := g.NewNode([]string{ls.DocumentNodeTerm}, map[string]interface{}{
		ls.EntitySchemaTerm: ls.StringPropertyValue("sch"),
	})
	n2 := g.NewNode([]string{ls.DocumentNodeTerm}, map[string]interface{}{
		ls.ValueTypeTerm:    ls.StringPropertyValue(MeasureTerm),
		ls.NodeValueTerm:    ls.StringPropertyValue("123"),
		MeasureUnitNodeTerm: ls.StringPropertyValue("schUnitNode"),
	})
	n3 := g.NewNode([]string{ls.DocumentNodeTerm}, map[string]interface{}{
		ls.NodeValueTerm:    ls.StringPropertyValue("cm"),
		ls.SchemaNodeIDTerm: ls.StringPropertyValue("schUnitNode"),
	})

	g.NewEdge(n1, n2, ls.HasTerm, nil)
	g.NewEdge(n1, n3, ls.HasTerm, nil)

	m, err := GetNodeMeasureValue(n2)
	if err != nil {
		t.Error(err)
		return
	}
	if m.Value != "123" || m.Unit != "cm" {
		t.Errorf("Wrong measure: %+v", m)
	}
}

func TestMeasureValueInSchemaNodeUnitInSchemaNode(t *testing.T) {
	g := graph.NewOCGraph()
	n1 := g.NewNode([]string{ls.DocumentNodeTerm}, map[string]interface{}{
		ls.EntitySchemaTerm: ls.StringPropertyValue("sch"),
	})
	n2 := g.NewNode([]string{ls.DocumentNodeTerm}, map[string]interface{}{
		ls.ValueTypeTerm:     ls.StringPropertyValue(MeasureTerm),
		MeasureValueNodeTerm: ls.StringPropertyValue("schMeasureNode"),
		MeasureUnitNodeTerm:  ls.StringPropertyValue("schUnitNode"),
	})
	n3 := g.NewNode([]string{ls.DocumentNodeTerm}, map[string]interface{}{
		ls.NodeValueTerm:    ls.StringPropertyValue("cm"),
		ls.SchemaNodeIDTerm: ls.StringPropertyValue("schUnitNode"),
	})
	n4 := g.NewNode([]string{ls.DocumentNodeTerm}, map[string]interface{}{
		ls.NodeValueTerm:    ls.StringPropertyValue("123"),
		ls.SchemaNodeIDTerm: ls.StringPropertyValue("schMeasureNode"),
	})

	g.NewEdge(n1, n2, ls.HasTerm, nil)
	g.NewEdge(n2, n3, ls.HasTerm, nil)
	g.NewEdge(n2, n4, ls.HasTerm, nil)

	m, err := GetNodeMeasureValue(n2)
	if err != nil {
		t.Error(err)
		return
	}
	if m.Value != "123" || m.Unit != "cm" {
		t.Errorf("Wrong measure: %+v", m)
	}
}

func TestMeasureValueInNodeUnitInPath(t *testing.T) {
	g := graph.NewOCGraph()
	n1 := g.NewNode([]string{ls.DocumentNodeTerm}, map[string]interface{}{
		ls.EntitySchemaTerm: ls.StringPropertyValue("sch"),
	})
	n2 := g.NewNode([]string{ls.DocumentNodeTerm}, map[string]interface{}{
		ls.ValueTypeTerm:    ls.StringPropertyValue(MeasureTerm),
		ls.NodeValueTerm:    ls.StringPropertyValue("123"),
		MeasureUnitPathTerm: ls.StringPropertyValue("(this)<-[]-()-[]->(target {`" + ls.SchemaNodeIDTerm + "`:\"schUnitNode\"})"),
	})
	n3 := g.NewNode([]string{ls.DocumentNodeTerm}, map[string]interface{}{
		ls.NodeValueTerm:    ls.StringPropertyValue("cm"),
		ls.SchemaNodeIDTerm: ls.StringPropertyValue("schUnitNode"),
	})

	g.NewEdge(n1, n2, ls.HasTerm, nil)
	g.NewEdge(n1, n3, ls.HasTerm, nil)

	m, err := GetNodeMeasureValue(n2)
	if err != nil {
		t.Error(err)
		return
	}
	if m.Value != "123" || m.Unit != "cm" {
		t.Errorf("Wrong measure: %+v", m)
	}
}
