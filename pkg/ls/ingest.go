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

import ()

// // Ingester keeps the schema and the ingestion options
// type Ingester struct {
// 	// The schema variant to use during ingestion
// 	Schema *Layer

// 	// If true, schema node properties are embedded into document
// 	// nodes. If false, schema nodes are preserved as separate nodes,
// 	// with an instanceOf edge between the document node to the schema
// 	// node.
// 	EmbedSchemaNodes bool

// 	// If OnlySchemaAttributes is true, only ingest data points if there is a schema for it.
// 	// If OnlySchemaAttributes is false, ingest whether or not there is a schema for it.
// 	OnlySchemaAttributes bool

// 	// IngestEmptyValues is true if the value to ingest contains data, otherwise default to false
// 	IngestEmptyValues bool

// 	ValuesetFunc func(ValuesetLookupRequest) (ValuesetLookupResponse, error)

// 	// SchemaNodeMap is used to keep a mapping of schema nodes copied into the
// 	// target graph. The key is a schema node. The value is the node in
// 	// target graph.
// 	SchemaNodeMap map[graph.Node]graph.Node

// 	// The target graph
// 	Graph graph.Graph
// }

// // DefaultNodeIDGenerator returns Ingester.Schema.ID + join(path,".")
// func DefaultNodeIDGenerator(path NodePath, schemaNode graph.Node) string {
// 	return path.String()
// }

// // IngestionContext keeps the ingestion state
// type IngestionContext struct {
// 	*Context
// 	// SourcePath is the path in the input document that is being
// 	// ingested. This is mainly useful for diagnostic messages as it
// 	// shows which field is being processed
// 	SourcePath NodePath
// 	// GraphPath is the path in the target graph. At any given point,
// 	// the last element of GraphPath gives the parent element
// 	GraphPath []graph.Node
// 	// SchemaPath is the path to the current schema node. Elements can be nil
// 	SchemaPath []graph.Node
// }

// // NewLevel return a new context with a new document node in graph path
// func (ctx IngestionContext) NewLevel(docNode graph.Node) IngestionContext {
// 	ctx.GraphPath = append(ctx.GraphPath, docNode)
// 	return ctx
// }

// // New returns a new ingestion context that is a copy of the original with one more level added
// func (ctx IngestionContext) New(key interface{}, schemaNode graph.Node) IngestionContext {
// 	ctx.SourcePath = ctx.SourcePath.Append(key)
// 	ctx.SchemaPath = append(ctx.SchemaPath, schemaNode)
// 	return ctx
// }

// // GetParentNode returns the last element of the graph path
// func (ctx IngestionContext) GetParentNode() graph.Node {
// 	if len(ctx.GraphPath) == 0 {
// 		return nil
// 	}
// 	return ctx.GraphPath[len(ctx.GraphPath)-1]
// }

// // GetSchemaNode returns the current schema node
// func (ctx IngestionContext) GetSchemaNode() graph.Node {
// 	if len(ctx.SchemaPath) == 0 {
// 		return nil
// 	}
// 	return ctx.SchemaPath[len(ctx.SchemaPath)-1]
// }

// // GetEntityRootNode returns the root node of the current entity. Returns nil if it cannot be determined.
// func (ctx IngestionContext) GetEntityRootNode() graph.Node {
// 	for i := len(ctx.GraphPath) - 1; i >= 0; i-- {
// 		_, root := ctx.GraphPath[i].GetProperty(EntitySchemaTerm)
// 		if root {
// 			return ctx.GraphPath[i]
// 			break
// 		}
// 	}
// 	return nil
// }

// func newIngestionContext(context *Context, baseID string, schemaRoot graph.Node) IngestionContext {
// 	ctx := IngestionContext{
// 		Context: context,
// 	}
// 	if len(baseID) > 0 {
// 		ctx.SourcePath = ctx.SourcePath.Append(baseID)
// 	}
// 	if schemaRoot != nil {
// 		ctx.SchemaPath = []graph.Node{schemaRoot}
// 	}
// 	return ctx
// }

