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
	//	"github.com/cloudprivacylabs/opencypher/graph"
)

func TestBasicVS(t *testing.T) {
	schText := `{
"@context": "../../schemas/ls.json",
"@id":"http://1",
"@type": "Schema",
"valueType": "test",
"layer" :{
  "@type": "Object",
 "@id": "http://schroot",
  "attributes": {
    "src": {
      "@type": "Value",
      "attributeName": "src"
    },
    "tgt": {
      "@type": "Value",
      "attributeName": "tgt",
      "https://lschema.org/vs/source":"src"
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

	ingester := Ingester{
		Schema:           layer,
		EmbedSchemaNodes: true,
		Graph:            NewDocumentGraph(),
		ValuesetFunc: func(req ValuesetLookupRequest) (ValuesetLookupResponse, error) {
			ret := ValuesetLookupResponse{
				KeyValues: map[string]string{"": "X"},
			}
			return ret, nil
		},
	}

	ctx := DefaultContext()
	ictx := ingester.Start(ctx, "http://base")
	_, _, rootNode, err := ingester.Object(ictx)
	if err != nil {
		t.Error(err)
		return
	}
	newLevel := ictx.NewLevel(rootNode)
	schNode := layer.GetAttributeByID("src")
	ingester.Value(newLevel.New("src", schNode), "a")
	// Graph must have 2 nodes
	if ingester.Graph.NumNodes() != 2 {
		t.Errorf("NumNodes: %d", ingester.Graph.NumNodes())
	}
	if err := ingester.Finish(ictx, rootNode); err != nil {
		t.Error(err)
		return
	}
	// Graph must have 3 nodes
	if ingester.Graph.NumNodes() != 3 {
		t.Errorf("NumNodes: %d", ingester.Graph.NumNodes())
	}

	nodes := FindChildInstanceOf(rootNode, "tgt")
	if len(nodes) != 1 {
		t.Errorf("Child nodes: %v", nodes)
	}

}

func TestStructuredVS(t *testing.T) {
	schText := `{
"@context": "../../schemas/ls.json",
"@id":"http://1",
"@type": "Schema",
"valueType": "test",
"layer" :{
  "@type": "Object",
 "@id": "http://schroot",
  "attributes": {
    "src": {
      "@type": "Object",
      "attributeName": "src",
      "attributes": {
        "code": {
          "@type": "Value",
          "attributeName": "code"
        },
        "system": {
          "@type": "Value",
          "attributeName": "system"
        }
      }
    },
    "tgt": {
      "@type": "Object",
      "attributeName": "tgt",
      "https://lschema.org/vs/source":"src",
      "https://lschema.org/vs/requestKeys": ["c","s"],
      "https://lschema.org/vs/requestValues": ["code","system"],
      "https://lschema.org/vs/resultKeys": ["tc","ts"],
      "https://lschema.org/vs/resultValues": ["tgtcode","tgtsystem"],
      "attributes": {
        "tgtcode": {
          "@type": "Value",
          "attributeName": "tgtcode"
        },
        "tgtsystem": {
          "@type": "Value",
          "attributeName": "tgtsystem"
        }
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

	ingester := Ingester{
		Schema:           layer,
		EmbedSchemaNodes: true,
		Graph:            NewDocumentGraph(),
		ValuesetFunc: func(req ValuesetLookupRequest) (ValuesetLookupResponse, error) {
			ret := ValuesetLookupResponse{}
			if req.KeyValues["c"] == "a" && req.KeyValues["s"] == "b" {
				ret.KeyValues = map[string]string{"tc": "aa", "ts": "bb"}
			}
			return ret, nil
		},
	}

	ctx := DefaultContext()
	ictx := ingester.Start(ctx, "http://base")
	_, _, rootNode, err := ingester.Object(ictx)
	if err != nil {
		t.Error(err)
		return
	}
	newLevel := ictx.NewLevel(rootNode)
	schNode := layer.GetAttributeByID("src")
	codeNode := layer.GetAttributeByID("code")
	systemNode := layer.GetAttributeByID("system")

	{
		ctx := ictx.New("src", schNode)
		_, _, srcNode, _ := ingester.Object(newLevel.New("src", schNode))
		ctx = ctx.NewLevel(srcNode)
		ingester.Value(ctx.New("code", codeNode), "a")
		ingester.Value(ctx.New("system", systemNode), "b")
	}

	// Graph must have 4 nodes
	if ingester.Graph.NumNodes() != 4 {
		t.Errorf("NumNodes: %d", ingester.Graph.NumNodes())
	}
	if err := ingester.Finish(ictx, rootNode); err != nil {
		t.Error(err)
		return
	}

	// Graph must have 7 nodes
	if ingester.Graph.NumNodes() != 7 {
		t.Errorf("NumNodes: %d", ingester.Graph.NumNodes())
	}

	nodes := FindChildInstanceOf(rootNode, "tgt")
	if len(nodes) != 1 {
		t.Errorf("Child nodes: %v", nodes)
	}
	tgtCodeNodes := FindChildInstanceOf(nodes[0], "tgtcode")
	if len(tgtCodeNodes) != 1 {
		t.Errorf("No tgtcode")
	}
	tgtSystemNodes := FindChildInstanceOf(nodes[0], "tgtsystem")
	if len(tgtSystemNodes) != 1 {
		t.Errorf("No tgtsystem")
	}
}
