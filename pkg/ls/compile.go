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

	"github.com/cloudprivacylabs/lpg"
)

// A CompiledGraph is a graph of compiled schemas
type CompiledGraph interface {
	GetCompiledSchema(string) *Layer
	PutCompiledSchema(*Context, string, *Layer) (*Layer, error)
	GetGraph() *lpg.Graph
}

// DefaultCompiledGraph keeps compiled graphs in a map. Zero value of
// DefaultCompiledGraph is ready to use
type DefaultCompiledGraph struct {
	layers map[string]*Layer
	g      *lpg.Graph
	// schemaNodeMap contains the map of the source layers -> compiled graph nodes
	schemaNodeMap map[*lpg.Node]*lpg.Node
}

func (d DefaultCompiledGraph) GetGraph() *lpg.Graph { return d.g }

// GetCompiledSchema returns a compiled schema for the reference if known
func (d DefaultCompiledGraph) GetCompiledSchema(ref string) *Layer {
	if d.layers == nil {
		return nil
	}
	return d.layers[ref]
}

func (d *DefaultCompiledGraph) copyNode(source *lpg.Node) *lpg.Node {
	newNode := lpg.CopyNode(source, d.g, ClonePropertyValueFunc)
	SetNodeID(newNode, GetNodeID(source))
	return newNode
}

func (d *DefaultCompiledGraph) copyEdge(from, to *lpg.Node, edge *lpg.Edge) {
	d.g.NewEdge(from, to, edge.GetLabel(), CloneProperties(edge))
}

// PutCompiledSchema adds the copy of the schema to the common graph
func (d *DefaultCompiledGraph) PutCompiledSchema(context *Context, ref string, layer *Layer) (*Layer, error) {
	if d.layers == nil {
		d.layers = make(map[string]*Layer)
		d.schemaNodeMap = make(map[*lpg.Node]*lpg.Node)
	}
	if d.g == nil {
		d.g = NewLayerGraph()
	}

	// This algorithm relies on unique attribute IDs
	newLayer := NewLayerInGraph(d.g)
	newLayer.SetID(layer.GetID())
	newLayer.SetLayerType(SchemaTerm)
	// Copy the root node of the layer into the compiled graph.
	layerRoot := layer.GetSchemaRootNode()
	nodeSlice := layer.NodeSlice()
	// attributeMap keeps track of copied attribute nodes. Key belongs
	// to layer, value belongs to d.g
	attributeMap := make(map[*lpg.Node]*lpg.Node, len(nodeSlice))

	newRoot := d.copyNode(layerRoot)
	d.g.NewEdge(newLayer.GetLayerRootNode(), newRoot, LayerRootTerm, nil)
	attributeMap[layerRoot] = newRoot
	// newAttributes contains only those attributes that are copied. The
	// key belongs to layer
	newAttributes := make(map[*lpg.Node]struct{}, len(nodeSlice))

	// Copy the attributes in this layer
	type nodeCopy struct {
		layerNode    *lpg.Node
		newAttribute bool
		compiledNode *lpg.Node
	}
	layerNodes := make([]nodeCopy, 0, len(nodeSlice))
	for _, node := range nodeSlice {
		nc := nodeCopy{
			layerNode: node,
		}
		attrID := GetNodeID(node)
		existingAttr := newLayer.GetAttributeByID(attrID)
		if existingAttr != nil && existingAttr != newRoot {
			return nil, ErrInvalidSchema(fmt.Sprintf("Node %s is duplicated by %s. If same schema appears under a different name, change namespaces", attrID, ref))
		}
		if existingAttr == nil {
			existingAttr = d.copyNode(node)
			nc.newAttribute = true
			newAttributes[node] = struct{}{}
			d.schemaNodeMap[node] = existingAttr
		}
		nc.compiledNode = existingAttr
		attributeMap[node] = existingAttr
		layerNodes = append(layerNodes, nc)
	}
	// Iterate all nodes again. This time, link them
	for _, nc := range layerNodes {
		for edges := nc.layerNode.GetEdges(lpg.OutgoingEdge); edges.Next(); {
			layerEdge := edges.Edge()
			if IsAttributeNode(layerEdge.GetTo()) {
				// If either the from or to is in newAttributes, then we need to add this edge
				_, toNew := newAttributes[layerEdge.GetTo()]
				if nc.newAttribute || toNew {
					d.copyEdge(nc.compiledNode, attributeMap[layerEdge.GetTo()], layerEdge)
				}
			} else {
				// Link to a non-attribute node
				// If this is a new node, we have to copy this subtree
				if nc.newAttribute {
					lpg.CopySubgraph(layerEdge.GetTo(), d.g, ClonePropertyValueFunc, d.schemaNodeMap)
					d.g.NewEdge(nc.compiledNode, d.schemaNodeMap[layerEdge.GetTo()], layerEdge.GetLabel(), nil)
				}
			}
		}
	}

	d.layers[ref] = newLayer
	return newLayer, nil
}

