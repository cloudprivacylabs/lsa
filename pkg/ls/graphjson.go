package ls

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

// JSONGraphNode includes a node index, labels, and properties of the node
type JSONGraphNode struct {
	N          int                       `json:"n"`
	Value      interface{}               `json:"value,omitempty"`
	ID         string                    `json:"id,omitempty"`
	Labels     []string                  `json:"labels,omitempty"`
	Properties map[string]*PropertyValue `json:"properties,omitempty"`
}

// JSONGraphEdge includes the from-to node indexes, the label and properties
type JSONGraphEdge struct {
	From       int                       `json:"from"`
	To         int                       `json:"to"`
	Label      string                    `json:"label,omitempty"`
	Properties map[string]*PropertyValue `json:"properties,omitempty"`
}

type JSONGraph struct {
	Nodes []JSONGraphNode `json:"nodes"`
	Edges []JSONGraphEdge `json:"edges"`
}

// JSONMarshaler marshals/unmarshals a graph
type JSONMarshaler struct {
	interner Interner
}

func (m *JSONMarshaler) copyProperties(properties map[string]*PropertyValue) map[string]interface{} {
	ret := make(map[string]interface{})
	for k, v := range properties {
		ret[m.interner.Intern(k)] = v
	}
	return ret
}

// Marshal marshals the graph as a JSON document
func (m *JSONMarshaler) Marshal(g graph.Graph) ([]byte, error) {
	buf := bytes.Buffer{}
	err := m.Encode(g, &buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Encode writes the graph as a JSON document to the given writer
func (m *JSONMarshaler) Encode(g graph.Graph, w io.Writer) error {
	encodeNode := func(node graph.Node, index int) error {
		v := JSONGraphNode{
			N:      index,
			Labels: node.GetLabels().Slice(),
		}
		node.ForEachProperty(func(key string, value interface{}) bool {
			p, ok := value.(*PropertyValue)
			if !ok {
				return true
			}
			if v.Properties == nil {
				v.Properties = make(map[string]*PropertyValue)
			}
			v.Properties[key] = p
			return true
		})
		v.Value = GetRawNodeValue(node)
		v.ID = GetNodeID(node)
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	}

	encodeEdge := func(edge graph.Edge, index map[graph.Node]int) error {
		v := JSONGraphEdge{
			From:  index[edge.GetFrom()],
			To:    index[edge.GetTo()],
			Label: edge.GetLabel(),
		}
		edge.ForEachProperty(func(key string, value interface{}) bool {
			p, ok := value.(*PropertyValue)
			if !ok {
				return true
			}
			if v.Properties == nil {
				v.Properties = make(map[string]*PropertyValue)
			}
			v.Properties[key] = p
			return true
		})
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
	nodeIx := make(map[graph.Node]int)
	for nodes := g.GetNodes(); nodes.Next(); {
		if index > 0 {
			if _, err := w.Write([]byte{','}); err != nil {
				return err
			}
		}
		node := nodes.Node()
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
	for edges := g.GetEdges(); edges.Next(); {
		edge := edges.Edge()
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
	if _, err := io.WriteString(w, `]}`); err != nil {
		return err
	}
	return nil
}

// Unmarshal unmarshals a graph from JSON input
func (m *JSONMarshaler) Unmarshal(in []byte, targetGraph graph.Graph) error {
	if m.interner == nil {
		m.interner = NewInterner()
	}
	var g JSONGraph
	if err := json.Unmarshal(in, &g); err != nil {
		return err
	}
	nodeIx := make(map[int]graph.Node)
	for _, node := range g.Nodes {
		for i := range node.Labels {
			node.Labels[i] = m.interner.Intern(node.Labels[i])
		}
		newNode := targetGraph.NewNode(node.Labels, m.copyProperties(node.Properties))
		if node.Value != nil {
			SetRawNodeValue(newNode, node.Value)
		}
		if len(node.ID) > 0 {
			SetNodeID(newNode, node.ID)
		}
		nodeIx[node.N] = newNode
	}
	for _, edge := range g.Edges {
		from := nodeIx[edge.From]
		if from == nil {
			return fmt.Errorf("Invalid source node index: %d", edge.From)
		}
		to := nodeIx[edge.To]
		if to == nil {
			return fmt.Errorf("Invalid target node index: %d", edge.To)
		}
		targetGraph.NewEdge(from, to, m.interner.Intern(edge.Label), m.copyProperties(edge.Properties))
	}
	return nil
}
