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

	"github.com/cloudprivacylabs/lsa/pkg/jsonld"
	"github.com/cloudprivacylabs/lsa/pkg/layers"
)

func TestMerge1(t *testing.T) {
	var v interface{}
	json.Unmarshal([]byte(`{
"@context": "../../../schemas/ls.jsonld",
"@type":"Schema",
"attributes": [
{
  "@id":  "attr1",
  "@type": "Value"
},
{
  "@id":  "attr2" ,
  "@type": "Value",
  "privacyClassification": [
    {
      "@value": "flg1"
    }
  ]
},
{
  "@id":"attr3",
  "@type": "Value",
  "privacyClassification": [
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
	json.Unmarshal([]byte(`{
"@context": ["../../../schemas/ls.jsonld",{"@vocab":"http://test/"}],
"@type":"Overlay",
"attributes":[
{
  "@id":"attr1",
  "@type": "Value",
   "someKey" : [
     {"@value": "someValue"}
   ]
},
{
  "@id":"attr2",
  "@type": "Value",
   "privacyClassification" : [
     {"@value": "addFlg1"}
   ]
},
{
  "@id": "attr3",
  "@type": "Value",
   "privacyClassification" : [
     {"@value": "addFlg2"},
     {"@value": "addFlg3"}
   ]
}
]
}`), &v)
	ovl, err := jsonld.UnmarshalLayer(v)
	if err != nil {
		panic(err)
	}
	err = base.Compose(layers.ComposeOptions{}, ovl)
	if err != nil {
		t.Error(err)
	}
	attrs := jsonld.MarshalLayer(base).([]interface{})[0].(map[string]interface{})[layers.TypeTerms.Attributes].([]interface{})
	t.Logf("%+v", attrs)

	find := func(key string) map[string]interface{} {
		for _, x := range attrs {
			m := x.(map[string]interface{})
			if m["@id"] == key {
				return m
			}
		}
		return nil
	}

	attr1 := find("attr1")
	if attr1 == nil {
		t.Errorf("No attr1")
		return
	}
	sk, ok := attr1["http://test/someKey"]
	if !ok {
		t.Errorf("Missing someKey")
		return
	}
	if sk.([]interface{})[0].(map[string]interface{})["@value"] != "someValue" {
		t.Errorf("Wrong value: %v", sk)
	}

	attr2 := find("attr2")
	if attr2 == nil {
		t.Errorf("No attr2")
		return
	}
	priv := attr2["https://layeredschemas.org/attr/privacyClassification"].([]interface{})
	if jsonld.GetNodeValue(priv[0]) != "flg1" || jsonld.GetNodeValue(priv[1]) != "addFlg1" {
		t.Errorf("Wrong flags: %v", priv)
	}

	attr3 := find("attr3")
	if attr3 == nil {
		t.Errorf("No attr3")
		return
	}
	priv = attr3["https://layeredschemas.org/attr/privacyClassification"].([]interface{})
	if jsonld.GetNodeValue(priv[0]) != "flg2" || jsonld.GetNodeValue(priv[1]) != "flg3" || jsonld.GetNodeValue(priv[2]) != "addFlg2" || jsonld.GetNodeValue(priv[3]) != "addFlg3" {
		t.Errorf("Wrong flags: %v", priv)
	}
}

func TestMergeArray(t *testing.T) {
	var v interface{}
	json.Unmarshal([]byte(`{
  "@context":"../../../schemas/ls.jsonld",
  "@type": "Schema",
  "attributes": {
    "array": {
      "@type": "Array",
      "items":  {
        "@id": "http://items",
        "@type": "Value"
      }
    }
  }
}`), &v)
	base, err := jsonld.UnmarshalLayer(v)
	if err != nil {
		panic(err)
	}
	json.Unmarshal([]byte(`{
  "@context": "../../../schemas/ls.jsonld",
  "@type": "Overlay",
  "attributes": {
    "array": {
     "@type": "Array",
      "items": {
       "@type": "Value",
       "@id": "http://items",
       "targetType":"string"
      }
    }
  }
}`), &v)
	ovl, err := jsonld.UnmarshalLayer(v)
	if err != nil {
		panic(err)
	}
	err = base.Compose(layers.ComposeOptions{}, ovl)
	if err != nil {
		t.Error(err)
	}
	attrs := jsonld.MarshalLayer(base).([]interface{})[0].(map[string]interface{})[layers.TypeTerms.Attributes].([]interface{})
	t.Logf("%+v", attrs)

	find := func(key string) map[string]interface{} {
		for _, x := range attrs {
			m := x.(map[string]interface{})
			if m["@id"] == key {
				return m
			}
		}
		return nil
	}

	array := find("array")
	if array == nil {
		t.Errorf("No array")
		return
	}
	items := array[layers.TypeTerms.ArrayItems].([]interface{})[0].(map[string]interface{})
	if s := jsonld.GetNodeID(items[layers.TargetType]); s != "string" {
		t.Errorf("Missing type: %+v %s", items, s)
	}
}

func TestMergeChoice(t *testing.T) {
	var v interface{}
	json.Unmarshal([]byte(`{
  "@context": "../../../schemas/ls.jsonld",
  "@type" : "Schema",
  "attributes": {
   "attr": {
    "@type": "Polymorphic",
	   "oneOf": [
	    {
        "@id": "id1",
        "@type": "Value"
	    }
	   ]
	 }
  }
}`), &v)
	base, err := jsonld.UnmarshalLayer(v)
	if err != nil {
		panic(err)
	}
	json.Unmarshal([]byte(`{
  "@context": "../../../schemas/ls.jsonld",
  "@type": "Overlay",
  "attributes": {
   "attr": {
    "@type": "Polymorphic",
	   "oneOf": [
	    {
        "@id": "id1",
        "@type":"Value",
        "targetType": "string"
	    }
	   ]
	 }
  }
}`), &v)
	ovl, err := jsonld.UnmarshalLayer(v)
	if err != nil {
		panic(err)
	}
	err = base.Compose(layers.ComposeOptions{}, ovl)
	if err != nil {
		t.Error(err)
	}

	item := base.Graph.AllNodesWithLabel("id1").All()[0]

	if item.Payload.(*layers.SchemaNode).Properties[layers.TargetType] != layers.IRI("string") {
		t.Errorf("Wrong value: %+v", item.Payload.(*layers.SchemaNode).Properties[layers.TargetType])
	}
}

// func TestOverride(t *testing.T) {
// 	base := expand(t, `{
// "@context": "../../schemas/ls.jsonld",
// "@type":"Schema",
// "attributes": [
// 	{
//  	"@id":  "attr1" ,
//   "@type": "Value",
// 	"type":"string"
//  }
//  ]
// }`)
// 	ovl := expand(t, `{
// "@context": "../../schemas/ls.jsonld",
// "@type":"Overlay",
// "attributes": [
// 	{
//  	"@id":  "attr1" ,
//   "@type": "Value",
// 	"type":"int"
//  }
//  ]
// }`)
// 	baseattr := getLayer(base)
// 	ovlattr := getLayer(ovl)
// 	err := baseattr.Compose(ComposeOptions{}, Terms, ovlattr)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	base = baseattr.MarshalExpanded().([]interface{})
// 	t.Logf("%+v", base[0])
// 	//attrBase := base[0].(map[string]interface{})[LayerTerms.Attributes.GetTerm()]
// 	item := baseattr.Index["attr1"]
// 	if GetNodeValue(item.Values[AttributeAnnotations.Type.GetTerm()].([]interface{})[0]) != "int" {
// 		t.Errorf("Expecting int")
// 	}
// 	if len(item.Values[AttributeAnnotations.Type.GetTerm()].([]interface{})) != 1 {
// 		t.Errorf("Expecting 1 elements")
// 	}
// }
