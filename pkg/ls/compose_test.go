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

func getLayer(in interface{}) *SchemaLayer {
	a := &SchemaLayer{}
	if err := a.UnmarshalExpanded(in); err != nil {
		panic(err)
	}
	return a
}

func TestMerge1(t *testing.T) {
	base := expand(t, `{
"@context": "../../schemas/layers.jsonld",
"@type":"Layer",
"attributes": [
{
  "@id":  "attr1" 
},
{
  "@id":  "attr2" ,
  "privacyClassification": [
    {
      "@value": "flg1"
    }
  ]
},
{
  "@id":"attr3",
  "privacyClassification": [
      {"@value": "flg2"},
      {"@value": "flg3"}
   ]
 }
]
}`)
	t.Logf("Base: %+v", base)
	ovl := expand(t, `{
"@context": ["../../schemas/layers.jsonld",{"@vocab":"http://test/"}],
"@type":"Layer",
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
	err := baseattr.Attributes.Compose(ComposeOptions{}, &ovlattr.Attributes)
	if err != nil {
		t.Error(err)
	}
	base = baseattr.MarshalExpanded().([]interface{})
	t.Logf("%+v", base[0])
	attrBase := base[0].(map[string]interface{})[AttributeStructure.Attributes.ID]

	attr1Arr := FindNodeByID(attrBase, "attr1")
	if attr1Arr == nil {
		t.Errorf("No attr1")
		return
	}
	attr1 := attr1Arr[0].(map[string]interface{})
	sk, ok := attr1["http://test/someKey"]
	if !ok {
		t.Errorf("Missing someKey")
		return
	}
	if sk.([]interface{})[0].(map[string]interface{})["@value"] != "someValue" {
		t.Errorf("Wrong value: %v", sk)
	}

	attr2Arr := FindNodeByID(attrBase, "attr2")
	if attr2Arr == nil {
		t.Errorf("No attr2")
		return
	}
	priv := attr2Arr[0].(map[string]interface{})[AttributeAnnotations.Privacy.ID].([]interface{})
	if GetNodeValue(priv[0]) != "flg1" || GetNodeValue(priv[1]) != "addFlg1" {
		t.Errorf("Wrong flags: %v", priv)
	}

	attr3Arr := FindNodeByID(attrBase, "attr3")
	if attr3Arr == nil {
		t.Errorf("No attr3")
		return
	}
	priv = attr3Arr[0].(map[string]interface{})[AttributeAnnotations.Privacy.ID].([]interface{})
	if GetNodeValue(priv[0]) != "flg2" || GetNodeValue(priv[1]) != "flg3" || GetNodeValue(priv[2]) != "addFlg2" || GetNodeValue(priv[3]) != "addFlg3" {
		t.Errorf("Wrong flags: %v", priv)
	}
}

func TestMergeArray(t *testing.T) {
	base := expand(t, `{
  "@context":"../../schemas/layers.jsonld",
  "@type": "Layer",
  "attributes": {
    "array": {
      "arrayItems":  {
      }
    }
  }
}`)
	ovl := expand(t, `{
  "@context": "../../schemas/layers.jsonld",
  "@type": "Layer",
  "attributes": {
    "array": {
      "arrayItems": {
       "type":"string"
      }
    }
  }
}`)
	t.Logf("Ovl: %v", ovl)
	baseattr := getLayer(base)
	ovlattr := getLayer(ovl)
	err := baseattr.Attributes.Compose(ComposeOptions{}, &ovlattr.Attributes)
	if err != nil {
		t.Error(err)
	}
	base = baseattr.MarshalExpanded().([]interface{})
	attrBase := base[0].(map[string]interface{})[AttributeStructure.Attributes.ID]

	array := FindNodeByID(attrBase, "array")
	if array == nil {
		t.Errorf("No array")
		return
	}
	items := array[0].(map[string]interface{})[AttributeStructure.ArrayItems.ID].([]interface{})[0].(map[string]interface{})
	if s := GetNodeValue(items[AttributeAnnotations.Type.ID].([]interface{})[0]); s != "string" {
		t.Errorf("Missing type: %+v %s", items, s)
	}
}

func TestMergeChoice(t *testing.T) {
	base := expand(t, `{
  "@context": "../../schemas/layers.jsonld",
  "@type" : "Layer",
  "attributes": {
   "attr": {
	   "oneOf": [
	    {
        "@id": "id1",
        "reference": "ref1"
	    }
	   ]
	 }
  }
}`)
	ovl := expand(t, `{
  "@context": "../../schemas/layers.jsonld",
  "@type": "Layer",
  "attributes": {
   "attr": {
	   "oneOf": [
	    {
        "@id": "id1",
        "type": "string"
	    }
	   ]
	 }
  }
}`)
	t.Logf("Ovl: %v", ovl)
	baseattr := getLayer(base)
	ovlattr := getLayer(ovl)
	err := baseattr.Attributes.Compose(ComposeOptions{}, &ovlattr.Attributes)
	if err != nil {
		t.Error(err)
	}
	base = baseattr.MarshalExpanded().([]interface{})
	attrBase := base[0].(map[string]interface{})[AttributeStructure.Attributes.ID]

	item := FindNodeByID(attrBase, "id1")
	if item == nil {
		t.Errorf("item not found")
	}
	t.Logf("%+v", item[0])
	if GetNodeValue(item[0].(map[string]interface{})[AttributeAnnotations.Type.ID].([]interface{})[0]) != "string" {
		t.Errorf("Wrong value: %+v", item)
	}
}

func TestOverride(t *testing.T) {
	base := expand(t, `{
"@context": "../../schemas/layers.jsonld",
"@type":"Layer",
"attributes": [
	{
 	"@id":  "attr1" ,
	"type":"string"
 }
 ]
}`)
	ovl := expand(t, `{
"@context": "../../schemas/layers.jsonld",
"@type":"Layer",
"attributes": [
	{
 	"@id":  "attr1" ,
	"type":"int"
 }
 ]
}`)
	baseattr := getLayer(base)
	ovlattr := getLayer(ovl)
	err := baseattr.Attributes.Compose(ComposeOptions{}, &ovlattr.Attributes)
	if err != nil {
		t.Error(err)
	}
	base = baseattr.MarshalExpanded().([]interface{})
	t.Logf("%+v", base[0])
	attrBase := base[0].(map[string]interface{})[AttributeStructure.Attributes.ID]
	item := FindNodeByID(attrBase, "attr1")
	if GetNodeValue(item[0].(map[string]interface{})[AttributeAnnotations.Type.ID].([]interface{})[0]) != "int" {
		t.Errorf("Expecting int")
	}
	if len(item[0].(map[string]interface{})[AttributeAnnotations.Type.ID].([]interface{})) != 1 {
		t.Errorf("Expecting 1 elements")
	}
}
