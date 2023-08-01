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

	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/lsa/pkg/ls"

	_ "github.com/cloudprivacylabs/lsa/pkg/types"
)

type reingestTestCase struct {
	Name     string          `json:"name"`
	Graph    json.RawMessage `json:"graph"`
	Layer    interface{}     `json:"layer"`
	Expected json.RawMessage `json:"expected"`
}

func (tc reingestTestCase) GetName() string { return tc.Name }

func (tc reingestTestCase) Run(t *testing.T) {
	t.Logf("Running %s", tc.Name)
	layer, err := ls.UnmarshalLayer(tc.Layer, nil)
	if err != nil {
		t.Errorf("Test case: %s Cannot unmarshal layer: %v", tc.Name, err)
		return
	}
	g := lpg.NewGraph()
	m := ls.JSONMarshaler{}
	if err := m.Unmarshal(tc.Graph, g); err != nil {
		t.Errorf("Test case: %s Cannot unmarshal  graph: %v", tc.Name, err)
		return
	}
	expectedGraph := lpg.NewGraph()
	if err := m.Unmarshal(tc.Expected, expectedGraph); err != nil {
		t.Errorf("Test case: %s Cannot unmarshal expected graph: %v", tc.Name, err)
		return
	}

	ctx := ls.DefaultContext()
	target := ls.NewGraphBuilder(nil, ls.GraphBuilderOptions{
		EmbedSchemaNodes: true,
	})
	root := lpg.Sources(g)[0]
	if err := Reingest(ctx, root, target, layer); err != nil {
		t.Errorf("Test case: %s Reingest error: %v", tc.Name, err)
		return
	}

	eq := lpg.CheckIsomorphism(target.GetGraph(), expectedGraph, func(n1, n2 *lpg.Node) bool {
		if !n1.GetLabels().IsEqual(n2.GetLabels()) {
			return false
		}
		t.Logf("Cmp: %+v %+v\n", n1, n2)
		s1, _ := ls.GetRawNodeValue(n1)
		s2, _ := ls.GetRawNodeValue(n2)
		if s1 != s2 {
			return false
		}
		// Expected properties must be a subset
		propertiesOK := true
		n2.ForEachProperty(func(k string, v interface{}) bool {
			pv, ok := v.(ls.PropertyValue)
			if !ok {
				return true
			}
			v2, ok := n1.GetProperty(k)
			if !ok {
				propertiesOK = false
				return false
			}
			pv2, ok := v2.(ls.PropertyValue)
			if !ok {
				propertiesOK = false
				return false
			}
			if !pv2.Equal(pv) {
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
	}, func(e1, e2 *lpg.Edge) bool {
		return e1.GetLabel() == e2.GetLabel() && ls.IsPropertiesEqual(ls.PropertiesAsMap(e1), ls.PropertiesAsMap(e2))
	})

	if !eq {
		result, _ := m.Marshal(target.GetGraph())
		expected, _ := m.Marshal(expectedGraph)
		t.Errorf("Test case: %s Result is different from the expected: Result:\n%s\nExpected:\n%s", tc.Name, string(result), string(expected))
	}
}

func TestReingest(t *testing.T) {
	run := func(in json.RawMessage) (ls.TestCase, error) {
		var c reingestTestCase
		err := json.Unmarshal(in, &c)
		return c, err
	}
	ls.RunTestsFromFile(t, "testdata/reingest.json", run)
}
