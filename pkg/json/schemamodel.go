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
	"strings"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
	"github.com/cloudprivacylabs/lsa/pkg/validators"
)

type schemaProperty struct {
	key          string
	reference    *CompiledEntity
	object       *objectSchema
	array        *arraySchema
	oneOf        []schemaProperty
	allOf        []schemaProperty
	typ          []string
	format       string
	enum         []interface{}
	pattern      string
	description  string
	defaultValue *string
	annotations  map[string]interface{}
}

type arraySchema struct {
	items schemaProperty
}

type objectSchema struct {
	properties map[string]schemaProperty
	required   []string
}

func (a arraySchema) itr(imp schemaImporter, name []string) (graph.Node, error) {
	return imp.schemaAttrs(append(name, "*"), a.items)
}

func (obj objectSchema) itr(imp schemaImporter, name []string) ([]graph.Node, error) {
	ret := make([]graph.Node, 0, len(obj.properties))
	for k, v := range obj.properties {
		nm := append(name, k)
		node, err := imp.schemaAttrs(nm, v)
		if err != nil {
			return nil, err
		}
		ret = append(ret, node)
	}
	return ret, nil
}

type schemaImporter struct {
	entityId string
	layer    *ls.Layer
	interner ls.Interner
	linkRefs LinkRefsBy
}

func (imp schemaImporter) schemaAttrs(name []string, attr schemaProperty) (graph.Node, error) {
	id := imp.entityId + "#" + strings.Join(name, ".")
	newNode := imp.layer.Graph.NewNode(nil, nil)
	ls.SetNodeID(newNode, id)
	return imp.buildSchemaAttrs(name, attr, newNode)
}

func (imp schemaImporter) buildSchemaAttrs(name []string, attr schemaProperty, newNode graph.Node) (graph.Node, error) {
	if len(attr.format) > 0 {
		newNode.SetProperty(validators.JsonFormatTerm, ls.StringPropertyValue(attr.format))
	}
	if len(attr.pattern) > 0 {
		newNode.SetProperty(validators.PatternTerm, ls.StringPropertyValue(attr.pattern))
	}
	if len(attr.description) > 0 {
		newNode.SetProperty(ls.DescriptionTerm, ls.StringPropertyValue(attr.description))
	}
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
			return nil, err
		}
	}

	if attr.reference != nil {
		newNode.SetLabels(newNode.GetLabels().Add(ls.AttributeTypeReference))
		newNode.SetProperty(ls.ReferenceTerm, ls.StringPropertyValue(attr.reference))
		return newNode, nil
	}

	if attr.object != nil {
		newNode.SetLabels(newNode.GetLabels().Add(ls.AttributeTypeObject))
		attrs, err := attr.object.itr(imp, name)
		if err != nil {
			return nil, err
		}
		for _, x := range attrs {
			imp.layer.Graph.NewEdge(newNode, x, ls.ObjectAttributeListTerm, nil)
		}
		if len(attr.object.required) > 0 {
			newNode.SetProperty(validators.RequiredTerm, ls.StringSlicePropertyValue(attr.object.required))
		}
		return newNode, nil
	}
	if attr.array != nil {
		newNode.SetLabels(newNode.GetLabels().Add(ls.AttributeTypeArray))
		n, err := attr.array.itr(imp, name)
		if err != nil {
			return nil, err
		}
		imp.layer.Graph.NewEdge(newNode, n, ls.ArrayItemsTerm, nil)
		return newNode, nil
	}

	buildChoices := func(arr []schemaProperty) ([]graph.Node, error) {
		elements := make([]graph.Node, 0, len(arr))
		for i, x := range arr {
			newName := append(name, fmt.Sprint(i))
			node, err := imp.schemaAttrs(newName, x)
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
			return nil, err
		}
		for _, x := range choices {
			imp.layer.Graph.NewEdge(newNode, x, ls.OneOfTerm, nil)
		}
		return newNode, nil
	}
	if len(attr.allOf) > 0 {
		newNode.SetLabels(newNode.GetLabels().Add(ls.AttributeTypeComposite))
		choices, err := buildChoices(attr.oneOf)
		if err != nil {
			return nil, err
		}
		for _, x := range choices {
			imp.layer.Graph.NewEdge(newNode, x, ls.AllOfTerm, nil)
		}
		return newNode, nil
	}
	newNode.SetLabels(newNode.GetLabels().Add(ls.AttributeTypeValue))
	return newNode, nil
}
