package memgraph

import (
	"container/list"

	"github.com/cloudprivacylabs/lsa/pkg/graph"
)

type Node struct {
	g          *Graph
	id         string
	el         *list.Element
	properties map[string]interface{}

	outgoing       list.List
	outgoingLabels map[string]map[*list.Element]struct{}
	incoming       list.List
	incomingLabels map[string]map[*list.Element]struct{}
}

func newNode(g *Graph, id string) *Node {
	if len(id) == 0 {
		id = g.newId()
	}
	return &Node{
		g:              g,
		id:             id,
		properties:     make(map[string]interface{}),
		outgoingLabels: make(map[string]map[*list.Element]struct{}),
		incomingLabels: make(map[string]map[*list.Element]struct{}),
	}
}

func (n *Node) GetGraph() graph.Graph { return n.g }

func (n *Node) NodeID() string { return n.id }

func (n *Node) Properties() map[string]interface{} { return n.properties }

type listNodeIterator struct {
	at *list.Element
}

func (l *listNodeIterator) Next() graph.Node {
	if l.at == nil {
		return nil
	}
	ret := l.at.Value.(*Node)
	l.at = l.at.Next()
	return ret
}

func (l *listNodeIterator) Rest() []graph.Node {
	ret := make([]graph.Node, 0)
	for ; l.at != nil; l.at = l.at.Next() {
		ret = append(ret, l.at.Value.(*Node))
	}
	return ret
}
