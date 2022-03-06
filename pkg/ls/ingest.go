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

	// If OnlySchemaAttributes is true, only ingest data points if there is a schema for it.
	// If OnlySchemaAttributes is false, ingest whether or not there is a schema for it.
	OnlySchemaAttributes bool

	// IngestEmptyValues is true if the value to ingest contains data, otherwise default to false
	IngestEmptyValues bool

	ExternalLookup func(lookupTableID string, dataNode graph.Node) (LookupResult, error)

	// SchemaNodeMap is used to keep a mapping of schema nodes copied into the
	// target graph. The key is a schema node. The value is the node in
	// target graph.
	SchemaNodeMap map[graph.Node]graph.Node

	// The target graph
	Graph graph.Graph
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

type ErrCannotDetermineEdgeLabel struct {
	Msg  string
	Path NodePath
}

type ErrCannotDeterminePropertyName struct {
	Path NodePath
}

func (e ErrCannotDeterminePropertyName) Error() string {
	return "Cannot determine property name: " + e.Path.String()
}

type ErrCannotFindAncestor struct {
	Path NodePath
}

func (e ErrCannotFindAncestor) Error() string {
	return "Cannot find ancestor: " + e.Path.String()
}

func (e ErrSchemaValidation) Error() string {
	ret := "Schema validation error: " + e.Msg
	if e.Path != nil {
		ret += " path:" + e.Path.String()
	}
	return ret
}

