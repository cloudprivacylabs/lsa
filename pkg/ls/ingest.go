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

	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

// Ingester keeps the schema and the ingestion options
type Ingester struct {
	// The schema variant to use during ingestion
	Schema *Layer

	// If true, schema node properties are embedded into document
	// nodes. If false, schema nodes are preserved as separate nodes,
	// with an instanceOf edge between the document node to the schema
	// node.
	EmbedSchemaNodes bool

	// If true, a map[Node][]interface{} is populated to preserve the
	// paths used to create nodes
	PreserveNodePaths bool

	// If PreserveNodePaths is true, this keeps the node paths after ingestion.
	// This map is reset when Start is called.
	NodePaths map[graph.Node]NodePath

	// If OnlySchemaAttributes is true, only ingest data points if there is a schema for it.
	// If OnlySchemaAttributes is false, ingest whether or not there is a schema for it.
	OnlySchemaAttributes bool
}

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

type ErrSchemaValidation struct {
	Msg  string
	Path NodePath
}

func (e ErrSchemaValidation) Error() string {
	ret := "Schema validation error: " + e.Msg
	if e.Path != nil {
		ret += " path:" + e.Path.String()
	}
	return ret
}

type ErrInvalidSchema string

func (e ErrInvalidSchema) Error() string { return "Invalid schema: " + string(e) }

type ErrDataIngestion struct {
	Key string
	Err error
}

func (e ErrDataIngestion) Error() string {
	return fmt.Sprintf("Data ingestion error: Key: %s - %s", e.Key, e.Err)
}

func (e ErrDataIngestion) Unwrap() error { return e.Err }

func pathToString(path []interface{}) string {
	components := make([]string, 0, len(path)+1)
	for _, x := range path {
		components = append(components, fmt.Sprint(x))
	}
	return strings.Join(components, ".")
}

// DefaultNodeIDGenerator returns Ingester.Schema.ID + join(path,".")
func DefaultNodeIDGenerator(path NodePath, schemaNode graph.Node) string {
	return path.String()
}

// Start ingestion. Returns the path initialized with the baseId, and
// the schema root.
func (ingester *Ingester) Start(baseID string) (path NodePath, schemaRoot graph.Node) {
	path = make(NodePath, 0, 16)
	path = append(path, baseID)
	if ingester.Schema != nil {
		schemaRoot = ingester.Schema.GetSchemaRootNode()
	}
	ingester.NodePaths = make(map[graph.Node]NodePath)
	return
}

// Validate the document node with the schema node
func (ingester *Ingester) Validate(documentNode, schemaNode graph.Node) error {
	if schemaNode != nil {
		if err := ValidateDocumentNodeBySchema(documentNode, schemaNode); err != nil {
			return err
		}
	}
	return nil
}

// Polymorphic tests all options in the schema by calling ingest func
func (ingester *Ingester) Polymorphic(g graph.Graph, path NodePath, schemaNode graph.Node, ingest func(targetGraph graph.Graph, p NodePath, optionNode graph.Node) (graph.Node, error)) (graph.Graph, graph.Node, error) {
	// Polymorphic node. Try each option
	var newChild graph.Node
	// iterate through all edges of the schema node which have a polymorphic attribute
	for edges := schemaNode.GetEdgesWithLabel(graph.OutgoingEdge, OneOfTerm); edges.Next(); {
		edge := edges.Edge()
		optionNode := edge.GetTo()
		newGraph := graph.NewOCGraph()
		childNode, err := ingest(newGraph, path, optionNode)
		if err == nil {
			if newChild != nil {
				return nil, nil, ErrSchemaValidation{Msg: "Multiple options of the polymorphic node matched:" + GetNodeID(schemaNode), Path: path}
			}
			newChild = childNode
		}
	}
	if newChild == nil {
		return nil, nil, ErrSchemaValidation{Msg: "None of the options of the polymorphic node matched:" + GetNodeID(schemaNode), Path: path}
	}
	return newChild.GetGraph(), newChild, nil
}

// GetObjectAttributeNodes returns the schema attribute nodes under a
// schema object. The returned map is keyed by the AttributeNameTerm
func (ingester *Ingester) GetObjectAttributeNodes(objectSchemaNode graph.Node) (map[string][]graph.Node, error) {
	nextNodes := make(map[string][]graph.Node)
	addNextNode := func(node graph.Node) error {
		key := AsPropertyValue(node.GetProperty(AttributeNameTerm)).AsString()
		if len(key) == 0 {
			return ErrInvalidSchema(fmt.Sprintf("No '%s' in schema at %s", AttributeNameTerm, GetNodeID(objectSchemaNode)))
		}
		nextNodes[key] = append(nextNodes[key], node)
		return nil
	}
	if objectSchemaNode != nil {
		for _, node := range graph.TargetNodes(objectSchemaNode.GetEdgesWithLabel(graph.OutgoingEdge, ObjectAttributesTerm)) {
			if err := addNextNode(node); err != nil {
				return nil, err
			}
		}
		for _, node := range graph.TargetNodes(objectSchemaNode.GetEdgesWithLabel(graph.OutgoingEdge, ObjectAttributeListTerm)) {
			if err := addNextNode(node); err != nil {
				return nil, err
			}
		}
	}
	return nextNodes, nil
}

