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

	"github.com/bserdar/digraph"
)

// A CompiledGraph is a graph of compiled schemas
type CompiledGraph interface {
	GetCompiledSchema(string) *Layer
	PutCompiledSchema(string, *Layer)
}

// DefaultCompiledGraph keeps compiled graphs in a map. Zero value of
// DefaultCompiledGraph is ready to use
type DefaultCompiledGraph struct {
	layers map[string]*Layer
}

// GetCompiledSchema returns a compiled schema for the reference if known
func (d DefaultCompiledGraph) GetCompiledSchema(ref string) *Layer {
	if d.layers == nil {
		return nil
	}
	return d.layers[ref]
}

// PutCompiledSchema adds the schema
func (d *DefaultCompiledGraph) PutCompiledSchema(ref string, layer *Layer) {
	if d.layers == nil {
		d.layers = make(map[string]*Layer)
	}
	d.layers[ref] = layer
}

type Compiler struct {
	// Loader loads a layer using a strong reference.
	Loader func(string) (*Layer, error)
	// CGraph keeps the compiled interlinked schemas. If this is
	// initalized before compilation, then it is used during compilation
	// and new schemas are added to it. If it is left uninitialized,
	// compilation initializes it to default compiled graph
	CGraph CompiledGraph
}

type compilerContext struct {
	loadedSchemas map[string]*Layer
	blankNodeID   uint
}

