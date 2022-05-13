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

package json

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/validators"
	"github.com/cloudprivacylabs/opencypher/graph"
)

type schemaProperty struct {
	ID           string
	location     string
	ref          string
	key          string
	reference    *CompiledEntity
	object       *objectSchema
	array        *arraySchema
	oneOf        []*schemaProperty
	allOf        []*schemaProperty
	typ          []string
	format       string
	enum         []interface{}
	pattern      string
	description  string
	defaultValue *string
	annotations  map[string]interface{}

	node graph.Node

	localReference *schemaProperty
	recurse        bool
}

type arraySchema struct {
	items *schemaProperty
}

func (a *arraySchema) resetNodes() {
	if a.items != nil {
		a.items.resetNodes()
	}
}

type objectSchema struct {
	properties map[string]*schemaProperty
	required   []string
}

func (o *objectSchema) resetNodes() {
	for _, x := range o.properties {
		if x != nil {
			x.resetNodes()
		}
	}
}

type schemaImporter struct {
	entityId string
	layer    *ls.Layer
	interner ls.Interner
	linkRefs LinkRefsBy
}

func (p *schemaProperty) resetNodes() {
	if p.recurse {
		return
	}
	p.recurse = true
	p.node = nil
	if p.object != nil {
		p.object.resetNodes()
	}
	if p.array != nil {
		p.array.resetNodes()
	}
	for _, x := range p.allOf {
		x.resetNodes()
	}
	for _, x := range p.oneOf {
		x.resetNodes()
	}
	if p.localReference != nil {
		p.localReference.resetNodes()
	}
	p.recurse = false
}

type schemaPath struct {
	name []string
	attr []*schemaProperty
}

// isLoop returns if adding the newAttr causes a loop
func (s schemaPath) isLoop(newAttr *schemaProperty) bool {
	for _, x := range s.attr {
		if x == newAttr {
			return true
		}
		if x.ref == newAttr.ref || x.location == newAttr.location {
			return true
		}
	}
	return false
}

func (s schemaPath) String() string {
	return strings.Join(s.name, "/")
}

func (imp schemaImporter) schemaAttrs(attr *schemaProperty, path schemaPath, key string) (graph.Node, error) {
	//if attr.node != nil {
	//	return attr.node, nil
	//}
	if path.isLoop(attr) {
		if attr.node != nil {
			return attr.node, nil
		}
	}
	path.name = append(path.name, key)
	path.attr = append(path.attr, attr)
	if attr.localReference != nil {
		// If this is a ref that is not a separate entity, simply point to
		// the contents of the referenced node
		attr.object = attr.localReference.object
		attr.array = attr.localReference.array
		attr.allOf = attr.localReference.allOf
		attr.oneOf = attr.localReference.oneOf
		attr.typ = attr.localReference.typ
		//attr.description = attr.localReference.description
		for k, v := range attr.localReference.annotations {
			if attr.annotations == nil {
				attr.annotations = make(map[string]interface{})
			}
			attr.annotations[k] = v
		}
		attr.localReference = nil
	}

	newNode, err := imp.newAttrNode(attr, path)
	if err != nil {
		return nil, err
	}
	return newNode, imp.buildChildAttrs(attr, newNode, path)
}

func (imp schemaImporter) newAttrNode(attr *schemaProperty, path schemaPath) (graph.Node, error) {
	newNode := imp.layer.Graph.NewNode([]string{ls.AttributeNodeTerm}, nil)
	attr.node = newNode
	return newNode, imp.setNodeProperties(attr, newNode, path)
}

