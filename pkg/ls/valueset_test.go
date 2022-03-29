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
	//	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
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

}