// // Start ingestion. Returns the path initialized with the baseId, and
// // the schema root.
// func (ingester *Ingester) Start(context *Context, baseID string) IngestionContext {
// 	var schRoot graph.Node
// 	if ingester.Schema != nil {
// 		schRoot = ingester.Schema.GetSchemaRootNode()
// 	}
// 	ctx := newIngestionContext(context, baseID, schRoot)
// 	if ingester.SchemaNodeMap == nil {
// 		ingester.SchemaNodeMap = make(map[graph.Node]graph.Node)
// 	}
// 	return ctx
// }

//  ValueAsEdge ingests a value using the following scheme:
//
//  input: (name: value)
//  output: --(label)-->(value:value, attributeName:name)
//
// where label=attributeName (in this case "name") if edgeLabel is not
// specified in schema.
// func (ingester *Ingester) ValueAsEdge(ictx IngestionContext, value string, types ...string) (graph.Edge, error) {
// 	schemaNode := ictx.GetSchemaNode()
// 	if schemaNode == nil {
// 		return nil, ErrInvalidInput{Msg: "missing schemaNode"}
// 	}
// 	if !ingester.IngestEmptyValues && len(value) == 0 {
// 		return nil, nil
// 	}
// 	if !schemaNode.GetLabels().Has(AttributeTypeValue) {
// 		return nil, ErrSchemaValidation{Msg: "A value attribute is expected here", Path: ictx.SourcePath.Copy()}
// 	}
// 	if len(ictx.GraphPath) == 0 {
// 		return nil, ErrDataIngestion{Key: ictx.SourcePath.String(), Err: fmt.Errorf("Document root value cannot be an edge")}
// 	}
// 	edgeLabel := determineEdgeLabel(schemaNode)
// 	if len(edgeLabel) == 0 {
// 		return nil, ErrCannotDetermineEdgeLabel{Path: ictx.SourcePath.Copy()}
// 	}
// 	node := ingester.NewNode(ictx)
// 	SetRawNodeValue(node, value)
// 	t := node.GetLabels()
// 	t.Add(types...)
// 	t.Add(AttributeTypeValue)
// 	node.SetLabels(t)
// 	edge := ingester.Graph.NewEdge(ictx.GetParentNode(), node, edgeLabel, nil)
// 	return edge, nil
// }

// // ValueAsNode creates a new value node. The new node has the given value
// // and the types
// func (ingester *Ingester) ValueAsNode(ictx IngestionContext, value string, types ...string) (graph.Edge, graph.Node, error) {
// 	schemaNode := ictx.GetSchemaNode()
// 	if ingester.OnlySchemaAttributes && schemaNode == nil {
// 		return nil, nil, nil
// 	}
// 	if schemaNode != nil {
// 		if !schemaNode.GetLabels().Has(AttributeTypeValue) {
// 			return nil, nil, ErrSchemaValidation{Msg: "A value is expected here", Path: ictx.SourcePath.Copy()}
// 		}
// 	}
// 	if !ingester.IngestEmptyValues && len(value) == 0 {
// 		return nil, nil, nil
// 	}
// 	newNode := ingester.NewNode(ictx)
// 	SetRawNodeValue(newNode, value)
// 	t := newNode.GetLabels()
// 	t.Add(types...)
// 	t.Add(AttributeTypeValue)
// 	newNode.SetLabels(t)
// 	var edge graph.Edge
// 	if len(ictx.GraphPath) > 0 {
// 		edge = ingester.Graph.NewEdge(ictx.GetParentNode(), newNode, HasTerm, nil)
// 	}
// 	return edge, newNode, nil
// }

