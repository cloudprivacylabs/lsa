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

package transform

import (
	"encoding/json"
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

type testCase struct {
	Name        string      `json:"name"`
	Target      interface{} `json:"target"`
	RootID      string      `json:"rootId"`
	SourceGraph interface{} `json:"sourceGraph"`
	Expected    interface{} `json:"expected"`
}

func (tc testCase) GetName() string { return tc.Name }

func (tc testCase) Run(t *testing.T) {
	t.Logf("Running %s", tc.Name)
	targetLayer, err := ls.UnmarshalLayer(tc.Target, nil)
	if err != nil {
		t.Errorf("Test case: %s Cannot unmarshal target layer: %v", tc.Name, err)
		return
	}
	if err := ls.CompileTerms(targetLayer); err != nil {
		t.Errorf("Test case: %s Cannot compile: %v", tc.Name, err)
		return
	}
	sourceGraph := graph.NewOCGraph()
	err = ls.UnmarshalJSONLDGraph(tc.SourceGraph, sourceGraph, nil)
	if err != nil {
		t.Errorf("Test case: %s Cannot unmarshal source graph: %v", tc.Name, err)
		return
	}
	var rootNode graph.Node
	for g := sourceGraph.GetNodes(); g.Next(); {
		node := g.Node()
		if ls.GetNodeID(node) == tc.RootID {
			rootNode = node
			break
		}
	}
	if rootNode == nil {
		t.Errorf(" Test case: %s No root node", tc.Name)
		return
	}
	reshaper := Reshaper{TargetSchema: targetLayer}
	resultGraph := graph.NewOCGraph()
	result, err := reshaper.Reshape(rootNode, resultGraph)
	if err != nil {
		t.Errorf("Test case: %s Reshaper error: %v", tc.Name, err)
		return
	}
	if result == nil {
		t.Errorf("Test case: %s nil reshaping", tc.Name)
		return
	}
	ls.SetNodeID(result, "root")

	expectedGraph := graph.NewOCGraph()
	err = ls.UnmarshalJSONLDGraph(tc.Expected, expectedGraph, nil)
	if err != nil {
		t.Errorf("Test case: %s Cannot unmarshal expected graph: %v", tc.Name, err)
		return
	}
	eq := graph.CheckIsomorphism(resultGraph, expectedGraph, func(n1, n2 graph.Node) bool {
		t.Logf("Cmp: %+v %+v\n", n1, n2)
		if ls.GetRawNodeValue(n1) != ls.GetRawNodeValue(n2) {
			return false
		}
		if !ls.IsPropertiesEqual(ls.PropertiesAsMap(n1), ls.PropertiesAsMap(n2)) {
			return false
		}
		t.Logf("True\n")
		return true
	}, func(e1, e2 graph.Edge) bool {
		return ls.IsPropertiesEqual(ls.PropertiesAsMap(e1), ls.PropertiesAsMap(e2))
	})

	if !eq {
		m := ls.JSONMarshaler{}
		result, _ := m.Marshal(resultGraph)
		expected, _ := m.Marshal(expectedGraph)
		t.Errorf("Test case: %s Result is different from the expected: Result:\n%s\nExpected:\n%s", tc.Name, string(result), string(expected))
	}
}

func TestBasicReshape(t *testing.T) {
	run := func(in json.RawMessage) (ls.TestCase, error) {
		var c testCase
		err := json.Unmarshal(in, &c)
		return c, err
	}
	ls.RunTestsFromFile(t, "testdata/basic.json", run)
}

func TestFHIRReshape(t *testing.T) {
	run := func(in json.RawMessage) (ls.TestCase, error) {
		var c testCase
		err := json.Unmarshal(in, &c)
		return c, err
	}
	ls.RunTestsFromFile(t, "testdata/fhir.json", run)
}
