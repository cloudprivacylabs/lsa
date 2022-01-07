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

	"github.com/bserdar/digraph"
)

// Ingester keeps the schema and the ingestion options
type Ingester struct {
	// The schema variant to use during ingestion
	Schema *Layer

	// NewNodeFunc will create a new node for the data graph. If
	// NewNodeFunc is nil, a ls.Node will be created using the default
	// ID generator. The function should not add the node to the graph.
	NewNodeFunc func(path []interface{}, schemaNode Node) Node

	// NewEdgeFunc will create a new edge for the data graph with the
	// given label. If NewEdgeFunc is nil, a ls.Edge will be
	// created. The function should not add the edge to the graph.
	NewEdgeFunc func(string) Edge

	// If true, schame node properties are embedded into document
	// nodes. If false, schema nodes are preserved as separate nodes,
	// with an instanceOf edge between the document node to the schema
	// node.
	EmbedSchemaNodes bool
}

type ingestedID struct {
	path []Node
}

// IngestionContext keeps contextual information during ingestion.
type IngestionContext struct {
	// BaseID contains the prefix for ID generation
	BaseID string

	// Path contains the string representation of the current full path up to current node
	path []interface{}
}

// PushToPath adds the given path component to the path
func (ictx *IngestionContext) AddToPath(v interface{}) {
	ictx.path = append(ictx.path, v)
}

// PopPath removes the last path component
func (ictx *IngestionContext) PopPath() {
	ictx.path = ictx.path[:len(ictx.path)-1]
}

// GetPath returns a string representation of the path
func (ictx *IngestionContext) GetPath() string {
	path := make([]string, 0, len(ictx.path))
	for _, x := range ictx.path {
		path = append(path, fmt.Sprint(x))
	}
	return strings.Join(path, ".")
}

type ErrSchemaValidation struct {
	Msg  string
	Path string
}

func (e ErrSchemaValidation) Error() string {
	ret := "Schema validation error: " + e.Msg
	if len(e.Path) > 0 {
		ret += " path:" + e.Path
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

// Start ingestion. Returns the path initialized with the baseId, and
// the schema root.
func (ingester *Ingester) Start(baseID string) (ictx *IngestionContext, schemaRoot Node) {
	ictx = &IngestionContext{
		BaseID: baseID,
		path:   make([]interface{}, 0, 32),
	}
	if len(baseID) > 0 {
		ictx.AddToPath(baseID)
	}
	if ingester.Schema != nil {
		schemaRoot = ingester.Schema.GetSchemaRootNode()
	}
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
func (ingester *Ingester) Polymorphic(ictx *IngestionContext, schemaNode Node, ingest func(ictx IngestionContext, optionNode Node) (Node, error)) (Node, error) {
	// Polymorphic node. Try each option
	var newChild Node
	for nodes := schemaNode.OutWith(LayerTerms.OneOf).Targets(); nodes.HasNext(); {
		optionNode := nodes.Next().(Node)
		childNode, err := ingest(ictx, optionNode)
		if err == nil {
			if newChild != nil {
				return nil, ErrSchemaValidation{Msg: "Multiple options of the polymorphic node matched:" + schemaNode.GetID(), Path: ictx.GetPath()}
			}
			newChild = childNode
		}
	}
	if newChild == nil {
		return nil, ErrSchemaValidation{Msg: "None of the options of the polymorphic node matched:" + schemaNode.GetID(), Path: ictx.GetPath()}
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
func (ingester *Ingester) Object(ictx *IngestionContext, schemaNode Node, elements []Node, types ...string) (Node, error) {
	// An object node
	// There is a schema node for this node. It must be an object
	if schemaNode != nil {
		if !schemaNode.GetTypes().Has(AttributeTypes.Object) {
			return nil, ErrSchemaValidation{Msg: "An object is not expected here", Path: ictx.GetPath()}
		}

	}
	ret := ingester.NewNode(ictx, schemaNode)
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
func (ingester *Ingester) Array(ictx IngestionContext, schemaNode Node, elements []Node, types ...string) (Node, error) {
	if schemaNode != nil {
		if !schemaNode.GetTypes().Has(AttributeTypes.Array) {
			return nil, ErrSchemaValidation{Msg: "An array is not expected here", Path: ictx.GetPath()}
		}
	}
	ret := ingester.NewNode(ictx, schemaNode)
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
func (ingester *Ingester) Value(ictx *IngestionContext, schemaNode Node, value interface{}, types ...string) (Node, error) {
	if schemaNode != nil {
		if !schemaNode.GetTypes().Has(AttributeTypes.Value) {
			return nil, ErrSchemaValidation{Msg: "A value is not expected here", Path: ictx.path}
		}
		// Is this an entity ID node?
		if _, exists := schemaNode.GetProperties()[EntityIDTerm]; exists {
			ictx.SetCurrentEntityID(value)
		}
	}
	newNode := ingester.NewNode(ictx, schemaNode)
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
func (ingester *Ingester) NewNode(ictx IngestionContext, schemaNode Node) Node {
	var node Node
	if ingester.NewNodeFunc != nil {
		node = ingester.NewNodeFunc(path, schemaNode)
	} else {
		node = NewNode(DefaultNodeIDGenerator(path, schemaNode))
	}
	node.GetTypes().Add(DocumentNodeTerm)
	if ingester.EmbedSchemaNodes && schemaNode != nil {
		ingester.EmbedSchemaNode(node, schemaNode)
	} else if schemaNode != nil {
		ingester.connect(node, schemaNode, InstanceOfTerm)
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
	targetNode.GetTypes().Add(FilterNonLayerTypes(schemaNode.GetTypes().Slice())...)
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

// DefaultNodeIDGenerator returns Ingester.Schema.ID + join(path,".")
func DefaultNodeIDGenerator(path []interface{}, schemaNode Node) string {
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
