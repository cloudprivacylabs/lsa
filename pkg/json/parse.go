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
	"encoding/json"
	"fmt"

	"github.com/bserdar/jsonom"

	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type ParsedDocNode struct {
	schemaNode *lpg.Node
	typeTerm   string
	value      string
	valueTypes []string
	children   []ls.ParsedDocNode
	name       string
	index      int
	id         string
}

func (i ParsedDocNode) GetSchemaNode() *lpg.Node              { return i.schemaNode }
func (i ParsedDocNode) GetTypeTerm() string                   { return i.typeTerm }
func (i ParsedDocNode) GetValue() string                      { return i.value }
func (i ParsedDocNode) GetValueTypes() []string               { return i.valueTypes }
func (i ParsedDocNode) GetChildren() []ls.ParsedDocNode       { return i.children }
func (i ParsedDocNode) GetID() string                         { return i.id }
func (i ParsedDocNode) GetProperties() map[string]interface{} { return nil }
func (i ParsedDocNode) GetAttributeIndex() int                { return i.index }
func (i ParsedDocNode) GetAttributeName() string              { return i.name }

type Parser struct {
	OnlySchemaAttributes bool
	IngestNullValues     bool
	Layer                *ls.Layer
	objectCache          map[*lpg.Node]map[string][]*lpg.Node
	discriminator        map[*lpg.Node][]*lpg.Node
}

type parserContext struct {
	context    *ls.Context
	path       ls.NodePath
	schemaNode *lpg.Node
}

func (ing *Parser) getObjectNodes(schemaNode *lpg.Node) (map[string][]*lpg.Node, error) {
	if ing.objectCache == nil {
		ing.objectCache = make(map[*lpg.Node]map[string][]*lpg.Node)
	}
	nodes, exists := ing.objectCache[schemaNode]
	if exists {
		return nodes, nil
	}
	nodes, err := ls.GetObjectAttributeNodesBy(schemaNode, ls.AttributeNameTerm.Name)
	if err != nil {
		return nil, err
	}
	ing.objectCache[schemaNode] = nodes
	return nodes, nil
}

func (ing *Parser) ParseDoc(context *ls.Context, baseID string, input jsonom.Node) (*ParsedDocNode, error) {
	ctx := parserContext{
		context:    context,
		path:       ls.NodePath{},
		schemaNode: ing.Layer.GetSchemaRootNode(),
	}
	if len(baseID) > 0 {
		ctx.path = append(ctx.path, baseID)
	}
	return ing.parseDoc(ctx, input)
}

func (ing *Parser) parseDoc(ctx parserContext, input jsonom.Node) (*ParsedDocNode, error) {
	if ctx.schemaNode == nil && ing.OnlySchemaAttributes {
		return nil, nil
	}
	if ctx.schemaNode != nil && ctx.schemaNode.HasLabel(ls.AttributeTypePolymorphic.Name) {
		return ing.parsePolymorphic(ctx, input)
	}
	switch next := input.(type) {
	case *jsonom.Object:
		return ing.parseObject(ctx, next)
	case *jsonom.Array:
		return ing.parseArray(ctx, next)
	}
	return ing.parseValue(ctx, input.(*jsonom.Value))
}

var PolyHintBlock = func(*jsonom.KeyValue) {}

func (ing *Parser) parseObject(ctx parserContext, input *jsonom.Object) (*ParsedDocNode, error) {
	// An object node
	if ctx.schemaNode != nil {
		if !ctx.schemaNode.HasLabel(ls.AttributeTypeObject.Name) {
			return nil, ls.ErrSchemaValidation{Msg: fmt.Sprintf("An object is expected here but found %s", ctx.schemaNode.GetLabels()), Path: ctx.path.Copy()}
		}
	}
	if ing.discriminator == nil {
		ing.discriminator = make(map[*lpg.Node][]*lpg.Node)
	}

	// if a cache exists for this schema node, parse nodes with type hint first
	if discrims, cached := ing.discriminator[ctx.schemaNode]; cached {
		for _, snode := range discrims {
			kv := input.Get(ls.AttributeNameTerm.PropertyValue(snode))
			newCtx := ctx
			newCtx.schemaNode = snode
			newCtx.path = newCtx.path.AppendString(kv.Key())
			_, err := ing.parseDoc(newCtx, kv.Value())
			if err != nil {
				return nil, err
			}
		}
	}

	// There is a schema node for this node. It must be an object
	nextNodes, err := ing.getObjectNodes(ctx.schemaNode)
	if err != nil {
		return nil, err
	}

	// if there is a discriminator for this schema node, cache other schema nodes with a type hint
	if _, ok := ing.discriminator[ctx.schemaNode]; !ok {
		ing.discriminator[ctx.schemaNode] = make([]*lpg.Node, 0)
		for _, sln := range nextNodes {
			for _, snode := range sln {
				if snode.HasLabel(ls.TypeDiscriminatorTerm.Name) {
					ing.discriminator[ctx.schemaNode] = append(ing.discriminator[ctx.schemaNode], snode)
				}
			}
		}
	}

	ret := ParsedDocNode{
		schemaNode: ctx.schemaNode,
		typeTerm:   ls.AttributeTypeObject.Name,
		children:   make([]ls.ParsedDocNode, 0, input.Len()),
		id:         ctx.path.String(),
	}

	processChildren := func(f func(*lpg.Node, jsonom.Node) bool) error {
		for i := 0; i < input.Len(); i++ {
			keyValue := input.N(i)
			schNodes := nextNodes[keyValue.Key()]
			if len(schNodes) > 1 {
				return ls.ErrInvalidSchema(fmt.Sprintf("Multiple elements with key '%s'", keyValue.Key()))
			}
			var schNode *lpg.Node
			if len(schNodes) == 1 {
				schNode = schNodes[0]
			}
			if !f(schNode, keyValue.Value()) {
				continue
			}
			newCtx := ctx
			newCtx.path = newCtx.path.AppendString(keyValue.Key())
			newCtx.schemaNode = schNode

			childNode, err := ing.parseDoc(newCtx, keyValue.Value())
			if err != nil {
				return ls.ErrDataIngestion{Key: keyValue.Key(), Err: err}
			}
			if childNode != nil {
				childNode.name = keyValue.Key()
				ret.children = append(ret.children, childNode)
			}
		}
		return nil
	}
	// Process value attributes with validators first, so if there are any validation errors, we terminate quickly
	if err := processChildren(func(schNode *lpg.Node, v jsonom.Node) bool {
		_, ok := v.(*jsonom.Value)
		return ok
	}); err != nil {
		return nil, err
	}
	if err := processChildren(func(_ *lpg.Node, v jsonom.Node) bool {
		_, ok := v.(*jsonom.Value)
		return !ok
	}); err != nil {
		return nil, err
	}

	return &ret, nil
}

