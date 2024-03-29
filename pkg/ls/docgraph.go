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
	"strings"

	"github.com/cloudprivacylabs/lpg/v2"
)

// NewDocumentGraph creates a new graph with the correct indexes for document ingestion
func NewDocumentGraph() *lpg.Graph {
	g := lpg.NewGraph()
	g.AddNodePropertyIndex(EntitySchemaTerm.Name, lpg.BtreeIndex)
	g.AddNodePropertyIndex(SchemaNodeIDTerm.Name, lpg.HashIndex)
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
	root      *lpg.Node
	sch       string
	valueType []string
	id        []string
}

func (e EntityInfo) GetRoot() *lpg.Node      { return e.root }
func (e EntityInfo) GetEntitySchema() string { return e.sch }
func (e EntityInfo) GetID() []string         { return e.id }
func (e EntityInfo) GetValueType() []string  { return e.valueType }

func (e EntityInfo) String() string {
	return fmt.Sprintf(`rootNode: %v
schema: %s"
id: %v
`, e.root, e.sch, e.id)
}

// GetEntityInfo returns all the nodes that are entity roots,
// i.e. nodes containing EntitySchemaTerm
//
// Deprecated
func GetEntityInfo(g *lpg.Graph) map[*lpg.Node]EntityInfo {
	ret := make(map[*lpg.Node]EntityInfo)
	for nodes := g.GetNodesWithProperty(EntitySchemaTerm.Name); nodes.Next(); {
		node := nodes.Node()
		sch := EntitySchemaTerm.PropertyValue(node)
		if len(sch) > 0 {
			types := FilterNonLayerTypes(node.GetLabels().Slice())
			ret[node] = EntityInfo{
				root:      node,
				sch:       sch,
				valueType: types,
				id:        EntityIDTerm.PropertyValue(node),
			}
		}
	}
	return ret
}

// GetEntityRootsByID returns all nodes that have entity id
func GetEntityRootsByID(g *lpg.Graph) map[*lpg.Node]EntityInfo {
	ret := make(map[*lpg.Node]EntityInfo)
	for nodes := g.GetNodesWithProperty(EntityIDTerm.Name); nodes.Next(); {
		node := nodes.Node()
		id := EntityIDTerm.PropertyValue(node)
		if len(id) > 0 {
			sch := EntitySchemaTerm.PropertyValue(node)
			types := FilterNonLayerTypes(node.GetLabels().Slice())
			ret[node] = EntityInfo{
				root:      node,
				sch:       sch,
				valueType: types,
				id:        id,
			}
		}
	}
	return ret
}

// GetEntityInfoIndex returns a fast-access entity info
func GetEntityInfoIndex(g *lpg.Graph) EntityInfoIndex {
	return IndexEntityInfo(GetEntityInfo(g))
}

type EntityInfoIndex struct {
	indexByType map[string]map[string][]*lpg.Node
}

func (e EntityInfoIndex) getFkHash(fk []string) string {
	if len(fk) == 1 {
		return fk[0]
	}
	return strings.Join(fk, " ")
}

func (e EntityInfoIndex) Find(entityName string, fk []string) []*lpg.Node {
	m := e.indexByType[entityName]
	if m == nil {
		return nil
	}
	h := e.getFkHash(fk)
	return m[h]
}

// IndexEntityInfo returns a fast-access version of entity info
func IndexEntityInfo(entityInfo map[*lpg.Node]EntityInfo) EntityInfoIndex {
	ix := EntityInfoIndex{
		indexByType: make(map[string]map[string][]*lpg.Node),
	}

	add := func(t, hash string, node *lpg.Node) {
		m := ix.indexByType[t]
		if m == nil {
			m = make(map[string][]*lpg.Node)
			ix.indexByType[t] = m
		}
		m[hash] = append(m[hash], node)
	}
	for node, ei := range entityInfo {
		hash := ix.getFkHash(ei.GetID())
		add(ei.sch, hash, node)
		for _, t := range ei.valueType {
			if t != ei.sch {
				add(t, hash, node)
			}
		}
	}
	return ix
}

