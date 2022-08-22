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
	"github.com/cloudprivacylabs/opencypher/graph"
)

// NewDocumentGraph creates a new graph with the correct indexes for document ingestion
func NewDocumentGraph() graph.Graph {
	g := graph.NewOCGraph()
	g.AddNodePropertyIndex(EntitySchemaTerm)
	g.AddNodePropertyIndex(SchemaNodeIDTerm)
	for _, f := range newDocGraphHooks {
		f(g)
	}
	return g
}

var newDocGraphHooks = []func(*graph.OCGraph){}

func RegisterNewDocGraphHook(f func(*graph.OCGraph)) {
	newDocGraphHooks = append(newDocGraphHooks, f)
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
func (e EntityInfo) GetValueType() []string {
	return FilterNonLayerTypes(e.root.GetLabels().Slice())
}

// GetEntityInfo returns all the nodes that are entity roots,
// i.e. nodes containing EntitySchemaTerm
func GetEntityInfo(g graph.Graph) map[graph.Node]EntityInfo {
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

// GetParentDocumentNodes returns the document nodes that have incoming edges to this node
func GetParentDocumentNodes(node graph.Node) []graph.Node {
	out := make(map[graph.Node]struct{})
	for edges := node.GetEdges(graph.IncomingEdge); edges.Next(); {
		edge := edges.Edge()
		ancestor := edge.GetFrom()
		if !ancestor.GetLabels().Has(DocumentNodeTerm) {
			continue
		}
		out[ancestor] = struct{}{}
	}
	ret := make([]graph.Node, 0, len(out))
	for x := range out {
		ret = append(ret, x)
	}
	return ret
}

// GetEntityRoot tries to find the entity containing this node by
// going backwards until a node with EntitySchemaTerm
func GetEntityRoot(node graph.Node) graph.Node {
	var find func(graph.Node) graph.Node
	seen := make(map[graph.Node]struct{})
	find = func(root graph.Node) graph.Node {
		if _, ok := root.GetProperty(EntitySchemaTerm); ok {
			return root
		}
		if _, ok := seen[root]; ok {
			return nil
		}
		seen[root] = struct{}{}
		seenAncestor := false
		var ret graph.Node
		for edges := root.GetEdges(graph.IncomingEdge); edges.Next(); {
			edge := edges.Edge()
			ancestor := edge.GetFrom()
			if !ancestor.GetLabels().Has(DocumentNodeTerm) {
				continue
			}
			if seenAncestor {
				return nil
			}
			seenAncestor = true
			ret = find(ancestor)
		}
		return ret
	}
	return find(node)
}

// GetEntityIDFields returns the value of the entity ID fields from a document node.
func GetEntityIDFields(node graph.Node) *PropertyValue {
	if node == nil {
		return nil
	}
	idFields, _ := GetNodeOrSchemaProperty(node, EntityIDFieldsTerm)
	return idFields
}

// GetNodesInstanceOf returns document nodes that are instance of the given attribute id
func GetNodesInstanceOf(g graph.Graph, attrId string) []graph.Node {
	pattern := graph.Pattern{{
		Labels: graph.NewStringSet(DocumentNodeTerm),
		Properties: map[string]interface{}{
			SchemaNodeIDTerm: StringPropertyValue(GetTermInfo(SchemaNodeIDTerm).Term, attrId),
		},
	}}
	nodes, err := pattern.FindNodes(g, nil)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IsInstanceOf returns true if g is an instance of the schema node
func IsInstanceOf(n graph.Node, schemaNodeID string) bool {
	p, ok := GetNodeOrSchemaProperty(n, SchemaNodeIDTerm)
	if !ok {
		return false
	}
	return p.AsString() == schemaNodeID
}
