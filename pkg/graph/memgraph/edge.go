package memgraph

import (
	"container/list"

	"github.com/cloudprivacylabs/lsa/pkg/graph"
)

type Edge struct {
	g          *Graph
	id         string
	label      string
	start      *Node
	end        *Node
	el         *list.Element
	elOutgoing *list.Element
	elIncoming *list.Element
	properties map[string]interface{}
}

func newEdge(g *Graph, id, label string, start, end graph.Node) *Edge {
	var (
		startNode *Node
		endNode   *Node
	)
	if start != nil {
		startNode = start.(*Node)
	}
	if end != nil {
		endNode = end.(*Node)
	}
	ret := &Edge{
		g:          g,
		id:         id,
		label:      label,
		start:      startNode,
		end:        endNode,
		properties: make(map[string]interface{}),
	}
	if ret.start != nil && ret.start.g != g {
		panic("Edge start is not in graph")
	}
	if ret.end != nil && ret.end.g != g {
		panic("Edge end is not in graph")
	}
	return ret
}

func (e *Edge) GetGraph() graph.Graph { return e.g }

func (e *Edge) EdgeID() string { return e.id }

func (e *Edge) Properties() map[string]interface{} { return e.properties }

func (e *Edge) EdgeLabel() string { return e.label }

func (e *Edge) StartNode() graph.Node { return e.start }

func (e *Edge) EndNode() graph.Node { return e.end }

type emptyEdgeIterator struct{}

func (emptyEdgeIterator) Next() graph.Edge   { return nil }
func (emptyEdgeIterator) Rest() []graph.Edge { return nil }

type listEdgeIterator struct {
	at *list.Element
}

func (l *listEdgeIterator) Next() graph.Edge {
	if l.at == nil {
		return nil
	}
	ret := l.at.Value.(*Edge)
	l.at = l.at.Next()
	return ret
}

func (l *listEdgeIterator) Rest() []graph.Edge {
	ret := make([]graph.Edge, 0)
	for ; l.at != nil; l.at = l.at.Next() {
		ret = append(ret, l.at.Value.(*Edge))
	}
	return ret
}

type arrayEdgeIterator struct {
	edges []*Edge
	i     int
}

func (l *arrayEdgeIterator) Next() graph.Edge {
	if l.i >= len(l.edges) {
		return nil
	}
	ret := l.edges[l.i]
	l.i++
	return ret
}

func (l *arrayEdgeIterator) Rest() []graph.Edge {
	ret := make([]graph.Edge, 0, len(l.edges)-l.i)
	for ; l.i < len(l.edges); l.i++ {
		ret = append(ret, l.edges[l.i])
	}
	return ret
}

func newMapEdgeIterator(m map[*list.Element]struct{}) *arrayEdgeIterator {
	ret := &arrayEdgeIterator{edges: make([]*Edge, 0, len(m))}
	for k := range m {
		ret.edges = append(ret.edges, k.Value.(*Edge))
	}
	return ret
}
