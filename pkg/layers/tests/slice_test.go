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
package tests

// Separate package to prevent import loops

import (
	"encoding/json"
	"testing"

	"github.com/bserdar/digraph"

	"github.com/cloudprivacylabs/lsa/pkg/jsonld"
	"github.com/cloudprivacylabs/lsa/pkg/layers"
)

func TestSlice1(t *testing.T) {
	var v interface{}
	json.Unmarshal([]byte(`{
"@context": ["../../../schemas/ls.jsonld",{"@vocab":"http://test/"}],
"@type":"Schema",
"@id": "schId",
"attributes": [
{
  "@id":  "attr1",
  "@type": "Value",
  "someKey" : [
     {"@value": "someValue"}
   ]
},
{
  "@id":  "attr2" ,
  "@type": "Value",
  "privacy": [
    {
      "@value": "flg1"
    }
  ]
},
{
  "@id":"attr3",
  "@type": "Value",
  "privacy": [
      {"@value": "flg2"},
      {"@value": "flg3"}
   ]
 }
]
}`), &v)
	base, err := jsonld.UnmarshalLayer(v)
	if err != nil {
		panic(err)
	}

	slice := base.Slice(layers.OverlayTerm, func(target *layers.Layer, node *digraph.Node) *digraph.Node {
		if v, ok := node.Payload.(*layers.SchemaNode).Properties["http://test/someKey"]; ok {
			ret := target.NewNode(node.Label(), node.Payload.(*layers.SchemaNode).GetTypes()...)
			ret.Payload.(*layers.SchemaNode).Properties["http://test/someKey"] = v
			return ret
		}
		return nil
	})
	data := jsonld.MarshalLayer(slice).([]interface{})[0].(map[string]interface{})
	{
		x, _ := json.MarshalIndent(data, "", "  ")
		t.Logf("%s", string(x))
	}
	if len(data[layers.TypeTerms.Attributes].([]interface{})) != 1 {
		t.Errorf("1 attr expected")
	}
	if data[layers.TypeTerms.Attributes].([]interface{})[0].(map[string]interface{})["@id"] != "attr1" {
		t.Errorf("attr1 expected")
	}

	slice = base.Slice(layers.OverlayTerm, func(target *layers.Layer, node *digraph.Node) *digraph.Node {
		if v, ok := node.Payload.(*layers.SchemaNode).Properties["http://test/privacy"]; ok {
			ret := target.NewNode(node.Label(), node.Payload.(*layers.SchemaNode).GetTypes()...)
			ret.Payload.(*layers.SchemaNode).Properties["http://test/privacy"] = v
			return ret
		}
		return nil
	})
	data = jsonld.MarshalLayer(slice).([]interface{})[0].(map[string]interface{})
	{
		x, _ := json.MarshalIndent(data, "", "  ")
		t.Logf("%s", string(x))
	}
	if len(data[layers.TypeTerms.Attributes].([]interface{})) != 2 {
		t.Errorf("2 attr expected")
	}
	if data[layers.TypeTerms.Attributes].([]interface{})[0].(map[string]interface{})["@id"] != "attr2" {
		t.Errorf("attr2 expected")
	}
	if data[layers.TypeTerms.Attributes].([]interface{})[1].(map[string]interface{})["@id"] != "attr3" {
		t.Errorf("attr3 expected")
	}

}
