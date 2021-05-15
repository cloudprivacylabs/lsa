package graph

type Graph interface {
	GetNodes() NodeIterator
	GetEdges() EdgeIterator
	GetNodeByID(string) Node
	GetEdgeByID(string) Edge

	// Make a new node with the ID. If ID is empty, it is generated
	NewNode(string) Node
	RemoveNode(Node)
	// Make a new edge with the given ID and label. If ID is empty, it is generated
	NewEdge(id, label string, start, end Node) Edge
	RemoveEdge(Edge)
}

type Node interface {
	GetGraph() Graph
	// Unique node ID
	NodeID() string

	// AllOutgoingEdges() EdgeIterator
	// OutgoingEdgesByLabel(string) EdgeIterator

	// AllIncomingEdges() EdgeIterator
	// IncomingEdgesByLabel(string) EdgeIterator
	Properties() map[string]interface{}
}

type Edge interface {
	GetGraph() Graph
	// Unique edge ID
	EdgeID() string
	// Edge label is the label of the edge. That is Start --label--> End
	EdgeLabel() string
	StartNode() Node
	EndNode() Node
	Properties() map[string]interface{}
}

type NodeIterator interface {
	// Return the next node. Returns nil if iteration is complete
	Next() Node
	// Return all remaining nodes as an array
	Rest() []Node
}

type EdgeIterator interface {
	// Returns the next edge. Returns nil if iteration is complete.
	Next() Edge
	// Return all remaining edges as an array
	Rest() []Edge
}
