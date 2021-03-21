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
	"encoding/json"
	"testing"

	"github.com/piprate/json-gold/ld"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
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

func TestIngestFlat(t *testing.T) {
	schStr := `{
 "@context": "../../schemas/layers.jsonld",
 "@type": "SchemaBase",
 "attributes": {
   "id1": {
     "attributeName":"field1"
   },
   "id2": {
     "attributeName":"field2",
     "required": true
   },
   "id3": {
     "attributeName": "field3"
   },
   "id4": {
    "attributeName":"field4" 
  }
 }
}`
	inputStr := `{
  "field1": "value1",
  "field2": 2,
  "field3": true,
  "field4": null,
  "field5": "extra"
}`

	inputStr2 := `{
  "field1": "value1",
  "field3": true,
  "field4": null,
  "field5": "extra"
}`

	var input map[string]interface{}
	if err := json.Unmarshal([]byte(inputStr), &input); err != nil {
		t.Fatal(err)
	}
	schema, err := ls.NewSchemaLayer(expand(t, schStr))
	if err != nil {
		t.Error(err)
	}
	schemastr, _ := json.MarshalIndent(schema.MarshalExpanded(), "", "  ")
	t.Logf("Schema: %s", string(schemastr))

	merged, err := Ingest("http://base", input, schema)
	if err != nil {
		t.Error(err)
	}
	output := merged.ToMap()
	t.Log(output)
	processor := ld.NewJsonLdProcessor()
	output, err = processor.Compact(output, nil, nil)
	if err != nil {
		t.Error(err)
	}
	{
		x, _ := json.MarshalIndent(output, "", "  ")
		t.Log(string(x))
	}

	attributes := output.(map[string]interface{})[ls.DocTerms.Attributes.ID].(map[string]interface{})
	if attributes["http://base.field1"].(map[string]interface{})[ls.DocTerms.Value.ID] != "value1" {
		t.Errorf("%+v", attributes)
	}
	if attributes["http://base.field1"].(map[string]interface{})[ls.AttributeAnnotations.Name.ID] != "field1" {
		t.Errorf("%+v", attributes)
	}
	if attributes["http://base.field2"].(map[string]interface{})[ls.DocTerms.Value.ID] != float64(2) {
		t.Errorf("%+v", attributes)
	}
	if attributes["http://base.field2"].(map[string]interface{})[ls.AttributeAnnotations.Name.ID] != "field2" {
		t.Errorf("%+v", attributes)
	}
	if attributes["http://base.field2"].(map[string]interface{})[ls.AttributeAnnotations.Required.ID] != true {
		t.Errorf("%+v", attributes)
	}

	input = nil
	if err := json.Unmarshal([]byte(inputStr2), &input); err != nil {
		t.Fatal(err)
	}

	output, err = Ingest("base", input, schema)
	if err == nil {
		t.Errorf("Validation error expected")
	}

	t.Log(output)
}

func TestIngestObject(t *testing.T) {
	schStr := `{
 "@context": "../../schemas/layers.jsonld",
 "@type":"SchemaBase",
 "attributes": {
   "id1": {
     "attributeName":"field1"
   },
   "id2": {
     "attributeName":"field2",
     "attributes": {
        "id3": {
           "attributeName": "field3"
        }
     }
   }
 }
}`
	inputStr := `{
  "field1": "value1",
  "field2": { "field3": "x"}
}`

	inputStr2 := `{
  "field1": "value1",
  "field2": true
}`

	var input map[string]interface{}
	if err := json.Unmarshal([]byte(inputStr), &input); err != nil {
		t.Fatal(err)
	}
	schema, err := ls.NewSchemaLayer(expand(t, schStr))
	if err != nil {
		t.Error(err)
	}
	merged, err := Ingest("http://base", input, schema)
	if err != nil {
		t.Error(err)
	}
	t.Log(merged)
	output := merged.ToMap()

	t.Log(output)
	processor := ld.NewJsonLdProcessor()
	output, err = processor.Compact(output, nil, nil)
	if err != nil {
		t.Error(err)
	}
	{
		x, _ := json.MarshalIndent(output, "", "  ")
		t.Log(string(x))
	}

	attributes := output.(map[string]interface{})[ls.DocTerms.Attributes.ID].(map[string]interface{})
	if attributes["http://base.field2"].(map[string]interface{})[ls.AttributeAnnotations.Name.ID] != "field2" {
		t.Errorf("%+v", attributes)
	}
	nested := attributes["http://base.field2"].(map[string]interface{})[ls.DocTerms.Attributes.ID].(map[string]interface{})
	if nested["http://base.field2.field3"].(map[string]interface{})[ls.DocTerms.Value.ID] != "x" {
		t.Errorf("%v", nested)
	}

	input = nil
	if err := json.Unmarshal([]byte(inputStr2), &input); err != nil {
		t.Fatal(err)
	}

	output, err = Ingest("base", input, schema)
	if err == nil {
		t.Errorf("Validation error expected")
	}
}

