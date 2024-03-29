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

	"github.com/cloudprivacylabs/lpg/v2"
)

type GraphBuilderOptions struct {
	// If true, schema node properties are embedded into document
	// nodes. If false, schema nodes are preserved as separate nodes,
	// with an instanceOf edge between the document node to the schema
	// node.
	EmbedSchemaNodes bool
	// If OnlySchemaAttributes is true, only ingest data points if there is a schema for it.
	// If OnlySchemaAttributes is false, ingest whether or not there is a schema for it.
	OnlySchemaAttributes bool
}

// GraphBuilder contains the methods to ingest a graph
type GraphBuilder struct {
	options *GraphBuilderOptions
	// SchemaNodeMap keeps the map of schema nodes copied into the target graph
	schemaNodeMap map[*lpg.Node]*lpg.Node
	targetGraph   *lpg.Graph
}

type ErrCannotInstantiateSchemaNode struct {
	SchemaNodeID string
	Reason       string
}

func (e ErrCannotInstantiateSchemaNode) Error() string {
	return fmt.Sprintf("Cannot instantiate schema node %s because: %s", e.SchemaNodeID, e.Reason)
}

// NewGraphBuilder returns a new builder with an optional graph. If g
// is nil, a new graph is initialized
func NewGraphBuilder(g *lpg.Graph, options GraphBuilderOptions) GraphBuilder {
	if g == nil {
		g = NewDocumentGraph()
	}
	ret := GraphBuilder{
		options:       &options,
		targetGraph:   g,
		schemaNodeMap: make(map[*lpg.Node]*lpg.Node),
	}
	return ret
}

func (gb GraphBuilder) GetOptions() GraphBuilderOptions {
	return *gb.options
}

func (gb GraphBuilder) GetGraph() *lpg.Graph {
	return gb.targetGraph
}

func determineEdgeLabel(schemaNode *lpg.Node) string {
	if str := EdgeLabelTerm.PropertyValue(schemaNode); len(str) > 0 {
		return str
	}
	if str := AttributeNameTerm.PropertyValue(schemaNode); len(str) > 0 {
		return str
	}
	return ""
}

// InstantiateSchemaNode creates a new node in targetGraph that is an
// instance of the given schemaNode. If embedSchemaNodes is true, the
// properties of the schema node will be embedded into the new schema
// node. If embedSchemaNodes is false, schema nodes will be kept
// separate. The schemaNodeMap will be filled with the map of schema
// nodes copied into the target graph. The key will be the original
// schema node, and the value will be the copied schema node in
// targetGraph. Returns the new node.
func InstantiateSchemaNode(targetGraph *lpg.Graph, schemaNode *lpg.Node, embedSchemaNodes bool, schemaNodeMap map[*lpg.Node]*lpg.Node) *lpg.Node {
	types := []string{DocumentNodeTerm.Name}
	for l := range schemaNode.GetLabels().M {
		if l != AttributeNodeTerm.Name {
			types = append(types, l)
		}
	}
	newNode := targetGraph.NewNode(types, nil)
	newNode.SetProperty(SchemaNodeIDTerm.Name, SchemaNodeIDTerm.MustPropertyValue(GetNodeID(schemaNode)))
	// If this is an entity boundary, mark it
	if pv, rootNode := schemaNode.GetProperty(EntitySchemaTerm.Name); rootNode {
		newNode.SetProperty(EntitySchemaTerm.Name, pv)
	}

	copyNodesAttachedToSchema := func(targetNode *lpg.Node) {
		for edges := schemaNode.GetEdges(lpg.OutgoingEdge); edges.Next(); {
			edge := edges.Edge()
			if IsAttributeTreeEdge(edge) {
				continue
			}
			lpg.CopySubgraph(edge.GetTo(), targetGraph, ClonePropertyValueFunc, schemaNodeMap)
			targetGraph.NewEdge(targetNode, schemaNodeMap[edge.GetTo()], edge.GetLabel(), nil)
		}
	}

	if embedSchemaNodes {
		schemaNode.ForEachProperty(func(k string, v interface{}) bool {
			if k == NodeIDTerm.Name {
				return true
			}
			if _, ok := newNode.GetProperty(k); !ok {
				newNode.SetProperty(k, v)
			}
			return true
		})
		copyNodesAttachedToSchema(newNode)
		return newNode
	}
	pat := lpg.Pattern{{
		Labels:     lpg.NewStringSet(AttributeNodeTerm.Name),
		Properties: map[string]interface{}{NodeIDTerm.Name: NodeIDTerm.MustPropertyValue(GetNodeID(schemaNode))},
	}}
	nodes, _ := pat.FindNodes(targetGraph, nil)
	// Copy the schema node into this
	// If the schema node already exists in the target graph, use it
	if len(nodes) != 0 {
		targetGraph.NewEdge(newNode, nodes[0], InstanceOfTerm.Name, nil)
	} else {
		// Copy the node
		newSchemaNode := lpg.CopyNode(schemaNode, targetGraph, ClonePropertyValueFunc)
		schemaNodeMap[schemaNode] = newSchemaNode
		targetGraph.NewEdge(newNode, newSchemaNode, InstanceOfTerm.Name, nil)
		copyNodesAttachedToSchema(newSchemaNode)
	}
	return newNode
}

