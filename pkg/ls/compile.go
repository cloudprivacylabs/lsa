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

type Compiler struct {
	// Loader loads a layer using a strong reference.
	Loader func(string) (*Layer, error)
}

type compilerContext struct {
	targetLayer   *Layer
	loadedSchemas map[string]*Layer
	compiled      map[string]Node
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
func (compiler Compiler) Compile(ref string) (*Layer, error) {
	ctx := &compilerContext{
		loadedSchemas: make(map[string]*Layer),
		compiled:      make(map[string]Node),
	}
	_, err := compiler.compile(ctx, ref, true)
	if err != nil {
		return nil, err
	}
	return ctx.targetLayer, nil
}

// CompileSchema compiles the loaded schema
func (compiler Compiler) CompileSchema(schema *Layer) (*Layer, error) {
	ctx := &compilerContext{
		loadedSchemas: map[string]*Layer{schema.GetID(): schema},
		compiled:      make(map[string]Node),
	}
	_, err := compiler.compile(ctx, schema.GetID(), true)
	if err != nil {
		return nil, err
	}
	return ctx.targetLayer, nil
}

func (compiler Compiler) compile(ctx *compilerContext, ref string, topLevel bool) (Node, error) {
	var err error
	// If compiled already, return the compiled node
	if c := ctx.compiled[ref]; c != nil {
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

	// Here, scheme is loaded but not compiled
	// If this is the top-level, we set the target layer as this schema
	var compileRoot Node
	schema.RenameBlankNodes(ctx.blankNodeNamer)
	if topLevel {
		ctx.targetLayer = schema
		compileRoot = ctx.targetLayer.GetSchemaRootNode()
	} else {
		compileRoot = schema.GetSchemaRootNode()
	}
	if compileRoot == nil {
		return nil, nil
	}
	ctx.compiled[ref] = compileRoot
	if err := compiler.resolveReferences(ctx, compileRoot); err != nil {
		return nil, err
	}
	if err := compiler.resolveCompositions(compileRoot); err != nil {
		return nil, err
	}
	if topLevel {
		if err := compiler.compileTerms(ctx.targetLayer); err != nil {
			return nil, err
		}
	}
	return compileRoot, nil
}

func (compiler Compiler) compileTerms(layer *Layer) error {
	for nodes := layer.GetAllNodes(); nodes.HasNext(); {
		node := nodes.Next().(Node)
		// Compile all non-attribute nodes
		if !IsAttributeNode(node) {
			if err := GetNodeCompiler(node.GetID()).CompileNode(node); err != nil {
				return err
			}
		}
		for k, v := range node.GetProperties() {
			result, err := GetTermCompiler(k).CompileTerm(k, v)
			if err != nil {
				return err
			}
			if result != nil {
				node.GetCompiledDataMap()[k] = result
			}
			for edges := node.Out(); edges.HasNext(); {
				edge := edges.Next().(Edge)
				if err := GetEdgeCompiler(edge.GetLabelStr()).CompileEdge(edge); err != nil {
					return err
				}
				for k, v := range edge.GetProperties() {
					result, err := GetTermCompiler(k).CompileTerm(k, v)
					if err != nil {
						return err
					}
					if result != nil {
						edge.GetCompiledDataMap()[k] = result
					}
				}
			}
		}
	}
	return nil
}

func (compiler Compiler) resolveReferences(ctx *compilerContext, root Node) error {
	// Collect all reference nodes
	references := make([]Node, 0)
	ForEachAttributeNode(root, func(n Node, _ []Node) bool {
		if n.GetTypes().Has(AttributeTypes.Reference) {
			references = append(references, n)
		}
		return true
	})
	// Resolve each reference
	for _, node := range references {
		if err := compiler.resolveReference(ctx, node); err != nil {
			return err
		}
	}
	return nil
}

func (compiler Compiler) resolveReference(ctx *compilerContext, node Node) error {
	properties := node.GetProperties()
	ref := properties[LayerTerms.Reference].AsString()
	delete(properties, LayerTerms.Reference)
	// already compiled, or being compiled?
	compiled := ctx.compiled[ref]
	if compiled == nil {
		var err error
		compiled, err = compiler.compile(ctx, ref, false)
		if err != nil {
			return err
		}
	}
	// This is no longer a reference node
	node.GetTypes().Remove(AttributeTypes.Reference)
	node.GetTypes().Add(compiled.GetTypes().Slice()...)
	// Compose the properties of the compiled root node with the referenced node
	if err := ComposeProperties(properties, node.GetProperties()); err != nil {
		return err
	}
	// Attach the node to all the children of the compiled node
	for edges := compiled.Out(); edges.HasNext(); {
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
