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
package layers

import (
	"github.com/bserdar/digraph"
	"github.com/cloudprivacylabs/lsa/pkg/term"
)

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
	compile := func(propertiesMap, compiledMap map[string]interface{}, key string, value interface{}) error {
		value, compiled, err := term.GetCompiler(term.GetTermMeta(key)).Compile(value)
		if err != nil {
			return err
		}
		propertiesMap[key] = value
		if compiled != nil {
			compiledMap[key] = compiled
		}
		return nil
	}

	for nodes := layer.AllNodes(); nodes.HasNext(); {
		node := nodes.Next()
		nodeData := node.Payload.(*SchemaNode)
		for k, v := range nodeData.Properties {
			if err := compile(nodeData.Properties, nodeData.Compiled, k, v); err != nil {
				return err
			}
			for edges := node.AllOutgoingEdges(); edges.HasNext(); {
				edge := edges.Next()
				edgeData := edge.Payload.(*SchemaEdge)
				for k, v := range edgeData.Properties {
					if err := compile(edgeData.Properties, edgeData.Compiled, k, v); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (compiler *Compiler) resolveReferences(layer *Layer) error {
	toRemove := make([]*digraph.Node, 0)
	for nodes := layer.AllNodes(); nodes.HasNext(); {
		node := nodes.Next()
		payload := node.Payload.(*SchemaNode)
		if payload.HasType(AttributeTypes.Reference) {
			ref := payload.Properties[TypeTerms.Reference]
			referencedLayer, err := compiler.Compile(string(ref.(IRI)))
			if err != nil {
				return err
			}

			// Attach all incoming edges to the layer's root
			for incoming := node.AllIncomingEdges(); incoming.HasNext(); {
				edge := incoming.Next()
				layer.NewEdge(edge.From(), referencedLayer.RootNode, edge.Label(), edge.Payload)
			}
			// Reattach all outgoing edges
			for outgoing := node.AllOutgoingEdges(); outgoing.HasNext(); {
				edge := outgoing.Next()
				layer.NewEdge(referencedLayer.RootNode, edge.To(), edge.Label(), edge.Payload)
			}
			// Copy over properties
			for k, v := range payload.Properties {
				if k != TypeTerms.Reference {
					referencedLayer.RootNode.Payload.(*SchemaNode).Properties[k] = v
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
		compositeNode := nodes.Next()
		payload := compositeNode.Payload.(*SchemaNode)
		if payload.HasType(AttributeTypes.Composite) {
			if err := compiler.resolveComposition(layer, compositeNode); err != nil {
				return err
			}
		}
	}
	return nil
}

func (compiler Compiler) resolveComposition(layer *Layer, compositeNode *digraph.Node) error {
	type removable interface{ Remove() }
	toDelete := make([]removable, 0)
	// At the end of this process, composite node will be converted into an object node
	for edges := compositeNode.AllOutgoingEdgesWithLabel(TypeTerms.AllOf); edges.HasNext(); {
		allOfEdge := edges.Next()
	top:
		componentPayload := allOfEdge.To().Payload.(*SchemaNode)
		if componentPayload.HasType(AttributeTypes.Object) {
			// link all the nodes to the main node
			for edges := allOfEdge.To().AllOutgoingEdges(); edges.HasNext(); {
				edge := edges.Next()
				edge.SetFrom(compositeNode)
			}
			toDelete = append(toDelete, allOfEdge.To())
		} else if componentPayload.HasType(AttributeTypes.Value) ||
			componentPayload.HasType(AttributeTypes.Array) ||
			componentPayload.HasType(AttributeTypes.Polymorphic) {
			// This node becomes an attribute of the main node.
			layer.NewEdge(compositeNode, allOfEdge.To(), TypeTerms.AttributeList, allOfEdge.Payload)
			toDelete = append(toDelete, allOfEdge)
		} else if componentPayload.HasType(AttributeTypes.Composite) {
			if err := compiler.resolveComposition(layer, allOfEdge.To()); err != nil {
				return err
			}
			goto top
		} else {
			return ErrInvalidComposition
		}
	}
	// Convert the node to an object
	payload := compositeNode.Payload.(*SchemaNode)
	payload.RemoveTypes(AttributeTypes.Composite)
	payload.AddTypes(AttributeTypes.Object)
	return nil
}