// NewNode creates a new graph node as an instance of SchemaNode. Then
// it either merges schema properties into the new node, or creates an
// instanceOf edge to the schema node.
func (gb GraphBuilder) NewNode(schemaNode *lpg.Node) *lpg.Node {
	if schemaNode == nil {
		return gb.targetGraph.NewNode([]string{DocumentNodeTerm.Name}, nil)
	}

	return InstantiateSchemaNode(gb.targetGraph, schemaNode, gb.options.EmbedSchemaNodes, gb.schemaNodeMap)
}

func (gb GraphBuilder) setEntityID(value string, parentDocumentNode, schemaNode *lpg.Node) error {
	entityRootNode := GetEntityRootNode(parentDocumentNode)
	if entityRootNode == nil {
		return nil
	}

	schemaNodeID := GetNodeID(schemaNode)
	return SetEntityIDVectorElement(entityRootNode, schemaNodeID, value)
}

// ValueSetAsEdge can be called to notify the graph builder that a value is set that was ingested as edge
func (gb GraphBuilder) ValueSetAsEdge(node, schemaNode, parentDocumentNode *lpg.Node) {
	if schemaNode != nil {
		value, _ := GetRawNodeValue(node)
		gb.setEntityID(value, parentDocumentNode, schemaNode)
	}
}

//	ValueAsEdge ingests a value using the following scheme:
//
//	input: (name: value)
//	output: --(label)-->(value:value, attributeName:name)
//
// where label=attributeName (in this case "name") if edgeLabel is not
// specified in schema.
func (gb GraphBuilder) RawValueAsEdge(schemaNode, parentDocumentNode *lpg.Node, value string, types ...string) (*lpg.Edge, error) {
	return gb.ValueAsEdge(schemaNode, parentDocumentNode, func(node *lpg.Node) error {
		SetRawNodeValue(node, value)
		return nil
	}, types...)
}

func (gb GraphBuilder) NativeValueAsEdge(schemaNode, parentDocumentNode *lpg.Node, value interface{}, types ...string) (*lpg.Edge, error) {
	return gb.ValueAsEdge(schemaNode, parentDocumentNode, func(node *lpg.Node) error {
		return SetNodeValue(node, value)
	}, types...)
}

func (gb GraphBuilder) ValueAsEdge(schemaNode, parentDocumentNode *lpg.Node, setValue func(*lpg.Node) error, types ...string) (*lpg.Edge, error) {
	var edgeLabel string
	if schemaNode != nil {
		if !schemaNode.HasLabel(AttributeTypeValue.Name) {
			return nil, ErrSchemaValidation{Msg: "A value is expected here"}
		}
		edgeLabel = determineEdgeLabel(schemaNode)
		if len(edgeLabel) == 0 {
			return nil, ErrCannotDetermineEdgeLabel{SchemaNodeID: GetNodeID(schemaNode)}
		}
	} else if gb.options.OnlySchemaAttributes {
		return nil, nil
	}
	node := gb.NewNode(schemaNode)
	if err := setValue(node); err != nil {
		return nil, err
	}
	if schemaNode != nil {
		rawValue, _ := GetRawNodeValue(node)
		gb.setEntityID(rawValue, parentDocumentNode, schemaNode)
	}
	t := node.GetLabels()
	t.Add(types...)
	t.Add(AttributeTypeValue.Name)
	node.SetLabels(t)
	edge := gb.targetGraph.NewEdge(parentDocumentNode, node, edgeLabel, nil)
	return edge, nil
}