// func setEntityID(ictx IngestionContext, value string) error {
// 	schemaNode := ictx.GetSchemaNode()
// 	if schemaNode == nil {
// 		return nil
// 	}
// 	attrId := GetNodeID(schemaNode)
// 	if len(attrId) == 0 {
// 		return nil
// 	}
// 	rootNode := ictx.GetEntityRootNode()
// 	if rootNode == nil {
// 		return nil
// 	}
// 	idFields := GetEntityIDFields(rootNode)
// 	if idFields == nil {
// 		return nil
// 	}
// 	entityID := AsPropertyValue(rootNode.GetProperty(EntityIDTerm))
// 	if idFields.IsString() {
// 		if idFields.AsString() == attrId {
// 			if entityID == nil {
// 				rootNode.SetProperty(EntityIDTerm, StringPropertyValue(value))
// 				return nil
// 			}
// 			if entityID.IsString() {
// 				if len(entityID.AsString()) > 0 {
// 					return ErrInvalidEntityID{Path: ictx.SourcePath.Copy()}
// 				}
// 				rootNode.SetProperty(EntityIDTerm, StringPropertyValue(value))
// 				return nil
// 			}
// 			return ErrInvalidEntityID{Path: ictx.SourcePath.Copy()}
// 		}
// 		return nil
// 	}
// 	if idFields.IsStringSlice() {
// 		found := false
// 		idf := idFields.AsStringSlice()
// 		for _, fld := range idf {
// 			if fld == attrId {
// 				found = true
// 				break
// 			}
// 		}
// 		if !found {
// 			return nil
// 		}
// 		if entityID == nil {
// 			slice := make([]string, len(idf))
// 			for x := range idf {
// 				if idf[x] == attrId {
// 					slice[x] = value
// 				}
// 			}
// 			rootNode.SetProperty(EntityIDTerm, StringSlicePropertyValue(slice))
// 			return nil
// 		}
// 		if entityID.IsStringSlice() {
// 			slice := entityID.AsStringSlice()
// 			if len(slice) != len(idf) {
// 				return ErrInvalidEntityID{Path: ictx.SourcePath.Copy()}
// 			}
// 			for x := range idf {
// 				if idf[x] == attrId {
// 					if len(slice[x]) > 0 {
// 						return ErrInvalidEntityID{Path: ictx.SourcePath.Copy()}
// 					}
// 					slice[x] = value
// 				}
// 			}
// 			rootNode.SetProperty(EntityIDTerm, StringSlicePropertyValue(slice))
// 			return nil
// 		}
// 		return ErrInvalidEntityID{Path: ictx.SourcePath.Copy()}
// 	}
// 	return ErrInvalidEntityID{Path: ictx.SourcePath.Copy()}
// }

// // ValueAsProperty ingests a value as a property of an ancestor node
// func (ingester *Ingester) ValueAsProperty(ictx IngestionContext, value string) error {
// 	// Schema node cannot be nil here
// 	schemaNode := ictx.GetSchemaNode()
// 	if schemaNode == nil {
// 		return ErrInvalidInput{Msg: "Missing schema node"}
// 	}
// 	asPropertyOf := AsPropertyValue(schemaNode.GetProperty(AsPropertyOfTerm)).AsString()
// 	propertyName := AsPropertyValue(schemaNode.GetProperty(PropertyNameTerm)).AsString()
// 	if len(propertyName) == 0 {
// 		propertyName = AsPropertyValue(schemaNode.GetProperty(AttributeNameTerm)).AsString()
// 	}
// 	if len(propertyName) == 0 {
// 		return ErrCannotDeterminePropertyName{Path: ictx.SourcePath.Copy()}
// 	}
// 	var targetNode graph.Node
// 	if len(asPropertyOf) == 0 {
// 		if len(ictx.GraphPath) > 0 {
// 			targetNode = ictx.GetParentNode()
// 		}
// 	} else {
// 		// Find ancestor that is instance of asPropertyOf
// 		for i := len(ictx.GraphPath) - 1; i >= 0; i-- {
// 			if AsPropertyValue(ictx.GraphPath[i].GetProperty(SchemaNodeIDTerm)).AsString() == asPropertyOf {
// 				targetNode = ictx.GraphPath[i]
// 				break
// 			}
// 		}
// 	}
// 	if targetNode == nil {
// 		return ErrCannotFindAncestor{Path: ictx.SourcePath.Copy()}
// 	}
// 	targetNode.SetProperty(propertyName, StringPropertyValue(value))
// 	return nil
// }

