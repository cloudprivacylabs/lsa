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

	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

// A CompiledGraph is a graph of compiled schemas
type CompiledGraph interface {
	GetCompiledSchema(string) *Layer
	PutCompiledSchema(*Context, string, *Layer) *Layer
	GetLayerNodes(string) []graph.Node
	GetGraph() graph.Graph
}

// DefaultCompiledGraph keeps compiled graphs in a map. Zero value of
// DefaultCompiledGraph is ready to use
type DefaultCompiledGraph struct {
	layers     map[string]*Layer
	g          graph.Graph
	layerNodes map[string][]graph.Node
}

func (d DefaultCompiledGraph) GetGraph() graph.Graph { return d.g }

// GetCompiledSchema returns a compiled schema for the reference if known
func (d DefaultCompiledGraph) GetCompiledSchema(ref string) *Layer {
	if d.layers == nil {
		return nil
	}
	return d.layers[ref]
}

// PutCompiledSchema adds the copy of the schema to the common graph
func (d *DefaultCompiledGraph) PutCompiledSchema(context *Context, ref string, layer *Layer) *Layer {
	if d.layers == nil {
		d.layers = make(map[string]*Layer)
		d.layerNodes = make(map[string][]graph.Node)
	}
	if d.g == nil {
		d.g = graph.NewOCGraph()
	}
	newLayer, nodeMap := layer.CloneInto(d.g)
	d.layers[ref] = newLayer
	nodes := make([]graph.Node, 0, len(nodeMap))
	for _, x := range nodeMap {
		nodes = append(nodes, x)
	}
	d.layerNodes[ref] = nodes
	return newLayer
}

func (d *DefaultCompiledGraph) GetLayerNodes(ref string) []graph.Node {
	return d.layerNodes[ref]
}

type Compiler struct {
	// Loader loads a layer using a strong reference.
	Loader func(ref string) (*Layer, error)
	// CGraph keeps the compiled interlinked schemas. If this is
	// initalized before compilation, then it is used during compilation
	// and new schemas are added to it. If it is left uninitialized,
	// compilation initializes it to default compiled graph
	CGraph CompiledGraph
}

type compilerContext struct {
	loadedSchemas map[string]*Layer
	blankNodeID   uint
	// This layer will not be cached
	doNotCache *Layer
}

func (c *compilerContext) blankNodeNamer(node graph.Node) {
	SetNodeID(node, fmt.Sprintf("_b:%d", c.blankNodeID))
	c.blankNodeID++
}

func (compiler Compiler) loadSchema(ctx *compilerContext, ref string) (*Layer, error) {
	var err error
	layer := ctx.loadedSchemas[ref]
	if layer != nil {
		return layer, nil
	}
	layer, err = compiler.Loader(ref)
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
	return compiler.compile(context, ctx, ref)
}

// CompileSchema compiles the loaded schema
func (compiler *Compiler) CompileSchema(context *Context, schema *Layer) (*Layer, error) {
	ctx := &compilerContext{
		loadedSchemas: map[string]*Layer{schema.GetID(): schema},
	}
	return compiler.compile(context, ctx, schema.GetID())
}

// RecompileSchema uses the cache to resolve the references of the
// schema, but does not put the schema back into the cache
func (compiler *Compiler) RecompileSchema(context *Context, schema *Layer) (*Layer, error) {
	ctx := &compilerContext{
		loadedSchemas: map[string]*Layer{schema.GetID(): schema},
		doNotCache:    schema,
	}
	return compiler.compile(context, ctx, schema.GetID())
}

func (compiler *Compiler) compile(context *Context, ctx *compilerContext, ref string) (*Layer, error) {
	context.GetLogger().Debug(map[string]interface{}{"mth": "compile", "ref": ref})
	if compiler.CGraph == nil {
		compiler.CGraph = &DefaultCompiledGraph{}
	}

	compiled := compiler.CGraph.GetCompiledSchema(ref)
	if compiled != nil {
		return compiled, nil
	}

	schema, err := compiler.compileRefs(context, ctx, ref)
	if err != nil {
		return nil, err
	}
	if err := compiler.resolveCompositions(context, schema.GetSchemaRootNode()); err != nil {
		return nil, err
	}
	if err := CompileTerms(schema); err != nil {
		return nil, err
	}
	return schema, nil
}

