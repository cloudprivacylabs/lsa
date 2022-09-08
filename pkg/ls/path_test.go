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

	"github.com/cloudprivacylabs/lpg"
)

func instantiatePathNodeFunc(g *lpg.Graph) func(parent, schemaNode *lpg.Node, seen map[string]struct{}) (*lpg.Node, error) {
	return func(parent, schemaNode *lpg.Node, seen map[string]struct{}) (*lpg.Node, error) {
		newNode := InstantiateSchemaNode(g, schemaNode, true, map[*lpg.Node]*lpg.Node{})
		g.NewEdge(parent, newNode, HasTerm, nil)
		return newNode, nil
	}
}

func TestInstantiatePathBasic(t *testing.T) {
	layer, err := UnmarshalLayerFromSlice([]byte(`{
  "@context": "../../schemas/ls.json",
  "@type": "Schema",
  "@id": "testSchema",
  "layer": {
    "@id": "schemaRoot",
    "@type": "Object",
    "attributes": {
       "attr1": { "@type": "Value" },
       "attr2": { "@type": "Value" }
    }
  }
}`))
	if err != nil {
		panic(err)
	}
	g := lpg.NewGraph()
	root := InstantiateSchemaNode(g, layer.GetSchemaRootNode(), true, map[*lpg.Node]*lpg.Node{})
	EnsurePath(root, nil, layer.GetSchemaRootNode(), layer.GetAttributeByID("attr1"), instantiatePathNodeFunc(g))
	nodes := lpg.NextNodesWith(root, HasTerm)
	if len(nodes) != 1 {
		t.Errorf("Expecting 1 node")
	}
	if GetNodeSchemaNodeID(nodes[0]) != "attr1" {
		t.Errorf("Wrong instance")
	}
}

func TestInstantiatePathNested(t *testing.T) {
	layer, err := UnmarshalLayerFromSlice([]byte(`{
  "@context": "../../schemas/ls.json",
  "@type": "Schema",
  "@id": "testSchema",
  "layer": {
    "@id": "schemaRoot",
    "@type": "Object",
    "attributes": {
       "attr1": { "@type": "Value" },
       "attr2": { 
          "@type": "Object",
          "attributes": {
             "attr3": { "@type": "Value"},
             "attr4": {
                "@type": "Object",
                "attributes": {
                   "attr5": { "@type": "Value" }
                }
             }
          }
       }
    }
  }
}`))
	if err != nil {
		panic(err)
	}
	g := lpg.NewGraph()
	root := InstantiateSchemaNode(g, layer.GetSchemaRootNode(), true, map[*lpg.Node]*lpg.Node{})
	EnsurePath(root, nil, layer.GetSchemaRootNode(), layer.GetAttributeByID("attr5"), instantiatePathNodeFunc(g))
	nodes := lpg.NextNodesWith(root, HasTerm)
	if len(nodes) != 1 {
		t.Errorf("Expecting 1 node")
	}
	nodes = lpg.NextNodesWith(nodes[0], HasTerm)
	if len(nodes) != 1 {
		t.Errorf("Expecting 1 node")
	}
	nodes = lpg.NextNodesWith(nodes[0], HasTerm)
	if len(nodes) != 1 {
		t.Errorf("Expecting 1 node")
	}
	if GetNodeSchemaNodeID(nodes[0]) != "attr5" {
		t.Errorf("Wrong instance")
	}
}

func TestInstantiatePathNestedWithAncestor(t *testing.T) {
	layer, err := UnmarshalLayerFromSlice([]byte(`{
  "@context": "../../schemas/ls.json",
  "@type": "Schema",
  "@id": "testSchema",
  "layer": {
    "@id": "schemaRoot",
    "@type": "Object",
    "attributes": {
       "attr1": { "@type": "Value" },
       "attr2": { 
          "@type": "Object",
          "attributes": {
             "attr3": { "@type": "Value"},
             "attr4": {
                "@type": "Object",
                "attributes": {
                   "attr5": { "@type": "Value" },
                   "attr6": {
                     "@type": "Object",
                     "attributes": {
                        "attr7": { "@type": "Value"}
                     }
                   }
                }
             }
          }
       }
    }
  }
}`))
	if err != nil {
		panic(err)
	}
	g := lpg.NewGraph()
	root := InstantiateSchemaNode(g, layer.GetSchemaRootNode(), true, map[*lpg.Node]*lpg.Node{})
	attr4, _ := EnsurePath(root, nil, layer.GetSchemaRootNode(), layer.GetAttributeByID("attr4"), instantiatePathNodeFunc(g))
	attr7, _ := EnsurePath(root, attr4, layer.GetSchemaRootNode(), layer.GetAttributeByID("attr7"), instantiatePathNodeFunc(g))
	nodes := lpg.NextNodesWith(root, HasTerm)
	if len(nodes) != 1 {
		t.Errorf("Expecting 1 node")
	}
	nodes = lpg.NextNodesWith(nodes[0], HasTerm)
	if len(nodes) != 1 {
		t.Errorf("Expecting 1 node")
	}
	nodes = lpg.NextNodesWith(nodes[0], HasTerm)
	if len(nodes) != 1 {
		t.Errorf("Expecting 1 node")
	}
	nodes = lpg.NextNodesWith(nodes[0], HasTerm)
	if len(nodes) != 1 {
		t.Errorf("Expecting 1 node")
	}
	if nodes[0] != attr7 {
		t.Errorf("Wrong instance")
	}
}