// // Value ingests a value as a node, edge-node, or as a property depending on the schema. The default is ingestion as node. Returns the node, and optionally, the edge going to that node
// func (ingester *Ingester) Value(ictx IngestionContext, value string, types ...string) (string, graph.Edge, graph.Node, error) {
// 	// Is this an ID value?
// 	if err := setEntityID(ictx, value); err != nil {
// 		return "", nil, nil, err
// 	}
// 	schemaNode := ictx.GetSchemaNode()
// 	if schemaNode == nil && ingester.OnlySchemaAttributes {
// 		return "", nil, nil, nil
// 	}
// 	switch GetIngestAs(schemaNode) {
// 	case IngestAsNode:
// 		e, g, err := ingester.ValueAsNode(ictx, value, types...)
// 		if err != nil {
// 			return "", nil, nil, err
// 		}
// 		return IngestAsNode, e, g, nil

// 	case IngestAsEdge:
// 		e, err := ingester.ValueAsEdge(ictx, value, types...)
// 		if err != nil {
// 			return "", nil, nil, err
// 		}
// 		return IngestAsEdge, e, e.GetTo(), nil

// 	case IngestAsProperty:
// 		err := ingester.ValueAsProperty(ictx, value)
// 		if err != nil {
// 			return "", nil, nil, err
// 		}
// 		return IngestAsProperty, nil, nil, nil
// 	}
// 	return "", nil, nil, nil
// }

// func (ingester *Ingester) collectionAsNode(ictx IngestionContext, typeTerm string, types ...string) (graph.Edge, graph.Node, error) {
// 	ret := ingester.NewNode(ictx)
// 	t := ret.GetLabels()
// 	t.Add(types...)
// 	// define that ret is an object
// 	t.Add(typeTerm)
// 	ret.SetLabels(t)
// 	var edge graph.Edge
// 	if len(ictx.GraphPath) > 0 {
// 		edge = ingester.Graph.NewEdge(ictx.GetParentNode(), ret, HasTerm, nil)
// 	}
// 	return edge, ret, nil
// }

// // ObjectAsNode creates a new object node
// func (ingester *Ingester) ObjectAsNode(ictx IngestionContext, types ...string) (graph.Edge, graph.Node, error) {
// 	// An object node
// 	// There is a schema node for this node. It must be an object
// 	if ictx.GetSchemaNode() != nil {
// 		if !ictx.GetSchemaNode().GetLabels().Has(AttributeTypeObject) {
// 			return nil, nil, ErrSchemaValidation{Msg: fmt.Sprintf("An object is expected here but found %s", ictx.GetSchemaNode().GetLabels()), Path: ictx.SourcePath.Copy()}
// 		}
// 	}
// 	if ictx.GetSchemaNode() == nil && ingester.OnlySchemaAttributes {
// 		return nil, nil, nil
// 	}
// 	return ingester.collectionAsNode(ictx, AttributeTypeObject, types...)
// }

// func (ingester *Ingester) ArrayAsNode(ictx IngestionContext, types ...string) (graph.Edge, graph.Node, error) {
// 	if ictx.GetSchemaNode() != nil {
// 		if !ictx.GetSchemaNode().GetLabels().Has(AttributeTypeArray) {
// 			return nil, nil, ErrSchemaValidation{Msg: fmt.Sprintf("An array is expected here but found %s", ictx.GetSchemaNode().GetLabels()), Path: ictx.SourcePath.Copy()}
// 		}
// 	}
// 	if ictx.GetSchemaNode() == nil && ingester.OnlySchemaAttributes {
// 		return nil, nil, nil
// 	}
// 	return ingester.collectionAsNode(ictx, AttributeTypeArray, types...)
// }

