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

package xml

import (
	"encoding/xml"
	"fmt"
	"io"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
)

type ParsedDocNode struct {
	schemaNode graph.Node
	typeTerm   string
	value      string
	valueTypes []string
	children   []ls.ParsedDocNode
	name       xml.Name
	index      int
	id         string
	properties map[string]interface{}
}

func (i ParsedDocNode) GetSchemaNode() graph.Node             { return i.schemaNode }
func (i ParsedDocNode) GetTypeTerm() string                   { return i.typeTerm }
func (i ParsedDocNode) GetValue() string                      { return i.value }
func (i ParsedDocNode) GetValueTypes() []string               { return i.valueTypes }
func (i ParsedDocNode) GetChildren() []ls.ParsedDocNode       { return i.children }
func (i ParsedDocNode) GetID() string                         { return i.id }
func (i ParsedDocNode) GetProperties() map[string]interface{} { return i.properties }
func (i ParsedDocNode) GetAttributeIndex() int                { return i.index }
func (i ParsedDocNode) GetAttributeName() string              { return "" }

type Parser struct {
	OnlySchemaAttributes bool
	IngestEmptyValues    bool
	SchemaNode           graph.Node
}

type parserContext struct {
	context    *ls.Context
	path       ls.NodePath
	schemaNode graph.Node
}

func (ing Parser) ParseStream(context *ls.Context, baseID string, input io.Reader) (*ParsedDocNode, error) {
	decoder := xml.NewDecoder(input)
	return ing.DecodeAndParse(context, baseID, decoder)
}

func (ing Parser) DecodeAndParse(context *ls.Context, baseID string, decoder *xml.Decoder) (*ParsedDocNode, error) {
	el, err := decode(decoder)
	if err != nil {
		return nil, err
	}
	return ing.ParseDoc(context, baseID, el)
}

func (ing Parser) ParseDoc(context *ls.Context, baseID string, input *xmlElement) (*ParsedDocNode, error) {
	ctx := parserContext{
		context:    context,
		path:       ls.NodePath{},
		schemaNode: ing.SchemaNode,
	}
	if len(baseID) > 0 {
		ctx.path = append(ctx.path, baseID)
	}

	return ing.element(ctx, input)
}

func (ing Parser) element(ctx parserContext, element *xmlElement) (*ParsedDocNode, error) {
	// If schemaNode is nil and we are only ingesting known nodes, ignore this node
	if ctx.schemaNode == nil && ing.OnlySchemaAttributes {
		return nil, nil
	}
	// What are we ingesting? If there is a schema, it dictates the type
	if ctx.schemaNode != nil {
		// Does element match the schema
		schemaName := GetXMLName(ctx.schemaNode)

		if !MatchName(element.name, schemaName) {
			return nil, nil
		}
		switch {
		case ctx.schemaNode.HasLabel(ls.AttributeTypeValue):
			return ing.parseValue(ctx, element)
		case ctx.schemaNode.HasLabel(ls.AttributeTypeObject):
			return ing.parseObject(ctx, element)
		case ctx.schemaNode.HasLabel(ls.AttributeTypeArray):
			return ing.parseArray(ctx, element)
		case ctx.schemaNode.HasLabel(ls.AttributeTypePolymorphic):
			return ing.parsePolymorphic(ctx, element)
		}
		return nil, ls.ErrInvalidSchema(fmt.Sprintf("Cannot determine attribute type for %s", ls.GetNodeID(ctx.schemaNode)))
	}
	// No schemas. If this is an element with a single text node, ingest as value
	if len(element.children) == 1 {
		if _, text := element.children[0].(*xmlText); text {
			return ing.parseValue(ctx, element)
		}
	}
	return ing.parseObject(ctx, element)
}

