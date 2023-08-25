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
	"bytes"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/cloudprivacylabs/lpg/v2"
)

func TestHashNested(t *testing.T) {
	layer, err := UnmarshalLayerFromSlice([]byte(`{
  "nodes": [
    {
      "n": 4,
      "labels": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "properties": {
        "https://lschema.org/attributeIndex": 0,
        "https://lschema.org/hash": [
          "attr5",
          "attr1"
        ],
        "https://lschema.org/nodeId": "attr3"
      }
    },
    {
      "n": 5,
      "labels": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Object"
      ],
      "properties": {
        "https://lschema.org/attributeIndex": 1,
        "https://lschema.org/nodeId": "attr4"
      },
      "edges": [
        {
          "to": 6,
          "label": "https://lschema.org/Object/attributes"
        }
      ]
    },
    {
      "n": 6,
      "labels": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "properties": {
        "https://lschema.org/attributeIndex": 0,
        "https://lschema.org/nodeId": "attr5"
      }
    },
    {
      "n": 0,
      "labels": [
        "https://lschema.org/Schema"
      ],
      "properties": {
        "https://lschema.org/nodeId": "testSchema"
      },
      "edges": [
        {
          "to": 1,
          "label": "https://lschema.org/layer"
        }
      ]
    },
    {
      "n": 1,
      "labels": [
        "https://lschema.org/Object",
        "https://lschema.org/Attribute"
      ],
      "properties": {
        "https://lschema.org/nodeId": "schemaRoot"
      },
      "edges": [
        {
          "to": 2,
          "label": "https://lschema.org/Object/attributes"
        },
        {
          "to": 3,
          "label": "https://lschema.org/Object/attributes"
        }
      ]
    },
    {
      "n": 2,
      "labels": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "properties": {
        "https://lschema.org/attributeIndex": 0,
        "https://lschema.org/nodeId": "attr1"
      }
    },
    {
      "n": 3,
      "labels": [
        "https://lschema.org/Object",
        "https://lschema.org/Attribute"
      ],
      "properties": {
        "https://lschema.org/attributeIndex": 1,
        "https://lschema.org/nodeId": "attr2"
      },
      "edges": [
        {
          "to": 4,
          "label": "https://lschema.org/Object/attributes"
        },
        {
          "to": 5,
          "label": "https://lschema.org/Object/attributes"
        }
      ]
    }
  ]
}
`))
	if err != nil {
		panic(err)
	}
	compiler := Compiler{}
	layer, err = compiler.CompileSchema(DefaultContext(), layer)
	if err != nil {
		t.Error(err)
		return
	}

	gb := NewGraphBuilder(nil, GraphBuilderOptions{EmbedSchemaNodes: true})
	_, schemaRoot, _ := gb.ObjectAsNode(layer.GetAttributeByID("schemaRoot"), nil)
	_, attr1, _ := gb.RawValueAsNode(layer.GetAttributeByID("attr1"), schemaRoot, "attr1")
	_, attr2, _ := gb.ObjectAsNode(layer.GetAttributeByID("attr2"), schemaRoot)
	if err := gb.PostIngest(layer.GetAttributeByID("schemaRoot"), schemaRoot); err != nil {
		t.Error(err)
	}

	buf := bytes.Buffer{}
	lpg.JSON{}.Encode(gb.GetGraph(), &buf)
	t.Log(buf.String())
	nodeIDMap := GetSchemaNodeIDMap(schemaRoot)
	attr3 := nodeIDMap["attr3"]
	if len(attr3) != 1 {
		t.Errorf("Expecting 1 node")
	}
	sum := fmt.Sprintf("%x", sha256.Sum256([]byte("attr1")))
	if v, _ := GetRawNodeValue(attr3[0]); v != sum {
		t.Errorf("Wrong value")
	}
	_ = attr1
	_ = attr2
}

func TestHashProperty(t *testing.T) {
	layer, err := UnmarshalLayerFromSlice([]byte(`{
  "nodes": [
    {
      "n": 6,
      "labels": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Object"
      ],
      "properties": {
        "https://lschema.org/attributeIndex": 1,
        "https://lschema.org/nodeId": "attr2"
      },
      "edges": [
        {
          "to": 2,
          "label": "https://lschema.org/Object/attributes"
        },
        {
          "to": 3,
          "label": "https://lschema.org/Object/attributes"
        }
      ]
    },
    {
      "n": 0,
      "labels": [
        "https://lschema.org/Schema"
      ],
      "properties": {
        "https://lschema.org/nodeId": "testSchema"
      },
      "edges": [
        {
          "to": 1,
          "label": "https://lschema.org/layer"
        }
      ]
    },
    {
      "n": 1,
      "labels": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Object"
      ],
      "properties": {
        "https://lschema.org/nodeId": "schemaRoot"
      },
      "edges": [
        {
          "to": 5,
          "label": "https://lschema.org/Object/attributes"
        },
        {
          "to": 6,
          "label": "https://lschema.org/Object/attributes"
        }
      ]
    },
    {
      "n": 2,
      "labels": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "properties": {
        "https://lschema.org/attributeIndex": 0,
        "https://lschema.org/hash": [
          "attr5",
          "attr1"
        ],
        "https://lschema.org/nodeId": "attr3"
      }
    },
    {
      "n": 3,
      "labels": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Object"
      ],
      "properties": {
        "https://lschema.org/attributeIndex": 1,
        "https://lschema.org/nodeId": "attr4"
      },
      "edges": [
        {
          "to": 4,
          "label": "https://lschema.org/Object/attributes"
        }
      ]
    },
    {
      "n": 4,
      "labels": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "properties": {
        "https://lschema.org/attributeIndex": 0,
        "https://lschema.org/nodeId": "attr5"
      }
    },
    {
      "n": 5,
      "labels": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "properties": {
        "https://lschema.org/attributeIndex": 0,
        "https://lschema.org/ingestAs": "property",
        "https://lschema.org/nodeId": "attr1"
      }
    }
  ]
}
`))
	if err != nil {
		panic(err)
	}
	compiler := Compiler{}
	layer, err = compiler.CompileSchema(DefaultContext(), layer)
	if err != nil {
		t.Error(err)
		return
	}

	gb := NewGraphBuilder(nil, GraphBuilderOptions{EmbedSchemaNodes: true})
	_, schemaRoot, _ := gb.ObjectAsNode(layer.GetAttributeByID("schemaRoot"), nil)
	attr1 := gb.RawValueAsProperty(layer.GetAttributeByID("attr1"), []*lpg.Node{schemaRoot}, "attr1")
	_, attr2, _ := gb.ObjectAsNode(layer.GetAttributeByID("attr2"), schemaRoot)
	if err := gb.PostIngest(layer.GetAttributeByID("schemaRoot"), schemaRoot); err != nil {
		t.Error(err)
	}

	buf := bytes.Buffer{}
	lpg.JSON{}.Encode(gb.GetGraph(), &buf)
	t.Log(buf.String())
	nodeIDMap := GetSchemaNodeIDMap(schemaRoot)
	attr3 := nodeIDMap["attr3"]
	if len(attr3) != 1 {
		t.Errorf("Expecting 1 node")
	}
	sum := fmt.Sprintf("%x", sha256.Sum256([]byte("attr1")))
	if v, _ := GetRawNodeValue(attr3[0]); v != sum {
		t.Errorf("Wrong value")
	}
	_ = attr1
	_ = attr2
}