// func (ingester *Ingester) collectionAsEdge(ictx IngestionContext, typeTerm string, types ...string) (graph.Edge, error) {
// 	if len(ictx.GraphPath) == 0 {
// 		return nil, ErrDataIngestion{Key: ictx.SourcePath.String(), Err: fmt.Errorf("Document root object cannot be an edge")}
// 	}
// 	blankNode := ingester.NewNode(ictx)
// 	edgeLabel := determineEdgeLabel(ictx.GetSchemaNode())
// 	if len(edgeLabel) == 0 {
// 		return nil, ErrCannotDetermineEdgeLabel{Path: ictx.SourcePath.Copy()}
// 	}
// 	t := blankNode.GetLabels()
// 	t.Add(types...)
// 	// define that newEdgeNode.Node is an object
// 	t.Add(typeTerm)
// 	blankNode.SetLabels(t)
// 	edge := ingester.Graph.NewEdge(ictx.GetParentNode(), blankNode, edgeLabel, nil)
// 	return edge, nil
// }

// // ObjectAsEdge creates an object node as an edge using the following scheme:
// //
// //  parent --object--> _blankNode --...
// func (ingester *Ingester) ObjectAsEdge(ictx IngestionContext, types ...string) (graph.Edge, error) {
// 	if ictx.GetSchemaNode() != nil {
// 		if !ictx.GetSchemaNode().GetLabels().Has(AttributeTypeObject) {
// 			return nil, ErrSchemaValidation{Msg: fmt.Sprintf("An object is expected here but found %s", ictx.GetSchemaNode().GetLabels()), Path: ictx.SourcePath.Copy()}
// 		}
// 	}
// 	if ictx.GetSchemaNode() == nil && ingester.OnlySchemaAttributes {
// 		return nil, nil
// 	}
// 	return ingester.collectionAsEdge(ictx, AttributeTypeObject, types...)
// }

// func (ingester *Ingester) ArrayAsEdge(ictx IngestionContext, types ...string) (graph.Edge, error) {
// 	if ictx.GetSchemaNode() != nil {
// 		if !ictx.GetSchemaNode().GetLabels().Has(AttributeTypeArray) {
// 			return nil, ErrSchemaValidation{Msg: fmt.Sprintf("An array is expected here but found %s", ictx.GetSchemaNode().GetLabels()), Path: ictx.SourcePath.Copy()}
// 		}
// 	}
// 	if ictx.GetSchemaNode() == nil && ingester.OnlySchemaAttributes {
// 		return nil, nil
// 	}
// 	return ingester.collectionAsEdge(ictx, AttributeTypeArray, types...)
// }

// // Object ingests an object as a node or edge
// func (ingester *Ingester) Object(ictx IngestionContext, types ...string) (string, graph.Edge, graph.Node, error) {
// 	schemaNode := ictx.GetSchemaNode()
// 	switch GetIngestAs(schemaNode) {
// 	case IngestAsNode:
// 		e, g, err := ingester.ObjectAsNode(ictx, types...)
// 		if err != nil {
// 			return "", nil, nil, err
// 		}
// 		return IngestAsNode, e, g, nil

// 	case "edge":
// 		e, err := ingester.ObjectAsEdge(ictx, types...)
// 		if err != nil {
// 			return "", nil, nil, err
// 		}
// 		return IngestAsEdge, e, e.GetTo(), nil
// 	}
// 	return "", nil, nil, nil
// }

// // Array ingests an array as a node or edge
// func (ingester *Ingester) Array(ictx IngestionContext, types ...string) (string, graph.Edge, graph.Node, error) {
// 	schemaNode := ictx.GetSchemaNode()
// 	switch GetIngestAs(schemaNode) {
// 	case IngestAsNode:
// 		e, g, err := ingester.ArrayAsNode(ictx, types...)
// 		if err != nil {
// 			return "", nil, nil, err
// 		}
// 		return IngestAsNode, e, g, nil

// 	case IngestAsEdge:
// 		e, err := ingester.ArrayAsEdge(ictx, types...)
// 		if err != nil {
// 			return "", nil, nil, err
// 		}
// 		return IngestAsEdge, e, e.GetTo(), nil
// 	}
// 	return "", nil, nil, nil
// }