func (ing Parser) parseValue(ctx parserContext, element *xmlElement) (*ParsedDocNode, error) {
	// element has at most one text node, or valueAttr is set
	var value string
	if ctx.schemaNode != nil {
		pvalue := ls.AsPropertyValue(ctx.schemaNode.GetProperty(ValueAttributeTerm)).AsString()
		if len(pvalue) > 0 {
			v, ok := element.findAttr(xml.Name{Local: pvalue})
			if ok {
				ret := &ParsedDocNode{
					name:       element.name,
					schemaNode: ctx.schemaNode,
					typeTerm:   ls.AttributeTypeValue,
					properties: make(map[string]interface{}),
					id:         ctx.path.Append(element.name.Local).Append(pvalue).String(),
					value:      v,
				}
				ret.properties[ls.AttributeNameTerm] = ls.StringPropertyValue(element.name.Local)
				if len(element.name.Space) > 0 {
					ret.properties[NamespaceTerm] = ls.StringPropertyValue(element.name.Space)
				}
				return ret, nil
			}
			return nil, nil
		}
	}

	if len(element.children) > 1 {
		return nil, ls.ErrSchemaValidation{Msg: "Cannot ingest element as a value because it has multiple child nodes", Path: ctx.path.Copy()}
	}
	if len(element.children) == 1 {
		t, ok := element.children[0].(*xmlText)
		if !ok {
			return nil, ls.ErrSchemaValidation{Msg: "Cannot ingest element as a value because it has elements as children", Path: ctx.path.Copy()}
		}
		value = string(t.text)
	}
	ret := &ParsedDocNode{
		name:       element.name,
		schemaNode: ctx.schemaNode,
		typeTerm:   ls.AttributeTypeValue,
		properties: make(map[string]interface{}),
		id:         ctx.path.Append(element.name.Local).String(),
		value:      value,
	}
	// Process attributes
	for _, attribute := range element.attributes {
		if !ing.IngestEmptyValues && len(attribute.value) == 0 {
			continue
		}
		ret.properties[makeFullName(attribute.name)] = ls.StringPropertyValue(attribute.value)
	}

	ret.properties[ls.AttributeNameTerm] = ls.StringPropertyValue(element.name.Local)
	if len(element.name.Space) > 0 {
		ret.properties[NamespaceTerm] = ls.StringPropertyValue(element.name.Space)
	}
	return ret, nil
}

func (ing Parser) attributes(ctx parserContext, element *xmlElement, childSchemaNodes []graph.Node) ([]ls.ParsedDocNode, error) {
	children := make([]ls.ParsedDocNode, 0)
	for _, attribute := range element.attributes {
		if !ing.IngestEmptyValues && len(attribute.value) == 0 {
			continue
		}
		var attrSchema graph.Node
		var err error
		if ctx.schemaNode != nil {
			attrSchema, err = findBestMatchingSchemaAttribute(attribute.name, childSchemaNodes, true)
			if err != nil {
				return nil, err
			}
		}
		attrNode := &ParsedDocNode{
			name:       attribute.name,
			value:      attribute.value,
			schemaNode: attrSchema,
			typeTerm:   ls.AttributeTypeValue,
			id:         fmt.Sprintf("%s#%s", ctx.path.String(), attribute.name.Local),
			properties: make(map[string]interface{}),
		}
		if len(attribute.name.Space) > 0 {
			attrNode.properties[NamespaceTerm] = ls.StringPropertyValue(ctx.context.GetInterner().Intern(attribute.name.Space))
		}
		attrNode.properties[ls.AttributeNameTerm] = ls.StringPropertyValue(ctx.context.GetInterner().Intern(attribute.name.Local))
		children = append(children, attrNode)
	}
	return children, nil
}

