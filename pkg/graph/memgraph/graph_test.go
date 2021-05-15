package memgraph

import (
	"testing"
)

func TestBasicGraph(t *testing.T) {
	g := NewGraph()

	n1 := g.NewNode("1")
	n2 := g.NewNode("2")
	g.NewEdge("", "label", n1, n2)

	edges := n1.AllOutgoingEdges().Rest()
	if len(edges) != 1 {
		t.Errorf("Expected 1 edge, %d", len(edges))
	}
	if edges[0].StartNode() != n1 {
		t.Error("Wrong start")
	}
	if edges[0].EndNode() != n2 {
		t.Error("Wrong end")
	}
	edges = n2.AllIncomingEdges().Rest()
	if len(edges) != 1 {
		t.Errorf("Expected 1 edge, %d", len(edges))
	}
	if edges[0].StartNode() != n1 {
		t.Error("Wrong start")
	}
	if edges[0].EndNode() != n2 {
		t.Error("Wrong end")
	}
	g.RemoveEdge(edges[0])
	if len(n1.AllOutgoingEdges().Rest()) != 0 {
		t.Error("There are still edges")
	}
	if len(n2.AllIncomingEdges().Rest()) != 0 {
		t.Error("There are still edges")
	}
}