// ValueSetAsNode can be called to notif the graph builder that a node value is set
func (gb GraphBuilder) ValueSetAsNode(node, schemaNode, parentDocumentNode *lpg.Node) {
	if schemaNode != nil {
		value, _ := GetRawNodeValue(node)
		gb.setEntityID(value, parentDocumentNode, schemaNode)
	}
}

// ValueAsNode creates a new value node. The new node has the given value
// and the types
func (gb GraphBuilder) RawValueAsNode(schemaNode, parentDocumentNode *lpg.Node, value string, types ...string) (*lpg.Edge, *lpg.Node, error) {
	return gb.ValueAsNode(schemaNode, parentDocumentNode, func(node *lpg.Node) error {
		SetRawNodeValue(node, value)
		return nil
	}, types...)
}

func (gb GraphBuilder) NativeValueAsNode(schemaNode, parentDocumentNode *lpg.Node, value interface{}, types ...string) (*lpg.Edge, *lpg.Node, error) {
	return gb.ValueAsNode(schemaNode, parentDocumentNode, func(node *lpg.Node) error {
		return SetNodeValue(node, value)
	}, types...)
}

func (gb GraphBuilder) ValueAsNode(schemaNode, parentDocumentNode *lpg.Node, setValue func(*lpg.Node) error, types ...string) (*lpg.Edge, *lpg.Node, error) {
	if schemaNode != nil {
		if !schemaNode.HasLabel(AttributeTypeValue.Name) {
			return nil, nil, ErrSchemaValidation{Msg: "A value expected here"}
		}
	} else {
		if gb.options.OnlySchemaAttributes {
			return nil, nil, nil
		}
	}
	newNode := gb.NewNode(schemaNode)
	if err := setValue(newNode); err != nil {
		return nil, nil, err
	}
	if schemaNode != nil {
		rawValue, _ := GetRawNodeValue(newNode)
		gb.setEntityID(rawValue, parentDocumentNode, schemaNode)
	}
	t := newNode.GetLabels()
	t.Add(types...)
	t.Add(AttributeTypeValue.Name)
	newNode.SetLabels(t)
	var edge *lpg.Edge
	if parentDocumentNode != nil {
		edge = gb.targetGraph.NewEdge(parentDocumentNode, newNode, HasTerm.Name, nil)
	}
	return edge, newNode, nil
}

// ValueSetAsProperty can be called to notify the graph builder that a node value is set
func (gb GraphBuilder) ValueSetAsProperty(schemaNode *lpg.Node, graphPath []*lpg.Node, value string) {
	if schemaNode != nil {
		gb.setEntityID(value, graphPath[len(graphPath)-1], schemaNode)
	}
}

func (gb GraphBuilder) RawValueAsProperty(schemaNode *lpg.Node, graphPath []*lpg.Node, value string) error {
	return gb.ValueAsProperty(schemaNode, graphPath, func(node *lpg.Node, key string) {
		node.SetProperty(key, NewPropertyValue(key, value))
	})
}

func (gb GraphBuilder) NativeValueAsProperty(schemaNode *lpg.Node, graphPath []*lpg.Node, value interface{}) error {
	return gb.ValueAsProperty(schemaNode, graphPath, func(node *lpg.Node, key string) {
		node.SetProperty(key, NewPropertyValue(key, fmt.Sprint(value)))
	})
}

// ValueAsProperty ingests a value as a property of an ancestor node. The ancestor
func (gb GraphBuilder) ValueAsProperty(schemaNode *lpg.Node, graphPath []*lpg.Node, setValue func(*lpg.Node, string)) error {
	// Schema node cannot be nil here
	if schemaNode == nil {
		return ErrInvalidInput{Msg: "Missing schema node"}
	}
	if !schemaNode.HasLabel(AttributeTypeValue.Name) {
		return ErrSchemaValidation{Msg: "A value expected here"}
	}
	asPropertyOf, propertyName := GetIngestAsProperty(schemaNode)
	if len(propertyName) == 0 {
		return ErrCannotDeterminePropertyName{SchemaNodeID: GetNodeID(schemaNode)}
	}
	var targetNode *lpg.Node
	if len(asPropertyOf) == 0 {
		targetNode = graphPath[len(graphPath)-1]
	} else {
		// Find ancestor that is instance of asPropertyOf
		for i := len(graphPath) - 1; i >= 0; i-- {
			if SchemaNodeIDTerm.PropertyValue(graphPath[i]) == asPropertyOf {
				targetNode = graphPath[i]
				break
			}
		}
	}
	if targetNode == nil {
		return ErrCannotFindAncestor{SchemaNodeID: GetNodeID(schemaNode)}
	}
	setValue(targetNode, propertyName)
	if v, ok := GetPropertyValueAs[string](targetNode, propertyName); ok {
		gb.setEntityID(v, graphPath[len(graphPath)-1], schemaNode)
	}
	return nil
}

