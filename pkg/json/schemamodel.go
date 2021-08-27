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
}

type arraySchema struct {
	items schemaProperty
}

type objectSchema struct {
	properties map[string]schemaProperty
	required   []string
}

func (a arraySchema) itr(entityId string, name []string, layer *ls.Layer) ls.Node {
	return schemaAttrs(entityId, append(name, "*"), a.items, layer)
}

func (obj objectSchema) itr(entityId string, name []string, layer *ls.Layer) []ls.Node {
	ret := make([]ls.Node, 0, len(obj.properties))
	for k, v := range obj.properties {
		nm := append(name, k)
		ret = append(ret, schemaAttrs(entityId, nm, v, layer))
	}
	return ret
}

func schemaAttrs(entityId string, name []string, attr schemaProperty, layer *ls.Layer) ls.Node {
	id := entityId + "#" + strings.Join(name, ".")
	newNode := layer.NewNode(id)
	if len(attr.format) > 0 {
		newNode.GetProperties()[validators.JsonFormatTerm] = ls.StringPropertyValue(attr.format)
	}
	if len(attr.pattern) > 0 {
		newNode.GetProperties()[validators.PatternTerm] = ls.StringPropertyValue(attr.pattern)
	}
	if len(attr.description) > 0 {
		newNode.GetProperties()[ls.DescriptionTerm] = ls.StringPropertyValue(attr.description)
	}
	if len(attr.typ) > 0 {
		newNode.GetProperties()[ls.TargetType] = ls.StringSlicePropertyValue(attr.typ)
	}
	if len(attr.key) > 0 {
		newNode.GetProperties()[ls.AttributeNameTerm] = ls.StringPropertyValue(attr.key)
	}
	if len(attr.enum) > 0 {
		elements := make([]string, 0, len(attr.enum))
		for _, v := range attr.enum {
			elements = append(elements, fmt.Sprint(v))
		}
		newNode.GetProperties()[validators.EnumTerm] = ls.StringSlicePropertyValue(elements)
	}
	if attr.defaultValue != nil {
		newNode.GetProperties()[ls.DefaultValueTerm] = ls.StringPropertyValue(*attr.defaultValue)
	}

	if len(attr.reference) > 0 {
		newNode.GetTypes().Add(ls.AttributeTypes.Reference)
		newNode.GetProperties()[ls.LayerTerms.Reference] = ls.StringPropertyValue(attr.reference)
		return newNode
	}

	if attr.object != nil {
		newNode.GetTypes().Add(ls.AttributeTypes.Object)
		attrs := attr.object.itr(entityId, name, layer)
		for _, x := range attrs {
			ls.Connect(newNode, x, ls.LayerTerms.AttributeList)
		}
		if len(attr.object.required) > 0 {
			newNode.GetProperties()[validators.RequiredTerm] = ls.StringSlicePropertyValue(attr.object.required)
		}
		return newNode
	}
	if attr.array != nil {
		newNode.GetTypes().Add(ls.AttributeTypes.Array)
		n := attr.array.itr(entityId, name, layer)
		ls.Connect(newNode, n, ls.LayerTerms.ArrayItems)
		return newNode
	}

	buildChoices := func(arr []schemaProperty) []ls.Node {
		elements := make([]ls.Node, 0, len(arr))
		for i, x := range arr {
			newName := append(name, fmt.Sprint(i))
			node := schemaAttrs(entityId, newName, x, layer)
			elements = append(elements, node)
		}
		return elements
	}
	if len(attr.oneOf) > 0 {
		newNode.GetTypes().Add(ls.AttributeTypes.Polymorphic)
		for _, x := range buildChoices(attr.oneOf) {
			ls.Connect(newNode, x, ls.LayerTerms.OneOf)
		}
		return newNode
	}
	if len(attr.allOf) > 0 {
		newNode.GetTypes().Add(ls.AttributeTypes.Composite)
		for _, x := range buildChoices(attr.oneOf) {
			ls.Connect(newNode, x, ls.LayerTerms.AllOf)
		}
		return newNode
	}
	newNode.GetTypes().Add(ls.AttributeTypes.Value)
	return newNode
}
