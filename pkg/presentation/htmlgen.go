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
package presentation

type TemplateBasedRenderer struct {
}

func (t *TemplateBasedRenderer) RenderValue(schemaNode *ls.Layer, docNode digraph.Node) error {
	value := ls.GetFilteredValue(schemaNode, docNode)

}

// RenderUsingSchema uses a presentation schema to build
// presentation. This will only build output for the fields in the
// presentation schema.
func RenderUsingSchema(presentationSchema *ls.Layer, doc *digraph.Graph, linkName string, renderer Renderer) error {
	rootNode := presentationSchema.GetObjectInfoNode()
	if rootNode == nil {
		return nil
	}
	nodes, err := ls.SelectNodes(doc, ls.NodeLinkedPredicate{TargetID: rootNode.GetID(), Label: &linkName})
	if err != nil {
		return err
	}
	if len(nodes) != 1 {
		return nil
	}
	switch {
	case rootNode.HasType(ls.AttributeTypes.Value):
		if err := renderer.RenderValue(rootNode, nodes[0]); err != nil {
			return err
		}
	case rootNode.HasType(ls.AttributeTypes.Object):

	case rootNode.HasType(ls.AttributeTypes.Array):
		panic("Arrays not supported yet")
	}
	return Err
}
