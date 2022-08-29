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

	"github.com/cloudprivacylabs/lpg"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/validators"
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

	node *lpg.Node

	localReference *schemaProperty
	recurse        bool
}

func (p *schemaProperty) size() int {
	n := p.object.size() + p.array.size()
	for _, x := range p.oneOf {
		n += x.size()
	}
	for _, x := range p.allOf {
		n += x.size()
	}
	return n
}

type arraySchema struct {
	items *schemaProperty
}

func (a *arraySchema) size() int {
	if a == nil {
		return 0
	}
	return a.items.size() + 1
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

func (o *objectSchema) size() int {
	if o == nil {
		return 0
	}
	n := 0
	for _, c := range o.properties {
		n += c.size() + 1
	}
	return n
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

func (imp schemaImporter) schemaAttrs(attr *schemaProperty, path schemaPath, key string) (*lpg.Node, error) {
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

func (imp schemaImporter) newAttrNode(attr *schemaProperty, path schemaPath) (*lpg.Node, error) {
	newNode := imp.layer.Graph.NewNode([]string{ls.AttributeNodeTerm}, nil)
	attr.node = newNode
	return newNode, imp.setNodeProperties(attr, newNode, path)
}

func (imp schemaImporter) setNodeProperties(attr *schemaProperty, newNode *lpg.Node, path schemaPath) error {
	//	if len(attr.ID) > 0 {
	//		ls.SetNodeID(newNode, attr.ID)
	//		fmt.Println(attr.ID)
	//	}
	labels := newNode.GetLabels()
	ls.SetNodeID(newNode, path.String())
	if len(attr.format) > 0 {
		newNode.SetProperty(validators.JsonFormatTerm, ls.StringPropertyValue(validators.JsonFormatTerm, attr.format))
	}
	if len(attr.pattern) > 0 {
		newNode.SetProperty(validators.PatternTerm, ls.StringPropertyValue(validators.PatternTerm, attr.pattern))
	}
	//if len(attr.description) > 0 {
	//	newNode.SetProperty(ls.DescriptionTerm, ls.StringPropertyValue(attr.description))
	//}
	if len(attr.typ) > 0 {
		newNode.SetProperty(ls.ValueTypeTerm, ls.StringSlicePropertyValue(ls.ValueTypeTerm, attr.typ))
	}
	if len(attr.key) > 0 {
		newNode.SetProperty(ls.AttributeNameTerm, ls.StringPropertyValue(ls.AttributeNameTerm, attr.key))
	}
	if len(attr.enum) > 0 {
		elements := make([]string, 0, len(attr.enum))
		for _, v := range attr.enum {
			elements = append(elements, imp.interner.Intern(fmt.Sprint(v)))
		}
		newNode.SetProperty(validators.EnumTerm, ls.StringSlicePropertyValue(validators.EnumTerm, elements))
	}
	if attr.defaultValue != nil {
		newNode.SetProperty(ls.DefaultValueTerm, ls.StringPropertyValue(ls.DefaultValueTerm, *attr.defaultValue))
	}
	for k, v := range attr.annotations {
		if err := ls.GetTermMarshaler(k).UnmarshalJSON(imp.layer, k, v, newNode, imp.interner); err != nil {
			return err
		}
	}

	if attr.reference != nil {
		labels.Add(ls.AttributeTypeReference)
		newNode.SetLabels(labels)
		switch imp.linkRefs {
		case LinkRefsBySchemaRef:
			newNode.SetProperty(ls.ReferenceTerm, ls.StringPropertyValue(ls.ReferenceTerm, attr.reference.Ref))
		case LinkRefsByLayerID:
			newNode.SetProperty(ls.ReferenceTerm, ls.StringPropertyValue(ls.ReferenceTerm, attr.reference.LayerID))
		case LinkRefsByValueType:
			newNode.SetProperty(ls.ReferenceTerm, ls.StringPropertyValue(ls.ReferenceTerm, attr.reference.ValueType))
		}
		return nil
	}
	if attr.object != nil {
		labels.Add(ls.AttributeTypeObject)
		newNode.SetLabels(labels)
		return nil
	}
	if attr.array != nil {
		labels.Add(ls.AttributeTypeArray)
		newNode.SetLabels(labels)
		return nil
	}
	if len(attr.oneOf) > 0 {
		labels.Add(ls.AttributeTypePolymorphic)
		newNode.SetLabels(labels)
		return nil
	}
	if len(attr.allOf) > 0 {
		labels.Add(ls.AttributeTypeComposite)
		newNode.SetLabels(labels)
		return nil
	}
	labels.Add(ls.AttributeTypeValue)
	newNode.SetLabels(labels)
	return nil
}

func (imp schemaImporter) buildChildAttrs(attr *schemaProperty, newNode *lpg.Node, path schemaPath) error {
	labels := newNode.GetLabels()
	if attr.object != nil {
		labels.Add(ls.AttributeTypeObject)
		newNode.SetLabels(labels)
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
			newNode.SetProperty(validators.RequiredTerm, ls.StringSlicePropertyValue(validators.RequiredTerm, req))
		}
		return nil
	}
	if attr.array != nil {
		labels.Add(ls.AttributeTypeArray)
		newNode.SetLabels(labels)
		n, err := imp.schemaAttrs(attr.array.items, path, "*")
		if err != nil {
			return err
		}
		imp.layer.Graph.NewEdge(newNode, n, ls.ArrayItemsTerm, nil)
		return nil
	}

	buildChoices := func(arr []*schemaProperty) ([]*lpg.Node, error) {
		elements := make([]*lpg.Node, 0, len(arr))
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
		labels.Add(ls.AttributeTypePolymorphic)
		newNode.SetLabels(labels)
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
		labels.Add(ls.AttributeTypeComposite)
		newNode.SetLabels(labels)
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

func (imp schemaImporter) linkChildAttrs(attr *schemaProperty, newNode, targetNode *lpg.Node) {
	for edges := targetNode.GetEdges(lpg.OutgoingEdge); edges.Next(); {
		edge := edges.Edge()
		imp.layer.Graph.NewEdge(newNode, edge.GetTo(), edge.GetLabel(), nil)
	}
}
