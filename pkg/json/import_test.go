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

package json

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/json/jsonschema"

	"github.com/cloudprivacylabs/lpg"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func TestAnnotations(t *testing.T) {
	compiler := jsonschema.NewCompiler()
	compiler.AddResource("/schema", strings.NewReader(`{
 "type":"object",
 "properties": {
   "p1": {
     "type": "string",
     "x-ls": {
       "field": "value"
     }
   }
 }
}`))
	compiled, err := CompileEntitiesWith(compiler, Entity{Ref: "/schema", LayerID: "id"})
	if err != nil {
		t.Error(err)
	}
	if compiled[0].Schema.Properties["p1"].Extensions[X_LS].(annotationExtSchema)["field"].(string) != "value" {
		t.Errorf("No extension")
	}
	targetGraph := lpg.NewGraph()
	layers, err := BuildEntityGraph(targetGraph, ls.SchemaTerm, LinkRefsBySchemaRef, compiled[0])
	if err != nil {
		t.Error(err)
		return
	}
	node := lpg.NextNodesWith(layers[0].Layer.GetSchemaRootNode(), ls.ObjectAttributeListTerm)[0]
	if ls.AsPropertyValue(node.GetProperty("field")).AsString() != "value" {
		t.Errorf("Wrong value: %+v", node)
	}
}

func TestRefs(t *testing.T) {
	td, err := ioutil.ReadFile("testdata/ref_schema.json")
	if err != nil {
		t.Fail()
		return
	}
	compiler := jsonschema.NewCompiler()
	compiler.AddResource("https://ref", bytes.NewReader(td))

	compiled, err := CompileEntitiesWith(compiler, Entity{Ref: "https://ref#/definitions/Array", LayerID: "http://array"},
		Entity{Ref: "https://ref#/definitions/Item", LayerID: "http://item"})
	if err != nil {
		t.Error(err)
		return
	}
	targetGraph := lpg.NewGraph()
	graphs, err := BuildEntityGraph(targetGraph, ls.SchemaTerm, LinkRefsByLayerID, compiled...)
	if err != nil {
		t.Error(err)
		return
	}
	// Array must have a reference to item
	root := graphs[0].Layer.GetSchemaRootNode()
	if !root.GetLabels().Has(ls.AttributeTypeArray) {
		t.Errorf("%s: Not an array", ls.GetNodeID(root))
	}
	items := lpg.NextNodesWith(root, ls.ArrayItemsTerm)
	if len(items) != 1 {
		t.Errorf("Wrong items")
	}
	itemNode := items[0]
	if !itemNode.GetLabels().Has(ls.AttributeTypeReference) {
		t.Errorf("Items not a ref")
	}
	if ls.AsPropertyValue(itemNode.GetProperty(ls.ReferenceTerm)).AsString() != "http://item" {
		t.Errorf("Wrong ref: %v", itemNode)
	}
}

func TestLoop(t *testing.T) {
	td, err := ioutil.ReadFile("testdata/loop_sch.json")
	if err != nil {
		t.Fail()
		return
	}
	compiler := jsonschema.NewCompiler()
	compiler.AddResource("https://loop", bytes.NewReader(td))

	compiled, err := CompileEntitiesWith(compiler, Entity{Ref: "https://loop#/definitions/Item", LayerID: "http://item"})
	if err != nil {
		t.Error(err)
		return
	}
	targetGraph := lpg.NewGraph()
	_, err = BuildEntityGraph(targetGraph, ls.SchemaTerm, LinkRefsBySchemaRef, compiled...)
	if err != nil {
		t.Error(err)
		return
	}
}