func TestIngestArray(t *testing.T) {
	schStr := `{
 "@context": "../../schemas/layers.jsonld",
 "@type":"SchemaBase",
 "attributes": {
   "id2": {
     "attributeName":"field2",
     "arrayItems": {
       "type": "string"
     }
   }
 }
}`
	inputStr := `{
  "field2": ["a","b"]
}`

	var input map[string]interface{}
	if err := json.Unmarshal([]byte(inputStr), &input); err != nil {
		t.Fatal(err)
	}
	schema, err := ls.NewSchemaLayer(expand(t, schStr))
	if err != nil {
		t.Error(err)
	}
	merged, err := Ingest("http://base", input, schema)
	if err != nil {
		t.Error(err)
	}
	t.Log(merged)
	output := merged.ToMap()

	t.Log(output)
	processor := ld.NewJsonLdProcessor()
	output, err = processor.Compact(output, nil, nil)
	if err != nil {
		t.Error(err)
	}
	{
		x, _ := json.MarshalIndent(output, "", "  ")
		t.Log(string(x))
	}

	attributes := output.(map[string]interface{})[ls.DocTerms.Attributes.ID].(map[string]interface{})
	t.Log(attributes)
	list := attributes["http://base.field2"].(map[string]interface{})[ls.DocTerms.ArrayElements.ID].(map[string]interface{})["@list"].([]interface{})
	if list[0].(map[string]interface{})[ls.AttributeAnnotations.Type.ID] != "string" {
		t.Errorf("%v", list[0])
	}
	if list[0].(map[string]interface{})[ls.DocTerms.Value.ID] != "a" {
		t.Errorf("%v", list[0])
	}
	if list[1].(map[string]interface{})[ls.AttributeAnnotations.Type.ID] != "string" {
		t.Errorf("%v", list[1])
	}
	if list[1].(map[string]interface{})[ls.DocTerms.Value.ID] != "b" {
		t.Errorf("%v", list[0])
	}
}

func TestIngestObjArray(t *testing.T) {
	schStr := `{
 "@context": "../../schemas/layers.jsonld",
 "@type":"SchemaBase",
 "attributes": {
   "id2": {
     "attributeName":"field2",
     "arrayItems": {
       "attributes": {
          "id3": {
            "type":"string"
          }
       }
     }
   }
 }
}`
	inputStr := `{
  "field2": [{"id3":"a"},{"id3":"b"}]
}`

	var input map[string]interface{}
	if err := json.Unmarshal([]byte(inputStr), &input); err != nil {
		t.Fatal(err)
	}
	schema, err := ls.NewSchemaLayer(expand(t, schStr))
	if err != nil {
		t.Error(err)
	}
	schemastr, _ := json.MarshalIndent(schema.MarshalExpanded(), "", "  ")
	t.Logf("Schema: %s", string(schemastr))
	merged, err := Ingest("http://base", input, schema)
	if err != nil {
		t.Error(err)
	}
	t.Log(merged)
	output := merged.ToMap()

	t.Log(output)
	processor := ld.NewJsonLdProcessor()
	output, err = processor.Compact(output, nil, nil)
	if err != nil {
		t.Error(err)
	}
	{
		x, _ := json.MarshalIndent(output, "", "  ")
		t.Log(string(x))
	}

	attributes := output.(map[string]interface{})[ls.DocTerms.Attributes.ID].(map[string]interface{})
	t.Log(attributes)
	list := attributes["http://base.field2"].(map[string]interface{})[ls.DocTerms.ArrayElements.ID].(map[string]interface{})["@list"].([]interface{})
	if list[0].(map[string]interface{})[ls.DocTerms.Attributes.ID].(map[string]interface{})["http://base.field2.0.id3"].(map[string]interface{})[ls.AttributeAnnotations.Type.ID] != "string" {
		t.Errorf("%v", list[0])
	}
	if list[0].(map[string]interface{})[ls.DocTerms.Attributes.ID].(map[string]interface{})["http://base.field2.0.id3"].(map[string]interface{})[ls.DocTerms.Value.ID] != "a" {
		t.Errorf("%v", list[0])
	}
	if list[1].(map[string]interface{})[ls.DocTerms.Attributes.ID].(map[string]interface{})["http://base.field2.1.id3"].(map[string]interface{})[ls.AttributeAnnotations.Type.ID] != "string" {
		t.Errorf("%v", list[1])
	}
	if list[1].(map[string]interface{})[ls.DocTerms.Attributes.ID].(map[string]interface{})["http://base.field2.1.id3"].(map[string]interface{})[ls.DocTerms.Value.ID] != "b" {
		t.Errorf("%v", list[0])
	}
}
