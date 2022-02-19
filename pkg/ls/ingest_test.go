package ls

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

type ingestTest struct {
	Labels    []string                `json:"labels"`
	Path      []string                `json:"path"`
	NodePaths map[graph.Node]NodePath `json:"node_paths"`
}

func TestEdges(t *testing.T) {
	var valueCase = ingestTest{
		Path:      nil,
		NodePaths: make(map[graph.Node]NodePath),
	}
	valueCase.testValueAsEdge(t)

	var objectCase = ingestTest{
		Path:      nil,
		NodePaths: make(map[graph.Node]NodePath),
	}
	objectCase.testObjectAsEdge(t)

	var arrayCase = ingestTest{
		Path:      nil,
		NodePaths: make(map[graph.Node]NodePath),
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
	root := schema.GetSchemaRootNode()
	root.SetProperty(EdgeLabelTerm, &PropertyValue{value: "attr3"})

	ing := Ingester{Schema: schema, EmbedSchemaNodes: true, PreserveNodePaths: true, NodePaths: tc.NodePaths}
	en, err := ing.ValueAsEdge(schema.Graph, tc.Path, root, "VAUs", AttributeTypeValue)
	if err != nil {
		t.Errorf("ingest err: %v", err)
		return
	}
	if en.Label != "attr3" {
		t.Errorf("invalid get property: %v", err)
		return
	}
	if GetRawNodeValue(en.Node) != "VAUs" {
		t.Errorf("Ingestion set value error: %v", err)
		return
	}
	root.RemoveProperty(EdgeLabelTerm)
	root.SetProperty(AttributeNameTerm, &PropertyValue{value: "attr1"})
	en, err = ing.ValueAsEdge(schema.Graph, tc.Path, root, "OTHERVALUE", AttributeTypeValue)
	if err != nil {
		t.Errorf("ingest err: %v", err)
		return
	}
	if en.Label != "attr1" {
		t.Errorf("invalid get property: %v", err)
		return
	}
	if GetRawNodeValue(en.Node) != "OTHERVALUE" {
		t.Errorf("Ingestion set value error: %v", err)
		return
	}
}

func (tc ingestTest) testObjectAsEdge(t *testing.T) {
	// g := graph.NewOCGraph()
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
	root := schema.GetSchemaRootNode()
	ing := Ingester{Schema: schema, EmbedSchemaNodes: true, PreserveNodePaths: true, NodePaths: tc.NodePaths}
	ix := schema.GetNodes()
	childNodes := make([]graph.Node, 0, schema.NumNodes())
	childEdgeNodes := make([]EdgeNode, 0, schema.NumNodes())
	for ix.Next() {
		childNodes = append(childNodes, ix.Node())
		if label, ok := ix.Node().GetProperty(IngestAsTerm); ok {
			childEdgeNodes = append(childEdgeNodes, EdgeNode{Node: ix.Node(), Label: label.(*PropertyValue).AsString(), Properties: make(map[string]interface{})})
		}
	}
	en, err := ing.ObjectAsEdge(schema.Graph, tc.Path, root, childNodes, childEdgeNodes, ObjectAttributeListTerm)
	if err != nil {
		t.Error(err)
	}
	checkEdgeNodeValue := func(nodeId string, childEdgeNodes []EdgeNode, expected interface{}) {
		for _, e := range childEdgeNodes {
			prop, ok := e.Node.GetProperty(IngestAsTerm)
			if ok && prop.(*PropertyValue).AsString() != e.Label {
				t.Errorf("Wrong value for %v: expecting: %v, got %v", nodeId, expected, prop)
			}
		}

	}
	checkEdgeNodeValue(EdgeLabelTerm, childEdgeNodes, en.Label)

	checkNodeValue := func(nodeId string, childNode graph.Node, expected interface{}) {
		prop, _ := childNode.GetProperty(AttributeNameTerm)
		if prop != expected {
			t.Errorf("Wrong value for %v: expecting: %v, got %v", nodeId, expected, prop)
		}
	}
	for _, cn := range childNodes {
		prop, _ := cn.GetProperty(AttributeNameTerm)
		checkNodeValue(GetAttributeID(cn), cn, prop)
	}
}

// attr2: (value_as_edge schema)
//
// --arr--> _ -->elem1
//            -->elem2
//
//  or?
//
//  --arr-->elem1
//       -->elem2
func (tc ingestTest) TestArrayAsEdge(t *testing.T) {
	// g := graph.NewOCGraph()
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
	root := schema.GetSchemaRootNode()
	root.SetLabels(graph.NewStringSet(AttributeTypeArray))
	ing := Ingester{Schema: schema, EmbedSchemaNodes: true, PreserveNodePaths: true, NodePaths: tc.NodePaths}
	ix := schema.GetNodes()
	childNodes := make([]graph.Node, 0, schema.NumNodes())
	childEdgeNodes := make([]EdgeNode, 0, schema.NumNodes())
	for ix.Next() {
		childNodes = append(childNodes, ix.Node())
		if label, ok := ix.Node().GetProperty(IngestAsTerm); ok {
			childEdgeNodes = append(childEdgeNodes, EdgeNode{Node: ix.Node(), Label: label.(*PropertyValue).AsString(), Properties: make(map[string]interface{})})
		}
	}
	en, err := ing.ArrayAsEdge(schema.Graph, tc.Path, root, childNodes, childEdgeNodes, AttributeTypeArray)
	if err != nil {
		t.Error(err)
	}
	checkEdgeNodeValue := func(nodeId string, childEdgeNodes []EdgeNode, expected interface{}) {
		for _, e := range childEdgeNodes {
			prop, ok := e.Node.GetProperty(IngestAsTerm)
			if ok && prop.(*PropertyValue).AsString() != e.Label {
				t.Errorf("Wrong value for %v: expecting: %v, got %v", nodeId, expected, prop)
			}
		}

	}
	checkEdgeNodeValue(EdgeLabelTerm, childEdgeNodes, en.Label)

	checkNodeValue := func(nodeId string, childNode graph.Node, expected interface{}) {
		prop, _ := childNode.GetProperty(AttributeNameTerm)
		if prop != expected {
			t.Errorf("Wrong value for %v: expecting: %v, got %v", nodeId, expected, prop)
		}
	}
	for _, cn := range childNodes {
		prop, _ := cn.GetProperty(AttributeNameTerm)
		checkNodeValue(GetAttributeID(cn), cn, prop)
	}
}
