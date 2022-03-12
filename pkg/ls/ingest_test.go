package ls

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	//	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
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

	ing := Ingester{Schema: schema, EmbedSchemaNodes: true, Graph: NewDocumentGraph()}
	ctx := ing.Start(DefaultContext(), "")
	_, _, rootNode, err := ing.Object(ctx)
	if err != nil {
		t.Errorf("Ingest err: %v", err)
		return
	}
	newctx := ctx.NewLevel(rootNode)

	attr3Node, _ := schema.FindAttributeByID("attr3")
	if attr3Node == nil {
		t.Errorf("Cannot find attr3 node")
		return
	}
	_, edge, node, err := ing.Value(newctx.New("attr3", attr3Node), "VAUs")
	if err != nil {
		t.Errorf("ingest err: %v", err)
		return
	}
	if edge.GetLabel() != "edgeLabel" {
		t.Errorf("invalid label: %v", edge.GetLabel())
		return
	}
	if s, _ := GetRawNodeValue(node); s != "VAUs" {
		t.Errorf("Ingestion set value error: %v", err)
		return
	}

	attr4Node, _ := schema.FindAttributeByID("attr4")
	if attr4Node == nil {
		t.Errorf("Cannot find attr4 node")
		return
	}
	_, edge, node, err = ing.Value(newctx.New("attr4", attr4Node), "b")
	if err != nil {
		t.Errorf("ingest err: %v", err)
		return
	}
	if edge.GetLabel() != "attr4" {
		t.Errorf("invalid get label: %v", edge.GetLabel())
		return
	}
	if s, _ := GetRawNodeValue(node); s != "b" {
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

	ing := Ingester{Schema: schema, EmbedSchemaNodes: true, Graph: NewDocumentGraph()}
	ctx := ing.Start(DefaultContext(), "")
	_, _, rootNode, err := ing.Object(ctx)
	if err != nil {
		t.Errorf("Ingest err: %v", err)
		return
	}
	newctx := ctx.NewLevel(rootNode)

	attr2Node, _ := schema.FindAttributeByID("https://www.example.com/id2")
	if attr2Node == nil {
		t.Errorf("Cannot find attr2 node")
		return
	}
	_, edge, node, err := ing.Object(newctx.New("obj2", attr2Node))
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
	if edge.GetTo() != node {
		t.Errorf("GetTo")
	}

	attr3Node, _ := schema.FindAttributeByID("https://www.example.com/id3")
	if attr3Node == nil {
		t.Errorf("Cannot find attr3 node")
		return
	}
	_, edge2, node2, err := ing.Value(newctx.New("field3", attr3Node), "3")
	if err != nil {
		t.Error(err)
	}
	if edge2.GetFrom() != node && edge2.GetLabel() != HasTerm && edge2.GetTo() != node2 {
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

	ing := Ingester{Schema: schema, EmbedSchemaNodes: true, Graph: NewDocumentGraph()}
	ctx := ing.Start(DefaultContext(), "")
	_, _, rootNode, err := ing.Object(ctx)
	if err != nil {
		t.Errorf("Ingest err: %v", err)
		return
	}
	newctx := ctx.NewLevel(rootNode)

	attr1Node, _ := schema.FindAttributeByID("https://attr1")
	if attr1Node == nil {
		t.Errorf("Cannot find attr1 node")
		return
	}
	_, edge, node, err := ing.Array(newctx.New("arr", attr1Node))
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
	if edge.GetTo() != node {
		t.Errorf("GetTo")
	}

	elemNode, _ := schema.FindAttributeByID("https://www.example.com/id0")
	if elemNode == nil {
		t.Errorf("Cannot find elem node")
		return
	}
	_, edge2, node2, err := ing.Value(newctx.New(0, elemNode), "3")
	if err != nil {
		t.Error(err)
	}
	if edge2.GetFrom() != node && edge2.GetLabel() != HasTerm && edge2.GetTo() != node2 {
		t.Errorf("Wrong path")
	}
}
