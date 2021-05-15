package memgraph

import (
	"container/list"

	"github.com/cloudprivacylabs/lsa/pkg/graph"
	"github.com/lithammer/shortuuid/v3"
)

type Graph struct {
	nodes indexedList
	edges indexedList
}

func (g *Graph) newId() string { return shortuuid.New() }

func (g *Graph) GetNodes() graph.NodeIterator { return &listNodeIterator{at: g.nodes.list.Front()} }
func (g *Graph) GetEdges() graph.EdgeIterator { return &listEdgeIterator{at: g.edges.list.Front()} }
func (g *Graph) GetNodeByID(id string) graph.Node {
	e := g.nodes.index[id]
	if e == nil {
		return nil
	}
	return e.Value.(*Node)
}

func (g *Graph) GetEdgeByID(id string) graph.Edge {
	e := g.edges.index[id]
	if e == nil {
		return nil
	}
	return e.Value.(*Edge)
}

func (g *Graph) NewNode(id string) graph.Node {
	node := newNode(g, id)
	el := g.nodes.add(node.id, node)
	node.el = el
	return node
}

func (g *Graph) NewEdge(id, label string, start, end graph.Node) graph.Edge {
	edge := newEdge(g, id, label, start, end)
	el := g.edges.add(edge.id, edge)
	edge.el = el
	if edge.start != nil {
		edge.elOutgoing = edge.start.outgoing.PushBack(edge)
		lst := edge.start.outgoingLabels[edge.label]
		if lst == nil {
			lst = make(map[*list.Element]struct{})
		}
		lst[edge.el] = struct{}{}
		edge.start.outgoingLabels[edge.label] = lst
	}
	if edge.end != nil {
		edge.elIncoming = edge.end.incoming.PushBack(edge)
		lst := edge.end.incomingLabels[edge.label]
		if lst == nil {
			lst = make(map[*list.Element]struct{})
		}
		lst[edge.el] = struct{}{}
		edge.end.incomingLabels[edge.label] = lst
	}
	return edge
}

func (g *Graph) RemoveNode(node graph.Node) {
	n, ok := node.(*Node)
	if !ok {
		panic("Not a node")
	}
	if n.g != g {
		panic("Node not in graph")
	}
	n.g = nil

	for e := n.outgoing.Front(); e != nil; e = n.outgoing.Front() {
		g.removeEdge(e.Value.(*Edge))
	}
	for e := n.incoming.Front(); e != nil; e = n.incoming.Front() {
		g.removeEdge(e.Value.(*Edge))
	}
	g.nodes.remove(n.id, n.el)
}

func (g *Graph) RemoveEdge(edge graph.Edge) {
	e, ok := edge.(*Edge)
	if !ok {
		panic("Not an edge")
	}
	if e.g != g {
		panic("Not in graph")
	}
	g.removeEdge(e)
}

func (g *Graph) removeEdge(edge *Edge) {
	g.edges.remove(edge.id, edge.el)
	if edge.start != nil {
		edge.start.outgoing.Remove(edge.elOutgoing)
		lst := edge.start.outgoingLabels[edge.label]
		delete(lst, edge.el)
	}
	if edge.end != nil {
		edge.end.incoming.Remove(edge.elIncoming)
		lst := edge.end.incomingLabels[edge.label]
		delete(lst, edge.el)
	}
}