func (ingester *Ingester) ConnectChildNodes(parent graph.Node, children []graph.Node) {
	for index := range children {
		SetNodeIndex(children[index], index)
		ingester.connect(parent, children[index], HasTerm)
	}
}

// Object creates a new object node
func (ingester *Ingester) Object(g graph.Graph, path NodePath, schemaNode graph.Node, elements []graph.Node, types ...string) (graph.Node, error) {
	// An object node
	// There is a schema node for this node. It must be an object
	if schemaNode != nil {
		if !schemaNode.GetLabels().Has(AttributeTypeObject) {
			return nil, ErrSchemaValidation{Msg: "An object is not expected here", Path: path}
		}
	}
	ret := ingester.NewNode(g, path, schemaNode)
	if ingester.PreserveNodePaths {
		ingester.NodePaths[ret] = path.Copy()
	}
	t := ret.GetLabels()
	t.Add(types...)
	// define that ret is an object
	t.Add(AttributeTypeObject)
	ret.SetLabels(t)
	ingester.ConnectChildNodes(ret, elements)
	return ret, nil
}

// GetArrayElementNode returns the array element node from an array node
func (ingester *Ingester) GetArrayElementNode(arraySchemaNode graph.Node) graph.Node {
	if arraySchemaNode == nil {
		return nil
	}
	n := graph.TargetNodes(arraySchemaNode.GetEdgesWithLabel(graph.OutgoingEdge, ArrayItemsTerm))
	if len(n) == 1 {
		return n[0]
	}
	return nil
}

// Array creates a new array node.
func (ingester *Ingester) Array(g graph.Graph, path NodePath, schemaNode graph.Node, elements []graph.Node, types ...string) (graph.Node, error) {
	if schemaNode != nil {
		if !schemaNode.GetLabels().Has(AttributeTypeArray) {
			return nil, ErrSchemaValidation{Msg: "An array is not expected here", Path: path}
		}
	}
	ret := ingester.NewNode(g, path, schemaNode)
	if ingester.PreserveNodePaths {
		ingester.NodePaths[ret] = path.Copy()
	}
	t := ret.GetLabels()
	t.Add(types...)
	t.Add(AttributeTypeArray)
	ret.SetLabels(t)
	ingester.ConnectChildNodes(ret, elements)
	return ret, nil
}

// Value creates a new value node. The new node has the given value
// and the types
func (ingester *Ingester) Value(g graph.Graph, path NodePath, schemaNode graph.Node, value interface{}, types ...string) (graph.Node, error) {
	if schemaNode != nil {
		if !schemaNode.GetLabels().Has(AttributeTypeValue) {
			return nil, ErrSchemaValidation{Msg: "A value is not expected here", Path: path}
		}
	}
	newNode := ingester.NewNode(g, path, schemaNode)
	if ingester.PreserveNodePaths {
		ingester.NodePaths[newNode] = path.Copy()
	}
	if value != nil {
		SetRawNodeValue(newNode, value)
	}
	t := newNode.GetLabels()
	t.Add(types...)
	t.Add(AttributeTypeValue)
	newNode.SetLabels(t)
	return newNode, nil
}

// NewNode creates a new graph node, either by using the NewNodeFunc
// or by creating a new node using DefaultNodeIDGenerator. Then it
// either merges schema properties into the new node, or creates an
// instanceOf edge to the schema node.
func (ingester *Ingester) NewNode(g graph.Graph, path NodePath, schemaNode graph.Node) graph.Node {
	node := g.NewNode([]string{DocumentNodeTerm}, nil)
	SetNodeID(node, DefaultNodeIDGenerator(path, schemaNode))
	if schemaNode != nil {
		types := node.GetLabels()
		types.Add(FilterNonLayerTypes(schemaNode.GetLabels().Slice())...)
		node.SetLabels(types)
		node.SetProperty(SchemaNodeIDTerm, StringPropertyValue(GetNodeID(schemaNode)))
		if ingester.EmbedSchemaNodes {
			ingester.EmbedSchemaNode(node, schemaNode)
		} else {
			ingester.connect(node, schemaNode, InstanceOfTerm)
		}
	}
	return node
}

// EmbedSchemaNode merges the schema node properties with the target
// node properties. No properties are overwritten in the target
// node. The schema node types that are not schema node types are also
// merged with the target node types.
func (ingester *Ingester) EmbedSchemaNode(targetNode, schemaNode graph.Node) {
	schemaNode.ForEachProperty(func(k string, v interface{}) bool {
		if _, exists := targetNode.GetProperty(k); !exists {
			if pv, ok := v.(*PropertyValue); ok {
				targetNode.SetProperty(k, pv.Clone())
			} else {
				targetNode.SetProperty(k, v)
			}
		}
		return true
	})
}