func (e ErrCannotDetermineEdgeLabel) Error() string {
	ret := "Cannot determine edge label: " + e.Msg
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

// NewDocumentGraph creates a new graph with the correct indexes for document ingestion
func NewDocumentGraph() graph.Graph {
	g := graph.NewOCGraph()
	g.AddNodePropertyIndex(EntitySchemaTerm)
	g.AddNodePropertyIndex(SchemaNodeIDTerm)
	return g
}

// DefaultNodeIDGenerator returns Ingester.Schema.ID + join(path,".")
func DefaultNodeIDGenerator(path NodePath, schemaNode graph.Node) string {
	return path.String()
}

// Start ingestion. Returns the path initialized with the baseId, and
// the schema root.
func (ingester *Ingester) Start(context *Context, baseID string) (path NodePath, schemaRoot graph.Node) {
	path = make(NodePath, 0, 16)
	if len(baseID) > 0 {
		path = append(path, baseID)
	}
	if ingester.Schema != nil {
		schemaRoot = ingester.Schema.GetSchemaRootNode()
	}
	if ingester.SchemaNodeMap == nil {
		ingester.SchemaNodeMap = make(map[graph.Node]graph.Node)
	}
	return
}

func determineEdgeLabel(schemaNode graph.Node) string {
	if x, ok := schemaNode.GetProperty(EdgeLabelTerm); ok {
		if label := x.(*PropertyValue).AsString(); len(label) > 0 {
			return label
		}
	}
	if x, ok := schemaNode.GetProperty(AttributeNameTerm); ok {
		if label := x.(*PropertyValue).AsString(); len(label) > 0 {
			return label
		}
	}
	return ""
}

//  ValueAsEdge ingests a value using the following scheme:
//
//  input: (name: value)
//  output: --(label)-->(value:value, attributeName:name)
//
// where label=attributeName (in this case "name") if edgeLabel is not
// specified in schema.
func (ingester *Ingester) ValueAsEdge(context *Context, path NodePath, ingestionPath []graph.Node, schemaNode graph.Node, value string, types ...string) (graph.Edge, error) {
	if schemaNode == nil {
		return nil, ErrInvalidInput{Msg: "missing schemaNode"}
	}
	if !ingester.IngestEmptyValues && len(value) == 0 {
		return nil, nil
	}
	if !schemaNode.GetLabels().Has(AttributeTypeValue) {
		return nil, ErrSchemaValidation{Msg: "A value attribute is expected here", Path: path}
	}
	if len(ingestionPath) == 0 {
		return nil, ErrDataIngestion{Key: path.String(), Err: fmt.Errorf("Document root value cannot be an edge")}
	}
	edgeLabel := determineEdgeLabel(schemaNode)
	if len(edgeLabel) == 0 {
		return nil, ErrCannotDetermineEdgeLabel{Path: path.Copy()}
	}
	node := ingester.NewNode(context, path, schemaNode)
	SetRawNodeValue(node, value)
	t := node.GetLabels()
	t.Add(types...)
	t.Add(AttributeTypeValue)
	node.SetLabels(t)
	edge := ingester.Graph.NewEdge(ingestionPath[len(ingestionPath)-1], node, edgeLabel, nil)
	return edge, nil
}

// ValueAsNode creates a new value node. The new node has the given value
// and the types
func (ingester *Ingester) ValueAsNode(context *Context, path NodePath, ingestionPath []graph.Node, schemaNode graph.Node, value string, types ...string) (graph.Edge, graph.Node, error) {
	if schemaNode != nil {
		if !schemaNode.GetLabels().Has(AttributeTypeValue) {
			return nil, nil, ErrSchemaValidation{Msg: "A value is not expected here", Path: path}
		}
	}
	if !ingester.IngestEmptyValues && len(value) == 0 {
		return nil, nil, nil
	}
	newNode := ingester.NewNode(context, path, schemaNode)
	SetRawNodeValue(newNode, value)
	t := newNode.GetLabels()
	t.Add(types...)
	t.Add(AttributeTypeValue)
	newNode.SetLabels(t)
	var edge graph.Edge
	if len(ingestionPath) > 0 {
		edge = ingester.Graph.NewEdge(ingestionPath[len(ingestionPath)-1], newNode, HasTerm, nil)
	}
	return edge, newNode, nil
}

// Value ingests a value as a node, edge-node, or as a property depending on the schema. The default is ingestion as node. Returns the node, and optionally, the edge going to that node
func (ingester *Ingester) Value(context *Context, path NodePath, ingestionPath []graph.Node, schemaNode graph.Node, value string, types ...string) (graph.Edge, graph.Node, error) {
	switch GetIngestAs(schemaNode) {
	case "node":
		return ingester.ValueAsNode(context, path, ingestionPath, schemaNode, value, types...)

	case "edge":
		e, err := ingester.ValueAsEdge(context, path, ingestionPath, schemaNode, value, types...)
		if err != nil {
			return nil, nil, err
		}
		return e, e.GetTo(), nil

	case "property":
		// Schema node cannot be nil here
		asPropertyOf := AsPropertyValue(schemaNode.GetProperty(AsPropertyOfTerm)).AsString()
		propertyName := AsPropertyValue(schemaNode.GetProperty(PropertyNameTerm)).AsString()
		if len(propertyName) == 0 {
			propertyName = AsPropertyValue(schemaNode.GetProperty(AttributeNameTerm)).AsString()
		}
		if len(propertyName) == 0 {
			return nil, nil, ErrCannotDeterminePropertyName{Path: path.Copy()}
		}
		var targetNode graph.Node
		if len(asPropertyOf) == 0 {
			if len(ingestionPath) > 0 {
				targetNode = ingestionPath[len(ingestionPath)-1]
			}
		} else {
			// Find ancestor that is instance of asPropertyOf
			for i := len(ingestionPath) - 1; i >= 0; i-- {
				if AsPropertyValue(ingestionPath[i].GetProperty(SchemaNodeIDTerm)).AsString() == asPropertyOf {
					targetNode = ingestionPath[i]
					break
				}
			}
		}
		if targetNode == nil {
			return nil, nil, ErrCannotFindAncestor{Path: path.Copy()}
		}
		targetNode.SetProperty(propertyName, StringPropertyValue(value))
	}
	return nil, nil, nil
}

func (ingester *Ingester) collectionAsNode(context *Context, typeTerm string, path NodePath, ingestionPath []graph.Node, schemaNode graph.Node, types ...string) (graph.Edge, graph.Node, error) {
	ret := ingester.NewNode(context, path, schemaNode)
	t := ret.GetLabels()
	t.Add(types...)
	// define that ret is an object
	t.Add(typeTerm)
	ret.SetLabels(t)
	var edge graph.Edge
	if len(ingestionPath) > 0 {
		edge = ingester.Graph.NewEdge(ingestionPath[len(ingestionPath)-1], ret, HasTerm, nil)
	}
	return edge, ret, nil
}

// ObjectAsNode creates a new object node
func (ingester *Ingester) ObjectAsNode(context *Context, path NodePath, ingestionPath []graph.Node, schemaNode graph.Node, types ...string) (graph.Edge, graph.Node, error) {
	// An object node
	// There is a schema node for this node. It must be an object
	if schemaNode != nil {
		if !schemaNode.GetLabels().Has(AttributeTypeObject) {
			return nil, nil, ErrSchemaValidation{Msg: fmt.Sprintf("An object is expected here but found %s", schemaNode.GetLabels()), Path: path}
		}
	}
	return ingester.collectionAsNode(context, AttributeTypeObject, path, ingestionPath, schemaNode, types...)
}

func (ingester *Ingester) collectionAsEdge(context *Context, typeTerm string, path NodePath, ingestionPath []graph.Node, schemaNode graph.Node, types ...string) (graph.Edge, error) {
	if len(ingestionPath) == 0 {
		return nil, ErrDataIngestion{Key: path.String(), Err: fmt.Errorf("Document root object cannot be an edge")}
	}
	blankNode := ingester.NewNode(context, path, schemaNode)
	edgeLabel := determineEdgeLabel(schemaNode)
	if len(edgeLabel) == 0 {
		return nil, ErrCannotDetermineEdgeLabel{Path: path.Copy()}
	}
	t := blankNode.GetLabels()
	t.Add(types...)
	// define that newEdgeNode.Node is an object
	t.Add(typeTerm)
	blankNode.SetLabels(t)
	edge := ingester.Graph.NewEdge(ingestionPath[len(ingestionPath)-1], blankNode, edgeLabel, nil)
	return edge, nil
}

// ObjectAsEdge creates an object node as an edge using the following scheme:
//
//  parent --object--> _blankNode --...
func (ingester *Ingester) ObjectAsEdge(context *Context, path NodePath, ingestionPath []graph.Node, schemaNode graph.Node, types ...string) (graph.Edge, error) {
	if schemaNode != nil {
		if !schemaNode.GetLabels().Has(AttributeTypeObject) {
			return nil, ErrSchemaValidation{Msg: fmt.Sprintf("An object is expected here but found %s", schemaNode.GetLabels()), Path: path}
		}
	}
	return ingester.collectionAsEdge(context, AttributeTypeObject, path, ingestionPath, schemaNode, types...)
}

// Validate the document node with the schema node
func (ingester *Ingester) Validate(context *Context, documentNode, schemaNode graph.Node) error {
	if schemaNode != nil {
		if err := ValidateDocumentNodeBySchema(documentNode, schemaNode); err != nil {
			return err
		}
	}
	return nil
}

// NewNode creates a new graph node, either by using the NewNodeFunc
// or by creating a new node using DefaultNodeIDGenerator. Then it
// either merges schema properties into the new node, or creates an
// instanceOf edge to the schema node.
func (ingester *Ingester) NewNode(context *Context, path NodePath, schemaNode graph.Node) graph.Node {
	node := ingester.Graph.NewNode([]string{DocumentNodeTerm}, nil)
	SetNodeID(node, DefaultNodeIDGenerator(path, schemaNode))
	if schemaNode == nil {
		return node
	}
	types := node.GetLabels()
	types.Add(FilterNonLayerTypes(schemaNode.GetLabels().Slice())...)
	node.SetLabels(types)
	node.SetProperty(SchemaNodeIDTerm, StringPropertyValue(GetNodeID(schemaNode)))

	pat := graph.Pattern{{
		Labels:     graph.NewStringSet(AttributeNodeTerm),
		Properties: map[string]interface{}{NodeIDTerm: GetNodeID(schemaNode)},
	}}
	acc := graph.DefaultMatchAccumulator{}
	pat.Run(ingester.Graph, nil, &acc)
	nodes := acc.GetHeadNodes()

	if ingester.EmbedSchemaNodes {
		ingester.EmbedSchemaNode(context, node, schemaNode)
		// Copy the subtrees for all nodes connected to the schema node
		for edges := schemaNode.GetEdges(graph.OutgoingEdge); edges.Next(); {
			edge := edges.Edge()
			if IsAttributeTreeEdge(edge) {
				continue
			}
			graph.CopySubgraph(edge.GetTo(), ingester.Graph, ClonePropertyValueFunc, ingester.SchemaNodeMap)
			ingester.Graph.NewEdge(node, ingester.SchemaNodeMap[edge.GetTo()], edge.GetLabel(), nil)
		}
	} else {
		// Copy the schema node into this
		// If the schema node already exists in the target graph, use it
		if len(nodes) != 0 {
			ingester.Graph.NewEdge(node, nodes[0], InstanceOfTerm, nil)
		} else {
			// Copy the subtree
			graph.CopySubgraph(schemaNode, ingester.Graph, ClonePropertyValueFunc, ingester.SchemaNodeMap)
			ingester.Graph.NewEdge(node, ingester.SchemaNodeMap[schemaNode], InstanceOfTerm, nil)
		}
	}
	// If this is an entity boundary, mark it
	pv, rootNode := schemaNode.GetProperty(EntitySchemaTerm)
	if rootNode {
		node.SetProperty(EntitySchemaTerm, pv)
	}
	return node
}

// EmbedSchemaNode merges the schema node properties with the target
// node properties. No properties are overwritten in the target
// node. The schema node types that are not schema node types are also
// merged with the target node types.
func (ingester *Ingester) EmbedSchemaNode(context *Context, targetNode, schemaNode graph.Node) {
	schemaNode.ForEachProperty(func(k string, v interface{}) bool {
		if k == NodeIDTerm {
			return true
		}
		if pv, ok := v.(*PropertyValue); ok {
			targetNode.SetProperty(k, pv.Clone())
		} else {
			targetNode.SetProperty(k, v)
		}
		return true
	})
}

// GetIngestAs returns "node", "edge", or "property" based on IngestAsTerm
func GetIngestAs(schemaNode graph.Node) string {
	if schemaNode == nil {
		return ""
	}
	p, ok := schemaNode.GetProperty(IngestAsTerm)
	if !ok {
		return "node"
	}
	s := AsPropertyValue(p, ok).AsString()
	if s == "edge" || s == "property" {
		return s
	}
	return "node"
}

// Polymorphic tests all options in the schema by calling ingest func
func (ingester *Ingester) Polymorphic(context *Context, g graph.Graph, path NodePath, schemaNode graph.Node, ingest func(targetGraph graph.Graph, p NodePath, optionNode graph.Node) (graph.Node, error)) (graph.Graph, graph.Node, error) {
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

// Finish ingesting by assigning node IDs and linking nodes to their
// entity root nodes. If generateIDFunc is nil, the default ID
// generation function is used
func (ingester *Ingester) Finish(context *Context, root graph.Node) error {
	if ingester.Schema != nil {
		if generateIDFunc == nil {
			generateIDFunc = ingester.DefaultEntityNodeIDGenerationFunc
		}
		AssignEntityIDs(context, root, generateIDFunc)
	}
	lpc := LookupProcessor{
		Graph:          root.GetGraph(),
		ExternalLookup: ingester.ExternalLookup,
	}
	for nodes := lpc.Graph.GetNodes(); nodes.Next(); {
		if err := lpc.ProcessLookup(nodes.Node()); err != nil {
			return err
		}
	}
	return nil
}

func assignEntityIDs(context *Context, root graph.Node) {
	// Find all entity root nodes, and assign ids
	var f func(graph.Node)
	seen := map[graph.Node]struct{}{}
	type stackItem struct {
		entityRoot graph.Node
		idAttrs    []string
	}
	stack := make([]stackItem, 0, 16)
	f = func(node graph.Node) {
		if _, exists := seen[node]; exists {
			return
		}
		if !node.GetLabels().Has(DocumentNodeTerm) {
			return
		}
		seen[node] = struct{}{}

		pv, root := node.GetLabels().Has(EntitySchemaTerm)
		if root {
			// Store the ID attributes and root node in stack
			pv.MustStringSlice()
			stack = append(stack, node)
		}
		if len(stack) > 0 && node.GetLabels().Has(AttributeTypeValue) {
			schId := AsPropertyValue(node.GetProperty(SchemaNodeIDTerm)).AsString()
			if len(schId) > 0 {

			}
		}

		if root {
			stack = stack[:len(stack)-1]
		}
	}
}

// AssignEntityIDs assigns the entityId property to all the entity
// root nodes
func AssignEntityIDs(context *Context, root graph.Node) {
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
