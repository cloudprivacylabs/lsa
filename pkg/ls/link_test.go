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
	"fmt"
	"reflect"
	"testing"

	"github.com/cloudprivacylabs/lpg/v2"
)

// fail
func TestBasicLink(t *testing.T) {
	schemas := make([]*Layer, 2)
	for i, x := range []string{"testdata/link_1/root.json", "testdata/link_1/2.json"} {
		var err error
		schemas[i], err = ReadLayerFromFile(x)
		if err != nil {
			t.Error(err)
			return
		}
	}
	compiler := Compiler{
		Loader: SchemaLoaderFunc(func(ref string) (*Layer, error) {
			for i := range schemas {
				/*
					https://root
					https://2
				*/
				if ref == schemas[i].GetID() {
					return schemas[i], nil
				}
			}
			return nil, fmt.Errorf("Not found: %s", ref)
		}),
	}
	layer0, err := compiler.Compile(DefaultContext(), schemas[0].GetID())
	if err != nil {
		t.Error(err)
		return
	}
	layer2, err := compiler.Compile(DefaultContext(), schemas[1].GetID())
	if err != nil {
		t.Error(err)
		return
	}

	builder := NewGraphBuilder(nil, GraphBuilderOptions{
		EmbedSchemaNodes: true,
	})
	_, root1, _ := builder.ObjectAsNode(layer0.GetSchemaRootNode(), nil)
	builder.RawValueAsNode(layer0.GetAttributeByID("https://idField"), root1, "123")

	_, root2, _ := builder.ObjectAsNode(layer2.GetSchemaRootNode(), nil)
	builder.RawValueAsNode(layer2.GetAttributeByID("idField"), root2, "456")
	builder.RawValueAsNode(layer2.GetAttributeByID("https://rootid"), root2, "123")

	builder.LinkNodes(DefaultContext(), layer2)
	// There must be an edge from root1 to root2
	found := false
	for edges := root1.GetEdges(lpg.OutgoingEdge); edges.Next(); {
		edge := edges.Edge()
		if edge.GetTo() == root2 {
			found = true
		}
	}
	if !found {
		t.Errorf("No edges from root1 to root2")
	}

}

// fail
func TestValueLink(t *testing.T) {
	schemas := make([]*Layer, 2)
	for i, x := range []string{"testdata/link_1/root.json", "testdata/link_1/3.json"} {
		var err error
		schemas[i], err = ReadLayerFromFile(x)
		if err != nil {
			t.Error(err)
			return
		}
	}
	compiler := Compiler{
		Loader: SchemaLoaderFunc(func(ref string) (*Layer, error) {
			for i := range schemas {
				if ref == schemas[i].GetID() {
					return schemas[i], nil
				}
			}
			return nil, fmt.Errorf("Not found: %s", ref)
		}),
	}
	layer0, err := compiler.Compile(DefaultContext(), schemas[0].GetID())
	if err != nil {
		t.Error(err)
		return
	}
	for nodes := schemas[1].Graph.GetNodes(); nodes.Next(); {
		fmt.Println(nodes.Node())
	}
	layer3, err := compiler.Compile(DefaultContext(), schemas[1].GetID())
	if err != nil {
		t.Error(err)
		return
	}

	builder := NewGraphBuilder(nil, GraphBuilderOptions{
		EmbedSchemaNodes: true,
	})
	_, root1, _ := builder.ObjectAsNode(layer0.GetSchemaRootNode(), nil)
	builder.RawValueAsNode(layer0.GetAttributeByID("https://idField"), root1, "123")

	_, root3, _ := builder.ObjectAsNode(layer3.GetSchemaRootNode(), nil)
	nd := layer3.GetAttributeByID("https://rootid")
	fmt.Println(nd)
	builder.RawValueAsNode(nd, root3, "123")

	builder.LinkNodes(DefaultContext(), layer3)
	fkValFound := false
	for nodeItr := builder.GetGraph().GetNodes(); nodeItr.Next(); {
		node := nodeItr.Node()
		if val, ok := node.GetProperty(ReferenceFK.Name); ok {
			if !reflect.DeepEqual(val.(PropertyValue).AsStringSlice(), []string{"123"}) {
				t.Errorf("Wrong fk val")
			}
			fkValFound = true
		}
	}
	if !fkValFound {
		t.Errorf("No fk val found")
	}
	// There must be an edge from root1 to root3
	found := false
	for edges := root1.GetEdges(lpg.OutgoingEdge); edges.Next(); {
		edge := edges.Edge()
		/*
			edge.GetTo() (:https://lschema.org/DocumentNode:https://lschema.org/Value {0:https://idField 3:0 4:123})
			root3 (:3:https://lschema.org/DocumentNode:https://lschema.org/Object {0:3 1:https://3})
		*/
		if edge.GetTo() == root3 {
			found = true
		}
	}
	if !found {
		t.Errorf("No edges from root1 to root3")
	}

}