// SchemaLoader interface defines the LoadSchema method that loads schemas by reference
type SchemaLoader interface {
	LoadSchema(ref string) (*Layer, error)
}

// SchemaLoaderFunc is the function type that load schemas. It also
// implements SchemaLoader interface
type SchemaLoaderFunc func(string) (*Layer, error)

func (s SchemaLoaderFunc) LoadSchema(ref string) (*Layer, error) { return s(ref) }

type Compiler struct {
	// Loader loads a layer using a strong reference.
	Loader SchemaLoader
	// CGraph keeps the compiled interlinked schemas. If this is
	// initialized before compilation, then it is used during compilation
	// and new schemas are added to it. If it is left uninitialized,
	// compilation initializes it to default compiled graph
	CGraph CompiledGraph
}

type compilerContext struct {
	loadedSchemas map[string]*Layer
	blankNodeID   uint
}

// IsCompilationArtifact returns true if the edge is a compilation artifact
func IsCompilationArtifact(edge *lpg.Edge) bool {
	return AsPropertyValue(edge.GetProperty("compilationArtifact")).AsString() == "true"
}

func (c *compilerContext) blankNodeNamer(node *lpg.Node) {
	SetNodeID(node, fmt.Sprintf("_b:%d", c.blankNodeID))
	c.blankNodeID++
}

func (compiler Compiler) loadSchema(ctx *compilerContext, ref string) (*Layer, error) {
	var err error
	layer := ctx.loadedSchemas[ref]
	if layer != nil {
		return layer, nil
	}
	layer, err = compiler.Loader.LoadSchema(ref)
	if err != nil {
		return nil, err
	}
	ctx.loadedSchemas[ref] = layer
	return layer, nil
}

// Compile compiles the schema by resolving all references and
// building all compositions.
func (compiler *Compiler) Compile(context *Context, ref string) (*Layer, error) {
	ctx := &compilerContext{
		loadedSchemas: make(map[string]*Layer),
	}
	layer, err := compiler.compile(context, ctx, ref)
	if err != nil {
		return nil, err
	}
	ProcessLabeledAs(layer.Graph)
	return layer, nil
}

// CompileSchema compiles the loaded schema
func (compiler *Compiler) CompileSchema(context *Context, schema *Layer) (*Layer, error) {
	ctx := &compilerContext{
		loadedSchemas: map[string]*Layer{schema.GetID(): schema},
	}
	layer, err := compiler.compile(context, ctx, schema.GetID())
	if err != nil {
		return nil, err
	}
	ProcessLabeledAs(layer.Graph)
	return layer, nil
}

func (compiler *Compiler) compile(context *Context, ctx *compilerContext, ref string) (*Layer, error) {
	context.GetLogger().Debug(map[string]interface{}{"mth": "compile", "ref": ref})
	if compiler.CGraph == nil {
		compiler.CGraph = &DefaultCompiledGraph{}
	}

	compiled := compiler.CGraph.GetCompiledSchema(ref)
	if compiled != nil {
		context.GetLogger().Debug(map[string]interface{}{"mth": "compile", "ref": ref, "stage": "Already compiled"})
		return compiled, nil
	}
	compiled, err := compiler.loadSchema(ctx, ref)
	if err != nil {
		return nil, err
	}
	compiled, err = compiler.CGraph.PutCompiledSchema(context, ref, compiled)
	if err != nil {
		return nil, err
	}
	compiled.GetSchemaRootNode().SetProperty(EntitySchemaTerm, StringPropertyValue(EntitySchemaTerm, compiled.GetID()))

	if err := compiler.compileReferences(context, ctx); err != nil {
		return nil, err
	}
	if err := compiler.resolveCompositions(context, compiled.GetSchemaRootNode()); err != nil {
		return nil, err
	}
	if err := CompileTerms(compiled); err != nil {
		return nil, err
	}
	return compiled, nil
}

