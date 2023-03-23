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
	"strings"
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/json/jsonschema"
	"github.com/piprate/json-gold/ld"

	"github.com/cloudprivacylabs/lpg"
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
 "@context": "../../schemas/ls.json",
 "@id":"http://example.org/id",
 "@type": "Schema",
 "layer": {
  "@type": "Object",
  "@id": "root",
  "required": [ "id2"],
 "attributes": {
   "id1": {
    "@type": "Value",
     "attributeName":"field1"
   },
   "id2": {
    "@type": "Value",
     "attributeName":"field2"
   },
   "id3": {
    "@type": "Value",
     "attributeName": "field3"
   },
   "id4": {
    "@type": "Value",
    "attributeName":"field4"
  }
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

	var schMap interface{}
	if err := json.Unmarshal([]byte(schStr), &schMap); err != nil {
		t.Fatal(err)
	}
	schema, err := ls.UnmarshalLayer(schMap, nil)
	if err != nil {
		t.Error(err)
	}
	// strtable
	bldr := ls.NewGraphBuilder(nil, ls.GraphBuilderOptions{
		OnlySchemaAttributes: false,
	})
	parser := Parser{
		Layer: schema,
	}
	_, err = IngestBytes(ls.DefaultContext(), "http://base", []byte(inputStr), parser, bldr, &ls.Ingester{Schema: schema})
	if err != nil {
		t.Error(err)
	}

	findNodes := func(nodeId string) []*lpg.Node {
		nodes := []*lpg.Node{}
		for nx := bldr.GetGraph().GetNodes(); nx.Next(); {
			node := nx.Node()
			if ls.GetNodeID(node) == nodeId {
				nodes = append(nodes, node)
			}
		}
		return nodes
	}

	checkNodeValue := func(nodeId string, expected interface{}) {
		nodes := findNodes(nodeId)
		if len(nodes) == 0 {
			t.Errorf("node not found: %s", nodeId)
		}
		s, ok := ls.GetRawNodeValue(nodes[0])
		if (expected == nil && ok) ||
			(expected != nil && !ok) ||
			(expected != nil && s != expected) {
			t.Errorf("Wrong value for %s: %s", nodeId, s)
		}
	}
	checkNodeValue("http://base.field1", "value1")
	checkNodeValue("http://base.field2", "2")
	checkNodeValue("http://base.field3", "true")
	checkNodeValue("http://base.field5", "extra")

	bldr = ls.NewGraphBuilder(nil, ls.GraphBuilderOptions{
		OnlySchemaAttributes: true,
	})
	parser = Parser{
		Layer: schema,
	}
	_, err = IngestBytes(ls.DefaultContext(), "http://base", []byte(inputStr), parser, bldr, &ls.Ingester{Schema: schema})
	if err != nil {
		t.Error(err)
	}

	parser.IngestNullValues = true
	_, err = IngestBytes(ls.DefaultContext(), "http://base", []byte(inputStr), parser, bldr, &ls.Ingester{Schema: schema})
	if err != nil {
		t.Error(err)
	}
	checkNodeValue("http://base.field1", "value1")
	checkNodeValue("http://base.field2", "2")
	checkNodeValue("http://base.field3", "true")
	checkNodeValue("http://base.field4", "")

	if len(findNodes("http://base.field5")) != 0 {
		t.Errorf("Unexpected node found")
	}

}

func TestIngestPoly(t *testing.T) {
	schStr := `{
 "@context": "../../schemas/ls.json",
 "@id":"http://example.org/id",
 "@type": "Schema",
 "layer": {
  "@type": "Object",
  "@id": "root",
 "attributes": {
   "id1": {
     "@type": "Value",
     "attributeName":"field1"
   },
   "id2": {
     "@type": "Polymorphic",
     "attributeName":"field2",
     "oneOf": [
       {
         "@id": "option1",
         "@type": "Object",
         "attributes": {
           "objType1ss": {
             "@type": "Value",
             "@id": "objType1",
             "attributeName": "t",
             "enumeration": "type1"
           }
         }
       },
       {
         "@id": "option2",
         "@type": "Object",
         "attributes": {
           "objType2": {
             "@type": "Value",
             "@id": "objType2",
             "attributeName": "t",
             "enumeration": "type2"
           }
         }
       }
     ]
   }
  }
 }
}`
	inputStr := `{
  "field1": "value1",
  "field2": {
     "t": "type1"
  }
}`

	var schMap interface{}
	if err := json.Unmarshal([]byte(schStr), &schMap); err != nil {
		t.Fatal(err)
	}
	schema, err := ls.UnmarshalLayer(schMap, nil)
	if err != nil {
		t.Error(err)
	}
	bldr := ls.NewGraphBuilder(nil, ls.GraphBuilderOptions{
		OnlySchemaAttributes: false,
	})
	parser := Parser{
		Layer: schema,
	}
	_, err = IngestBytes(ls.DefaultContext(), "http://base", []byte(inputStr), parser, bldr, &ls.Ingester{Schema: schema})
	if err != nil {
		t.Error(err)
	}

	findNodes := func(nodeId string) []*lpg.Node {
		nodes := []*lpg.Node{}
		for nx := bldr.GetGraph().GetNodes(); nx.Next(); {
			node := nx.Node()
			if ls.GetNodeID(node) == nodeId {
				nodes = append(nodes, node)
			}
		}
		return nodes
	}

	nodes := findNodes("objType1")
	t.Logf("%+v", nodes)
	if len(nodes) != 1 {
		t.Errorf("Expecting 1 type node")
	}
}

func TestIngestRootAnnotation(t *testing.T) {
	schStr := `{
   "definitions": {
      "a": {
         "type": "object",
         "x-ls": {
            "https://consentgrid.com/SmartConsent": "test"
         },
         "properties": {
            "field1": {"type": "number"},
           "field2":  {"type": "string"}
      }
   }
  }
}`
	inputStr := `{
  "field1": 1,
  "field2": "2"
}`

	compiler := jsonschema.NewCompiler()
	compiler.AddResource("http://test.json", strings.NewReader(schStr))
	compiled, err := CompileEntitiesWith(compiler, Entity{Ref: "http://test.json#/definitions/a", LayerID: "lid"})
	if err != nil {
		t.Error(err)
		return
	}
	layers, err := BuildEntityGraph(lpg.NewGraph(), ls.SchemaTerm, LinkRefsBySchemaRef, compiled...)
	if err != nil {
		t.Error(err)
		return
	}

	bldr := ls.NewGraphBuilder(nil, ls.GraphBuilderOptions{
		OnlySchemaAttributes: false,
	})
	parser := Parser{
		Layer: layers[0].Layer,
	}
	_, err = IngestBytes(ls.DefaultContext(), "http://base", []byte(inputStr), parser, bldr, &ls.Ingester{Schema: layers[0].Layer})
	if err != nil {
		t.Error(err)
	}

	findNodes := func(nodeId string) []*lpg.Node {
		nodes := []*lpg.Node{}
		for nx := bldr.GetGraph().GetNodes(); nx.Next(); {
			node := nx.Node()
			t.Logf("%s", ls.GetNodeID(node))
			if ls.GetNodeID(node) == nodeId {
				nodes = append(nodes, node)
			}
		}
		return nodes
	}
	nodes := findNodes("http://base")
	t.Logf("%+v", nodes[0])
}

// func TestIngestObject(t *testing.T) {
// 	schStr := `{
//  "@context": "../../schemas/layers.jsonld",
//  "@type":"SchemaBase",
//  "attributes": {
//    "id1": {
//      "attributeName":"field1"
//    },
//    "id2": {
//      "attributeName":"field2",
//      "attributes": {
//         "id3": {
//            "attributeName": "field3"
//         }
//      }
//    }
//  }
// }`
// 	inputStr := `{
//   "field1": "value1",
//   "field2": { "field3": "x"}
// }`

// 	inputStr2 := `{
//   "field1": "value1",
//   "field2": true
// }`

// 	var input map[string]interface{}
// 	if err := json.Unmarshal([]byte(inputStr), &input); err != nil {
// 		t.Fatal(err)
// 	}
// 	schema, err := ls.NewSchemaLayer(expand(t, schStr))
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	merged, err := Ingest("http://base", input, schema)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	t.Log(merged)
// 	output := merged.ToMap()

// 	t.Log(output)
// 	processor := ld.NewJsonLdProcessor()
// 	output, err = processor.Compact(output, nil, nil)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	{
// 		x, _ := json.MarshalIndent(output, "", "  ")
// 		t.Log(string(x))
// 	}

// 	attributes := output.(map[string]interface{})[ls.DocTerms.Attributes.ID].(map[string]interface{})
// 	if attributes["http://base.field2"].(map[string]interface{})[ls.AttributeAnnotations.Name.ID] != "field2" {
// 		t.Errorf("%+v", attributes)
// 	}
// 	nested := attributes["http://base.field2"].(map[string]interface{})[ls.DocTerms.Attributes.ID].(map[string]interface{})
// 	if nested["http://base.field2.field3"].(map[string]interface{})[ls.DocTerms.Value.ID] != "x" {
// 		t.Errorf("%v", nested)
// 	}

// 	input = nil
// 	if err := json.Unmarshal([]byte(inputStr2), &input); err != nil {
// 		t.Fatal(err)
// 	}

// 	output, err = Ingest("base", input, schema)
// 	if err == nil {
// 		t.Errorf("Validation error expected")
// 	}
// }

// func TestIngestArray(t *testing.T) {
// 	schStr := `{
//  "@context": "../../schemas/layers.jsonld",
//  "@type":"SchemaBase",
//  "attributes": {
//    "id2": {
//      "attributeName":"field2",
//      "arrayItems": {
//        "type": "string"
//      }
//    }
//  }
// }`
// 	inputStr := `{
//   "field2": ["a","b"]
// }`

// 	var input map[string]interface{}
// 	if err := json.Unmarshal([]byte(inputStr), &input); err != nil {
// 		t.Fatal(err)
// 	}
// 	schema, err := ls.NewSchemaLayer(expand(t, schStr))
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	merged, err := Ingest("http://base", input, schema)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	t.Log(merged)
// 	output := merged.ToMap()

// 	t.Log(output)
// 	processor := ld.NewJsonLdProcessor()
// 	output, err = processor.Compact(output, nil, nil)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	{
// 		x, _ := json.MarshalIndent(output, "", "  ")
// 		t.Log(string(x))
// 	}

// 	attributes := output.(map[string]interface{})[ls.DocTerms.Attributes.ID].(map[string]interface{})
// 	t.Log(attributes)
// 	list := attributes["http://base.field2"].(map[string]interface{})[ls.DocTerms.ArrayElements.ID].(map[string]interface{})["@list"].([]interface{})
// 	if list[0].(map[string]interface{})[ls.AttributeAnnotations.Type.ID] != "string" {
// 		t.Errorf("%v", list[0])
// 	}
// 	if list[0].(map[string]interface{})[ls.DocTerms.Value.ID] != "a" {
// 		t.Errorf("%v", list[0])
// 	}
// 	if list[1].(map[string]interface{})[ls.AttributeAnnotations.Type.ID] != "string" {
// 		t.Errorf("%v", list[1])
// 	}
// 	if list[1].(map[string]interface{})[ls.DocTerms.Value.ID] != "b" {
// 		t.Errorf("%v", list[0])
// 	}
// }

// func TestIngestObjArray(t *testing.T) {
// 	schStr := `{
//  "@context": "../../schemas/layers.jsonld",
//  "@type":"SchemaBase",
//  "attributes": {
//    "id2": {
//      "attributeName":"field2",
//      "arrayItems": {
//        "attributes": {
//           "id3": {
//             "type":"string"
//           }
//        }
//      }
//    }
//  }
// }`
// 	inputStr := `{
//   "field2": [{"id3":"a"},{"id3":"b"}]
// }`

// 	var input map[string]interface{}
// 	if err := json.Unmarshal([]byte(inputStr), &input); err != nil {
// 		t.Fatal(err)
// 	}
// 	schema, err := ls.NewSchemaLayer(expand(t, schStr))
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	schemastr, _ := json.MarshalIndent(schema.MarshalExpanded(), "", "  ")
// 	t.Logf("Schema: %s", string(schemastr))
// 	merged, err := Ingest("http://base", input, schema)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	t.Log(merged)
// 	output := merged.ToMap()

// 	t.Log(output)
// 	processor := ld.NewJsonLdProcessor()
// 	output, err = processor.Compact(output, nil, nil)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	{
// 		x, _ := json.MarshalIndent(output, "", "  ")
// 		t.Log(string(x))
// 	}

// 	attributes := output.(map[string]interface{})[ls.DocTerms.Attributes.ID].(map[string]interface{})
// 	t.Log(attributes)
// 	list := attributes["http://base.field2"].(map[string]interface{})[ls.DocTerms.ArrayElements.ID].(map[string]interface{})["@list"].([]interface{})
// 	if list[0].(map[string]interface{})[ls.DocTerms.Attributes.ID].(map[string]interface{})["http://base.field2.0.id3"].(map[string]interface{})[ls.AttributeAnnotations.Type.ID] != "string" {
// 		t.Errorf("%v", list[0])
// 	}
// 	if list[0].(map[string]interface{})[ls.DocTerms.Attributes.ID].(map[string]interface{})["http://base.field2.0.id3"].(map[string]interface{})[ls.DocTerms.Value.ID] != "a" {
// 		t.Errorf("%v", list[0])
// 	}
// 	if list[1].(map[string]interface{})[ls.DocTerms.Attributes.ID].(map[string]interface{})["http://base.field2.1.id3"].(map[string]interface{})[ls.AttributeAnnotations.Type.ID] != "string" {
// 		t.Errorf("%v", list[1])
// 	}
// 	if list[1].(map[string]interface{})[ls.DocTerms.Attributes.ID].(map[string]interface{})["http://base.field2.1.id3"].(map[string]interface{})[ls.DocTerms.Value.ID] != "b" {
// 		t.Errorf("%v", list[0])
// 	}
// }
