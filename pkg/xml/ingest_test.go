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

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
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
		layer, err = c.CompileSchema(ls.DefaultContext(), layer)
		if err != nil {
			return fmt.Errorf("%s: %w", schemaName, err)
		}
		schema = layer
	}

	parser := Parser{}
	if schema != nil {
		parser.SchemaNode = schema.GetSchemaRootNode()
	}
	builder := ls.NewGraphBuilder(nil, ls.GraphBuilderOptions{
		EmbedSchemaNodes: true,
	})

	parsed, err := parser.ParseStream(ls.DefaultContext(), "a", f)
	if err != nil {
		return fmt.Errorf("%s: %w", xmlname, err)
	}
	_, err = ls.Ingest(builder, parsed)
	if err != nil {
		return fmt.Errorf("%s: %w", graphname, err)
	}

	d, err := ioutil.ReadFile("testdata/" + graphname + ".json")
	if err != nil {
		return fmt.Errorf("%s: %w", graphname, err)
	}
	expected := graph.NewOCGraph()
	m := ls.JSONMarshaler{}
	if err := m.Unmarshal(d, expected); err != nil {
		return fmt.Errorf("%s: %s", graphname, err)
	}
	if !graph.CheckIsomorphism(builder.GetGraph(), expected, func(n1, n2 graph.Node) bool {
		s1, _ := ls.GetRawNodeValue(n1)
		s2, _ := ls.GetRawNodeValue(n2)
		if s1 != s2 {
			return false
		}
		if !ls.IsPropertiesEqual(ls.PropertiesAsMap(n1), ls.PropertiesAsMap(n2)) {
			return false
		}
		return true
	}, func(e1, e2 graph.Edge) bool {
		return ls.IsPropertiesEqual(ls.PropertiesAsMap(e1), ls.PropertiesAsMap(e2))
	}) {
		d, _ := m.Marshal(builder.GetGraph())
		fmt.Println("got:" + string(d))
		d, _ = m.Marshal(expected)
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

func TestAttrValue(t *testing.T) {
	if err := xmlIngestAndCheck("attrTest", "attrTestSchema", "attrTestExpected"); err != nil {
		t.Error(err)
	}
}

func TestEffectiveTime(t *testing.T) {
	if err := xmlIngestAndCheck("effectiveTime", "effectiveTimeSchema", "effectiveTimeExpected"); err != nil {
		t.Error(err)
	}
}