func (gb GraphBuilder) CollectionAsNode(schemaNode, parentNode *lpg.Node, typeTerm string, types ...string) (*lpg.Edge, *lpg.Node, error) {
	if schemaNode != nil {
		if !schemaNode.HasLabel(typeTerm) {
			return nil, nil, ErrSchemaValidation{Msg: fmt.Sprintf("A %s is expected here but found %s", typeTerm, schemaNode.GetLabels())}
		}
	}
	if schemaNode == nil && gb.options.OnlySchemaAttributes {
		return nil, nil, nil
	}
	ret := gb.NewNode(schemaNode)
	t := ret.GetLabels()
	t.Add(types...)
	t.Add(typeTerm)
	ret.SetLabels(t)
	var edge *lpg.Edge
	if parentNode != nil {
		edge = gb.targetGraph.NewEdge(parentNode, ret, GetOutputEdgeLabel(schemaNode), nil)
	}
	return edge, ret, nil
}

func (gb GraphBuilder) CollectionAsEdge(schemaNode, parentNode *lpg.Node, typeTerm string, types ...string) (*lpg.Edge, error) {
	if schemaNode != nil {
		if !schemaNode.HasLabel(typeTerm) {
			return nil, ErrSchemaValidation{Msg: fmt.Sprintf("A %s is expected here but found %s", typeTerm, schemaNode.GetLabels())}
		}
	}
	if schemaNode == nil && gb.options.OnlySchemaAttributes {
		return nil, nil
	}
	if parentNode == nil {
		return nil, ErrDataIngestion{Err: fmt.Errorf("Document root object cannot be an edge")}
	}
	blankNode := gb.NewNode(schemaNode)
	edgeLabel := determineEdgeLabel(schemaNode)
	if len(edgeLabel) == 0 {
		return nil, ErrCannotDetermineEdgeLabel{SchemaNodeID: GetNodeID(schemaNode)}
	}
	t := blankNode.GetLabels()
	t.Add(types...)
	t.Add(typeTerm)
	blankNode.SetLabels(t)
	edge := gb.targetGraph.NewEdge(parentNode, blankNode, edgeLabel, nil)
	return edge, nil
}

// ObjectAsNode creates a new object node
func (gb GraphBuilder) ObjectAsNode(schemaNode, parentNode *lpg.Node, types ...string) (*lpg.Edge, *lpg.Node, error) {
	return gb.CollectionAsNode(schemaNode, parentNode, AttributeTypeObject.Name, types...)
}

func (gb GraphBuilder) ArrayAsNode(schemaNode, parentNode *lpg.Node, types ...string) (*lpg.Edge, *lpg.Node, error) {
	return gb.CollectionAsNode(schemaNode, parentNode, AttributeTypeArray.Name, types...)
}

// ObjectAsEdge creates an object node as an edge using the following scheme:
//
//	parent --object--> _blankNode --...
func (gb GraphBuilder) ObjectAsEdge(schemaNode, parentNode *lpg.Node, types ...string) (*lpg.Edge, error) {
	return gb.CollectionAsEdge(schemaNode, parentNode, AttributeTypeObject.Name, types...)
}

func (gb GraphBuilder) ArrayAsEdge(schemaNode, parentNode *lpg.Node, types ...string) (*lpg.Edge, error) {
	return gb.CollectionAsEdge(schemaNode, parentNode, AttributeTypeArray.Name, types...)
}

// PostNodeIngest calls the post node ingestion functions for properties that has one
func (gp GraphBuilder) PostNodeIngest(schemaNode, docNode *lpg.Node) error {
	if schemaNode == nil {
		return nil
	}
	var err error
	schemaNode.ForEachProperty(func(key string, value interface{}) bool {
		pv, ok := value.(PropertyValue)
		if !ok {
			return true
		}
		ni, ok := pv.Sem().Metadata.(PostNodeIngest)
		if !ok {
			return true
		}
		if e := ni.ProcessNodePostIngest(pv, docNode, schemaNode); e != nil {
			err = e
			return false
		}
		return true
	})
	return err
	return nil
}

