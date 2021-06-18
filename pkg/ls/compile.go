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

type Compiler struct {
	// Resolver resolves an ID and returns a strong reference. If
	// resolver func is nil, references are directly sent to loader func
	Resolver func(string) (string, error)
	// Loader loads a layer using a strong reference. This cannot be nil
	Loader func(string) (*Layer, error)
}

type compilerContext struct {
	targetLayer   *Layer
	loadedSchemas map[string]*Layer
	compiled      map[string]LayerNode
}

func (compiler Compiler) loadSchema(ctx *compilerContext, ref string) (*Layer, error) {
	var err error
	if compiler.Resolver != nil {
		ref, err = compiler.Resolver(ref)
		if err != nil {
			return nil, err
		}
	}
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
		compiled:      make(map[string]LayerNode),
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
		compiled:      make(map[string]LayerNode),
	}
	_, err := compiler.compile(ctx, schema.GetID(), true)
	if err != nil {
		return nil, err
	}
	return ctx.targetLayer, nil
}

func (compiler Compiler) compile(ctx *compilerContext, ref string, topLevel bool) (LayerNode, error) {
	var err error
	// Resolve weak references
	if compiler.Resolver != nil {
		ref, err = compiler.Resolver(ref)
		if err != nil {
			return nil, err
		}
	}
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
	// If this is the top-leve, we set the target layer as this schema
	var compileRoot LayerNode
	if topLevel {
		ctx.targetLayer = schema.Clone()
		compileRoot = ctx.targetLayer.GetObjectInfoNode()
	} else {
		c := schema.Clone()
		compileRoot = c.GetObjectInfoNode()
		ctx.targetLayer.Import(c.Graph)
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
		compiler.compileTerms(ctx.targetLayer)
	}
	return compileRoot, nil
}

func (compiler Compiler) compileTerms(layer *Layer) error {
	for nodes := layer.AllNodes(); nodes.HasNext(); {
		node := nodes.Next().(LayerNode)
		// Compile all non-attribute nodes
		if !node.IsAttributeNode() {
			if err := GetNodeCompiler(node.GetID()).CompileNode(node); err != nil {
				return err
			}
		}
		for k, v := range node.GetPropertyMap() {
			result, err := GetTermCompiler(k).CompileTerm(k, v)
			if err != nil {
				return err
			}
			if result != nil {
				node.GetCompiledDataMap()[k] = result
			}
			for edges := node.AllOutgoingEdges(); edges.HasNext(); {
				edge := edges.Next().(LayerEdge)
				if err := GetEdgeCompiler(edge.GetLabel()).CompileEdge(edge); err != nil {
					return err
				}
				for k, v := range edge.GetPropertyMap() {
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

func (compiler Compiler) resolveReferences(ctx *compilerContext, root LayerNode) error {
	// Collect all reference nodes
	references := make([]LayerNode, 0)
	ForEachAttributeNode(root, func(n LayerNode) bool {
		if n.HasType(AttributeTypes.Reference) {
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

func (compiler Compiler) resolveReference(ctx *compilerContext, node LayerNode) error {
	properties := node.GetPropertyMap()
	ref := properties[LayerTerms.Reference].(string)
	// already compiled, or being compiled?
	compiled := ctx.compiled[ref]
	if compiled == nil {
		var err error
		compiled, err = compiler.compile(ctx, ref, false)
		if err != nil {
			return err
		}
	}
	// Here, compiled is already imported into the target graph

	// This is no longer a reference node
	node.RemoveTypes(AttributeTypes.Reference)
	node.AddTypes(compiled.GetTypes()...)
	// Compose the properties of the compiled root node with the referenced node
	if err := ComposeProperties(properties, node.GetPropertyMap()); err != nil {
		return err
	}
	// Attach the node to all the children of the compiled node
	for edges := compiled.AllOutgoingEdges(); edges.HasNext(); {
		edge := edges.Next().(LayerEdge)
		ctx.targetLayer.AddEdge(node, edge.To(), edge.Clone())
	}
	return nil
}

func (compiler Compiler) resolveCompositions(root LayerNode) error {
	// Process all composition nodes
	completed := map[LayerNode]struct{}{}
	var err error
	ForEachAttributeNode(root, func(n LayerNode) bool {
		if n.HasType(AttributeTypes.Composite) {
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

func (compiler Compiler) resolveComposition(compositeNode LayerNode, completed map[LayerNode]struct{}) error {
	completed[compositeNode] = struct{}{}
	// At the end of this process, composite node will be converted into an object node
	for edges := compositeNode.AllOutgoingEdgesWithLabel(LayerTerms.AllOf); edges.HasNext(); {
		allOfEdge := edges.Next().(LayerEdge)
	top:
		component := allOfEdge.To().(LayerNode)
		switch {
		case component.HasType(AttributeTypes.Object):
			//  Input:
			//    compositeNode ---> component --> attributes
			//  Output:
			//    compositeNode --> attributes
			for edges := component.AllOutgoingEdges(); edges.HasNext(); {
				edge := edges.Next()
				edge.SetFrom(compositeNode)
			}
			// Copy all properties of the component node to the composite node
			if err := ComposeProperties(compositeNode.GetPropertyMap(), component.GetPropertyMap()); err != nil {
				return err
			}
			// Copy all types
			compositeNode.AddTypes(component.GetTypes()...)
			// Copy compiled items
			copyCompiled(compositeNode.GetCompiledDataMap(), component.GetCompiledDataMap())

		case component.HasType(AttributeTypes.Value) ||
			component.HasType(AttributeTypes.Array) ||
			component.HasType(AttributeTypes.Polymorphic):
			// This node becomes an attribute of the main node.
			allOfEdge.SetLabel(LayerTerms.AttributeList)

		case component.HasType(AttributeTypes.Composite):
			if err := compiler.resolveComposition(component, completed); err != nil {
				return err
			}
			goto top
		default:
			return ErrInvalidComposition
		}
	}
	// Convert the node to an object
	compositeNode.RemoveTypes(AttributeTypes.Composite)
	compositeNode.AddTypes(AttributeTypes.Object)
	return nil
}
