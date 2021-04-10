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
	"testing"

	"github.com/piprate/json-gold/ld"
)

func expand(t *testing.T, in string) []interface{} {
	proc := ld.NewJsonLdProcessor()
	var v interface{}
	if err := json.Unmarshal([]byte(in), &v); err != nil {
		t.Error(err)
		t.Fail()
	}
	ret, err := proc.Expand(v, nil)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	return ret
}

func getLayer(in interface{}) *Layer {
	a := &Layer{}
	if err := a.UnmarshalExpanded(in); err != nil {
		panic(err)
	}
	return a
}

func TestMerge1(t *testing.T) {
	base := expand(t, `{
"@context": "../../schemas/ls.jsonld",
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
}`)
	t.Logf("Base: %+v", base)
	ovl := expand(t, `{
"@context": ["../../schemas/ls.jsonld",{"@vocab":"http://test/"}],
"@type":"Overlay",
"attributes":[
{
  "@id":"attr1",
   "someKey" : [
     {"@value": "someValue"}
   ]
},
{
  "@id":"attr2",
   "privacyClassification" : [
     {"@value": "addFlg1"}
   ]
},
{
  "@id": "attr3",
   "privacyClassification" : [
     {"@value": "addFlg2"},
     {"@value": "addFlg3"}
   ]
}
]
}`)
	t.Logf("Ovl: %v", ovl)
	baseattr := getLayer(base)
	ovlattr := getLayer(ovl)
	err := baseattr.Compose(ComposeOptions{}, Terms, ovlattr)
	if err != nil {
		t.Error(err)
	}
	base = baseattr.MarshalExpanded().([]interface{})
	t.Logf("%+v", base[0])
	//	attrBase := base[0].(map[string]interface{})[LayerTerms.Attributes.GetTerm()]

	attr1 := baseattr.Index["attr1"]
	if attr1 == nil {
		t.Errorf("No attr1")
		return
	}
	sk, ok := attr1.Values["http://test/someKey"]
	if !ok {
		t.Errorf("Missing someKey")
		return
	}
	if sk.([]interface{})[0].(map[string]interface{})["@value"] != "someValue" {
		t.Errorf("Wrong value: %v", sk)
	}

	attr2 := baseattr.Index["attr2"]
	if attr2 == nil {
		t.Errorf("No attr2")
		return
	}
	priv := attr2.Values[AttributeAnnotations.Privacy.GetTerm()].([]interface{})
	if GetNodeValue(priv[0]) != "flg1" || GetNodeValue(priv[1]) != "addFlg1" {
		t.Errorf("Wrong flags: %v", priv)
	}

	attr3 := baseattr.Index["attr3"]
	if attr3 == nil {
		t.Errorf("No attr3")
		return
	}
	priv = attr3.Values[AttributeAnnotations.Privacy.GetTerm()].([]interface{})
	if GetNodeValue(priv[0]) != "flg2" || GetNodeValue(priv[1]) != "flg3" || GetNodeValue(priv[2]) != "addFlg2" || GetNodeValue(priv[3]) != "addFlg3" {
		t.Errorf("Wrong flags: %v", priv)
	}
}

func TestMergeArray(t *testing.T) {
	base := expand(t, `{
  "@context":"../../schemas/ls.jsonld",
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
}`)
	ovl := expand(t, `{
  "@context": "../../schemas/ls.jsonld",
  "@type": "Overlay",
  "attributes": {
    "array": {
     "@type": "Array",
      "items": {
       "@id": "http://items",
       "type":"string"
      }
    }
  }
}`)
	t.Logf("Base: %v", base)
	t.Logf("Ovl: %v", ovl)
	baseattr := getLayer(base)
	ovlattr := getLayer(ovl)
	err := baseattr.Compose(ComposeOptions{}, Terms, ovlattr)
	if err != nil {
		t.Error(err)
	}
	base = baseattr.MarshalExpanded().([]interface{})
	//	attrBase := base[0].(map[string]interface{})[LayerTerms.Attributes.GetTerm()]

	array := baseattr.Index["array"]
	if array == nil {
		t.Errorf("No array")
		return
	}
	items := array.Type.(*ArrayType).Values
	if s := GetNodeValue(items[AttributeAnnotations.Type.GetTerm()].([]interface{})[0]); s != "string" {
		t.Errorf("Missing type: %+v %s", items, s)
	}
}

func TestMergeChoice(t *testing.T) {
	base := expand(t, `{
  "@context": "../../schemas/ls.jsonld",
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
}`)
	ovl := expand(t, `{
  "@context": "../../schemas/ls.jsonld",
  "@type": "Overlay",
  "attributes": {
   "attr": {
    "@type": "Polymorphic",
	   "oneOf": [
	    {
        "@id": "id1",
        "@type":"Value",
        "type": "string"
	    }
	   ]
	 }
  }
}`)
	t.Logf("Ovl: %v", ovl)
	baseattr := getLayer(base)
	ovlattr := getLayer(ovl)
	err := baseattr.Compose(ComposeOptions{}, Terms, ovlattr)
	if err != nil {
		t.Error(err)
	}
	base = baseattr.MarshalExpanded().([]interface{})
	//	attrBase := base[0].(map[string]interface{})[LayerTerms.Attributes.GetTerm()]

	item := baseattr.Index["id1"]
	if item == nil {
		t.Errorf("item not found")
	}
	t.Logf("%+v", item)
	if GetNodeValue(item.Values[AttributeAnnotations.Type.GetTerm()].([]interface{})[0]) != "string" {
		t.Errorf("Wrong value: %+v", item)
	}
}

func TestOverride(t *testing.T) {
	base := expand(t, `{
"@context": "../../schemas/ls.jsonld",
"@type":"Schema",
"attributes": [
	{
 	"@id":  "attr1" ,
  "@type": "Value",
	"type":"string"
 }
 ]
}`)
	ovl := expand(t, `{
"@context": "../../schemas/ls.jsonld",
"@type":"Overlay",
"attributes": [
	{
 	"@id":  "attr1" ,
  "@type": "Value",
	"type":"int"
 }
 ]
}`)
	baseattr := getLayer(base)
	ovlattr := getLayer(ovl)
	err := baseattr.Compose(ComposeOptions{}, Terms, ovlattr)
	if err != nil {
		t.Error(err)
	}
	base = baseattr.MarshalExpanded().([]interface{})
	t.Logf("%+v", base[0])
	//attrBase := base[0].(map[string]interface{})[LayerTerms.Attributes.GetTerm()]
	item := baseattr.Index["attr1"]
	if GetNodeValue(item.Values[AttributeAnnotations.Type.GetTerm()].([]interface{})[0]) != "int" {
		t.Errorf("Expecting int")
	}
	if len(item.Values[AttributeAnnotations.Type.GetTerm()].([]interface{})) != 1 {
		t.Errorf("Expecting 1 elements")
	}
}
