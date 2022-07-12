package ls

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	//	"github.com/cloudprivacylabs/opencypher/graph"
)

type ingestTest struct {
	Labels []string `json:"labels"`
	Path   []string `json:"path"`
}

func TestEdges(t *testing.T) {
	var valueCase = ingestTest{
		Path: nil,
	}
	valueCase.testValueAsEdge(t)

	var objectCase = ingestTest{
		Path: nil,
	}
	objectCase.testObjectAsEdge(t)

	var arrayCase = ingestTest{
		Path: nil,
	}
	arrayCase.TestArrayAsEdge(t)
}

// attr1:
//    -- attr1 -- > (value: "..")
// attr3:
//    -- edgeLabel --> (value: "..")
func (tc ingestTest) testValueAsEdge(t *testing.T) {
	// g := graph.NewOCGraph()
	var schMap interface{}
	schStr, err := ioutil.ReadFile("testdata/value_as_edge_test.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(schStr), &schMap); err != nil {
		t.Fatal(err)
	}
	schema, err := UnmarshalLayer(schMap, nil)
	if err != nil {
		t.Error(err)
	}

	builder := NewGraphBuilder(nil, GraphBuilderOptions{
		EmbedSchemaNodes: true,
	})

	schemaRoot := schema.GetSchemaRootNode()
	_, rootNode, err := builder.ObjectAsNode(schemaRoot, nil)
	if err != nil {
		t.Errorf("Ingest err: %v", err)
		return
	}

	attr3Node, _ := schema.FindAttributeByID("attr3")
	if attr3Node == nil {
		t.Errorf("Cannot find attr3 node")
		return
	}
	edge, err := builder.RawValueAsEdge(attr3Node, rootNode, "VAUs")
	if err != nil {
		t.Errorf("ingest err: %v", err)
		return
	}
	if edge.GetLabel() != "edgeLabel" {
		t.Errorf("invalid label: %v", edge.GetLabel())
		return
	}
	if s, _ := GetRawNodeValue(edge.GetTo()); s != "VAUs" {
		t.Errorf("Ingestion set value error: %v", err)
		return
	}

	attr4Node, _ := schema.FindAttributeByID("attr4")
	if attr4Node == nil {
		t.Errorf("Cannot find attr4 node")
		return
	}
	edge, err = builder.RawValueAsEdge(attr4Node, rootNode, "b")
	if err != nil {
		t.Errorf("ingest err: %v", err)
		return
	}
	if edge.GetLabel() != "attr4" {
		t.Errorf("invalid get label: %v", edge.GetLabel())
		return
	}
	if s, _ := GetRawNodeValue(edge.GetTo()); s != "b" {
		t.Errorf("Ingestion set value error: %v", err)
		return
	}

}

func (tc ingestTest) testObjectAsEdge(t *testing.T) {
	var schMap interface{}
	schStr, err := ioutil.ReadFile("testdata/object_as_edge_test.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(schStr), &schMap); err != nil {
		t.Fatal(err)
	}
	schema, err := UnmarshalLayer(schMap, nil)
	if err != nil {
		t.Error(err)
	}

	builder := NewGraphBuilder(nil, GraphBuilderOptions{
		EmbedSchemaNodes: true,
	})
	schemaRoot := schema.GetSchemaRootNode()
	_, rootNode, err := builder.ObjectAsNode(schemaRoot, nil)
	if err != nil {
		t.Errorf("Ingest err: %v", err)
		return
	}

	attr2Node, _ := schema.FindAttributeByID("https://www.example.com/id2")
	if attr2Node == nil {
		t.Errorf("Cannot find attr2 node")
		return
	}
	edge, err := builder.ObjectAsEdge(attr2Node, rootNode)
	if err != nil {
		t.Error(err)
	}
	if edge.GetLabel() != "theObjectEdge" {
		t.Errorf("Wrong edge label: %s", edge.GetLabel())
	}
	// There must be a blank node
	if edge.GetFrom() != rootNode {
		t.Errorf("Wrong from")
	}

	attr3Node, _ := schema.FindAttributeByID("https://www.example.com/id3")
	if attr3Node == nil {
		t.Errorf("Cannot find attr3 node")
		return
	}
	edge2, err := builder.RawValueAsEdge(attr3Node, edge.GetTo(), "3")
	if err != nil {
		t.Error(err)
	}
	if edge2.GetFrom() != edge.GetTo() && edge2.GetLabel() != HasTerm {
		t.Errorf("Wrong path")
	}
}

// // attr2: (value_as_edge schema)
// //
// // --arr--> _ -->elem1
// //            -->elem2
// //
// //  or?
// //
// //  --arr-->elem1
// //       -->elem2
func (tc ingestTest) TestArrayAsEdge(t *testing.T) {
	var schMap interface{}
	schStr, err := ioutil.ReadFile("testdata/object_as_edge_test.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(schStr), &schMap); err != nil {
		t.Fatal(err)
	}
	schema, err := UnmarshalLayer(schMap, nil)
	if err != nil {
		t.Error(err)
	}

	builder := NewGraphBuilder(nil, GraphBuilderOptions{
		EmbedSchemaNodes: true,
	})
	schemaRoot := schema.GetSchemaRootNode()
	_, rootNode, err := builder.ObjectAsNode(schemaRoot, nil)
	if err != nil {
		t.Errorf("Ingest err: %v", err)
		return
	}

	attr1Node, _ := schema.FindAttributeByID("https://attr1")
	if attr1Node == nil {
		t.Errorf("Cannot find attr1 node")
		return
	}
	edge, err := builder.ArrayAsEdge(attr1Node, rootNode)
	if err != nil {
		t.Error(err)
	}
	if edge.GetLabel() != "arr" {
		t.Errorf("Wrong edge label: %s", edge.GetLabel())
	}
	// There must be a blank node
	if edge.GetFrom() != rootNode {
		t.Errorf("Wrong from")
	}

	elemNode, _ := schema.FindAttributeByID("https://www.example.com/id0")
	if elemNode == nil {
		t.Errorf("Cannot find elem node")
		return
	}
	edge2, err := builder.RawValueAsEdge(elemNode, edge.GetTo(), "3")
	if err != nil {
		t.Error(err)
	}
	if edge2.GetFrom() != edge.GetTo() && edge2.GetLabel() != HasTerm {
		t.Errorf("Wrong path")
	}
}