func (ing Parser) parseObject(ctx parserContext, element *xmlElement) (*ParsedDocNode, error) {
	// Get all the possible child nodes from the schema. If the
	// schemaNode is nil, the returned schemaNodes will be empty
	childSchemaNodes := ls.GetObjectAttributeNodes(ctx.schemaNode)
	ctx.path = ctx.path.Append(element.name.Local)
	ret := &ParsedDocNode{
		name:       element.name,
		schemaNode: ctx.schemaNode,
		typeTerm:   ls.AttributeTypeObject,
		properties: make(map[string]interface{}),
		id:         ctx.path.String(),
	}
	ret.properties[ls.AttributeNameTerm] = ls.StringPropertyValue(element.name.Local)
	if len(element.name.Space) > 0 {
		ret.properties[NamespaceTerm] = ls.StringPropertyValue(element.name.Space)
	}

	ch, err := ing.attributes(ctx, element, childSchemaNodes)
	if err != nil {
		return nil, err
	}
	ret.children = append(ret.children, ch...)

	for index, child := range element.children {
		var newChildNode *ParsedDocNode
		switch childNode := child.(type) {
		case *xmlElement:
			childSchema, err := findBestMatchingSchemaAttribute(childNode.name, childSchemaNodes, false)
			if err != nil {
				return nil, err
			}
			// If nothing was found, check if there are polymorphic nodes
			if childSchema == nil {
				for _, childSchemaNode := range childSchemaNodes {
					if !childSchemaNode.HasLabel(ls.AttributeTypePolymorphic) {
						continue
					}
					newCtx := ctx
					newCtx.schemaNode = childSchemaNode
					// A polymorphic node
					newChildNode, err = ing.parsePolymorphic(newCtx, childNode)
					if err != nil {
						return nil, err
					}
				}
			}
			if newChildNode == nil {
				newCtx := ctx
				//newCtx.path = newCtx.path.Append(childNode.name.Local)
				newCtx.schemaNode = childSchema
				newChildNode, err = ing.element(newCtx, childNode)
				if err != nil {
					return nil, err
				}
			}
		case *xmlText:
			newChildNode = &ParsedDocNode{
				typeTerm: ls.AttributeTypeValue,
				value:    string(childNode.text),
			}
		}
		if newChildNode != nil {
			newChildNode.index = index
			ret.children = append(ret.children, newChildNode)
		}
	}
	return ret, nil
}

func (ing Parser) parseArray(ctx parserContext, element *xmlElement) (*ParsedDocNode, error) {
	elementNode := ls.GetArrayElementNode(ctx.schemaNode)
	ret := &ParsedDocNode{
		name:       element.name,
		schemaNode: ctx.schemaNode,
		typeTerm:   ls.AttributeTypeArray,
		properties: make(map[string]interface{}),
		id:         ctx.path.String(),
	}
	ret.properties[ls.AttributeNameTerm] = ls.StringPropertyValue(element.name.Local)
	if len(element.name.Space) > 0 {
		ret.properties[NamespaceTerm] = ls.StringPropertyValue(element.name.Space)
	}

	ch, err := ing.attributes(ctx, element, nil)
	if err != nil {
		return nil, err
	}
	ret.children = append(ret.children, ch...)

	for index, child := range element.children {
		var newNode *ParsedDocNode
		switch childNode := child.(type) {
		case *xmlElement:
			newCtx := ctx
			newCtx.path = append(newCtx.path, childNode.name.Local)
			newCtx.schemaNode = elementNode
			newNode, err = ing.element(newCtx, childNode)
			if err != nil {
				return nil, err
			}
		case *xmlText:
			newNode = &ParsedDocNode{
				typeTerm: ls.AttributeTypeValue,
				value:    string(childNode.text),
			}
		}
		if newNode != nil {
			newNode.properties[ls.AttributeIndexTerm] = ls.IntPropertyValue(index)
			ret.children = append(ret.children, newNode)
		}
	}
	return ret, nil
}

func (ing Parser) testOption(option graph.Node, ctx parserContext, element *xmlElement) bool {
	ctx.schemaNode = option
	out, err := ing.element(ctx, element)
	return out != nil && err == nil
}

func (ing Parser) parsePolymorphic(ctx parserContext, element *xmlElement) (*ParsedDocNode, error) {
	options := ls.GetPolymorphicOptions(ctx.schemaNode)
	var found graph.Node
	for _, option := range options {
		if ing.testOption(option, ctx, element) {
			if found != nil {
				return nil, ls.ErrSchemaValidation{Msg: "Multiple options of the polymorphic node matched:" + ls.GetNodeID(ctx.schemaNode), Path: ctx.path.Copy()}

			}
			found = option
		}
	}
	ctx.context.GetLogger().Info(map[string]interface{}{"xml.parse.polymorphic": "None of the options of the polymorphic node matched", "nodeId": ls.GetNodeID(ctx.schemaNode), "elem": element, "path": ctx.path.Copy()})
	//if found == nil {
	//	return nil, ls.ErrSchemaValidation{Msg: fmt.Sprintf("None of the options of the polymorphic node matched: %s (%+v)"+ls.GetNodeID(ctx.schemaNode), element), Path: ctx.path.Copy()}
	//}

	ctx.schemaNode = found
	return ing.element(ctx, element)
}