func (compiler *Compiler) compileRefs(context *Context, ctx *compilerContext, ref string) (*Layer, error) {
	context.GetLogger().Debug(map[string]interface{}{"mth": "compileRefs", "ref": ref})
	var err error
	// If compiled already, return the compiled node
	if c := compiler.CGraph.GetCompiledSchema(ref); c != nil {
		return c, nil
	}
	// Load the schema
	context.GetLogger().Debug(map[string]interface{}{"mth": "compileRefs", "ref": ref, "stage": "Loading"})
	schema, err := compiler.loadSchema(ctx, ref)
	if err != nil {
		return nil, err
	}
	if schema == nil {
		return nil, ErrNotFound(ref)
	}
	context.GetLogger().Debug(map[string]interface{}{"mth": "compileRefs", "ref": ref, "stage": "Loaded"})
	// Here, schema is loaded but not compiled
	// If this is the top-level, we set the target layer as this schema
	var compileRoot graph.Node
	schema.RenameBlankNodes(ctx.blankNodeNamer)
	compileRoot = schema.GetSchemaRootNode()
	if compileRoot == nil {
		return nil, ErrNotFound(ref)
	}
	// Record the schema ID in the entity root
	context.GetLogger().Debug(map[string]interface{}{"mth": "compileRefs", "entitySchema": schema.GetID()})
	compileRoot.SetProperty(EntitySchemaTerm, StringPropertyValue(schema.GetID()))

	// Resolve all references
	if schema != ctx.doNotCache {
		schema = compiler.CGraph.PutCompiledSchema(context, ref, schema)
	}
	schemaNodes := compiler.CGraph.GetLayerNodes(ref)
	context.GetLogger().Debug(map[string]interface{}{"mth": "compileRefs", "schemaNodes": len(schemaNodes)})
	if err := compiler.resolveReferences(context, ctx, schema, schemaNodes); err != nil {
		return nil, err
	}
	return schema, nil
}

func (compiler *Compiler) resolveReferences(context *Context, ctx *compilerContext, layer *Layer, nodes []graph.Node) error {
	// Collect all reference nodes
	references := make([]graph.Node, 0)
	for _, nd := range nodes {
		if nd.GetLabels().Has(AttributeTypeReference) {
			references = append(references, nd)
		}
	}
	// Resolve each reference
	for _, node := range references {
		if err := compiler.resolveReference(context, ctx, layer, node); err != nil {
			return err
		}
	}
	return nil
}

func (compiler *Compiler) resolveReference(context *Context, ctx *compilerContext, layer *Layer, node graph.Node) error {
	ref := AsPropertyValue(node.GetProperty(ReferenceTerm)).AsString()
	context.GetLogger().Debug(map[string]interface{}{"mth": "resolveReference", "ref": ref})
	// already compiled, or being compiled?
	compiledSchema := compiler.CGraph.GetCompiledSchema(ref)
	if compiledSchema == nil {
		var err error
		compiledSchema, err = compiler.compileRefs(context, ctx, ref)
		if err != nil {
			return err
		}
	}
	rootNode := compiledSchema.GetSchemaRootNode()
	// This is no longer a reference node
	types := node.GetLabels()
	types.Remove(AttributeTypeReference)
	types.Add(rootNode.GetLabels().Slice()...)
	node.SetLabels(types)
	// Compose the properties of the compiled root node with the referenced node
	if err := ComposeProperties(context, node, rootNode); err != nil {
		return err
	}
	node.SetProperty(ReferenceTerm, StringPropertyValue(ref))
	// Attach the node to all the children of the compiled node
	for edges := rootNode.GetEdges(graph.OutgoingEdge); edges.Next(); {
		edge := edges.Edge()
		CloneEdge(node, edge.GetTo(), edge, compiler.CGraph.GetGraph())
	}
	return nil
}

func (compiler Compiler) resolveCompositions(context *Context, root graph.Node) error {
	// Process all composition nodes
	completed := map[graph.Node]struct{}{}
	var err error
	ForEachAttributeNode(root, func(n graph.Node, _ []graph.Node) bool {
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

func (compiler Compiler) resolveComposition(context *Context, compositeNode graph.Node, completed map[graph.Node]struct{}) error {
	completed[compositeNode] = struct{}{}
	// At the end of this process, composite node will be converted into an object node
	for edges := compositeNode.GetEdgesWithLabel(graph.OutgoingEdge, AllOfTerm); edges.Next(); {
		allOfEdge := edges.Edge()
	top:
		component := allOfEdge.GetTo()
		switch {
		case component.GetLabels().Has(AttributeTypeObject):
			//  Input:
			//    compositeNode ---> component --> attributes
			//  Output:
			//    compositeNode --> attributes
			rmv := make([]graph.Edge, 0)
			for edges := component.GetEdges(graph.OutgoingEdge); edges.Next(); {
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
	IterateDescendants(layer.GetSchemaRootNode(), func(node graph.Node, _ []graph.Node) bool {
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
		for edges := node.GetEdges(graph.OutgoingEdge); edges.Next(); {
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
	}, func(edge graph.Edge, _ []graph.Node) EdgeFuncResult {
		return FollowEdgeResult
	}, false)
	return err
}