// // Instantiate the latest schema node element in the context, and
// // connect it to its parent. Returns the new ingestion context to add
// // nodes after the new node. If the new node is an object or array,
// // the returned ingestion context is a level deeper.
// func (ingester *Ingester) Instantiate(ictx IngestionContext) (string, graph.Edge, graph.Node, IngestionContext, error) {
// 	schemaNode := ictx.GetSchemaNode()
// 	if schemaNode == nil {
// 		return "", nil, nil, ictx, nil
// 	}
// 	// Create new node
// 	switch {
// 	case schemaNode.GetLabels().Has(AttributeTypeValue):
// 		t := ingester.IngestEmptyValues
// 		ingester.IngestEmptyValues = true
// 		s, e, n, err := ingester.Value(ictx, "")
// 		ingester.IngestEmptyValues = t
// 		if err != nil {
// 			return "", nil, nil, ictx, err
// 		}
// 		return s, e, n, ictx, nil

// 	case schemaNode.GetLabels().Has(AttributeTypeObject):
// 		t := ingester.IngestEmptyValues
// 		ingester.IngestEmptyValues = true
// 		s, e, n, err := ingester.Object(ictx)
// 		ingester.IngestEmptyValues = t
// 		if err != nil {
// 			return "", nil, nil, ictx, err
// 		}
// 		ictx = ictx.NewLevel(n)
// 		return s, e, n, ictx, nil
// 	case schemaNode.GetLabels().Has(AttributeTypeArray):
// 		t := ingester.IngestEmptyValues
// 		ingester.IngestEmptyValues = true
// 		s, e, n, err := ingester.Array(ictx)
// 		ingester.IngestEmptyValues = t
// 		if err != nil {
// 			return "", nil, nil, ictx, err
// 		}
// 		ictx = ictx.NewLevel(n)
// 		return s, e, n, ictx, nil
// 	}
// 	// Cannot instantiate
// 	return "", nil, nil, ictx, fmt.Errorf("Cannot instantiate node/edge")
// }

// // Validate the document node with the schema node
// func (ingester *Ingester) Validate(ictx IngestionContext, documentNode graph.Node) error {
// 	if ictx.GetSchemaNode() != nil {
// 		if err := ValidateDocumentNodeBySchema(documentNode, ictx.GetSchemaNode()); err != nil {
// 			return err
// 		}
// 	}
// 	return nil
// }

// // NewNode creates a new graph node, either by using the NewNodeFunc
// // or by creating a new node using DefaultNodeIDGenerator. Then it
// // either merges schema properties into the new node, or creates an
// // instanceOf edge to the schema node.
// func (ingester *Ingester) NewNode(ictx IngestionContext) graph.Node {
// 	node := ingester.Graph.NewNode([]string{DocumentNodeTerm}, nil)
// 	schemaNode := ictx.GetSchemaNode()
// 	SetNodeID(node, DefaultNodeIDGenerator(ictx.SourcePath, schemaNode))
// 	if schemaNode == nil {
// 		return node
// 	}
// 	types := node.GetLabels()
// 	types.Add(FilterNonLayerTypes(schemaNode.GetLabels().Slice())...)
// 	node.SetLabels(types)
// 	node.SetProperty(SchemaNodeIDTerm, StringPropertyValue(GetNodeID(schemaNode)))

// 	pat := graph.Pattern{{
// 		Labels:     graph.NewStringSet(AttributeNodeTerm),
// 		Properties: map[string]interface{}{NodeIDTerm: GetNodeID(schemaNode)},
// 	}}
// 	acc := graph.DefaultMatchAccumulator{}
// 	pat.Run(ingester.Graph, nil, &acc)
// 	nodes := acc.GetHeadNodes()

