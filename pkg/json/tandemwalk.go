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

package json

import (
	"fmt"

	"github.com/bserdar/jsonom"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
)

// TWCursor is the tandem walk cursor.
type TWCursor struct {
	SchemaPath []graph.Node
	DocPath    []jsonom.Node
	Path       ls.NodePath
}

func NewTWCursor(schemaNode graph.Node, docNode jsonom.Node) TWCursor {
	return TWCursor{
		SchemaPath: []graph.Node{schemaNode},
		DocPath:    []jsonom.Node{docNode},
	}
}

func (tw TWCursor) GetDocNode() jsonom.Node {
	return tw.DocPath[len(tw.DocPath)-1]
}

func (tw TWCursor) GetSchemaNode() graph.Node {
	return tw.SchemaPath[len(tw.SchemaPath)-1]
}

func (tw TWCursor) Push(schemaNode graph.Node, docNode jsonom.Node, path interface{}) TWCursor {
	return TWCursor{
		SchemaPath: append(tw.SchemaPath, schemaNode),
		DocPath:    append(tw.DocPath, docNode),
		Path:       tw.Path.Append(path),
	}
}

type WalkType int

const (
	// WalkDocNodes will walk all doc nodes, even if there is not a
	// matching schema node
	WalkDocNodes WalkType = iota

	// WalkSchemaNodes will walk all schema nodes, even if there is
	// not a matching doc node. This will not process the schema
	// recursively
	WalkSchemaNodes

	// WalkPairs will only walk doc nodes with matching schema
	WalkPairs
)

// TandemWalk will walk the doc and schema at the current root,
// calling `each` for each entry. Processing stops if `each` returns
// false
func TandemWalk(tw TWCursor, w WalkType, each func(TWCursor) (bool, error)) (bool, error) {
	switch w {
	case WalkDocNodes:
		return TandemWalkDocNodes(tw, each)
	case WalkSchemaNodes:
		return false, fmt.Errorf("Tandem walk based on schema nodes is not implemented yet")
	case WalkPairs:
		return TandemWalkPairs(tw, each)
	}
	return false, nil
}

type walkFunc func(TWCursor, func(TWCursor) (bool, error)) (bool, error)

func TandemWalkDocNodes(tw TWCursor, each func(TWCursor) (bool, error)) (bool, error) {
	switch tw.GetDocNode().(type) {
	case *jsonom.Object:
		return tandemWalkObjectDocNodes(tw, each, TandemWalkDocNodes)
	case *jsonom.Array:
		return tandemWalkArrayDocNodes(tw, each, TandemWalkDocNodes)
	case *jsonom.Value:
		return tandemWalkValueDocNodes(tw, each)
	}
	return true, nil
}

func TandemWalkPairs(tw TWCursor, each func(TWCursor) (bool, error)) (bool, error) {
	if tw.GetSchemaNode() == nil {
		return true, nil
	}
	switch tw.GetDocNode().(type) {
	case *jsonom.Object:
		return tandemWalkObjectDocNodes(tw, each, TandemWalkPairs)
	case *jsonom.Array:
		return tandemWalkArrayDocNodes(tw, each, TandemWalkPairs)
	case *jsonom.Value:
		return tandemWalkValueDocNodes(tw, each)
	}
	return true, nil
}

func tandemWalkObjectDocNodes(tw TWCursor, each func(TWCursor) (bool, error), walk walkFunc) (bool, error) {
	// doc node keys drive the walk
	node := tw.GetDocNode().(*jsonom.Object)
	schNode := tw.GetSchemaNode()
	if schNode != nil && !schNode.GetLabels().Has(ls.AttributeTypeObject) {
		return false, ls.ErrSchemaValidation{Msg: "An object attribute is expected here", Path: tw.Path.Copy()}
	}
	if r, err := each(tw); !r || err != nil {
		return r, err
	}

	nextNodes, err := ls.GetObjectAttributeNodesBy(schNode, ls.AttributeNameTerm)
	if err != nil {
		return false, err
	}

	for i := 0; i < node.Len(); i++ {
		nextKV := node.N(i)
		schNodes := nextNodes[nextKV.Key()]
		if len(schNodes) > 1 {
			return false, ls.ErrInvalidSchema(fmt.Sprintf("Multiple elements with key '%s'", nextKV.Key()))
		}
		var childSchemaNode graph.Node
		if len(schNodes) == 1 {
			childSchemaNode = schNodes[0]
		}
		if r, err := walk(tw.Push(childSchemaNode, nextKV.Value(), nextKV.Key()), each); !r || err != nil {
			return r, err
		}
	}
	return true, nil
}

func tandemWalkArrayDocNodes(tw TWCursor, each func(TWCursor) (bool, error), walk walkFunc) (bool, error) {
	// doc node keys drive the walk
	node := tw.GetDocNode().(*jsonom.Array)
	schNode := tw.GetSchemaNode()
	if schNode != nil && !schNode.GetLabels().Has(ls.AttributeTypeArray) {
		return false, ls.ErrSchemaValidation{Msg: "An array attribute is expected here", Path: tw.Path.Copy()}
	}
	if r, err := each(tw); !r || err != nil {
		return r, err
	}

	elementsNode := ls.GetArrayElementNode(schNode)
	for index := 0; index < node.Len(); index++ {
		if r, err := walk(tw.Push(elementsNode, node.N(index), index), each); !r || err != nil {
			return r, err
		}
	}
	return true, nil
}

func tandemWalkValueDocNodes(tw TWCursor, each func(TWCursor) (bool, error)) (bool, error) {
	// doc node keys drive the walk
	schNode := tw.GetSchemaNode()
	if schNode != nil && !schNode.GetLabels().Has(ls.AttributeTypeValue) {
		return false, ls.ErrSchemaValidation{Msg: "A value attribute is expected here", Path: tw.Path.Copy()}
	}
	return each(tw)
}
