// Copyright 2021 Cloud Privacy Labs, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ls

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudprivacylabs/opencypher/graph"
)

// IngestAs constants
const (
	IngestAsNode     = "node"
	IngestAsEdge     = "edge"
	IngestAsProperty = "property"
)

// NodePath contains the name components identifying a node. For JSON,
// this is the components of a JSON pointer
type NodePath []string

// String returns '.' combined path components
func (n NodePath) String() string {
	return strings.Join([]string(n), ".")
}

// Create a deep-copy of the nodepath
func (n NodePath) Copy() NodePath {
	ret := make(NodePath, len(n))
	copy(ret, n)
	return ret
}

func (n NodePath) AppendString(s string) NodePath {
	return append(n, s)
}

func (n NodePath) AppendInt(i int) NodePath {
	return append(n, strconv.Itoa(i))
}

func (n NodePath) Append(i interface{}) NodePath {
	return n.AppendString(fmt.Sprint(i))
}

type ParsedDocNode interface {
	GetSchemaNode() graph.Node

	// Returns value, object, array
	GetTypeTerm() string

	GetID() string

	GetValue() string
	GetValueTypes() []string

	GetAttributeIndex() int
	GetAttributeName() string

	GetChildren() []ParsedDocNode

	GetProperties() map[string]interface{}
}

type ingestCursor struct {
	input  []ParsedDocNode
	output []graph.Node
}

func (i ingestCursor) getInput() ParsedDocNode {
	return i.input[len(i.input)-1]
}

func (i ingestCursor) getOutput() graph.Node {
	if len(i.output) == 0 {
		return nil
	}
	return i.output[len(i.output)-1]
}

func Ingest(builder GraphBuilder, root ParsedDocNode) (graph.Node, error) {
	cursor := ingestCursor{
		input: []ParsedDocNode{root},
	}
	return ingestWithCursor(builder, cursor)
}

// GetIngestAs returns "node", "edge", "property", or "none" based on IngestAsTerm
func GetIngestAs(schemaNode graph.Node) string {
	if schemaNode == nil {
		return "node"
	}
	p, ok := schemaNode.GetProperty(IngestAsTerm)
	if !ok {
		return "node"
	}
	s := AsPropertyValue(p, ok).AsString()
	if s == "edge" || s == "property" || s == "none" {
		return s
	}
	return "node"
}

func ingestWithCursor(builder GraphBuilder, cursor ingestCursor) (graph.Node, error) {
	root := cursor.getInput()
	schemaNode := root.GetSchemaNode()
	typeTerm := root.GetTypeTerm()
	setID := func(node graph.Node) {
		if node != nil {
			if id := root.GetID(); len(id) > 0 {
				SetNodeID(node, id)
			}
		}
	}
	setProp := func(node graph.Node) {
		node.SetProperty(AttributeIndexTerm, StringPropertyValue(strconv.Itoa(root.GetAttributeIndex())))
		if s := root.GetAttributeName(); len(s) > 0 {
			node.SetProperty(AttributeNameTerm, StringPropertyValue(s))
		}
		for k, v := range root.GetProperties() {
			node.SetProperty(k, v)
		}
	}
	if typeTerm == AttributeTypeValue {
		switch GetIngestAs(schemaNode) {
		case "node":
			_, node, err := builder.ValueAsNode(schemaNode, cursor.getOutput(), root.GetValue(), root.GetValueTypes()...)
			if node != nil {
				setID(node)
				setProp(node)
			}
			return node, err
		case "edge":
			edge, err := builder.ValueAsEdge(schemaNode, cursor.getOutput(), root.GetValue(), root.GetValueTypes()...)
			if err != nil {
				return nil, err
			}
			if edge == nil {
				return nil, nil
			}
			setID(edge.GetTo())
			setProp(edge.GetTo())
			return edge.GetTo(), nil
		case "property":
			err := builder.ValueAsProperty(schemaNode, cursor.output, root.GetValue())
			if err != nil {
				return nil, err
			}
			return cursor.getOutput(), nil
		case "none":
			return cursor.getOutput(), nil
		}
		return nil, nil
	}
	var parentNode graph.Node
	switch GetIngestAs(schemaNode) {
	case "node":
		_, node, err := builder.CollectionAsNode(schemaNode, cursor.getOutput(), typeTerm, root.GetValueTypes()...)
		if err != nil {
			return nil, err
		}
		setID(node)
		setProp(node)
		parentNode = node
	case "edge":
		edge, err := builder.CollectionAsEdge(schemaNode, cursor.getOutput(), typeTerm, root.GetValueTypes()...)
		if err != nil {
			return nil, err
		}
		setID(edge.GetTo())
		setProp(edge.GetTo())
		parentNode = edge.GetTo()
	case "none":
		parentNode = cursor.getOutput()
	}
	if parentNode != nil {
		newCursor := cursor
		newCursor.input = append(newCursor.input, nil)
		if parentNode != cursor.getOutput() { // This happens if ingestAs==none.
			newCursor.output = append(newCursor.output, parentNode)
		}
		for _, child := range root.GetChildren() {
			newCursor.input[len(newCursor.input)-1] = child
			node, err := ingestWithCursor(builder, newCursor)
			if err != nil {
				return nil, err
			}
			if node != nil {
				n := child.GetAttributeIndex()
				if parentNode != nil {
					n = parentNode.GetEdges(graph.OutgoingEdge).MaxSize() - 1
				}
				if n == -1 {
					n = child.GetAttributeIndex()
				}
				node.SetProperty(AttributeIndexTerm, IntPropertyValue(n))
			}
		}
		if schemaNode != nil {
			switch AsPropertyValue(schemaNode.GetProperty(ConditionalTerm)).AsString() {
			case "mustHaveChildren":
				if parentNode.GetEdges(graph.OutgoingEdge).MaxSize() == 0 {
					parentNode.DetachAndRemove()
					parentNode = nil
				}
			}
		}
		return parentNode, nil
	}
	return nil, ErrInvalidSchema(fmt.Sprintf("Invalid input node type: %s", typeTerm))
}
