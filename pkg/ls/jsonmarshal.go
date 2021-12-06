package ls

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/bserdar/digraph"
)

type jsonGraphNode struct {
	N          int                       `json:"n"`
	Type       []string                  `json:"type,omitempty"`
	ID         string                    `json:"id,omitempty"`
	Value      interface{}               `json:"val,omitempty"`
	Properties map[string]*PropertyValue `json:"properties,omitempty"`
}

func (node jsonGraphNode) makeNode(interner Interner) Node {
	for i := range node.Type {
		node.Type[i] = interner.Intern(node.Type[i])
	}
	newNode := NewNode(node.ID, node.Type...)
	if node.Value != nil {
		newNode.SetValue(node.Value)
	}
	if len(node.Properties) > 0 {
		target := newNode.GetProperties()
		for k, v := range node.Properties {
			target[interner.Intern(k)] = v
		}
	}
	return newNode
}

type jsonGraphEdge struct {
	From       int                       `json:"fn"`
	To         int                       `json:"tn"`
	Label      string                    `json:"label,omitempty"`
	Properties map[string]*PropertyValue `json:"properties,omitempty"`
}

func (edge jsonGraphEdge) makeEdge(interner Interner) Edge {
	newEdge := NewEdge(interner.Intern(edge.Label))
	if len(edge.Properties) > 0 {
		target := newEdge.GetProperties()
		for k, v := range edge.Properties {
			target[interner.Intern(k)] = v
		}
	}
	return newEdge
}

// MarshalGraphJSON marshals the graph as a JSON document
func MarshalGraphJSON(graph *digraph.Graph) ([]byte, error) {
	buf := bytes.Buffer{}
	err := EncodeGraphJSON(graph, &buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// EncodeGraphJSON writes the graph as a JSON document to the given writer
func EncodeGraphJSON(graph *digraph.Graph, w io.Writer) error {
	encodeNode := func(node Node, index int) error {
		v := jsonGraphNode{
			N:    index,
			Type: node.GetTypes().Slice(),
			ID:   node.GetID(),
		}
		if val := node.GetValue(); val != nil {
			v.Value = val
		}
		properties := node.GetProperties()
		if len(properties) > 0 {
			v.Properties = properties
		}
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	}

	encodeEdge := func(edge Edge, index map[Node]int) error {
		v := jsonGraphEdge{
			From:  index[edge.GetFrom().(Node)],
			To:    index[edge.GetTo().(Node)],
			Label: edge.GetLabelStr(),
		}
		properties := edge.GetProperties()
		if len(properties) > 0 {
			v.Properties = properties
		}
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	}

	if _, err := io.WriteString(w, `{"nodes":[`); err != nil {
		return err
	}
	index := 0
	nodeIx := make(map[Node]int)
	for nodes := graph.GetAllNodes(); nodes.HasNext(); {
		if index > 0 {
			if _, err := w.Write([]byte{','}); err != nil {
				return err
			}
		}
		node := nodes.Next().(Node)
		nodeIx[node] = index
		if err := encodeNode(node, index); err != nil {
			return err
		}
		index++
	}
	if _, err := io.WriteString(w, `],"edges":[`); err != nil {
		return err
	}
	first := true
	for node, _ := range nodeIx {
		for edges := node.Out(); edges.HasNext(); {
			edge := edges.Next().(Edge)
			if !first {
				if _, err := w.Write([]byte{','}); err != nil {
					return err
				}
			} else {
				first = false
			}
			if err := encodeEdge(edge, nodeIx); err != nil {
				return err
			}
		}
	}
	if _, err := io.WriteString(w, `]}`); err != nil {
		return err
	}
	return nil
}

// UnmarshalGraphJSON unmarshals a graph from JSON input
func UnmarshalGraphJSON(in []byte, targetGraph *digraph.Graph, interner Interner) error {
	type graph struct {
		Nodes []jsonGraphNode `json:"nodes"`
		Edges []jsonGraphEdge `json:"edges"`
	}
	if interner == nil {
		interner = NewInterner()
	}
	var g graph
	if err := json.Unmarshal(in, &g); err != nil {
		return err
	}
	nodeIx := make(map[int]Node)
	for _, node := range g.Nodes {
		newNode := node.makeNode(interner)
		targetGraph.AddNode(newNode)
		nodeIx[node.N] = newNode
	}
	for _, edge := range g.Edges {
		newEdge := edge.makeEdge(interner)
		from := nodeIx[edge.From]
		if from == nil {
			return fmt.Errorf("Invalid source node index: %d", edge.From)
		}
		to := nodeIx[edge.To]
		if to == nil {
			return fmt.Errorf("Invalid target node index: %d", edge.To)
		}
		digraph.Connect(from, to, newEdge)
	}
	return nil
}