// PostIngestSchemaNode calls the post ingest functions for properties
// of the document nodes for the given schema node.
func (gb GraphBuilder) PostIngestSchemaNode(schemaRootNode, schemaNode, docRootNode *lpg.Node, nodeIDMap map[string][]*lpg.Node) error {
	var err error
	schemaNodeID := GetNodeID(schemaNode)
	schemaNode.ForEachProperty(func(key string, value interface{}) bool {
		pv, ok := value.(PropertyValue)
		if !ok {
			return true
		}
		ni, ok := pv.Sem().Metadata.(PostIngest)
		if !ok {
			return true
		}

		// First process all doc nodes that already exist
		for _, docNode := range nodeIDMap[schemaNodeID] {
			if e := ni.ProcessNodePostDocIngest(schemaRootNode, schemaNode, pv, docNode); e != nil {
				err = e
				return false
			}
		}

		// Now process schema nodes for which there are missing nodes To
		// do that, find the document nodes that are instances of the
		// parent of the schema node
		if IsEntityRoot(schemaNode) {
			return true
		}
		parentSchemaNode := GetParentAttribute(schemaNode)
		for _, parentDocNode := range nodeIDMap[GetNodeID(parentSchemaNode)] {
			docNode, e := EnsurePath(docRootNode, parentDocNode, schemaRootNode, schemaNode, func(parentNode, childSchemaNode *lpg.Node) (*lpg.Node, error) {
				n := gb.NewNode(childSchemaNode)
				gb.targetGraph.NewEdge(parentNode, n, HasTerm.Name, nil)
				return n, nil
			})
			if e != nil {
				err = e
				return false
			}
			if e := ni.ProcessNodePostDocIngest(schemaRootNode, schemaNode, pv, docNode); e != nil {
				err = e
				return false
			}
		}
		return true
	})
	return err
}

// PostIngest calls the post ingestion functions for properties that has one
func (gb GraphBuilder) PostIngest(schemaRootNode, docRootNode *lpg.Node) error {
	if schemaRootNode == nil {
		return nil
	}
	nodeIDMap := GetSchemaNodeIDMap(docRootNode)
	var err error
	ForEachAttributeNode(schemaRootNode, func(schemaNode *lpg.Node, _ []*lpg.Node) bool {
		err = gb.PostIngestSchemaNode(schemaRootNode, schemaNode, docRootNode, nodeIDMap)
		return err == nil
	})
	gb.AddDefaults(schemaRootNode, docRootNode)
	return err
}

// NewUniqueEdge create a new edge if one does not exist.
func (gb GraphBuilder) NewUniqueEdge(fromNode, toNode *lpg.Node, label string, properties map[string]interface{}) *lpg.Edge {
	fromItr := fromNode.GetEdgesWithLabel(lpg.OutgoingEdge, label)
	toItr := toNode.GetEdgesWithLabel(lpg.IncomingEdge, label)
	itr := fromItr
	if fromItr.MaxSize() != -1 && toItr.MaxSize() != -1 {
		if fromItr.MaxSize() > toItr.MaxSize() {
			itr = toItr
		}
	}
	for itr.Next() {
		edge := itr.Edge()
		if edge.GetTo() != toNode ||
			edge.GetFrom() != fromNode {
			continue
		}
		if properties != nil {
			for k, v := range properties {
				edge.SetProperty(k, v)
			}
		}
		return edge
	}
	return gb.targetGraph.NewEdge(fromNode, toNode, label, properties)
}

