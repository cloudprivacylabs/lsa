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

	"github.com/cloudprivacylabs/lsa/pkg/layers"
	"github.com/cloudprivacylabs/lsa/pkg/validators"
)

type schemaProperty struct {
	key         string
	reference   string
	object      *objectSchema
	array       *arraySchema
	oneOf       []schemaProperty
	allOf       []schemaProperty
	typ         []string
	format      string
	enum        []interface{}
	required    bool
	pattern     string
	description string
}

type arraySchema struct {
	items schemaProperty
}

type objectSchema struct {
	properties map[string]schemaProperty
}

func (a arraySchema) itr(entityId string, name []string, layer *layers.Layer) *layers.SchemaNode {
	return schemaAttrs(entityId, append(name, "*"), a.items, layer)
}

func (obj objectSchema) itr(entityId string, name []string, layer *layers.Layer) []*layers.SchemaNode {
	ret := make([]*layers.SchemaNode, 0, len(obj.properties))
	for k, v := range obj.properties {
		nm := append(name, k)
		ret = append(ret, schemaAttrs(entityId, nm, v, layer))
	}
	return ret
}

func schemaAttrs(entityId string, name []string, attr schemaProperty, layer *layers.Layer) *layers.SchemaNode {
	id := entityId + "#" + strings.Join(name, ".")
	newNode := layers.NewSchemaNode(id)
	layer.AddNode(newNode)
	if len(attr.format) > 0 {
		newNode.Properties[validators.FormatTerm] = attr.format
	}
	if len(attr.pattern) > 0 {
		newNode.Properties[validators.PatternTerm] = attr.pattern
	}
	if len(attr.description) > 0 {
		newNode.Properties[layers.DescriptionTerm] = attr.description
	}
	if attr.required {
		newNode.Properties[validators.RequiredTerm] = true
	}
	if len(attr.typ) > 0 {
		arr := make([]interface{}, 0, len(attr.typ))
		for _, x := range attr.typ {
			arr = append(arr, x)
		}
		newNode.Properties[layers.TargetType] = arr
	}
	if len(attr.key) > 0 {
		newNode.Properties[layers.AttributeNameTerm] = attr.key
	}
	if len(attr.enum) > 0 {
		elements := make([]interface{}, 0, len(attr.enum))
		for _, v := range attr.enum {
			elements = append(elements, v)
		}
		newNode.Properties[validators.EnumTerm] = elements
	}

	if len(attr.reference) > 0 {
		newNode.AddTypes(layers.AttributeTypes.Reference)
		return newNode
	}

	if attr.object != nil {
		newNode.AddTypes(layers.AttributeTypes.Object)
		attrs := attr.object.itr(entityId, name, layer)
		for _, x := range attrs {
			layer.AddEdge(newNode, x, layers.NewSchemaEdge(layers.TypeTerms.AttributeList))
		}
		return newNode
	}
	if attr.array != nil {
		newNode.AddTypes(layers.AttributeTypes.Array)
		n := attr.array.itr(entityId, name, layer)
		layer.AddEdge(newNode, n, layers.NewSchemaEdge(layers.TypeTerms.ArrayItems))
		return newNode
	}

	buildChoices := func(arr []schemaProperty) []*layers.SchemaNode {
		elements := make([]*layers.SchemaNode, 0, len(arr))
		for i, x := range arr {
			newName := append(name, fmt.Sprint(i))
			node := schemaAttrs(entityId, newName, x, layer)
			elements = append(elements, node)
		}
		return elements
	}
	if len(attr.oneOf) > 0 {
		newNode.AddTypes(layers.AttributeTypes.Polymorphic)
		for _, x := range buildChoices(attr.oneOf) {
			layer.AddEdge(newNode, x, layers.NewSchemaEdge(layers.TypeTerms.OneOf))
		}
		return newNode
	}
	if len(attr.allOf) > 0 {
		newNode.AddTypes(layers.AttributeTypes.Composite)
		for _, x := range buildChoices(attr.oneOf) {
			layer.AddEdge(newNode, x, layers.NewSchemaEdge(layers.TypeTerms.AllOf))
		}
		return newNode
	}
	newNode.AddTypes(layers.AttributeTypes.Value)
	return newNode
}