// GetAsPropertyValue returns if the node should be a property of a
// predecessor node. If not, returns nil
func GetAsProperty(schemaNode graph.Node) (of string, name string) {
	if schemaNode == nil {
		return
	}
	of = AsPropertyValue(schemaNode.GetProperty(AsPropertyOfTerm)).AsString()
	name = AsPropertyValue(schemaNode.GetProperty(AsPropertyTerm)).AsString()
	return
}

func (ingester *Ingester) connect(srcNode, targetNode graph.Node, edgeLabel string) graph.Edge {
	return srcNode.GetGraph().NewEdge(srcNode, targetNode, edgeLabel, nil)
}

func (ingester *Ingester) DefaultEntityNodeIDGenerationFunc(entity string, ID string, node graph.Node, path []graph.Node) string {
	nodePath := ingester.NodePaths[node]
	eid := fmt.Sprintf("%s/%s", entity, ID)
	if len(nodePath) > 1 {
		eid += "/" + NodePath(nodePath[1:]).String()
	}
	return eid
}

// Finish ingesting by assigning node IDs and linking nodes to their
// entity root nodes. If generateIDFunc is nil, the default ID
// generation function is used
func (ingester *Ingester) Finish(root graph.Node, generateIDFunc func(entity string, ID string, node graph.Node, path []graph.Node) string) {
	if ingester.Schema != nil {
		if generateIDFunc == nil {
			generateIDFunc = ingester.DefaultEntityNodeIDGenerationFunc
		}
		AssignEntityIDs(root, generateIDFunc)
	}
}

// // AssignEntityRoots will iterate the document and assign entity roots to each node
// func AssignEntityRoots(root graph.Node) {
// 	entityBoundaries := make(map[graph.Node]struct{})
// 	// If there are no schemas, link everything to root
// 	entityBoundaries[root] = struct{}{}
// 	IterateDescendants(root, func(node graph.Node, path []graph.Node) bool {
// 		// Pass root
// 		if node == root {
// 			return true
// 		}

// 		// Is this an entity boundary
// 		_, boundary := GetNodeOrSchemaProperty(node, EntitySchemaTerm)
// 		if boundary {
// 			// Put node into boundaries
// 			entityBoundaries[node] = struct{}{}
// 		}
// 		// Skip last entry in path, thats the node
// 		for ix := len(path) - 2; ix >= 0; ix-- {
// 			if _, boundary := entityBoundaries[path[ix]]; boundary {
// 				node.SetProperty(entityRootTerm, path[ix])
// 				break
// 			}
// 		}
// 		return true
// 	}, OnlyDocumentNodes, false)
// }

// AssignEntityIDs traverses all the nodes under root, and reassigns
// IDs to the nodes based on the discovered entity boundaries and
// entity IDs. If there is no schema information, or if there are no
// entity IDs, the IDs are unchanged.
func AssignEntityIDs(root graph.Node, generateIDFunc func(entity string, ID string, node graph.Node, path []graph.Node) string) {
	// entityMap: map of nodes to their schemas. These nodes are the entity roots
	entityNodeMap := make(map[graph.Node]string)
	// entityIDMap: ID of the entity root
	entityIDMap := make(map[graph.Node]string)
	IterateDescendants(root, func(node graph.Node, path []graph.Node) bool {
		_, root := GetNodeOrSchemaProperty(node, EntitySchemaTerm)
		if root {
			types := graph.NewStringSet()
			types.Add(FilterNonLayerTypes(node.GetLabels().Slice())...)
			types.Remove(DocumentNodeTerm)
			typesSlice := types.Slice()
			if len(typesSlice) == 1 {
				entityNodeMap[node] = typesSlice[0]
			}
		}
		_, hasId := GetNodeOrSchemaProperty(node, EntityIDTerm)
		var closest graph.Node
		if hasId {
			// Find the closest entity root. Must be in entityNodeMap
			for ix := len(path) - 1; ix >= 0; ix-- {
				if _, root := entityNodeMap[path[ix]]; root {
					closest = path[ix]
					break
				}
			}
			if closest != nil {
				entityIDMap[closest] = fmt.Sprint(GetRawNodeValue(node))
			}
		}
		return true
	}, OnlyDocumentNodes, false)
	if len(entityIDMap) == 0 {
		return
	}
	// Iterate the nodes again and assign IDs
	IterateDescendants(root, func(node graph.Node, path []graph.Node) bool {
		for ix := len(path) - 1; ix >= 0; ix-- {
			id, exists := entityIDMap[path[ix]]
			if exists {
				// Generate ID
				// There must be an entity
				entity := entityNodeMap[path[ix]]
				SetNodeID(node, generateIDFunc(entity, id, node, path))
				break
			}
		}
		return true
	}, OnlyDocumentNodes, false)
}
