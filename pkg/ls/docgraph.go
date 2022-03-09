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
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

// NewDocumentGraph creates a new graph with the correct indexes for document ingestion
func NewDocumentGraph() graph.Graph {
	g := graph.NewOCGraph()
	g.AddNodePropertyIndex(EntitySchemaTerm)
	g.AddNodePropertyIndex(SchemaNodeIDTerm)
	return g
}

// EntityInfo contains the entity information in the doc graph
type EntityInfo struct {
	root graph.Node
	sch  string
}

func (e EntityInfo) GetRoot() graph.Node     { return e.root }
func (e EntityInfo) GetEntitySchema() string { return e.sch }
func (e EntityInfo) GetID() []string {
	return AsPropertyValue(e.root.GetProperty(EntityIDTerm)).MustStringSlice()
}

// GetEntityRootNodes returns all the nodes that are entity roots
func GetEntityRootNodes(g graph.Graph) map[graph.Node]EntityInfo {
	ret := make(map[graph.Node]EntityInfo)
	for nodes := g.GetNodesWithProperty(EntitySchemaTerm); nodes.Next(); {
		node := nodes.Node()
		sch := AsPropertyValue(node.GetProperty(EntitySchemaTerm)).AsString()
		if len(sch) > 0 {
			ret[node] = EntityInfo{root: node, sch: sch}
		}
	}
	return ret
}

// GetEntityIDFields returns the value of the entity ID fields from a document node
func GetEntityIDFields(node graph.Node) *PropertyValue {
	if node == nil {
		return nil
	}
	idFields, _ := GetNodeOrSchemaProperty(node, EntityIDFieldsTerm)
	return idFields
}
