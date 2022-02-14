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

	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

func TestLDMarshal(t *testing.T) {
	schText := `{
"@context": "../../schemas/ls.json",
"@id":"http://1",
"@type": "Schema",
"layer" :{
  "@type": "Object",
  "attributes": {
    "a1": {
      "@type": "Value",
      "lookupTable": {
        "elements": [
           {
              "options": ["a"],
              "value":"a"
           },
           {  
              "options": ["b","c"],
              "value":"b"
           }
        ]
      }
    }
  }
}
}`
	var v interface{}
	if err := json.Unmarshal([]byte(schText), &v); err != nil {
		t.Error(err)
		return
	}
	layer, err := UnmarshalLayer(v, nil)
	if err != nil {
		t.Error(err)
		return
	}
	attr, _ := layer.FindAttributeByID("a1")
	tableRoot := graph.NextNodesWith(attr, LookupTableTerm)[0]
	t.Log(tableRoot)
	// Expect two elements
	elements := graph.NextNodesWith(tableRoot, LookupTableElementsTerm)
	if len(elements) != 2 {
		t.Errorf("Expecting 2 elements: %d", len(elements))
	}
	SortNodes(elements)
	if AsPropertyValue(elements[0].GetProperty(LookupTableElementOptionsTerm)).AsStringSlice()[0] != "a" {
		t.Errorf("Wrong options: %v", elements[0])
	}
	if AsPropertyValue(elements[0].GetProperty(LookupTableElementValueTerm)).AsString() != "a" {
		t.Errorf("Wrong value: %v", elements[0])
	}
	if AsPropertyValue(elements[1].GetProperty(LookupTableElementOptionsTerm)).AsStringSlice()[0] != "b" ||
		AsPropertyValue(elements[1].GetProperty(LookupTableElementOptionsTerm)).AsStringSlice()[1] != "c" {
		t.Errorf("Wrong options: %v", elements[1])
	}
	if AsPropertyValue(elements[1].GetProperty(LookupTableElementValueTerm)).AsString() != "b" {
		t.Errorf("Wrong value: %v", elements[1])
	}
}

func TestLDMarshalTables(t *testing.T) {
	schText := `{
"@context": "../../schemas/ls.json",
"@id":"http://1",
"@type": "Schema",
"lookupTable": [
  {
    "@id": "http://tbl1",
        "elements": [
           {
              "options": ["a"],
              "value":"a"
           },
           {  
              "options": ["b","c"],
              "value":"b"
           }
        ]
  }
],
"layer" :{
  "@type": "Object",
  "attributes": {
    "a1": {
      "@type": "Value",
      "lookupTable": {
        "ref": "http://tbl1"
      }
    }
  }
}
}`
	var v interface{}
	if err := json.Unmarshal([]byte(schText), &v); err != nil {
		t.Error(err)
		return
	}
	layer, err := UnmarshalLayer(v, nil)
	if err != nil {
		t.Error(err)
		return
	}
	attr, _ := layer.FindAttributeByID("a1")
	tableRoot := graph.NextNodesWith(attr, LookupTableTerm)[0]
	t.Log(tableRoot)
	// Expect two elements
	elements := graph.NextNodesWith(tableRoot, LookupTableElementsTerm)
	if len(elements) != 2 {
		t.Errorf("Expecting 2 elements: %d", len(elements))
	}
	SortNodes(elements)
	if AsPropertyValue(elements[0].GetProperty(LookupTableElementOptionsTerm)).AsStringSlice()[0] != "a" {
		t.Errorf("Wrong options: %v", elements[0])
	}
	if AsPropertyValue(elements[0].GetProperty(LookupTableElementValueTerm)).AsString() != "a" {
		t.Errorf("Wrong value: %v", elements[0])
	}
	if AsPropertyValue(elements[1].GetProperty(LookupTableElementOptionsTerm)).AsStringSlice()[0] != "b" ||
		AsPropertyValue(elements[1].GetProperty(LookupTableElementOptionsTerm)).AsStringSlice()[1] != "c" {
		t.Errorf("Wrong options: %v", elements[1])
	}
	if AsPropertyValue(elements[1].GetProperty(LookupTableElementValueTerm)).AsString() != "b" {
		t.Errorf("Wrong value: %v", elements[1])
	}
}

func TestLDMarshalExt(t *testing.T) {
	schText := `{
"@context": "../../schemas/ls.json",
"@id":"http://1",
"@type": "Schema",
"layer" :{
  "@type": "Object",
  "attributes": {
    "a1": {
      "@type": "Value",
      "lookupTable": {
        "ref": "http://tbl1"
      }
    }
  }
}
}`
	var v interface{}
	if err := json.Unmarshal([]byte(schText), &v); err != nil {
		t.Error(err)
		return
	}
	layer, err := UnmarshalLayer(v, nil)
	if err != nil {
		t.Error(err)
		return
	}
	attr, _ := layer.FindAttributeByID("a1")
	tableRoot := graph.NextNodesWith(attr, LookupTableTerm)[0]
	t.Log(tableRoot)
	// Expect no elements
	elements := graph.NextNodesWith(tableRoot, LookupTableElementsTerm)
	if len(elements) != 0 {
		t.Errorf("Expecting 0 elements: %d", len(elements))
	}
	if GetNodeID(tableRoot) != "http://tbl1" {
		t.Errorf("Wrong ID: %+v", tableRoot)
	}
}
