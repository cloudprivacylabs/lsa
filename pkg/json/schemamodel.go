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
	reference    string
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

func (a arraySchema) itr(entityId string, name []string, layer *ls.Layer, interner ls.Interner) (graph.Node, error) {
	return schemaAttrs(entityId, append(name, "*"), a.items, layer, interner)
}

func (obj objectSchema) itr(entityId string, name []string, layer *ls.Layer, interner ls.Interner) ([]graph.Node, error) {
	ret := make([]graph.Node, 0, len(obj.properties))
	for k, v := range obj.properties {
		nm := append(name, k)
		node, err := schemaAttrs(entityId, nm, v, layer, interner)
		if err != nil {
			return nil, err
		}
		ret = append(ret, node)
	}
	return ret, nil
}

func schemaAttrs(entityId string, name []string, attr schemaProperty, layer *ls.Layer, interner ls.Interner) (graph.Node, error) {
	id := entityId + "#" + strings.Join(name, ".")
	newNode := layer.Graph.NewNode(nil, nil)
	ls.SetNodeID(newNode, id)
	return buildSchemaAttrs(entityId, name, attr, layer, newNode, interner)
}

func buildSchemaAttrs(entityId string, name []string, attr schemaProperty, layer *ls.Layer, newNode graph.Node, interner ls.Interner) (graph.Node, error) {
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
		newNode.SetProperty(ls.TargetType, ls.StringSlicePropertyValue(attr.typ))
	}
	if len(attr.key) > 0 {
		newNode.SetProperty(ls.AttributeNameTerm, ls.StringPropertyValue(attr.key))
	}
	if len(attr.enum) > 0 {
		elements := make([]string, 0, len(attr.enum))
		for _, v := range attr.enum {
			elements = append(elements, interner.Intern(fmt.Sprint(v)))
		}
		newNode.SetProperty(validators.EnumTerm, ls.StringSlicePropertyValue(elements))
	}
	if attr.defaultValue != nil {
		newNode.SetProperty(ls.DefaultValueTerm, ls.StringPropertyValue(*attr.defaultValue))
	}
	for k, v := range attr.annotations {
		if err := ls.GetTermMarshaler(k).UnmarshalJSON(layer, k, v, newNode, interner); err != nil {
			return nil, err
		}
	}

	if len(attr.reference) > 0 {
		newNode.SetLabels(newNode.GetLabels().Add(ls.AttributeTypeReference))
		newNode.SetProperty(ls.ReferenceTerm, ls.StringPropertyValue(attr.reference))
		return newNode, nil
	}

	if attr.object != nil {
		newNode.SetLabels(newNode.GetLabels().Add(ls.AttributeTypeObject))
		attrs, err := attr.object.itr(entityId, name, layer, interner)
		if err != nil {
			return nil, err
		}
		for _, x := range attrs {
			layer.Graph.NewEdge(newNode, x, ls.ObjectAttributeListTerm, nil)
		}
		if len(attr.object.required) > 0 {
			newNode.SetProperty(validators.RequiredTerm, ls.StringSlicePropertyValue(attr.object.required))
		}
		return newNode, nil
	}
	if attr.array != nil {
		newNode.SetLabels(newNode.GetLabels().Add(ls.AttributeTypeArray))
		n, err := attr.array.itr(entityId, name, layer, interner)
		if err != nil {
			return nil, err
		}
		layer.Graph.NewEdge(newNode, n, ls.ArrayItemsTerm, nil)
		return newNode, nil
	}

	buildChoices := func(arr []schemaProperty) ([]graph.Node, error) {
		elements := make([]graph.Node, 0, len(arr))
		for i, x := range arr {
			newName := append(name, fmt.Sprint(i))
			node, err := schemaAttrs(entityId, newName, x, layer, interner)
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
			layer.Graph.NewEdge(newNode, x, ls.OneOfTerm, nil)
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
			layer.Graph.NewEdge(newNode, x, ls.AllOfTerm, nil)
		}
		return newNode, nil
	}
	newNode.SetLabels(newNode.GetLabels().Add(ls.AttributeTypeValue))
	return newNode, nil
}
