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

	"github.com/cloudprivacylabs/opencypher/graph"
)

type ParsedDocNode interface {
	GetSchemaNode() graph.Node

	// Returns value, object, array
	GetTypeTerm() string

	GetValue() string
	GetValueTypes() []string

	GetChildren() []ParsedDocNode
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

func ingestWithCursor(builder GraphBuilder, cursor ingestCursor) (graph.Node, error) {
	root := cursor.getInput()
	schemaNode := root.GetSchemaNode()
	typeTerm := root.GetTypeTerm()
	if typeTerm == AttributeTypeValue {
		switch GetIngestAs(schemaNode) {
		case "node":
			_, node, err := builder.ValueAsNode(schemaNode, cursor.getOutput(), root.GetValue(), root.GetValueTypes()...)
			return node, err
		case "edge":
			edge, err := builder.ValueAsEdge(schemaNode, cursor.getOutput(), root.GetValue(), root.GetValueTypes()...)
			if err != nil {
				return nil, err
			}
			if edge == nil {
				return nil, nil
			}
			return edge.GetTo(), nil
		case "property":
			err := builder.ValueAsProperty(schemaNode, cursor.output, root.GetValue())
			if err != nil {
				return nil, err
			}
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
		parentNode = node
	case "edge":
		edge, err := builder.CollectionAsEdge(schemaNode, cursor.getOutput(), typeTerm, root.GetValueTypes()...)
		if err != nil {
			return nil, err
		}
		parentNode = edge.GetTo()
	}
	if parentNode != nil {
		newCursor := cursor
		newCursor.input = append(newCursor.input, nil)
		newCursor.output = append(newCursor.output, parentNode)
		for index, child := range root.GetChildren() {
			newCursor.input[len(newCursor.input)-1] = child
			node, err := ingestWithCursor(builder, cursor)
			if err != nil {
				return nil, err
			}
			if node != nil {
				node.SetProperty(AttributeIndexTerm, IntPropertyValue(index))
			}
		}
	}
	return nil, ErrInvalidSchema(fmt.Sprintf("Invalid input node type: %s", typeTerm))
}