func (ing *Parser) parseArray(ctx parserContext, input *jsonom.Array) (*ParsedDocNode, error) {
	// An array node
	if ctx.schemaNode != nil {
		if !ctx.schemaNode.HasLabel(ls.AttributeTypeArray.Name) {
			return nil, ls.ErrSchemaValidation{Msg: fmt.Sprintf("An array is expected here but found %s", ctx.schemaNode.GetLabels()), Path: ctx.path.Copy()}
		}
	}
	ret := ParsedDocNode{
		schemaNode: ctx.schemaNode,
		typeTerm:   ls.AttributeTypeArray.Name,
		children:   make([]ls.ParsedDocNode, 0, input.Len()),
		id:         ctx.path.String(),
	}
	elementsNode := ls.GetArrayElementNode(ctx.schemaNode)

	for i := 0; i < input.Len(); i++ {
		child := input.N(i)
		newCtx := ctx
		newCtx.path = newCtx.path.AppendInt(i)
		newCtx.schemaNode = elementsNode

		childNode, err := ing.parseDoc(newCtx, child)
		if err != nil {
			return nil, ls.ErrDataIngestion{Key: ctx.path.String(), Err: err}
		}
		if childNode != nil {
			childNode.index = i
			ret.children = append(ret.children, childNode)
		}
	}
	return &ret, nil
}

func (ing *Parser) testOption(option *lpg.Node, ctx parserContext, input jsonom.Node) (*ParsedDocNode, bool) {
	ctx.schemaNode = option
	out, err := ing.parseDoc(ctx, input)
	return out, out != nil && err == nil
}

func (ing *Parser) parsePolymorphic(ctx parserContext, input jsonom.Node) (*ParsedDocNode, error) {
	var found *lpg.Node
	var ret *ParsedDocNode
	for edges := ctx.schemaNode.GetEdgesWithLabel(lpg.OutgoingEdge, ls.OneOfTerm.Name); edges.Next(); {
		edge := edges.Edge()
		option := edge.GetTo()
		// fmt.Println("Option", option)
		pdn, ok := ing.testOption(option, ctx, input)
		// fmt.Println("Testing option", input.(*jsonom.Object).Get("resourceType").Value(), ok, option)
		if ok {
			if found != nil {
				return nil, ls.ErrSchemaValidation{Msg: "Multiple options of the polymorphic node matched:" + ls.GetNodeID(ctx.schemaNode), Path: ctx.path.Copy()}
			}
			found = option
			ret = pdn
		}
	}
	if found == nil {
		return nil, ls.ErrSchemaValidation{Msg: "None of the options of the polymorphic node matched:" + ls.GetNodeID(ctx.schemaNode), Path: ctx.path.Copy()}
	}
	ctx.schemaNode = found
	return ret, nil
}

func (ing *Parser) parseValue(ctx parserContext, input *jsonom.Value) (*ParsedDocNode, error) {
	if input.Value() == nil {
		if !ing.IngestNullValues {
			return nil, nil
		}
	}
	if ctx.schemaNode != nil {
		if !ctx.schemaNode.HasLabel(ls.AttributeTypeValue.Name) {
			return nil, ls.ErrSchemaValidation{Msg: fmt.Sprintf("A value is expected here but found %s", ctx.schemaNode.GetLabels()), Path: ctx.path.Copy()}
		}
	}
	var value string
	var typ string
	if input.Value() != nil {
		switch v := input.Value().(type) {
		case bool:
			value = fmt.Sprint(v)
			typ = BooleanTypeTerm.Name
		case string:
			value = v
			typ = StringTypeTerm.Name
		case uint8, uint16, uint32, uint64, int8, int16, int32, int64, int, uint, float32, float64:
			value = fmt.Sprint(input.Value())
			typ = NumberTypeTerm.Name
		case json.Number:
			value = string(v)
			typ = NumberTypeTerm.Name
		default:
			value = fmt.Sprint(v)
		}
	}
	if ctx.schemaNode != nil {
		if err := ls.ValidateValueBySchema(&value, ctx.schemaNode); err != nil {
			return nil, err
		}
	}
	ret := ParsedDocNode{
		schemaNode: ctx.schemaNode,
		typeTerm:   ls.AttributeTypeValue.Name,
		value:      value,
		valueTypes: []string{typ},
		id:         ctx.path.String(),
	}
	return &ret, nil
}
