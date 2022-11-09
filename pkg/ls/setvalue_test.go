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
	"bytes"
	"os"
	"testing"

	"github.com/cloudprivacylabs/lpg"
)

func TestSetValue(t *testing.T) {
	d, err := os.ReadFile("testdata/setvaluetest-1.json")
	if err != nil {
		t.Error(err)
		return
	}
	layer, err := UnmarshalLayerFromSlice(d)
	if err != nil {
		panic(err)
	}
	compiler := Compiler{}
	layer, err = compiler.CompileSchema(DefaultContext(), layer)
	if err != nil {
		t.Error(err)
		return
	}

	gb := NewGraphBuilder(nil, GraphBuilderOptions{EmbedSchemaNodes: true})
	_, schemaRoot, _ := gb.ObjectAsNode(layer.GetAttributeByID("schemaRoot"), nil)
	_, attr1, _ := gb.RawValueAsNode(layer.GetAttributeByID("attr1"), schemaRoot, "attr1")
	_, attr2, _ := gb.ObjectAsNode(layer.GetAttributeByID("attr2"), schemaRoot)
	_, attr3, _ := gb.RawValueAsNode(layer.GetAttributeByID("attr3"), attr2, "attr3")

	if err := gb.PostIngest(layer.GetAttributeByID("schemaRoot"), schemaRoot); err != nil {
		t.Error(err)
	}

	buf := bytes.Buffer{}
	lpg.JSON{}.Encode(gb.GetGraph(), &buf)
	t.Log(buf.String())
	nodeIDMap := GetSchemaNodeIDMap(schemaRoot)
	attr4 := nodeIDMap["attr4"]
	if len(attr4) != 1 {
		t.Errorf("Expecting 1 node")
	}
	if v, _ := GetRawNodeValue(attr3); v != "attr3" {
		t.Errorf("Wrong value")
	}
	_ = attr1
	_ = attr2
}
