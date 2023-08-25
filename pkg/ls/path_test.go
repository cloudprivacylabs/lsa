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
	"testing"

	"github.com/cloudprivacylabs/lpg/v2"
)

func instantiatePathNodeFunc(g *lpg.Graph) func(parent, schemaNode *lpg.Node) (*lpg.Node, error) {
	return func(parent, schemaNode *lpg.Node) (*lpg.Node, error) {
		newNode := InstantiateSchemaNode(g, schemaNode, true, map[*lpg.Node]*lpg.Node{})
		g.NewEdge(parent, newNode, HasTerm.Name, nil)
		return newNode, nil
	}
}

func TestInstantiatePathBasic(t *testing.T) {
	layer, err := UnmarshalLayerFromSlice([]byte(`{
  "nodes": [
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
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "properties": {
        "https://lschema.org/attributeIndex": 1,
        "https://lschema.org/nodeId": "attr2"
      }
    }
  ]
}
`))
	if err != nil {
		panic(err)
	}
	g := lpg.NewGraph()
	root := InstantiateSchemaNode(g, layer.GetSchemaRootNode(), true, map[*lpg.Node]*lpg.Node{})
	EnsurePath(root, nil, layer.GetSchemaRootNode(), layer.GetAttributeByID("attr1"), instantiatePathNodeFunc(g))
	nodes := lpg.NextNodesWith(root, HasTerm.Name)
	if len(nodes) != 1 {
		t.Errorf("Expecting 1 node")
	}
	if GetNodeSchemaNodeID(nodes[0]) != "attr1" {
		t.Errorf("Wrong instance")
	}
}

func TestInstantiatePathNested(t *testing.T) {
	layer, err := UnmarshalLayerFromSlice([]byte(`
{
  "nodes": [
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
        "https://lschema.org/nodeId": "attr1"
      }
    },
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
    }
  ]
}
`))
	if err != nil {
		panic(err)
	}
	g := lpg.NewGraph()
	root := InstantiateSchemaNode(g, layer.GetSchemaRootNode(), true, map[*lpg.Node]*lpg.Node{})
	EnsurePath(root, nil, layer.GetSchemaRootNode(), layer.GetAttributeByID("attr5"), instantiatePathNodeFunc(g))
	nodes := lpg.NextNodesWith(root, HasTerm.Name)
	if len(nodes) != 1 {
		t.Errorf("Expecting 1 node")
	}
	nodes = lpg.NextNodesWith(nodes[0], HasTerm.Name)
	if len(nodes) != 1 {
		t.Errorf("Expecting 1 node")
	}
	nodes = lpg.NextNodesWith(nodes[0], HasTerm.Name)
	if len(nodes) != 1 {
		t.Errorf("Expecting 1 node")
	}
	if GetNodeSchemaNodeID(nodes[0]) != "attr5" {
		t.Errorf("Wrong instance")
	}
}

func TestInstantiatePathNestedWithAncestor(t *testing.T) {
	layer, err := UnmarshalLayerFromSlice([]byte(`
{
  "nodes": [
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
      "n": 3,
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
      "n": 4,
      "labels": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "properties": {
        "https://lschema.org/attributeIndex": 0,
        "https://lschema.org/nodeId": "attr7"
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
        "https://lschema.org/nodeId": "attr3"
      }
    },
    {
      "n": 7,
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
      "n": 6,
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
          "to": 7,
          "label": "https://lschema.org/Object/attributes"
        },
        {
          "to": 8,
          "label": "https://lschema.org/Object/attributes"
        }
      ]
    },
    {
      "n": 8,
      "labels": [
        "https://lschema.org/Object",
        "https://lschema.org/Attribute"
      ],
      "properties": {
        "https://lschema.org/attributeIndex": 1,
        "https://lschema.org/nodeId": "attr6"
      },
      "edges": [
        {
          "to": 4,
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
	g := lpg.NewGraph()
	root := InstantiateSchemaNode(g, layer.GetSchemaRootNode(), true, map[*lpg.Node]*lpg.Node{})
	attr4, _ := EnsurePath(root, nil, layer.GetSchemaRootNode(), layer.GetAttributeByID("attr4"), instantiatePathNodeFunc(g))
	attr7, _ := EnsurePath(root, attr4, layer.GetSchemaRootNode(), layer.GetAttributeByID("attr7"), instantiatePathNodeFunc(g))
	nodes := lpg.NextNodesWith(root, HasTerm.Name)
	if len(nodes) != 1 {
		t.Errorf("Expecting 1 node")
	}
	nodes = lpg.NextNodesWith(nodes[0], HasTerm.Name)
	if len(nodes) != 1 {
		t.Errorf("Expecting 1 node")
	}
	nodes = lpg.NextNodesWith(nodes[0], HasTerm.Name)
	if len(nodes) != 1 {
		t.Errorf("Expecting 1 node")
	}
	nodes = lpg.NextNodesWith(nodes[0], HasTerm.Name)
	if len(nodes) != 1 {
		t.Errorf("Expecting 1 node")
	}
	if nodes[0] != attr7 {
		t.Errorf("Wrong instance")
	}
}