// 	if ingester.EmbedSchemaNodes {
// 		ingester.EmbedSchemaNode(node, schemaNode)
// 		// Copy the subtrees for all nodes connected to the schema node
// 		for edges := schemaNode.GetEdges(graph.OutgoingEdge); edges.Next(); {
// 			edge := edges.Edge()
// 			if IsAttributeTreeEdge(edge) {
// 				continue
// 			}
// 			graph.CopySubgraph(edge.GetTo(), ingester.Graph, ClonePropertyValueFunc, ingester.SchemaNodeMap)
// 			ingester.Graph.NewEdge(node, ingester.SchemaNodeMap[edge.GetTo()], edge.GetLabel(), nil)
// 		}
// 	} else {
// 		// Copy the schema node into this
// 		// If the schema node already exists in the target graph, use it
// 		if len(nodes) != 0 {
// 			ingester.Graph.NewEdge(node, nodes[0], InstanceOfTerm, nil)
// 		} else {
// 			// Copy the node
// 			newNode := graph.CopyNode(schemaNode, ingester.Graph, ClonePropertyValueFunc)
// 			ingester.SchemaNodeMap[schemaNode] = newNode
// 			ingester.Graph.NewEdge(node, newNode, InstanceOfTerm, nil)
// 			// Copy the subtrees for all nodes connected to the schema node
// 			for edges := schemaNode.GetEdges(graph.OutgoingEdge); edges.Next(); {
// 				edge := edges.Edge()
// 				if IsAttributeTreeEdge(edge) {
// 					continue
// 				}
// 				graph.CopySubgraph(edge.GetTo(), ingester.Graph, ClonePropertyValueFunc, ingester.SchemaNodeMap)
// 				ingester.Graph.NewEdge(node, ingester.SchemaNodeMap[edge.GetTo()], edge.GetLabel(), nil)
// 			}
// 		}
// 	}
// 	// If this is an entity boundary, mark it
// 	pv, rootNode := schemaNode.GetProperty(EntitySchemaTerm)
// 	if rootNode {
// 		node.SetProperty(EntitySchemaTerm, pv)
// 	}
// 	return node
// }

// // EmbedSchemaNode merges the schema node properties with the target
// // node properties. No properties are overwritten in the target
// // node. The schema node types that are not schema node types are also
// // merged with the target node types.
// func (ingester *Ingester) EmbedSchemaNode(targetNode, schemaNode graph.Node) {
// 	schemaNode.ForEachProperty(func(k string, v interface{}) bool {
// 		if k == NodeIDTerm {
// 			return true
// 		}
// 		if pv, ok := v.(*PropertyValue); ok {
// 			targetNode.SetProperty(k, pv.Clone())
// 		} else {
// 			targetNode.SetProperty(k, v)
// 		}
// 		return true
// 	})
// }

// type PolymorphicOption struct {
// 	SchemaNode     graph.Node
// 	Graph          graph.Graph
// 	IngestedNode   graph.Node
// 	IngestionError error
// }

// // TestPolymorphicOptions ingests data at the current location using
// // all polymorphic options with OnlySchemaAttributes set to true, and
// // returns the results of those ingestions
// func (ingester *Ingester) TestPolymorphicOptions(ictx IngestionContext, ingest func(*Ingester, IngestionContext) (graph.Node, error)) ([]PolymorphicOption, error) {
// 	if ictx.GetSchemaNode() == nil {
// 		return nil, ErrDataIngestion{Key: ictx.SourcePath.String(), Err: fmt.Errorf("A schema is required to ingest polymorphic nodes")}
// 	}
// 	// Polymorphic node. Try each option
// 	// iterate through all edges of the schema node which have a polymorphic attribute
// 	ret := make([]PolymorphicOption, 0)
// 	for edges := ictx.GetSchemaNode().GetEdgesWithLabel(graph.OutgoingEdge, OneOfTerm); edges.Next(); {
// 		edge := edges.Edge()
// 		optionNode := edge.GetTo()

// 		pOption := PolymorphicOption{SchemaNode: optionNode}

// 		newIngester := *ingester
// 		newIngester.SchemaNodeMap = make(map[graph.Node]graph.Node)
// 		newIngester.OnlySchemaAttributes = true
// 		newIngester.Graph = NewDocumentGraph()
// 		pOption.Graph = newIngester.Graph
// 		newContext := IngestionContext{
// 			Context:    ictx.Context,
// 			SchemaPath: []graph.Node{optionNode},
// 		}
// 		pOption.IngestedNode, pOption.IngestionError = ingest(&newIngester, newContext)
// 		ret = append(ret, pOption)
// 	}
// 	return ret, nil
// }

