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
package jsonld

import (
	"encoding/json"
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func TestUnmarshalJsonld(t *testing.T) {
	var input interface{}
	err := json.Unmarshal([]byte(`{
"@context": ["../../schemas/ls.jsonld",
  {"@vocab":"http://test#"}],
"@type":"Schema",
"attributes": [
 {
  "@id": "attr1",
  "@type": "Value"
 },
 {
  "@id":  "attr2",
  "@type": "Value",
  "flag": [
    {
      "@value": "flg1"
    }
  ]
 },
 {
  "@id": "attr3",
  "@type": "Reference",
  "reference": "ref1"
 }
]
}`), &input)
	if err != nil {
		t.Error(err)
		return
	}

	layer, err := UnmarshalLayer(input)
	if err != nil {
		t.Error(err)
	}
	n := layer.AllNodesWithLabel("attr1").All()[0]
	v, ok := n.(*ls.SchemaNode)
	if !ok {
		t.Errorf("Not a value")
	}
	if !v.HasType(ls.AttributeTypes.Value) {
		t.Errorf("Not a value")
	}
	n = layer.AllNodesWithLabel("attr3").All()[0]
	ref := n.(*ls.SchemaNode)
	if ref.Properties[ls.TypeTerms.Reference] != ls.IRI("ref1") {
		t.Errorf("Wrong ref: %v", ref.Properties)
	}
	edges := layer.GetRoot().AllOutgoingEdgesWithLabel(ls.TypeTerms.Attributes).All()
	if len(edges) != 3 {
		t.Errorf("Expected 3 got %d", len(edges))
	}

	n2 := layer.AllNodesWithLabel("attr2").All()[0]
	if n2.(*ls.SchemaNode).Properties["http://test#flag"] != "flg1" {
		t.Errorf("Wrong label: %v", n2)
	}
}

func TestMarshalJsonld(t *testing.T) {
	var input interface{}
	err := json.Unmarshal([]byte(`{
"@context": ["../../schemas/ls.jsonld",
  {"@vocab":"http://test#"}],
"@type":"Schema",
"attributes": [
 {
  "@id": "attr1",
  "@type": "Value"
 },
 {
  "@id":  "attr2",
  "@type": "Value",
  "flag": [
    {
      "@value": "flg1"
    }
  ]
 },
 {
  "@id": "attr3",
  "@type": "Reference",
  "reference": "ref1"
 }
]
}`), &input)
	if err != nil {
		t.Error(err)
		return
	}

	layer, err := UnmarshalLayer(input)
	if err != nil {
		t.Error(err)
	}
	out := MarshalLayer(layer)
	x, _ := json.MarshalIndent(out, "", "")
	t.Log(string(x))
	expected := `[
{
"@id": "_:b0",
"@type": [
"https://lschema.org/Attribute",
"https://lschema.org/Schema",
"https://lschema.org/Object"
],
"https://lschema.org/Object#attributes": [
{
"@id": "attr1",
"@type": [
"https://lschema.org/Value",
"https://lschema.org/Attribute"
]
},
{
"@id": "attr2",
"@type": [
"https://lschema.org/Value",
"https://lschema.org/Attribute"
],
"http://test#flag": [
{
"@value": "flg1"
}
]
},
{
"@id": "attr3",
"@type": [
"https://lschema.org/Reference",
"https://lschema.org/Attribute"
],
"https://lschema.org/Reference#reference": [
{
"@id": "ref1"
}
]
}
]
}
]`

	if string(x) != expected {
		t.Errorf("Got %s Expected: %s", string(x), expected)
	}
}