func (imp schemaImporter) setNodeProperties(attr *schemaProperty, newNode graph.Node, path schemaPath) error {
	//	if len(attr.ID) > 0 {
	//		ls.SetNodeID(newNode, attr.ID)
	//		fmt.Println(attr.ID)
	//	}
	ls.SetNodeID(newNode, path.String())
	if len(attr.format) > 0 {
		newNode.SetProperty(validators.JsonFormatTerm, ls.StringPropertyValue(attr.format))
	}
	if len(attr.pattern) > 0 {
		newNode.SetProperty(validators.PatternTerm, ls.StringPropertyValue(attr.pattern))
	}
	//if len(attr.description) > 0 {
	//	newNode.SetProperty(ls.DescriptionTerm, ls.StringPropertyValue(attr.description))
	//}
	if len(attr.typ) > 0 {
		newNode.SetProperty(ls.ValueTypeTerm, ls.StringSlicePropertyValue(attr.typ))
	}
	if len(attr.key) > 0 {
		newNode.SetProperty(ls.AttributeNameTerm, ls.StringPropertyValue(attr.key))
	}
	if len(attr.enum) > 0 {
		elements := make([]string, 0, len(attr.enum))
		for _, v := range attr.enum {
			elements = append(elements, imp.interner.Intern(fmt.Sprint(v)))
		}
		newNode.SetProperty(validators.EnumTerm, ls.StringSlicePropertyValue(elements))
	}
	if attr.defaultValue != nil {
		newNode.SetProperty(ls.DefaultValueTerm, ls.StringPropertyValue(*attr.defaultValue))
	}
	for k, v := range attr.annotations {
		if err := ls.GetTermMarshaler(k).UnmarshalJSON(imp.layer, k, v, newNode, imp.interner); err != nil {
			return err
		}
	}

	if attr.reference != nil {
		newNode.SetLabels(newNode.GetLabels().Add(ls.AttributeTypeReference))
		switch imp.linkRefs {
		case LinkRefsBySchemaRef:
			newNode.SetProperty(ls.ReferenceTerm, ls.StringPropertyValue(attr.reference.Ref))
		case LinkRefsByLayerID:
			newNode.SetProperty(ls.ReferenceTerm, ls.StringPropertyValue(attr.reference.LayerID))
		case LinkRefsByValueType:
			newNode.SetProperty(ls.ReferenceTerm, ls.StringPropertyValue(attr.reference.ValueType))
		}
		return nil
	}
	if attr.object != nil {
		newNode.SetLabels(newNode.GetLabels().Add(ls.AttributeTypeObject))
		return nil
	}
	if attr.array != nil {
		newNode.SetLabels(newNode.GetLabels().Add(ls.AttributeTypeArray))
		return nil
	}
	if len(attr.oneOf) > 0 {
		newNode.SetLabels(newNode.GetLabels().Add(ls.AttributeTypePolymorphic))
		return nil
	}
	if len(attr.allOf) > 0 {
		newNode.SetLabels(newNode.GetLabels().Add(ls.AttributeTypeComposite))
		return nil
	}
	newNode.SetLabels(newNode.GetLabels().Add(ls.AttributeTypeValue))
	return nil
}

func (imp schemaImporter) buildChildAttrs(attr *schemaProperty, newNode graph.Node, path schemaPath) error {
	if attr.object != nil {
		newNode.SetLabels(newNode.GetLabels().Add(ls.AttributeTypeObject))
		attrIds := make(map[string]string)
		for _, v := range attr.object.properties {
			node, err := imp.schemaAttrs(v, path, v.key)
			if err != nil {
				return err
			}
			if len(v.key) > 0 {
				attrIds[v.key] = ls.GetNodeID(node)
			}
			imp.layer.Graph.NewEdge(newNode, node, ls.ObjectAttributeListTerm, nil)
		}
		if len(attr.object.required) > 0 {
			req := make([]string, 0, len(attr.object.required))
			for _, x := range attr.object.required {
				id := attrIds[x]
				if len(id) > 0 {
					req = append(req, id)
				}
			}
			newNode.SetProperty(validators.RequiredTerm, ls.StringSlicePropertyValue(req))
		}
		return nil
	}
	if attr.array != nil {
		newNode.SetLabels(newNode.GetLabels().Add(ls.AttributeTypeArray))
		n, err := imp.schemaAttrs(attr.array.items, path, "*")
		if err != nil {
			return err
		}
		imp.layer.Graph.NewEdge(newNode, n, ls.ArrayItemsTerm, nil)
		return nil
	}

	buildChoices := func(arr []*schemaProperty) ([]graph.Node, error) {
		elements := make([]graph.Node, 0, len(arr))
		for i, x := range arr {
			node, err := imp.schemaAttrs(x, path, strconv.Itoa(i))
			if err != nil {
				return nil, err
			}
			elements = append(elements, node)
		}
		return elements, nil
	}
	if len(attr.oneOf) > 0 {
		newNode.SetLabels(newNode.GetLabels().Add(ls.AttributeTypePolymorphic))
		choices, err := buildChoices(attr.oneOf)
		if err != nil {
			return err
		}
		for _, x := range choices {
			imp.layer.Graph.NewEdge(newNode, x, ls.OneOfTerm, nil)
		}
		return nil
	}
	if len(attr.allOf) > 0 {
		newNode.SetLabels(newNode.GetLabels().Add(ls.AttributeTypeComposite))
		choices, err := buildChoices(attr.allOf)
		if err != nil {
			return err
		}
		for _, x := range choices {
			imp.layer.Graph.NewEdge(newNode, x, ls.AllOfTerm, nil)
		}
		return nil
	}
	return nil
}

func (imp schemaImporter) linkChildAttrs(attr *schemaProperty, newNode, targetNode graph.Node) {
	for edges := targetNode.GetEdges(graph.OutgoingEdge); edges.Next(); {
		edge := edges.Edge()
		imp.layer.Graph.NewEdge(newNode, edge.GetTo(), edge.GetLabel(), nil)
	}
}