// // Polymorphic tests all options in the schema by calling ingest func to find the actual type. Then ingests using that option
// func (ingester *Ingester) Polymorphic(ictx IngestionContext, test, ingest func(*Ingester, IngestionContext) (graph.Node, error)) (graph.Node, error) {
// 	if ictx.GetSchemaNode() == nil {
// 		return nil, ErrDataIngestion{Key: ictx.SourcePath.String(), Err: fmt.Errorf("A schema is required to ingest polymorphic nodes")}
// 	}
// 	optionResults, err := ingester.TestPolymorphicOptions(ictx, test)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// Look at the results and find the actual type
// 	numMatches := 0
// 	var matched *PolymorphicOption
// 	for ix, r := range optionResults {
// 		if r.IngestionError == nil && r.IngestedNode != nil {
// 			numMatches++
// 			matched = &optionResults[ix]
// 		}
// 	}
// 	if numMatches == 0 {
// 		if ingester.OnlySchemaAttributes {
// 			return nil, nil
// 		}
// 		return nil, ErrSchemaValidation{Msg: "None of the options of the polymorphic node matched:" + GetNodeID(ictx.GetSchemaNode()), Path: ictx.SourcePath.Copy()}
// 	}
// 	if numMatches > 1 {
// 		return nil, ErrSchemaValidation{Msg: "Multiple options of the polymorphic node matched:" + GetNodeID(ictx.GetSchemaNode()), Path: ictx.SourcePath.Copy()}
// 	}

// 	// Only one matched
// 	// Reingest
// 	return ingest(ingester, ictx.New("", matched.SchemaNode))
// }

// // Finish ingesting by assigning node IDs and linking nodes to their
// // entity root nodes. If generateIDFunc is nil, the default ID
// // generation function is used
// func (ingester *Ingester) Finish(ictx IngestionContext, root graph.Node) error {
// 	if root != nil {
// 		if ingester.Schema != nil {
// 			for nodes := ingester.Schema.Graph.GetNodes(); nodes.Next(); {
// 				node := nodes.Node()
// 				if err := ingester.ProcessValueset(ictx, root, node); err != nil {
// 					return err
// 				}
// 			}
// 		}
// 	}
// 	var entityInfo map[graph.Node]EntityInfo
// 	if ingester.Schema != nil {
// 		entityInfo = GetEntityRootNodes(ingester.Graph)
// 		for nodes := ingester.Schema.Graph.GetNodes(); nodes.Next(); {
// 			attrNode := nodes.Node()
// 			ls, err := GetLinkSpec(attrNode)
// 			if err != nil {
// 				return err
// 			}
// 			if ls == nil {
// 				continue
// 			}
// 			attrId := GetNodeID(attrNode)
// 			// Found a link spec. Find corresponding nodes in the document, or find parents of those nodes
// 			parentSchemaNode := GetParentAttribute(attrNode)
// 			// Find nodes that are instance of this node
// 			parentDocNodes := GetNodesInstanceOf(ingester.Graph, GetNodeID(parentSchemaNode))
// 			for _, parent := range parentDocNodes {
// 				// Each parent node has at least one reference node child
// 				childFound := false
// 				for edges := parent.GetEdges(graph.OutgoingEdge); edges.Next(); {
// 					childNode := edges.Edge().GetTo()
// 					if !childNode.GetLabels().Has(DocumentNodeTerm) {
// 						continue
// 					}
// 					if AsPropertyValue(childNode.GetProperty(SchemaNodeIDTerm)).AsString() != attrId {
// 						continue
// 					}
// 					// childNode is an instance of attrNode, which is a link
// 					childFound = true
// 					if err := ingester.Link(ictx, ls, childNode, parent, entityInfo); err != nil {
// 						return err
// 					}
// 				}
// 				if !childFound {
// 					if err := ingester.Link(ictx, ls, nil, parent, entityInfo); err != nil {
// 						return err
// 					}
// 				}
// 			}
// 		}
// 	}

// 	return nil
// }