// GetParentDocumentNodes returns the document nodes that have incoming edges to this node
func GetParentDocumentNodes(node *lpg.Node) []*lpg.Node {
	out := make(map[*lpg.Node]struct{})
	for edges := node.GetEdges(lpg.IncomingEdge); edges.Next(); {
		edge := edges.Edge()
		ancestor := edge.GetFrom()
		if !ancestor.GetLabels().Has(DocumentNodeTerm.Name) {
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
		if _, ok := root.GetProperty(EntitySchemaTerm.Name); ok {
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
			if !ancestor.GetLabels().Has(DocumentNodeTerm.Name) {
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
	_, ok := node.GetProperty(EntitySchemaTerm.Name)
	return ok
}

// GetEntityIDFields returns the value of the entity ID fields from a document node.
func GetEntityIDFields(node *lpg.Node) PropertyValue {
	if node == nil {
		return PropertyValue{}
	}
	idFields, _ := GetNodeOrSchemaProperty(node, EntityIDFieldsTerm.Name)
	return idFields
}

// GetNodesInstanceOf returns document nodes that are instance of the given attribute id
func GetNodesInstanceOf(g *lpg.Graph, attrId string) []*lpg.Node {
	pattern := lpg.Pattern{{
		Properties: map[string]any{
			SchemaNodeIDTerm.Name: SchemaNodeIDTerm.MustPropertyValue(attrId),
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
	p, ok := GetNodeOrSchemaProperty(n, SchemaNodeIDTerm.Name)
	if !ok {
		return false
	}
	return p.Value() == schemaNodeID
}

// GetSchemaNodeIDMap returns a map of schema node IDs to slices of
// nodes that are instances of those schema nodes
func GetSchemaNodeIDMap(docRoot *lpg.Node) map[string][]*lpg.Node {
	ret := make(map[string][]*lpg.Node)
	IterateDescendants(docRoot, func(node *lpg.Node) bool {
		nodeId := SchemaNodeIDTerm.PropertyValue(node)
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
	if idFieldsProp.Value() == nil {
		return nil
	}
	idFields := EntityIDFieldsTerm.PropertyValue(entityRootNode)
	if len(idFields) == 0 {
		return nil
	}
	if len(idFields) == 1 {
		if schemaNodeID != idFields[0] {
			return nil
		}
		entityRootNode.SetProperty(EntityIDTerm.Name, EntityIDTerm.MustPropertyValue(value))
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
	existingEntityIDSlice := EntityIDTerm.PropertyValue(entityRootNode)
	for len(existingEntityIDSlice) <= idIndex {
		existingEntityIDSlice = append(existingEntityIDSlice, "")
	}
	existingEntityIDSlice[idIndex] = value
	entityRootNode.SetProperty(EntityIDTerm.Name, EntityIDTerm.MustPropertyValue(existingEntityIDSlice))
	return nil
}

// SetEntityIDVectorElementFromNode sets the entity Id component of
// the entity root node if the schema node is part of an entity
// id. The schema node ID and entity root node are found based on the
// given node
func SetEntityIDVectorElementFromNode(docNode *lpg.Node, value string) error {
	schemaNodeID := SchemaNodeIDTerm.PropertyValue(docNode)
	if len(schemaNodeID) == 0 {
		return nil
	}
	entityRootNode := GetEntityRootNode(docNode)
	if entityRootNode == nil {
		return nil
	}
	return SetEntityIDVectorElement(entityRootNode, schemaNodeID, value)
}

// AttributeReference points to an attribute in a document. The
// attribute can be a node, or a property of a node
type AttributeReference struct {
	Node     *lpg.Node
	Property string
}

func (a AttributeReference) IsProperty() bool { return len(a.Property) > 0 }

func (a AttributeReference) AsPropertyValue() (PropertyValue, bool) {
	return GetPropertyValue(a.Node, a.Property)
}

// GetAttributeReferenceBySchemaNode returns the attribute reference
// whose instance is under docContextNode
func GetAttributeReferenceBySchemaNode(schemaRootNode, schemaNode *lpg.Node, docContextNode *lpg.Node) (AttributeReference, bool) {
	path := GetPathFromRoot(schemaNode)
	return GetAttributeReferenceBySchemaPath(path, docContextNode)
}

func GetAttributeReferenceBySchemaPath(schemaPath []*lpg.Node, docContextNode *lpg.Node) (AttributeReference, bool) {
	if len(schemaPath) == 0 {
		return AttributeReference{}, false
	}
	attr := GetNodeID(schemaPath[len(schemaPath)-1])
	// If ingestAs for the schema node is "property", get the parent node
	switch GetIngestAs(schemaPath[len(schemaPath)-1]) {
	case "node", "edge":
		var found *lpg.Node
		IterateDescendants(docContextNode, func(node *lpg.Node) bool {
			instance := SchemaNodeIDTerm.PropertyValue(node)
			if instance == attr {
				found = node
				return false
			}
			return true
		}, FollowEdgesInEntity, false)
		if found != nil {
			return AttributeReference{
				Node: found,
			}, true
		}
		return AttributeReference{}, false

	case "property":
		asPropertyOf, propertyName := GetIngestAsProperty(schemaPath[len(schemaPath)-1])
		var targetSchemaNode *lpg.Node
		if len(asPropertyOf) == 0 {
			if len(schemaPath) < 2 {
				return AttributeReference{}, false
			}
			targetSchemaNode = schemaPath[len(schemaPath)-2]
		} else {
			// Find ancestor that is instance of asPropertyOf
			for i := len(schemaPath) - 2; i >= 0; i-- {
				if SchemaNodeIDTerm.PropertyValue(schemaPath[i]) == asPropertyOf {
					targetSchemaNode = schemaPath[i]
					break
				}
			}
		}
		if targetSchemaNode != nil {
			var found *lpg.Node
			targetSchemaNodeID := GetNodeID(targetSchemaNode)
			IterateDescendants(docContextNode, func(node *lpg.Node) bool {
				if SchemaNodeIDTerm.PropertyValue(node) == targetSchemaNodeID {
					found = node
					return false
				}
				return true
			}, OnlyDocumentNodes, false)
			if found != nil {
				return AttributeReference{
					Node:     found,
					Property: propertyName,
				}, true
			}
			return AttributeReference{}, false
		}
	}
	return AttributeReference{}, false
}
