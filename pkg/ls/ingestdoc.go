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

	"github.com/cloudprivacylabs/lpg"
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
	GetSchemaNode() *lpg.Node

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

// HasNativeValue is implemented by parsed doc nodes if the node knows its native value
type HasNativeValue interface {
	GetNativeValue() (interface{}, bool)
}

type HasLabels interface {
	GetLabels() []string
}

type ingestCursor struct {
	input  []ParsedDocNode
	output []*lpg.Node
}

func (i ingestCursor) getInput() ParsedDocNode {
	return i.input[len(i.input)-1]
}

func (i ingestCursor) getOutput() *lpg.Node {
	if len(i.output) == 0 {
		return nil
	}
	return i.output[len(i.output)-1]
}

func Ingest(builder GraphBuilder, root ParsedDocNode) (*lpg.Node, error) {
	cursor := ingestCursor{
		input: []ParsedDocNode{root},
	}
	_, n, err := ingestWithCursor(builder, cursor)
	return n, err
}

// GetIngestAs returns "node", "edge", "property", or "none" based on IngestAsTerm
func GetIngestAs(schemaNode *lpg.Node) string {
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

func ingestWithCursor(builder GraphBuilder, cursor ingestCursor) (bool, *lpg.Node, error) {
	root := cursor.getInput()
	schemaNode := root.GetSchemaNode()
	typeTerm := root.GetTypeTerm()
	setID := func(node *lpg.Node) {
		if node != nil {
			if id := root.GetID(); len(id) > 0 {
				SetNodeID(node, id)
			}
		}
	}
	setLabels := func(node *lpg.Node) {
		lbl, ok := root.(HasLabels)
		if !ok {
			return
		}
		labels := node.GetLabels()
		labels.Add(lbl.GetLabels()...)
		node.SetLabels(labels)
	}
	setProp := func(node *lpg.Node) {
		node.SetProperty(AttributeIndexTerm, StringPropertyValue(AttributeIndexTerm, strconv.Itoa(root.GetAttributeIndex())))
		if s := root.GetAttributeName(); len(s) > 0 {
			node.SetProperty(AttributeNameTerm, StringPropertyValue(AttributeNameTerm, s))
		}
		for k, v := range root.GetProperties() {
			node.SetProperty(k, v)
		}
	}
	hasData := false
	if typeTerm == AttributeTypeValue {
		setValue := func(node *lpg.Node) error {
			SetRawNodeValue(node, root.GetValue())
			return nil
		}
		hasNativeValue := false
		var nativeValue interface{}
		nvi, hn := root.(HasNativeValue)
		if hn {
			nativeValue, hasNativeValue = nvi.GetNativeValue()
			if hasNativeValue {
				setValue = func(node *lpg.Node) error {
					return SetNodeValue(node, nativeValue)
				}
			}
		}
		switch GetIngestAs(schemaNode) {
		case "node":
			_, node, err := builder.ValueAsNode(schemaNode, cursor.getOutput(), setValue)
			if err != nil {
				return false, nil, err
			}
			if node != nil {
				setID(node)
				setProp(node)
				setLabels(node)
				hasData = true
			}
			return hasData, node, err
		case "edge":
			edge, err := builder.ValueAsEdge(schemaNode, cursor.getOutput(), setValue)
			if err != nil {
				return false, nil, err
			}
			if edge == nil {
				return false, nil, nil
			}
			setID(edge.GetTo())
			setProp(edge.GetTo())
			return true, edge.GetTo(), nil
		case "property":
			var err error
			if hasNativeValue {
				err = builder.NativeValueAsProperty(schemaNode, cursor.output, nativeValue)
			} else {
				err = builder.RawValueAsProperty(schemaNode, cursor.output, root.GetValue())
			}
			if err != nil {
				return false, nil, err
			}
			return true, nil, nil
		case "none":
			return false, nil, nil
		}
		return false, nil, nil
	}
	newCursor := cursor
	switch GetIngestAs(schemaNode) {
	case "node":
		_, node, err := builder.CollectionAsNode(schemaNode, cursor.getOutput(), typeTerm)
		if err != nil {
			return false, nil, err
		}
		setID(node)
		setProp(node)
		setLabels(node)
		newCursor.output = append(newCursor.output, node)
		hasData = true
	case "edge":
		edge, err := builder.CollectionAsEdge(schemaNode, cursor.getOutput(), typeTerm)
		if err != nil {
			return false, nil, err
		}
		setID(edge.GetTo())
		setProp(edge.GetTo())
		newCursor.output = append(newCursor.output, edge.GetTo())
		hasData = true
	case "none":
	}
	newCursor.input = append(newCursor.input, nil)
	hasChildren := false
	for _, child := range root.GetChildren() {
		newCursor.input[len(newCursor.input)-1] = child
		ch, node, err := ingestWithCursor(builder, newCursor)
		if ch {
			hasChildren = true
		}
		if err != nil {
			return hasData, nil, err
		}
		if node != nil {
			n := child.GetAttributeIndex()
			if newCursor.getOutput() != nil {
				n = newCursor.getOutput().GetEdges(lpg.OutgoingEdge).MaxSize() - 1
			}
			if n == -1 {
				n = child.GetAttributeIndex()
			}
			node.SetProperty(AttributeIndexTerm, IntPropertyValue(AttributeIndexTerm, n))
		}
	}
	if schemaNode != nil && hasData {
		switch AsPropertyValue(schemaNode.GetProperty(ConditionalTerm)).AsString() {
		case "mustHaveChildren":
			if !hasChildren {
				newCursor.getOutput().DetachAndRemove()
				return false, nil, nil
			}
		}
	}
	return hasData, newCursor.getOutput(), nil
}