func (compiler *Compiler) compileReferences(context *Context, ctx *compilerContext) error {
	context.GetLogger().Debug(map[string]interface{}{"mth": "compileReferences"})
	// Process until there are reference nodes left
	refset := lpg.NewStringSet(AttributeTypeReference)
	for {
		refNodes := compiler.CGraph.GetGraph().GetNodesWithAllLabels(refset)
		if !refNodes.Next() {
			break
		}
		context.GetLogger().Debug(map[string]interface{}{"mth": "compileReferences", "nReferences": refNodes.MaxSize(), "nNodes": compiler.CGraph.GetGraph().NumNodes()})

		refNode := refNodes.Node()
		ref := AsPropertyValue(refNode.GetProperty(ReferenceTerm)).AsString()
		context.GetLogger().Debug(map[string]interface{}{"mth": "compileReferences", "ref": ref})
		// already loaded and added to the graph?
		compiledSchema := compiler.CGraph.GetCompiledSchema(ref)
		if compiledSchema != nil {
			context.GetLogger().Debug(map[string]interface{}{"mth": "compileReferences", "ref": ref, "stage": "already compiled, linking"})
			if err := compiler.linkReference(context, refNode, compiledSchema, ref); err != nil {
				return err
			}
			continue
		}
		// Schema is not yet loaded
		context.GetLogger().Debug(map[string]interface{}{"mth": "compileReferences", "ref": ref, "stage": "Loading"})

		schema, err := compiler.loadSchema(ctx, ref)
		if err != nil {
			return err
		}
		if schema == nil {
			return ErrNotFound(ref)
		}
		context.GetLogger().Debug(map[string]interface{}{"mth": "compileReferences", "ref": ref, "stage": "Loaded"})
		compileRoot := schema.GetSchemaRootNode()
		if compileRoot == nil {
			return ErrNotFound(ref)
		}
		// Record the schema ID in the entity root
		context.GetLogger().Debug(map[string]interface{}{"mth": "compileReferences", "entitySchema": schema.GetID()})

		newLayer, err := compiler.CGraph.PutCompiledSchema(context, ref, schema)
		if err != nil {
			return err
		}
		if err := compiler.linkReference(context, refNode, newLayer, ref); err != nil {
			return err
		}
	}
	return nil
}

func (compiler *Compiler) linkReference(context *Context, refNode *lpg.Node, schema *Layer, ref string) error {
	// This is no longer a reference node
	linkTo := schema.GetSchemaRootNode()
	types := refNode.GetLabels()
	types.Remove(AttributeTypeReference)
	types.AddSet(linkTo.GetLabels())
	refNode.SetLabels(types)
	// Compose the properties of the compiled root node with the referenced node
	if err := ComposeProperties(context, refNode, linkTo); err != nil {
		return err
	}
	refNode.SetProperty(ReferenceTerm, StringPropertyValue(ReferenceTerm, ref))
	// Attach the node to all the children of the compiled node
	for edges := linkTo.GetEdges(lpg.OutgoingEdge); edges.Next(); {
		edge := edges.Edge()
		CloneEdge(refNode, edge.GetTo(), edge, compiler.CGraph.GetGraph())
		// Mark all edges that connect the original schema node as
		// "compilationArtifact", so we can trace back the schema nodes
		// correctly
		edge.SetProperty("compilationArtifact", StringPropertyValue("compilationArtifact", "true"))
	}
	refNode.SetProperty(EntitySchemaTerm, StringPropertyValue(EntitySchemaTerm, schema.GetID()))
	return nil
}

