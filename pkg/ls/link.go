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
	"errors"

	"github.com/bserdar/digraph"
)

var ErrInvalidLink = errors.New("Invalid link")
var ErrMultipleLinkMatch = errors.New("Multiple values match for link")
var ErrMultipleValuesForLink = errors.New("Link ID vector has multiple values")

// LinkField links an ingested document node to the node it is linked
// to in the graph.The `documentRoot` is the root node of the document
// containing this link. The `referenceSchemaNode` is the compiled
// schema node that has the reference annotations. This function
// creates a node that is linked to the instance of the target
// object. The target object is found by constructing a vector
// []referenceIDValue from the document, and then locating an object
// whose contents of []referenceIDField matched []referenceIDValue.
func LinkField(documentRoot, documentNode Node, referenceSchemaNode Node, graph *digraph.Graph) error {
	referenceIDField := referenceSchemaNode.GetProperties()[ReferenceIDFieldTerm]
	if referenceIDField == nil {
		return nil
	}
	referenceIDValue := referenceSchemaNode.GetProperties()[ReferenceIDValueTerm]
	if referenceIDValue == nil {
		return nil
	}
	var idFields []string
	if referenceIDField.IsString() {
		idFields = []string{referenceIDField.AsString()}
	} else {
		idFields = referenceIDField.AsStringSlice()
	}
	var idValueFields []string
	if referenceIDValue.IsString() {
		idValueFields = []string{referenceIDValue.AsString()}
	} else {
		idValueFields = referenceIDValue.AsStringSlice()
	}
	if len(idValueFields) != len(idFields) {
		return ErrInvalidLink
	}

	idVector, err := GetLinkIDVector(documentRoot, idValueFields...)
	if err != nil {
		return err
	}

	var foundNode Node
	types := FilterNonLayerTypes(referenceSchemaNode.GetTypes().Slice())
	for nodes := graph.GetAllNodes(); nodes.HasNext(); {
		node := nodes.Next().(Node)
		nodeTypes := node.GetTypes()
		if !nodeTypes.Has(DocumentNodeTerm) {
			continue
		}
		// If the node has all the types of 'types', then check the ID vector
		found := false
		for _, t := range types {
			found = true
			if !nodeTypes.Has(t) {
				found = false
				break
			}
		}
		if !found {
			continue
		}
		id, err := GetLinkIDVector(node, idFields...)
		if err != nil {
			return err
		}
		eq := true
		for i := range id {
			if id[i] != idVector[i] {
				eq = false
				break
			}
		}
		if eq {
			if foundNode != nil {
				return ErrMultipleLinkMatch
			}
			foundNode = node
		}
	}
	if foundNode != nil {
		digraph.Connect(documentNode, foundNode, NewEdge(HasTerm))
	}
	return nil
}

// GetLinkIDVector returns a vector of values filled in from
// idValueFields, each of which are schema node IDs. The nodes under
// documentRoot that are instance of these schema node IDs are used to
// construct the link ID vector
func GetLinkIDVector(documentRoot Node, idValueFields ...string) ([]string, error) {
	idVector := make([]string, len(idValueFields))
	var err error
	IterateDescendants(documentRoot, func(node Node, _ []Node) bool {
		// If this node is an instance of one of the referenceIDValues, get its value
		inst := InstanceOfID(node)
		for _, t := range inst {
			for i := range idValueFields {
				if t == idValueFields[i] {
					value := node.GetValue()
					if value != nil {
						if len(idVector[i]) > 0 {
							err = ErrMultipleValuesForLink
							return false
						}
						idVector[i] = node.GetValue().(string)
					}
				}
			}
		}
		return true
	}, func(edge Edge, _ []Node) EdgeFuncResult {
		if edge.GetLabel() == HasTerm {
			return FollowEdgeResult
		}
		return SkipEdgeResult
	}, false)
	if err != nil {
		return nil, err
	}
	return idVector, nil
}
