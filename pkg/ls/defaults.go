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
	"github.com/cloudprivacylabs/lpg/v2"
)

// PostIngestSchemaNode calls the post ingest functions for properties
// of the document nodes for the given schema node.
func (gb GraphBuilder) AddDefaults(schemaRootNode, docRootNode *lpg.Node) {
	if schemaRootNode == nil {
		return
	}
	ForEachAttributeNode(schemaRootNode, func(node *lpg.Node, path []*lpg.Node) bool {
		defValue, exists := GetPropertyValueAs[string](node, DefaultValueTerm.Name)
		if !exists {
			return true
		}
		if len(path) == 1 {
			return true
		}
		parentSchemaNode := path[len(path)-2]
		for _, parentDocNode := range GetNodesInstanceOf(docRootNode.GetGraph(), GetNodeID(parentSchemaNode)) {
			switch GetIngestAs(node) {
			case "node", "edge":
				child := FindChildInstanceOf(parentDocNode, GetNodeID(node))
				if len(child) == 0 {
					// Ingest default as node
					gb.RawValueAsNode(node, parentDocNode, defValue)
				}
			case "property":
				_, propertyName := GetIngestAsProperty(node)
				if _, exists := parentDocNode.GetProperty(propertyName); !exists {
					// Ingest default as property
					parentDocNode.SetProperty(propertyName, NewPropertyValue(NodeValueTerm.Name, defValue))
				}
			}
		}
		return true
	})
}