// Link the given node, or create a link from the parent node.
//
// `spec` is the link spec. `docNode` contains the ingested document
// node that will be linked. It can be nil. `parentNode` is the
// document node containing the docNode.
func (gb GraphBuilder) linkNode(spec *LinkSpec, docNode, parentNode, entityRoot *lpg.Node, foreignKeys []ForeignKeyInfo, entityInfo EntityInfoIndex) error {

	var linkNode *lpg.Node
	specIsValueNode := spec.SchemaNode.HasLabel(AttributeTypeValue.Name)
	if specIsValueNode {
		if len(spec.LinkNode) != 0 {
			WalkNodesInEntity(entityRoot, func(n *lpg.Node) bool {
				if IsInstanceOf(n, spec.LinkNode) {
					linkNode = n
					return false
				}
				return true
			})
		}
		if linkNode == nil {
			linkNode = entityRoot
		}
	}

	if docNode != nil {
		docNode.SetProperty(ReferenceFKFor.Name, ReferenceFKFor.MustPropertyValue(spec.TargetEntity))
		docNode.SetProperty(ReferenceFK.Name, ReferenceFK.MustPropertyValue(foreignKeys[0].ForeignKey))
	}
	var nodeProperties map[string]interface{}
	if spec.IngestAs == IngestAsEdge && docNode != nil {
		// This document node is removed and a link from the parent to the target is created
		nodeProperties = CloneProperties(docNode)
		docNode.DetachAndRemove()
	}

	link := func(ref []*lpg.Node) {
		for _, linkRef := range ref {
			if specIsValueNode {
				if spec.Forward {
					gb.NewUniqueEdge(linkNode, linkRef, spec.Label, nodeProperties)
				} else {
					gb.NewUniqueEdge(linkRef, linkNode, spec.Label, nodeProperties)
				}
			} else {
				if spec.IngestAs == IngestAsEdge {
					// Node is already removed. Make an edge
					if spec.Forward {
						gb.NewUniqueEdge(parentNode, linkRef, spec.Label, nodeProperties)
					} else {
						gb.NewUniqueEdge(linkRef, parentNode, spec.Label, nodeProperties)
					}
				} else {
					if docNode == nil {
						docNode = gb.NewNode(spec.SchemaNode)
						gb.NewUniqueEdge(parentNode, docNode, HasTerm.Name, nil)
					}
					// A link from this document node to target is created
					if spec.Forward {
						gb.NewUniqueEdge(docNode, linkRef, spec.Label, nil)
					} else {
						gb.NewUniqueEdge(linkRef, docNode, spec.Label, nil)
					}
				}
			}
		}
	}

	// Find remote references
	if len(spec.FK) == 0 {
		ref, err := spec.FindReference(entityInfo, spec.FK)
		if err != nil {
			return err
		}
		link(ref)
		return nil
	}
	for _, fk := range foreignKeys {
		ref, err := spec.FindReference(entityInfo, fk.ForeignKey)
		if err != nil {
			return err
		}
		if len(ref) == 0 {
			continue
		}
		link(ref)
	}

	return nil
}

func (gb GraphBuilder) LinkNodes(ctx *Context, schema *Layer) error {
	entityInfo := GetEntityInfo(gb.GetGraph())
	eix := IndexEntityInfo(entityInfo)
	specs, err := schema.GetLinkSpecs()
	if err != nil {
		return err
	}
	if err := gb.LinkNodesWithSpecs(ctx, specs, eix); err != nil {
		return err
	}
	return nil
}

func (gb GraphBuilder) LinkNodesWithSpecs(ctx *Context, specs []*LinkSpec, eix EntityInfoIndex) error {
	for _, spec := range specs {
		attrId := GetNodeID(spec.SchemaNode)
		ctx.GetLogger().Debug(map[string]interface{}{"graphBuilder": "linkNodes", "linking": attrId})
		// Find nodes that are instance of this node
		parentSchemaNodeID := GetNodeID(spec.ParentSchemaNode)
		parentDocNodes := GetNodesInstanceOf(gb.targetGraph, parentSchemaNodeID)
		ctx.GetLogger().Debug(map[string]interface{}{"graphBuilder": "linkNodes", "parentSchemaNodeID": parentSchemaNodeID})

		fkMap := make(map[*lpg.Node][]ForeignKeyInfo)

		for _, parent := range parentDocNodes {
			entityRoot := GetEntityRoot(parent)
			if entityRoot == nil {
				return ErrCannotResolveLink(*spec)
			}
			foreignKeys, ok := fkMap[entityRoot]
			if !ok {
				var err error
				foreignKeys, err = spec.GetForeignKeys(entityRoot)
				if err != nil {
					return err
				}
				fkMap[entityRoot] = foreignKeys
			}
			if len(foreignKeys) == 0 && len(spec.FK) != 0 {
				// Nothing to link
				continue
			}

			// Each parent node has at least one reference node child
			childFound := false
			for edges := parent.GetEdges(lpg.OutgoingEdge); edges.Next(); {
				childNode := edges.Edge().GetTo()
				if !IsDocumentNode(childNode) {
					continue
				}
				if SchemaNodeIDTerm.PropertyValue(childNode) != attrId {
					continue
				}
				// childNode is an instance of attrNode, which is a link
				childFound = true
				if err := gb.linkNode(spec, childNode, parent, entityRoot, foreignKeys, eix); err != nil {
					return err
				}
			}
			if !childFound {
				if err := gb.linkNode(spec, nil, parent, entityRoot, foreignKeys, eix); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
