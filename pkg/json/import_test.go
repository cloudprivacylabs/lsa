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

	"github.com/santhosh-tekuri/jsonschema/v5"

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
	compiled, err := CompileEntitiesWith(compiler, Entity{Ref: "/schema", ID: "id"})
	if err != nil {
		t.Error(err)
	}
	if compiled[0].Schema.Properties["p1"].Extensions[X_LS].(annotationExtSchema)["field"].AsString() != "value" {
		t.Errorf("No extension")
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

	compiled, err := CompileEntitiesWith(compiler, Entity{Ref: "https://ref#/definitions/Array", ID: "http://array"},
		Entity{Ref: "https://ref#/definitions/Item", ID: "http://item"})
	if err != nil {
		t.Error(err)
		return
	}
	graphs, err := BuildEntityGraph(ls.SchemaTerm, compiled...)
	if err != nil {
		t.Error(err)
		return
	}
	// Array must have a reference to item
	root := graphs[0].Layer.GetSchemaRootNode()
	if !root.GetTypes().Has(ls.AttributeTypes.Array) {
		t.Errorf("%s: Not an array", root.GetID())
	}
	items := root.OutWith(ls.LayerTerms.ArrayItems).Targets().All()
	if len(items) != 1 {
		t.Errorf("Wrong items")
	}
	itemNode := items[0].(ls.Node)
	if !itemNode.GetTypes().Has(ls.AttributeTypes.Reference) {
		t.Errorf("Items not a ref")
	}
	if itemNode.GetProperties()[ls.LayerTerms.Reference].AsString() != "http://item" {
		t.Errorf("Wrong ref: %v", itemNode.GetProperties())
	}
}
