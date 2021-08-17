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
package project

import (
	"encoding/json"
	"io/ioutil"
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

func runTestSuite(t *testing.T, file string) {
	data, err := ioutil.ReadFile("testdata/" + file)
	if err != nil {
		t.Errorf("File: %s %v", file, err)
		return
	}
	var cases []testCase
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Errorf("File: %s %v", file, err)
		return
	}
	for _, tc := range cases {
		t.Logf("Running %s", tc.Name)
		targetLayer, err := ls.UnmarshalLayer(tc.Target)
		if err != nil {
			t.Errorf("File: %s Test case: %s Cannot unmarshal target layer: %v", file, tc.Name, err)
			return
		}
		sourceGraph, err := ls.UnmarshalGraph(tc.SourceGraph)
		if err != nil {
			t.Errorf("File: %s Test case: %s Cannot unmarshal source graph: %v", file, tc.Name, err)
			return
		}
		nix := sourceGraph.GetNodeIndex()
		rootNode := nix.NodesByLabel(tc.RootID).All()
		if len(rootNode) != 1 {
			t.Errorf("File: %s Test case: %s No root node", file, tc.Name)
			return
		}
		projector := Projector{TargetSchema: targetLayer}
		result, err := projector.Project(rootNode[0].(ls.Node))
		if err != nil {
			t.Errorf("File: %s Test case: %s Projection error: %v", file, tc.Name, err)
			return
		}
		ldMarshaler := ls.LDMarshaler{}
		resultGraph := digraph.New()
		resultGraph.AddNode(result)
		resultMarshaled := ldMarshaler.Marshal(resultGraph)
		t.Logf("Projected: %v", ls.ToMap(resultMarshaled))
		if err := ls.DeepEqual(ls.ToMap(resultMarshaled), ls.ToMap(tc.Expected)); err != nil {
			t.Errorf("File: %s Test case: %s Result is different from the expected: Result: %v Expected: %v", file, tc.Name, ls.ToMap(resultMarshaled), ls.ToMap(tc.Expected))
		}
	}
}

func TestBasicProjection(t *testing.T) {
	runTestSuite(t, "basic.json")
}