func (compiler Compiler) resolveCompositions(context *Context, root *lpg.Node) error {
	// Process all composition nodes
	completed := map[*lpg.Node]struct{}{}
	var err error
	ForEachAttributeNode(root, func(n *lpg.Node, _ []*lpg.Node) bool {
		if n.GetLabels().Has(AttributeTypeComposite) {
			if _, processed := completed[n]; !processed {
				if x := compiler.resolveComposition(context, n, completed); x != nil {
					err = x
					return false
				}
			}
		}
		return true
	})
	return err
}

func (compiler Compiler) resolveComposition(context *Context, compositeNode *lpg.Node, completed map[*lpg.Node]struct{}) error {
	completed[compositeNode] = struct{}{}
	// At the end of this process, composite node will be converted into an object node
	for edges := compositeNode.GetEdgesWithLabel(lpg.OutgoingEdge, AllOfTerm); edges.Next(); {
		allOfEdge := edges.Edge()
	top:
		component := allOfEdge.GetTo()
		switch {
		case component.GetLabels().Has(AttributeTypeObject):
			//  Input:
			//    compositeNode ---> component --> attributes
			//  Output:
			//    compositeNode --> attributes
			rmv := make([]*lpg.Edge, 0)
			for edges := component.GetEdges(lpg.OutgoingEdge); edges.Next(); {
				edge := edges.Edge()
				CloneEdge(compositeNode, edge.GetTo(), edge, compiler.CGraph.GetGraph())
				rmv = append(rmv, edge)
			}
			for _, e := range rmv {
				e.Remove()
			}
			// Copy all properties of the component node to the composite node
			if err := ComposeProperties(context, compositeNode, component); err != nil {
				return err
			}
			// Copy all types
			types := compositeNode.GetLabels()
			types.AddSet(component.GetLabels())
			compositeNode.SetLabels(types)
			// Copy non-property items
			component.ForEachProperty(func(key string, value interface{}) bool {
				if _, ok := value.(*PropertyValue); !ok {
					compositeNode.SetProperty(key, value)
				}
				return true
			})

		case component.GetLabels().Has(AttributeTypeValue) ||
			component.GetLabels().Has(AttributeTypeArray) ||
			component.GetLabels().Has(AttributeTypePolymorphic):
			// This node becomes an attribute of the main node.
			allOfEdge.Remove()
			compiler.CGraph.GetGraph().NewEdge(compositeNode, component, ObjectAttributeListTerm, nil)

		case component.GetLabels().Has(AttributeTypeComposite):
			if err := compiler.resolveComposition(context, component, completed); err != nil {
				return err
			}
			goto top
		default:
			return ErrInvalidComposition
		}
	}
	// Convert the node to an object
	types := compositeNode.GetLabels()
	types.Remove(AttributeTypeComposite)
	types.Add(AttributeTypeObject)
	compositeNode.SetLabels(types)
	return nil
}

// CompileTerms compiles all node and edge terms of the layer
func CompileTerms(layer *Layer) error {
	var err error
	IterateDescendants(layer.GetSchemaRootNode(), func(node *lpg.Node) bool {
		// Compile all non-attribute nodes
		if !IsAttributeNode(node) {
			if err = GetNodeCompiler(GetNodeID(node)).CompileNode(layer, node); err != nil {
				return false
			}
		}
		node.ForEachProperty(func(k string, val interface{}) bool {
			if v, ok := val.(*PropertyValue); ok {
				err = GetTermCompiler(k).CompileTerm(node, k, v)
				if err != nil {
					return false
				}
			}
			return true
		})
		if err != nil {
			return false
		}
		for edges := node.GetEdges(lpg.OutgoingEdge); edges.Next(); {
			edge := edges.Edge()
			if err = GetEdgeCompiler(edge.GetLabel()).CompileEdge(layer, edge); err != nil {
				return false
			}
			edge.ForEachProperty(func(k string, val interface{}) bool {
				if v, ok := val.(*PropertyValue); ok {
					err = GetTermCompiler(k).CompileTerm(edge, k, v)
					if err != nil {
						return false
					}
				}
				return true
			})
		}
		if err != nil {
			return false
		}
		return true
	}, func(edge *lpg.Edge) EdgeFuncResult {
		return FollowEdgeResult
	}, false)
	return err
}
