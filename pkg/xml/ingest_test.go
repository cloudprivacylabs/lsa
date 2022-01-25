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

package xml

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/bserdar/digraph"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func xmlIngestAndCheck(xmlname, schemaName, graphname string) error {
	f, err := os.Open("testdata/" + xmlname + ".xml")
	if err != nil {
		return fmt.Errorf("%s: %w", xmlname, err)
	}
	defer f.Close()

	var schema *ls.Layer
	if schemaName != "" {
		s, err := ioutil.ReadFile("testdata/" + schemaName + ".json")
		if err != nil {
			return fmt.Errorf("%s: %w", schemaName, err)
		}
		var v interface{}
		if err := json.Unmarshal(s, &v); err != nil {
			return fmt.Errorf("%s: %w", schemaName, err)
		}
		layer, err := ls.UnmarshalLayer(v, nil)
		if err != nil {
			return fmt.Errorf("%s: %w", schemaName, err)
		}
		c := ls.Compiler{}
		layer, err = c.CompileSchema(layer)
		if err != nil {
			return fmt.Errorf("%s: %w", schemaName, err)
		}
		schema = layer
	}

	ingester := &Ingester{}
	ingester.Schema = schema
	ingester.EmbedSchemaNodes = true
	root, err := IngestStream(ingester, "a", f)
	if err != nil {
		return fmt.Errorf("%s: %w", xmlname, err)
	}

	d, err := ioutil.ReadFile("testdata/" + graphname + ".json")
	if err != nil {
		return fmt.Errorf("%s: %w", graphname, err)
	}
	expected := digraph.New()
	if err := ls.UnmarshalGraphJSON(d, expected, nil); err != nil {
		return fmt.Errorf("%s: %s", graphname, err)
	}

	got := digraph.New()
	got.AddNode(root)
	if !digraph.CheckIsomorphism(got.GetIndex(), expected.GetIndex(), func(n1, n2 digraph.Node) bool {
		if n1.(ls.Node).GetValue() != n2.(ls.Node).GetValue() {
			fmt.Printf("Different values: '%v' '%v'\n", n1.(ls.Node).GetValue(), n2.(ls.Node).GetValue())
			return false
		}
		if !n1.(ls.Node).GetTypes().IsEqual(*n2.(ls.Node).GetTypes()) {
			fmt.Printf("Different types: '%v' '%v'\n", n1.(ls.Node).GetTypes(), n2.(ls.Node).GetTypes())
			return false
		}
		if !ls.IsPropertiesEqual(n1.(ls.Node).GetProperties(), n2.(ls.Node).GetProperties()) {
			fmt.Printf("Different properties: '%v' '%v'\n", n1.(ls.Node).GetProperties(), n2.(ls.Node).GetProperties())
			return false
		}
		return true
	}, func(e1, e2 digraph.Edge) bool {
		return ls.IsPropertiesEqual(e1.(ls.Edge).GetProperties(), e2.(ls.Edge).GetProperties())
	}) {
		d, _ := ls.MarshalGraphJSON(got)
		fmt.Println("got:" + string(d))
		d, _ = ls.MarshalGraphJSON(expected)
		fmt.Println("expected:" + string(d))
		return fmt.Errorf("%s: Not isomorphic", xmlname)
	}
	return nil
}

func TestBasicIngest(t *testing.T) {
	if err := xmlIngestAndCheck("basicIngest", "", "basicIngest"); err != nil {
		t.Error(err)
	}
}

func TestBasicIngestWS(t *testing.T) {
	if err := xmlIngestAndCheck("basicIngest_ws", "", "basicIngest"); err != nil {
		t.Error(err)
	}
}

func TestBasicIngestWithSchema1(t *testing.T) {
	if err := xmlIngestAndCheck("basicIngest_ws", "basicIngestSchema1", "basicIngestSchema1Expected"); err != nil {
		t.Error(err)
	}
}
