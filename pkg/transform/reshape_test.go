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

	"github.com/bserdar/digraph"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
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
	sourceGraph, err := ls.UnmarshalGraph(tc.SourceGraph, nil)
	if err != nil {
		t.Errorf("Test case: %s Cannot unmarshal source graph: %v", tc.Name, err)
		return
	}
	nix := sourceGraph.GetIndex()
	rootNode := nix.NodesByLabel(tc.RootID).All()
	if len(rootNode) != 1 {
		t.Errorf(" Test case: %s No root node", tc.Name)
		return
	}
	reshaper := Reshaper{TargetSchema: targetLayer}
	result, err := reshaper.Reshape(rootNode[0].(ls.Node))
	if err != nil {
		t.Errorf("Test case: %s Reshaper error: %v", tc.Name, err)
		return
	}
	if result == nil {
		t.Errorf("Test case: %s nil reshaping", tc.Name)
		return
	}
	result.SetLabel("root")
	ldMarshaler := ls.LDMarshaler{}
	resultGraph := digraph.New()
	resultGraph.AddNode(result)

	expectedGraph, err := ls.UnmarshalGraph(tc.Expected, nil)
	if err != nil {
		t.Errorf("Test case: %s Cannot unmarshal expected graph: %v", tc.Name, err)
		return
	}
	resultMarshaled := ldMarshaler.Marshal(resultGraph)
	t.Logf("Projected: %v", ls.ToMap(resultMarshaled))
	eq := digraph.CheckIsomorphism(resultGraph.GetIndex(), expectedGraph.GetIndex(), func(n1, n2 digraph.Node) bool { return true }, func(e1, e2 digraph.Edge) bool { return true })

	if !eq {
		t.Errorf("Test case: %s Result is different from the expected: Result: %v Expected: %v", tc.Name, ls.ToMap(resultMarshaled), ls.ToMap(tc.Expected))
	}
}

func TestBasicReshape(t *testing.T) {
	run := func(in json.RawMessage) (ls.TestCase, error) {
		var c testCase
		err := json.Unmarshal(in, &c)
		return c, err
	}
	ls.RunTestsFromFile(t, "testdata/basic.json", run)
	ls.RunTestsFromFile(t, "testdata/fhir.json", run)
}
