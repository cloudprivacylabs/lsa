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
	"github.com/cloudprivacylabs/opencypher/graph"
)

type testCase struct {
	Name          string          `json:"name"`
	Target        interface{}     `json:"target"`
	RootID        string          `json:"rootId"`
	SourceGraph   json.RawMessage `json:"sourceGraph"`
	SourceLDGraph interface{}     `json:"sourceLdGraph"`
	ExpectedLD    interface{}     `json:"expectedLd"`
	Expected      json.RawMessage `json:"expected"`
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

	if tc.SourceGraph != nil {
		m := ls.JSONMarshaler{}
		err = m.Unmarshal(tc.SourceGraph, sourceGraph)
	} else {
		err = ls.UnmarshalJSONLDGraph(tc.SourceLDGraph, sourceGraph, nil)
	}
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
	reshaper := Reshaper{}
	reshaper.TargetSchema = targetLayer
	reshaper.Builder = ls.NewGraphBuilder(nil, ls.GraphBuilderOptions{
		EmbedSchemaNodes: true,
	})
	ls.DefaultLogLevel = ls.LogLevelDebug
	err = reshaper.Reshape(ls.DefaultContext(), sourceGraph)
	if err != nil {
		t.Errorf("Test case: %s Reshaper error: %v", tc.Name, err)
		return
	}
	result := graph.Sources(reshaper.Builder.GetGraph())[0]
	ls.SetNodeID(result, "root")

	expectedGraph := graph.NewOCGraph()
	if tc.Expected != nil {
		m := ls.JSONMarshaler{}
		err = m.Unmarshal(tc.Expected, expectedGraph)
	} else {
		err = ls.UnmarshalJSONLDGraph(tc.ExpectedLD, expectedGraph, nil)
	}
	if err != nil {
		t.Errorf("Test case: %s Cannot unmarshal expected graph: %v", tc.Name, err)
		return
	}
	eq := graph.CheckIsomorphism(reshaper.Builder.GetGraph(), expectedGraph, func(n1, n2 graph.Node) bool {
		t.Logf("Cmp: %+v %+v\n", n1, n2)
		s1, _ := ls.GetRawNodeValue(n1)
		s2, _ := ls.GetRawNodeValue(n2)
		if s1 != s2 {
			return false
		}
		// Expected properties must be a subset
		propertiesOK := true
		n2.ForEachProperty(func(k string, v interface{}) bool {
			pv, ok := v.(*ls.PropertyValue)
			if !ok {
				return true
			}
			v2, ok := n1.GetProperty(k)
			if !ok {
				propertiesOK = false
				return false
			}
			pv2, ok := v2.(*ls.PropertyValue)
			if !ok {
				propertiesOK = false
				return false
			}
			if !pv2.IsEqual(pv) {
				propertiesOK = false
				return false
			}
			return true
		})
		if !propertiesOK {
			return false
		}
		t.Logf("True\n")
		return true
	}, func(e1, e2 graph.Edge) bool {
		return ls.IsPropertiesEqual(ls.PropertiesAsMap(e1), ls.PropertiesAsMap(e2))
	})

	if !eq {
		m := ls.JSONMarshaler{}
		result, _ := m.Marshal(reshaper.Builder.GetGraph())
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

func TestBasicMap(t *testing.T) {
	run := func(in json.RawMessage) (ls.TestCase, error) {
		var c testCase
		err := json.Unmarshal(in, &c)
		return c, err
	}
	ls.RunTestsFromFile(t, "testdata/mapbasic.json", run)
}

func TestFHIRReshape(t *testing.T) {
	run := func(in json.RawMessage) (ls.TestCase, error) {
		var c testCase
		err := json.Unmarshal(in, &c)
		return c, err
	}
	ls.RunTestsFromFile(t, "testdata/fhir.json", run)
}
