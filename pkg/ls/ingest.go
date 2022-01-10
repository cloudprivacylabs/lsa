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

	"github.com/bserdar/digraph"
)

// Ingester keeps the schema and the ingestion options
type Ingester struct {
	// The schema variant to use during ingestion
	Schema *Layer

	// NewNodeFunc will create a new node for the data graph. If
	// NewNodeFunc is nil, a ls.Node will be created using the default
	// ID generator. The function should not add the node to the graph.
	NewNodeFunc func(path NodePath, schemaNode Node) Node

	// NewEdgeFunc will create a new edge for the data graph with the
	// given label. If NewEdgeFunc is nil, a ls.Edge will be
	// created. The function should not add the edge to the graph.
	NewEdgeFunc func(string) Edge

	// If true, schame node properties are embedded into document
	// nodes. If false, schema nodes are preserved as separate nodes,
	// with an instanceOf edge between the document node to the schema
	// node.
	EmbedSchemaNodes bool

	// If true, a map[Node][]interface{} is populated to preserve the
	// paths used to create nodes
	PreserveNodePaths bool

	// If PreserveNodePaths is true, this keeps the node paths after ingestion.
	// This map is reset when Start is called.
	NodePaths map[Node]NodePath
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
func DefaultNodeIDGenerator(path NodePath, schemaNode Node) string {
	return path.String()
}

// Start ingestion. Returns the path initialized with the baseId, and
// the schema root.
func (ingester *Ingester) Start(baseID string) (path NodePath, schemaRoot Node) {
	path = make(NodePath, 0, 16)
	path = append(path, baseID)
	if ingester.Schema != nil {
		schemaRoot = ingester.Schema.GetSchemaRootNode()
	}
	ingester.NodePaths = make(map[Node]NodePath)
	return
}

// Validate the document node with the schema node
func (ingester *Ingester) Validate(documentNode, schemaNode Node) error {
	if schemaNode != nil {
		if err := ValidateDocumentNodeBySchema(documentNode, schemaNode); err != nil {
			return err
		}
	}
	return nil
}

// Polymorphic tests all options in the schema by calling ingest func
func (ingester *Ingester) Polymorphic(path NodePath, schemaNode Node, ingest func(p NodePath, optionNode Node) (Node, error)) (Node, error) {
	// Polymorphic node. Try each option
	var newChild Node
	for nodes := schemaNode.OutWith(LayerTerms.OneOf).Targets(); nodes.HasNext(); {
		optionNode := nodes.Next().(Node)
		childNode, err := ingest(path, optionNode)
		if err == nil {
			if newChild != nil {
				return nil, ErrSchemaValidation{Msg: "Multiple options of the polymorphic node matched:" + schemaNode.GetID(), Path: path}
			}
			newChild = childNode
		}
	}
	if newChild == nil {
		return nil, ErrSchemaValidation{Msg: "None of the options of the polymorphic node matched:" + schemaNode.GetID(), Path: path}
	}
	return newChild, nil
}

// GetObjectAttributeNodes returns the schema attribute nodes under a schema object
func (ingester *Ingester) GetObjectAttributeNodes(objectSchemaNode Node) (map[string]Node, error) {
	nextNodes := make(map[string]Node)
	addNextNode := func(node Node) error {
		key := node.GetProperties()[AttributeNameTerm].AsString()
		if len(key) == 0 {
			return ErrInvalidSchema(fmt.Sprintf("No '%s' in schema at %s", AttributeNameTerm, objectSchemaNode.GetID()))
		}
		if _, ok := nextNodes[key]; ok {
			return ErrInvalidSchema(fmt.Sprintf("Multiple elements with key '%s'", key))
		}
		nextNodes[key] = node
		return nil
	}
	if objectSchemaNode != nil {
		for nodes := objectSchemaNode.OutWith(LayerTerms.Attributes).Targets(); nodes.HasNext(); {
			if err := addNextNode(nodes.Next().(Node)); err != nil {
				return nil, err
			}
		}
		for nodes := objectSchemaNode.OutWith(LayerTerms.AttributeList).Targets(); nodes.HasNext(); {
			if err := addNextNode(nodes.Next().(Node)); err != nil {
				return nil, err
			}
		}
	}
	return nextNodes, nil
}

// Object creates a new object node
func (ingester *Ingester) Object(path NodePath, schemaNode Node, elements []Node, types ...string) (Node, error) {
	// An object node
	// There is a schema node for this node. It must be an object
	if schemaNode != nil {
		if !schemaNode.GetTypes().Has(AttributeTypes.Object) {
			return nil, ErrSchemaValidation{Msg: "An object is not expected here", Path: path}
		}
	}
	ret := ingester.NewNode(path, schemaNode)
	if ingester.PreserveNodePaths {
		ingester.NodePaths[ret] = path.Copy()
	}
	ret.GetTypes().Add(types...)
	ret.GetTypes().Add(AttributeTypes.Object)
	for index := range elements {
		elements[index].GetProperties()[AttributeIndexTerm] = StringPropertyValue(fmt.Sprint(index))
		ingester.connect(ret, elements[index], HasTerm)
	}
	return ret, nil
}

// GetArrayElementNode returns the array element node from an array node
func (ingester *Ingester) GetArrayElementNode(arraySchemaNode Node) Node {
	if arraySchemaNode == nil {
		return nil
	}
	n := arraySchemaNode.NextWith(LayerTerms.ArrayItems)
	if len(n) == 1 {
		return n[0].(Node)
	}
	return nil
}

// Array creates a new array node.
func (ingester *Ingester) Array(path NodePath, schemaNode Node, elements []Node, types ...string) (Node, error) {
	if schemaNode != nil {
		if !schemaNode.GetTypes().Has(AttributeTypes.Array) {
			return nil, ErrSchemaValidation{Msg: "An array is not expected here", Path: path}
		}
	}
	ret := ingester.NewNode(path, schemaNode)
	if ingester.PreserveNodePaths {
		ingester.NodePaths[ret] = path.Copy()
	}
	ret.GetTypes().Add(types...)
	ret.GetTypes().Add(AttributeTypes.Array)
	for index := range elements {
		elements[index].GetProperties()[AttributeIndexTerm] = StringPropertyValue(fmt.Sprint(index))
		ingester.connect(ret, elements[index], HasTerm)
	}
	return ret, nil
}

// Value creates a new value node. The new node has the given value
// and the types
func (ingester *Ingester) Value(path NodePath, schemaNode Node, value interface{}, types ...string) (Node, error) {
	if schemaNode != nil {
		if !schemaNode.GetTypes().Has(AttributeTypes.Value) {
			return nil, ErrSchemaValidation{Msg: "A value is not expected here", Path: path}
		}
	}
	newNode := ingester.NewNode(path, schemaNode)
	if ingester.PreserveNodePaths {
		ingester.NodePaths[newNode] = path.Copy()
	}
	if value != nil {
		newNode.SetValue(value)
	}
	newNode.GetTypes().Add(types...)
	newNode.GetTypes().Add(AttributeTypes.Value)
	return newNode, nil
}

// NewNode creates a new graph node, either by using the NewNodeFunc
// or by creating a new node using DefaultNodeIDenerator. Then it
// either merges schema properties into the new node, or creates an
// instanceOf edge to the schema node.
func (ingester *Ingester) NewNode(path NodePath, schemaNode Node) Node {
	var node Node
	if ingester.NewNodeFunc != nil {
		node = ingester.NewNodeFunc(path, schemaNode)
	} else {
		node = NewNode(DefaultNodeIDGenerator(path, schemaNode))
	}
	node.GetTypes().Add(DocumentNodeTerm)
	if schemaNode != nil {
		node.GetTypes().Add(FilterNonLayerTypes(schemaNode.GetTypes().Slice())...)
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
func (ingester *Ingester) EmbedSchemaNode(targetNode, schemaNode Node) {
	targetProperties := targetNode.GetProperties()
	for k, v := range schemaNode.GetProperties() {
		if _, exists := targetProperties[k]; !exists {
			targetProperties[k] = v
		}
	}
}

// GetAsPropertyValue returns if the node should be a property of a
// predecessor node. If not, returns nil
func GetAsProperty(schemaNode Node) (of string, name string) {
	if schemaNode == nil {
		return
	}
	properties := schemaNode.GetProperties()
	of = properties[AsPropertyOfTerm].AsString()
	name = properties[AsPropertyTerm].AsString()
	return
}

func (ingester *Ingester) connect(srcNode, targetNode digraph.Node, edgeLabel string) digraph.Edge {
	var edge digraph.Edge
	if ingester.NewEdgeFunc != nil {
		edge = ingester.NewEdgeFunc(edgeLabel)
	} else {
		edge = NewEdge(edgeLabel)
	}
	digraph.Connect(srcNode, targetNode, edge)
	return edge
}

// AssignEntityIDs traverses all the nodes under root, and reassigns
// IDs to the nodes based on the discovered entity boundaries and
// entity IDs. If there is no schema information, or if there are no
// entity IDs, the IDs are unchanges.
func AssignEntityIDs(root Node, generateIDFunc func(entity string, ID string, node Node, path []Node) string) {
	onlyDocNodes := func(edge Edge, _ []Node) EdgeFuncResult {
		if !edge.GetTo().(Node).GetTypes().Has(DocumentNodeTerm) {
			return SkipEdgeResult
		}
		return FollowEdgeResult
	}
	// entityMap: map of nodes to their schemas. These nodes are the entity roots
	entityNodeMap := make(map[Node]string)
	// entityIDMap: ID of the entity root
	entityIDMap := make(map[Node]string)
	IterateDescendants(root, func(node Node, path []Node) bool {
		_, root := GetNodeOrSchemaProperty(node, EntitySchemaTerm)
		if root {
			types := Types{}
			types.Add(FilterNonLayerTypes(node.GetTypes().Slice())...)
			types.Remove(DocumentNodeTerm)
			typesSlice := types.Slice()
			if len(typesSlice) == 1 {
				entityNodeMap[node] = typesSlice[0]
			}
		}
		_, hasId := GetNodeOrSchemaProperty(node, EntityIDTerm)
		var closest Node
		if hasId {
			// Find the closest entity root. Must be in entityNodeMap
			for ix := len(path) - 1; ix >= 0; ix-- {
				if _, root := entityNodeMap[path[ix]]; root {
					closest = path[ix]
					break
				}
			}
			if closest != nil {
				entityIDMap[closest] = fmt.Sprint(node.GetValue())
			}
		}
		return true
	}, onlyDocNodes, false)
	if len(entityIDMap) == 0 {
		return
	}
	// Iterate the nodes again and assign IDs
	IterateDescendants(root, func(node Node, path []Node) bool {
		for ix := len(path) - 1; ix >= 0; ix-- {
			id, exists := entityIDMap[path[ix]]
			if exists {
				// Generate ID
				// There must be an entity
				entity := entityNodeMap[path[ix]]
				node.SetID(generateIDFunc(entity, id, node, path))
				break
			}
		}
		return true
	}, onlyDocNodes, false)
}