func (c *compilerContext) blankNodeNamer(node Node) {
	node.SetID(fmt.Sprintf("_b:%d", c.blankNodeID))
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
func (compiler *Compiler) Compile(ref string) (*Layer, error) {
	ctx := &compilerContext{
		loadedSchemas: make(map[string]*Layer),
	}
	return compiler.compile(ctx, ref)
}

// CompileSchema compiles the loaded schema
func (compiler *Compiler) CompileSchema(schema *Layer) (*Layer, error) {
	ctx := &compilerContext{
		loadedSchemas: map[string]*Layer{schema.GetID(): schema},
	}
	return compiler.compile(ctx, schema.GetID())
}

func (compiler *Compiler) compile(ctx *compilerContext, ref string) (*Layer, error) {
	if compiler.CGraph == nil {
		compiler.CGraph = &DefaultCompiledGraph{}
	}

	compiled := compiler.CGraph.GetCompiledSchema(ref)
	if compiled != nil {
		return compiled, nil
	}

	schema, err := compiler.compileRefs(ctx, ref)
	if err != nil {
		return nil, err
	}
	schema.ResetIndex()
	if err := compiler.resolveCompositions(schema.GetSchemaRootNode()); err != nil {
		return nil, err
	}
	if err := CompileTerms(schema); err != nil {
		return nil, err
	}
	return schema, nil
}

func (compiler *Compiler) compileRefs(ctx *compilerContext, ref string) (*Layer, error) {
	var err error
	// If compiled already, return the compiled node
	if c := compiler.CGraph.GetCompiledSchema(ref); c != nil {
		return c, nil
	}
	// Load the schema
	schema, err := compiler.loadSchema(ctx, ref)
	if err != nil {
		return nil, err
	}
	if schema == nil {
		return nil, ErrNotFound(ref)
	}

	// Here, schema is loaded but not compiled
	// If this is the top-level, we set the target layer as this schema
	var compileRoot Node
	schema.RenameBlankNodes(ctx.blankNodeNamer)
	schema.ResetIndex()
	compileRoot = schema.GetSchemaRootNode()
	if compileRoot == nil {
		return nil, ErrNotFound(ref)
	}
	// Resolve all references
	compiler.CGraph.PutCompiledSchema(ref, schema)
	if err := compiler.resolveReferences(ctx, schema.GetIndex().NodesSlice()); err != nil {
		return nil, err
	}
	return schema, nil
}

func (compiler *Compiler) resolveReferences(ctx *compilerContext, nodes []digraph.Node) error {
	// Collect all reference nodes
	references := make([]Node, 0)
	for _, n := range nodes {
		nd := n.(Node)
		if nd.GetTypes().Has(AttributeTypes.Reference) {
			references = append(references, nd)
		}
	}
	// Resolve each reference
	for _, node := range references {
		if err := compiler.resolveReference(ctx, node); err != nil {
			return err
		}
	}
	return nil
}

func (compiler *Compiler) resolveReference(ctx *compilerContext, node Node) error {
	properties := node.GetProperties()
	ref := properties[LayerTerms.Reference].AsString()
	delete(properties, LayerTerms.Reference)
	// already compiled, or being compiled?
	compiledSchema := compiler.CGraph.GetCompiledSchema(ref)
	if compiledSchema == nil {
		var err error
		compiledSchema, err = compiler.compileRefs(ctx, ref)
		if err != nil {
			return err
		}
	}
	rootNode := compiledSchema.GetSchemaRootNode()
	// This is no longer a reference node
	node.GetTypes().Remove(AttributeTypes.Reference)
	node.GetTypes().Add(rootNode.GetTypes().Slice()...)
	// Compose the properties of the compiled root node with the referenced node
	if err := ComposeProperties(properties, rootNode.GetProperties()); err != nil {
		return err
	}
	// Attach the node to all the children of the compiled node
	for edges := rootNode.Out(); edges.HasNext(); {
		edge := edges.Next().(Edge)
		digraph.Connect(node, edge.GetTo(), edge.Clone())
	}
	return nil
}

func (compiler Compiler) resolveCompositions(root Node) error {
	// Process all composition nodes
	completed := map[Node]struct{}{}
	var err error
	ForEachAttributeNode(root, func(n Node, _ []Node) bool {
		if n.GetTypes().Has(AttributeTypes.Composite) {
			if _, processed := completed[n]; !processed {
				if x := compiler.resolveComposition(n, completed); x != nil {
					err = x
					return false
				}
			}
		}
		return true
	})
	return err
}

func copyCompiled(target, source map[interface{}]interface{}) {
	for k, v := range source {
		target[k] = v
	}
}

func (compiler Compiler) resolveComposition(compositeNode Node, completed map[Node]struct{}) error {
	completed[compositeNode] = struct{}{}
	// At the end of this process, composite node will be converted into an object node
	for edges := compositeNode.OutWith(LayerTerms.AllOf); edges.HasNext(); {
		allOfEdge := edges.Next().(Edge)
	top:
		component := allOfEdge.GetTo().(Node)
		switch {
		case component.GetTypes().Has(AttributeTypes.Object):
			//  Input:
			//    compositeNode ---> component --> attributes
			//  Output:
			//    compositeNode --> attributes
			rmv := make([]Edge, 0)
			for edges := component.Out(); edges.HasNext(); {
				edge := edges.Next().(Edge)
				digraph.Connect(compositeNode, edge.GetTo(), edge.Clone())
				rmv = append(rmv, edge)
			}
			for _, e := range rmv {
				e.Disconnect()
			}
			// Copy all properties of the component node to the composite node
			if err := ComposeProperties(compositeNode.GetProperties(), component.GetProperties()); err != nil {
				return err
			}
			// Copy all types
			compositeNode.GetTypes().Add(component.GetTypes().Slice()...)
			// Copy compiled items
			copyCompiled(compositeNode.GetCompiledDataMap(), component.GetCompiledDataMap())

		case component.GetTypes().Has(AttributeTypes.Value) ||
			component.GetTypes().Has(AttributeTypes.Array) ||
			component.GetTypes().Has(AttributeTypes.Polymorphic):
			// This node becomes an attribute of the main node.
			newEdge := CloneWithLabel(allOfEdge, LayerTerms.AttributeList)
			allOfEdge.Disconnect()
			digraph.Connect(compositeNode, component, newEdge)

		case component.GetTypes().Has(AttributeTypes.Composite):
			if err := compiler.resolveComposition(component, completed); err != nil {
				return err
			}
			goto top
		default:
			return ErrInvalidComposition
		}
	}
	// Convert the node to an object
	compositeNode.GetTypes().Remove(AttributeTypes.Composite)
	compositeNode.GetTypes().Add(AttributeTypes.Object)
	return nil
}

// CompileTerms compiles all node and edge terms of the layer
func CompileTerms(layer *Layer) error {
	for _, n := range layer.GetIndex().NodesSlice() {
		node := n.(Node)
		// Compile all non-attribute nodes
		if !IsAttributeNode(node) {
			if err := GetNodeCompiler(node.GetID()).CompileNode(node); err != nil {
				return err
			}
		}
		for k, v := range node.GetProperties() {
			err := GetTermCompiler(k).CompileTerm(node, k, v)
			if err != nil {
				return err
			}
			for edges := node.Out(); edges.HasNext(); {
				edge := edges.Next().(Edge)
				if err := GetEdgeCompiler(edge.GetLabelStr()).CompileEdge(edge); err != nil {
					return err
				}
				for k, v := range edge.GetProperties() {
					err := GetTermCompiler(k).CompileTerm(edge, k, v)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
