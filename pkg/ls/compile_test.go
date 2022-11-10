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

package ls

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/cloudprivacylabs/lpg"
)

type compileTestCase struct {
	Name     string          `json:"name"`
	Schemas  []interface{}   `json:"schemas"`
	Expected json.RawMessage `json:"expected"`
}

func (tc compileTestCase) GetName() string { return tc.Name }

func (tc compileTestCase) Run(t *testing.T) {
	t.Logf("Running %s", tc.Name)
	schemas := make([]*Layer, 0)
	for _, sch := range tc.Schemas {
		l, err := UnmarshalLayer(sch, nil)
		if err != nil {
			t.Errorf("%s: Cannot unmarshal layer: %v", tc.Name, err)
			return
		}
		schemas = append(schemas, l)
	}
	compiler := Compiler{
		Loader: SchemaLoaderFunc(func(x string) (*Layer, error) {
			for _, s := range schemas {
				if s.GetID() == x {
					return s, nil
				}
			}
			return nil, nil
		}),
	}
	result, err := compiler.Compile(DefaultContext(), schemas[0].GetID())
	if err != nil {
		t.Errorf("Cannot compile %v", err)
		return
	}
	m := JSONMarshaler{}
	data, _ := m.Marshal(result.Graph)
	t.Logf(string(data))
}

func TestCompile(t *testing.T) {
	RunTestsFromFile(t, "testdata/compilecases.json", func(in json.RawMessage) (TestCase, error) {
		var c compileTestCase
		err := json.Unmarshal(in, &c)
		return c, err
	})
}

func TestCompileIncludeAttribute(t *testing.T) {
	compiler := Compiler{
		Loader: SchemaLoaderFunc(func(x string) (*Layer, error) {
			data, err := os.ReadFile(x)
			if err != nil {
				return nil, err
			}
			var v interface{}
			if err := json.Unmarshal([]byte(data), &v); err != nil {
				return nil, err
			}
			return UnmarshalLayer(v, nil)
		}),
		CGraph: &DefaultCompiledGraph{},
	}
	ctx := &compilerContext{
		loadedSchemas: make(map[string]*Layer),
	}
	ref := "testdata/includeschema.json"
	compiled, err := compiler.loadSchema(ctx, ref)
	if err != nil {
		t.Errorf("Cannot load schema %v", err)
	}
	compiled, err = compiler.CGraph.PutCompiledSchema(DefaultContext(), ref, compiled)
	err = compiler.compileIncludeAttribute(DefaultContext(), ctx)
	if err != nil {
		t.Errorf("Cannot compile %v", err)
		return
	}
	f, err := os.Open("testdata/includeschema_expected.json")
	if err != nil {
		t.Error(err)
	}
	expectedGraph := lpg.NewGraph()
	m := JSONMarshaler{}
	if err := m.Decode(expectedGraph, json.NewDecoder(f)); err != nil {
		t.Error(err)
	}
	fmt.Println(compiled.layerInfo.GetLabels().Slice())
	// var v interface{}
	// if err := json.Unmarshal([]byte(f), &v); err != nil {
	// 	t.Error(err)
	// }
	// layer, err := UnmarshalLayer(v, nil)
	// if err != nil {
	// 	t.Error(err)
	// }
	m = JSONMarshaler{}
	x, _ := os.Create("cgraph.json")
	m.Encode(compiler.CGraph.GetGraph(), x)
	//
	// x, _ = os.Create("exp.json")
	// m.Encode(layer.Graph, x)
	// //
	if !lpg.CheckIsomorphism(compiler.CGraph.GetGraph(), expectedGraph, checkNodeEquivalence, checkEdgeEquivalence) {
		log.Fatalf("Result:\n%s\nExpected:\n%s", testPrintGraph(compiler.CGraph.GetGraph()), testPrintGraph(expectedGraph))
	}
}

func testPrintGraph(g *lpg.Graph) string {
	m := JSONMarshaler{}
	result, _ := m.Marshal(g)
	return string(result)
}

func checkNodeEquivalence(n1, n2 *lpg.Node) bool {
	return isNodeIdentical(n1, n2)
}

func checkEdgeEquivalence(e1, e2 *lpg.Edge) bool {
	if e1.GetLabel() != e2.GetLabel() {
		return false
	}
	if !IsPropertiesEqual(PropertiesAsMap(e1), PropertiesAsMap(e2)) {
		return false
	}
	return true
}

// Return true if n1 is identical to n2
func isNodeIdentical(n1, n2 *lpg.Node) bool {
	if !n1.GetLabels().IsEqual(n2.GetLabels()) {
		return false
	}
	eq := true
	n1.ForEachProperty(func(k string, v interface{}) bool {
		pv, ok := v.(*PropertyValue)
		if !ok {
			return true
		}
		v2, ok := n2.GetProperty(k)
		if !ok {
			eq = false
			return false
		}
		pv2, ok := v2.(*PropertyValue)
		if !ok {
			eq = false
			return false
		}
		if !pv2.IsEqual(pv) {
			eq = false
			return false
		}
		return true
	})
	if !eq {
		return false
	}
	n2.ForEachProperty(func(k string, v interface{}) bool {
		_, ok := v.(*PropertyValue)
		if !ok {
			return true
		}
		v2, ok := n2.GetProperty(k)
		if !ok {
			eq = false
			return false
		}
		_, ok = v2.(*PropertyValue)
		if !ok {
			eq = false
			return false
		}
		return true
	})
	return eq
}
