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

	"github.com/cloudprivacylabs/lpg"
)

type ErrPathInstantiation struct {
	TargetNode string
	Msg        string
}

func (e ErrPathInstantiation) Error() string {
	return fmt.Sprintf("Path instantiation error while instantiating a path to %s: %s", e.TargetNode, e.Msg)
}

// EnsurePath finds or instantiates a path in the graph. The path will
// start at rootDocNode, which is an instance of rootSchemaNode, will
// pass through ancestorDocNode if ancestorDocNode is not nil, and
// will be an instance of schemaNode. Each node will be created using
// the instanate func that takes the parent document node, and the
// schema node to instantiate.
func EnsurePath(rootDocNode, ancestorDocNode *lpg.Node, rootSchemaNode, schemaNode *lpg.Node, instantiate func(parentDocNode, schemaNode *lpg.Node) (*lpg.Node, error)) (*lpg.Node, error) {
	// Find the path in schema from the root to the schema node. There
	// must be at most one.
	schemaPath := GetAttributePath(rootSchemaNode, schemaNode)
	if schemaPath == nil {
		return nil, ErrPathInstantiation{
			TargetNode: GetNodeID(schemaNode),
			Msg:        "Cannot find the path to target node in schema",
		}
	}
	docPathNode := rootDocNode
	// If ancestorDocNode exists, locate that node on the schema path.
	if ancestorDocNode != nil {
		ancestorSchema := GetNodeSchemaNodeID(ancestorDocNode)
		if len(ancestorSchema) == 0 {
			// Ancestor node is not an instance of an attribute in schema
			return nil, ErrPathInstantiation{
				TargetNode: GetNodeID(schemaNode),
				Msg:        "Ancestor node is not an instance of a schema node",
			}
		}
		found := false
		for i := range schemaPath {
			if GetNodeID(schemaPath[i]) == ancestorSchema {
				found = true
				schemaPath = schemaPath[i:]
				break
			}
		}
		if !found {
			// Ancestor not on path
			return nil, ErrPathInstantiation{
				TargetNode: GetNodeID(schemaNode),
				Msg:        "Ancestor not on path",
			}
		}
		docPathNode = ancestorDocNode
	}
	// Here, docPathNode points to a document node that is instance of schemaPath[0]
	// We descend from here
	if len(schemaPath) <= 1 {
		// We are already here
		return docPathNode, nil
	}

	for ; len(schemaPath) > 1; schemaPath = schemaPath[1:] {
		children := FindChildInstanceOf(docPathNode, GetNodeID(schemaPath[1]))
		switch len(children) {
		case 0:
			// Instantiate the schemaPath[1] under docPathNode
			newNode, err := instantiate(docPathNode, schemaPath[1])
			if err != nil {
				return nil, ErrPathInstantiation{
					TargetNode: GetNodeID(schemaNode),
					Msg:        fmt.Sprintf("Error while instantiating node for %s: %s", GetNodeID(schemaPath[1]), err.Error()),
				}
			}
			docPathNode = newNode

		case 1:
			docPathNode = children[0]

		default:
			return nil, ErrPathInstantiation{
				TargetNode: GetNodeID(schemaNode),
				Msg:        fmt.Sprintf("Multiple child nodes of %s", GetNodeID(schemaPath[1])),
			}
		}
	}
	return docPathNode, nil
}
