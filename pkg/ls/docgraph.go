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
	"github.com/cloudprivacylabs/lpg"
)

// NewDocumentGraph creates a new graph with the correct indexes for document ingestion
func NewDocumentGraph() *lpg.Graph {
	g := lpg.NewGraph()
	g.AddNodePropertyIndex(EntitySchemaTerm, lpg.BtreeIndex)
	g.AddNodePropertyIndex(SchemaNodeIDTerm, lpg.HashIndex)
	for _, f := range newDocGraphHooks {
		f(g)
	}
	return g
}

var newDocGraphHooks = []func(*lpg.Graph){}

func RegisterNewDocGraphHook(f func(*lpg.Graph)) {
	newDocGraphHooks = append(newDocGraphHooks, f)
}

// EntityInfo contains the entity information in the doc graph
type EntityInfo struct {
	root *lpg.Node
	sch  string
}

func (e EntityInfo) GetRoot() *lpg.Node      { return e.root }
func (e EntityInfo) GetEntitySchema() string { return e.sch }
func (e EntityInfo) GetID() []string {
	return AsPropertyValue(e.root.GetProperty(EntityIDTerm)).MustStringSlice()
}
func (e EntityInfo) GetValueType() []string {
	return FilterNonLayerTypes(e.root.GetLabels().Slice())
}

// GetEntityInfo returns all the nodes that are entity roots,
// i.e. nodes containing EntitySchemaTerm
func GetEntityInfo(g *lpg.Graph) map[*lpg.Node]EntityInfo {
	ret := make(map[*lpg.Node]EntityInfo)
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
func GetParentDocumentNodes(node *lpg.Node) []*lpg.Node {
	out := make(map[*lpg.Node]struct{})
	for edges := node.GetEdges(lpg.IncomingEdge); edges.Next(); {
		edge := edges.Edge()
		ancestor := edge.GetFrom()
		if !ancestor.GetLabels().Has(DocumentNodeTerm) {
			continue
		}
		out[ancestor] = struct{}{}
	}
	ret := make([]*lpg.Node, 0, len(out))
	for x := range out {
		ret = append(ret, x)
	}
	return ret
}

// GetEntityRoot tries to find the entity containing this node by
// going backwards until a node with EntitySchemaTerm
func GetEntityRoot(node *lpg.Node) *lpg.Node {
	var find func(*lpg.Node) *lpg.Node
	seen := make(map[*lpg.Node]struct{})
	find = func(root *lpg.Node) *lpg.Node {
		if _, ok := root.GetProperty(EntitySchemaTerm); ok {
			return root
		}
		if _, ok := seen[root]; ok {
			return nil
		}
		seen[root] = struct{}{}
		seenAncestor := false
		var ret *lpg.Node
		for edges := root.GetEdges(lpg.IncomingEdge); edges.Next(); {
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

// IsEntityRoot returns true if the node is an entity root
func IsEntityRoot(node *lpg.Node) bool {
	_, ok := node.GetProperty(EntitySchemaTerm)
	return ok
}

// GetEntityIDFields returns the value of the entity ID fields from a document node.
func GetEntityIDFields(node *lpg.Node) *PropertyValue {
	if node == nil {
		return nil
	}
	idFields, _ := GetNodeOrSchemaProperty(node, EntityIDFieldsTerm)
	return idFields
}

// GetNodesInstanceOf returns document nodes that are instance of the given attribute id
func GetNodesInstanceOf(g *lpg.Graph, attrId string) []*lpg.Node {
	pattern := lpg.Pattern{{
		Labels: lpg.NewStringSet(DocumentNodeTerm),
		Properties: map[string]interface{}{
			SchemaNodeIDTerm: StringPropertyValue(SchemaNodeIDTerm, attrId),
		},
	}}
	nodes, err := pattern.FindNodes(g, nil)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IsInstanceOf returns true if g is an instance of the schema node
func IsInstanceOf(n *lpg.Node, schemaNodeID string) bool {
	p, ok := GetNodeOrSchemaProperty(n, SchemaNodeIDTerm)
	if !ok {
		return false
	}
	return p.AsString() == schemaNodeID
}

// GetSchemaNodeIDMap returns a map of schema node IDs to slices of
// nodes that are instances of those schema nodes
func GetSchemaNodeIDMap(docRoot *lpg.Node) map[string][]*lpg.Node {
	ret := make(map[string][]*lpg.Node)
	IterateDescendants(docRoot, func(node *lpg.Node) bool {
		nodeId := AsPropertyValue(node.GetProperty(SchemaNodeIDTerm)).AsString()
		if len(nodeId) > 0 {
			ret[nodeId] = append(ret[nodeId], node)
		}
		return true
	}, OnlyDocumentNodes, false)
	return ret
}

// SetEntityIDVectorElement sets the entity Id component of the entity
// root node if the schema node is part of an entity id
func SetEntityIDVectorElement(entityRootNode *lpg.Node, schemaNodeID, value string) error {
	idFieldsProp := GetEntityIDFields(entityRootNode)
	if idFieldsProp == nil {
		return nil
	}
	if idFieldsProp.IsString() {
		if schemaNodeID != idFieldsProp.AsString() {
			return nil
		}
		entityRootNode.SetProperty(EntityIDTerm, StringPropertyValue(EntityIDTerm, value))
		return nil
	}

	idFields := idFieldsProp.MustStringSlice()
	if len(idFields) == 0 {
		return nil
	}
	idIndex := -1
	for i, idField := range idFields {
		if schemaNodeID == idField {
			idIndex = i
			break
		}
	}
	// Is this an ID field?
	if idIndex == -1 {
		return nil
	}

	// Get existing ID
	entityID := AsPropertyValue(entityRootNode.GetProperty(EntityIDTerm))
	existingEntityIDSlice := entityID.MustStringSlice()
	for len(existingEntityIDSlice) <= idIndex {
		existingEntityIDSlice = append(existingEntityIDSlice, "")
	}
	existingEntityIDSlice[idIndex] = value
	entityRootNode.SetProperty(EntityIDTerm, StringSlicePropertyValue(EntityIDTerm, existingEntityIDSlice))
	return nil
}

// SetEntityIDVectorElementFromNode sets the entity Id component of
// the entity root node if the schema node is part of an entity
// id. The schema node ID and entity root node are found based on the
// given node
func SetEntityIDVectorElementFromNode(docNode *lpg.Node, value string) error {
	schemaNodeID := AsPropertyValue(docNode.GetProperty(SchemaNodeIDTerm)).AsString()
	if len(schemaNodeID) == 0 {
		return nil
	}
	entityRootNode := GetEntityRootNode(docNode)
	if entityRootNode == nil {
		return nil
	}
	return SetEntityIDVectorElement(entityRootNode, schemaNodeID, value)
}
