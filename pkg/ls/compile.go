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
	// Resolver resolves an ID and returns a strong reference
	Resolver func(string) (string, error)
	// Loader loads a layer using strong reference
	Loader func(string) (*Layer, error)

	compiledSchemas map[string]*Layer
}

// Compile compiles the schema by resolving all references and
// computing all compositions. Compilation process directly modifies
// the schema
func (compiler *Compiler) Compile(ref string) (*Layer, error) {
	if compiler.compiledSchemas == nil {
		compiler.compiledSchemas = make(map[string]*Layer)
	}
	id, err := compiler.Resolver(ref)
	if err != nil {
		return nil, err
	}
	ret := compiler.compiledSchemas[id]
	if ret != nil {
		return ret, nil
	}
	schema, err := compiler.Loader(id)
	if err != nil {
		return nil, err
	}
	if schema == nil {
		return nil, ErrNotFound(ref)
	}
	schema = schema.Clone()
	// Put the compiled schema here, so if there are loops, we can refer to the
	// same object
	compiler.compiledSchemas[id] = schema
	if err := compiler.resolveReferences(schema); err != nil {
		return nil, err
	}
	if err := compiler.resolveCompositions(schema); err != nil {
		return nil, err
	}
	compiler.compileTerms(schema)
	return schema, nil
}

func (compiler *Compiler) compileTerms(layer *Layer) error {
	for nodes := layer.AllNodes(); nodes.HasNext(); {
		node := nodes.Next().(*SchemaNode)
		// Compile all non-attribute nodes
		if !node.IsAttributeNode() {
			if err := GetNodeCompiler(node.GetID()).CompileNode(node); err != nil {
				return err
			}
		}
		for k, v := range node.Properties {
			result, err := GetTermCompiler(k).CompileTerm(k, v)
			if err != nil {
				return err
			}
			if result != nil {
				node.Compiled[k] = result
			}
			for edges := node.AllOutgoingEdges(); edges.HasNext(); {
				edge := edges.Next().(*SchemaEdge)
				if err := GetEdgeCompiler(edge.GetLabel()).CompileEdge(edge); err != nil {
					return err
				}
				for k, v := range edge.Properties {
					result, err := GetTermCompiler(k).CompileTerm(k, v)
					if err != nil {
						return err
					}
					if result != nil {
						edge.Compiled[k] = result
					}
				}
			}
		}
	}
	return nil
}

func (compiler *Compiler) resolveReferences(layer *Layer) error {
	toRemove := make([]*SchemaNode, 0)
	for nodes := layer.AllNodes(); nodes.HasNext(); {
		node := nodes.Next().(*SchemaNode)
		if node.HasType(AttributeTypes.Reference) {
			ref := node.Properties[LayerTerms.Reference]
			referencedLayer, err := compiler.Compile(ref.(string))
			if err != nil {
				return err
			}

			// Attach all incoming edges to the layer's root
			for incoming := node.AllIncomingEdges(); incoming.HasNext(); {
				edge := incoming.Next().(*SchemaEdge)
				layer.AddEdge(edge.From(), referencedLayer.GetRoot(), edge.Clone())
			}
			// Reattach all outgoing edges
			for outgoing := node.AllOutgoingEdges(); outgoing.HasNext(); {
				edge := outgoing.Next().(*SchemaEdge)
				layer.AddEdge(referencedLayer.GetRoot(), edge.To(), edge.Clone())
			}
			// Copy over properties
			for k, v := range node.Properties {
				if k != LayerTerms.Reference {
					referencedLayer.GetRoot().Properties[k] = v
				}
			}
			// Remove the reference node
			toRemove = append(toRemove, node)
		}
	}
	for _, x := range toRemove {
		x.Remove()
	}
	return nil
}

func (compiler *Compiler) resolveCompositions(layer *Layer) error {
	for nodes := layer.AllNodes(); nodes.HasNext(); {
		compositeNode := nodes.Next().(*SchemaNode)
		if compositeNode.HasType(AttributeTypes.Composite) {
			if err := compiler.resolveComposition(layer, compositeNode); err != nil {
				return err
			}
		}
	}
	return nil
}

func (compiler Compiler) resolveComposition(layer *Layer, compositeNode *SchemaNode) error {
	type removable interface{ Remove() }
	toDelete := make([]removable, 0)
	// At the end of this process, composite node will be converted into an object node
	for edges := compositeNode.AllOutgoingEdgesWithLabel(LayerTerms.AllOf); edges.HasNext(); {
		allOfEdge := edges.Next().(*SchemaEdge)
	top:
		component := allOfEdge.To().(*SchemaNode)
		switch {
		case component.HasType(AttributeTypes.Object):
			// link all the nodes to the main node
			for edges := component.AllOutgoingEdges(); edges.HasNext(); {
				edge := edges.Next()
				edge.SetFrom(compositeNode)
			}
			// Copy all properties of the component node to the composite node
			if err := ComposeProperties(compositeNode.Properties, component.Properties); err != nil {
				return err
			}
			// Copy compiled items
			copyCompiled(compositeNode.Compiled, component.Compiled)
			toDelete = append(toDelete, component)

		case component.HasType(AttributeTypes.Value) ||
			component.HasType(AttributeTypes.Array) ||
			component.HasType(AttributeTypes.Polymorphic):
			// This node becomes an attribute of the main node.
			allOfEdge.SetLabel(LayerTerms.AttributeList)

		case component.HasType(AttributeTypes.Composite):
			if err := compiler.resolveComposition(layer, component); err != nil {
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
