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

package transform

import (
	"errors"
	"fmt"

	"github.com/cloudprivacylabs/lsa/pkg/gl"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type Reshaper struct {
	TargetSchema *ls.Layer

	// If true, adds the references to the target schema
	AddInstanceOfEdges bool

	// GenerateID will generate a node ID given schema path and target document path up to the new node
	GenerateID func(schemaPath, docPath []ls.Node) string
}

// GenerateID gets the schema path for the new field, and the path to
// the parent container of the generated node
func (respaher *Reshaper) generateID(schemaPath, docPath []ls.Node) string {
	if respaher.GenerateID != nil {
		return respaher.GenerateID(schemaPath, docPath)
	}
	return schemaPath[len(schemaPath)-1].GetID() + "d"
}

// ErrInvalidSchemaNodeType is returned if the schema node type cannot
// be projected (such as a reference, which cannot happen after
// compilation)
type ErrInvalidSchemaNodeType []string

func (e ErrInvalidSchemaNodeType) Error() string {
	return fmt.Sprintf("Invalid schema node type for reshaping: %v", []string(e))
}

var (
	ErrInvalidSource                = errors.New("Invalid source")
	ErrMultipleSourceNodesForObject = errors.New("Multiple source nodes specified for an object")
	ErrSourceMustBeString           = errors.New("source term value must be a string")
)

type ReshapeContext struct {
	// The expression language interpreter context
	glContext *gl.Scope
	// All schema nodes from the root to the current node
	schemaPath []ls.Node
	// Generated document paths from the root to the parent of the current node
	docPath []ls.Node

	// The root node to be used to reshape
	sourceNode ls.Node
}

func (p *ReshapeContext) CurrentSchemaNode() ls.Node {
	return p.schemaPath[len(p.schemaPath)-1]
}

func (p *ReshapeContext) nestedContext() *ReshapeContext {
	ret := *p
	ret.glContext = p.glContext.NewScope()
	return &ret
}

// Reshape the graph rooted at the rootNode to the targetSchema, using
// the getReshapeProperties function that will return reshaping
// properties for given schema nodes
func (respaher *Reshaper) Reshape(rootNode ls.Node) (ls.Node, error) {
	ctx := ReshapeContext{
		glContext:  gl.NewScope(),
		schemaPath: []ls.Node{respaher.TargetSchema.GetSchemaRootNode()},
		docPath:    []ls.Node{},
		sourceNode: rootNode,
	}
	ctx.glContext.Set("source", rootNode)
	return respaher.reshape(&ctx)
}

func (reshaper *Reshaper) reshape(context *ReshapeContext) (ls.Node, error) {
	context = context.nestedContext()
	schemaNode := context.CurrentSchemaNode()
	// Check conditionals first
	v, err := reshaper.checkConditionals(context, schemaNode)
	if err != nil {
		return nil, err
	}
	if !v {
		return nil, nil
	}
	// Declare the variables
	if err = reshaper.setupVariables(context, schemaNode); err != nil {
		return nil, err
	}
	switch {
	case schemaNode.GetTypes().Has(ls.AttributeTypes.Value):
		return reshaper.value(context)
	case schemaNode.GetTypes().Has(ls.AttributeTypes.Object):
		return reshaper.object(context)
	case schemaNode.GetTypes().Has(ls.AttributeTypes.Array):
	case schemaNode.GetTypes().Has(ls.AttributeTypes.Polymorphic):
	}
	return nil, ErrInvalidSchemaNodeType(schemaNode.GetTypes().Slice())
}

func (reshaper *Reshaper) object(context *ReshapeContext) (ls.Node, error) {
	schemaNode := context.CurrentSchemaNode()
	properties := schemaNode.GetProperties()
	attributes := ls.SortEdgesItr(schemaNode.OutWith(ls.LayerTerms.Attributes)).Targets().All()
	attributes = append(attributes, ls.SortEdgesItr(schemaNode.OutWith(ls.LayerTerms.AttributeList)).Targets().All()...)

	// Create a target node for this object node. If the object turns
	// out to be empty, this target node may be thrown away
	targetNode := ls.NewNode(reshaper.generateID(context.schemaPath, context.docPath), ls.DocumentNodeTerm)
	if reshaper.AddInstanceOfEdges {
		ls.Connect(targetNode, schemaNode, ls.InstanceOfTerm)
	}
	context.docPath = append(context.docPath, targetNode)

	source, err := getSource(context, schemaNode)
	if err != nil {
		return nil, err
	}
	// If there is a source node, there can be at most one
	if source != nil {
		sourceNodes, ok := source.(gl.NodeValue)
		if !ok {
			return nil, ErrInvalidSource
		}
		if sourceNodes.Nodes.Len() > 1 {
			return nil, ErrMultipleSourceNodesForObject
		}
		if sourceNodes.Nodes.Len() == 1 {
			for _, k := range sourceNodes.Nodes.Slice() {
				context.sourceNode = k
				context.glContext.Set("source", context.sourceNode)
			}
		}
	}

	empty := true
	for _, a := range attributes {
		schemaAttribute := a.(ls.Node)
		context.schemaPath = append(context.schemaPath, schemaAttribute)
		newNode, err := reshaper.reshape(context)
		context.schemaPath = context.schemaPath[:len(context.schemaPath)-1]
		if err != nil {
			return nil, err
		}
		if newNode != nil {
			ls.Connect(targetNode, newNode, ls.HasTerm)
			empty = false
		}
	}
	if empty {
		ifEmpty := properties[ReshapeTerms.IfEmpty]
		if ifEmpty != nil && ifEmpty.IsString() && ifEmpty.AsString() == "true" {
			return targetNode, nil
		}
		return nil, nil
	}
	return targetNode, nil
}

func (reshaper *Reshaper) value(context *ReshapeContext) (ls.Node, error) {
	schemaNode := context.CurrentSchemaNode()
	properties := schemaNode.GetProperties()
	// Create a target node for this object node. If the object turns
	// out to be empty, this target node may be thrown away
	targetNode := ls.NewNode(reshaper.generateID(context.schemaPath, context.docPath), ls.DocumentNodeTerm)
	if reshaper.AddInstanceOfEdges {
		ls.Connect(targetNode, schemaNode, ls.InstanceOfTerm)
	}
	context.docPath = append(context.docPath, targetNode)

	source, err := getSource(context, schemaNode)
	if err != nil {
		return nil, err
	}
	empty := true
	if source != nil {
		switch sourceValue := source.(type) {
		case gl.NodeValue:
			switch {
			case sourceValue.Nodes.Len() == 1:
				targetNode.SetValue(sourceValue.Nodes.Slice()[0].GetValue())
				empty = false
			case sourceValue.Nodes.Len() > 1:
				joinMethod := "join"
				prop := properties[ReshapeTerms.JoinMethod]
				if prop != nil && prop.IsString() {
					joinMethod = prop.AsString()
				}
				joinDelimiter := " "
				prop = properties[ReshapeTerms.JoinDelimiter]
				if prop != nil && prop.IsString() {
					joinDelimiter = prop.AsString()
				}
				result, err := JoinValues(sourceValue.Nodes.Slice(), joinMethod, joinDelimiter)
				if err != nil {
					return nil, err
				}
				targetNode.SetValue(result)
				empty = false
			}
		case gl.BoolValue, gl.NumberValue, gl.StringValue:
			str, err := source.AsString()
			if err != nil {
				return nil, err
			}
			targetNode.SetValue(str)
			empty = false
		}
	}

	if empty {
		ifEmpty := properties[ReshapeTerms.IfEmpty]
		if ifEmpty != nil && ifEmpty.IsString() && ifEmpty.AsString() == "true" {
			return targetNode, nil
		}
		return nil, nil
	}
	return targetNode, nil
}

func getSource(context *ReshapeContext, schemaNode ls.Node) (gl.Value, error) {
	data := schemaNode.GetCompiledDataMap()[ReshapeTerms.Source]
	expr, ok := data.(gl.Evaluatable)
	if !ok {
		return nil, nil
	}
	value, err := expr.Evaluate(context.glContext)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (reshaper *Reshaper) setupVariables(context *ReshapeContext, schemaNode ls.Node) error {
	data := schemaNode.GetCompiledDataMap()[ReshapeTerms.Vars]
	slice, ok := data.([]gl.Evaluatable)
	if !ok {
		return nil
	}
	for _, x := range slice {
		if _, err := x.Evaluate(context.glContext); err != nil {
			return err
		}
	}
	return nil
}

func (reshaper *Reshaper) checkConditionals(context *ReshapeContext, schemaNode ls.Node) (bool, error) {
	data := schemaNode.GetCompiledDataMap()[ReshapeTerms.If]
	slice, ok := data.([]gl.Evaluatable)
	if !ok {
		return true, nil
	}
	result := true
	for _, x := range slice {
		r, err := x.Evaluate(context.glContext)
		if err != nil {
			return false, err
		}
		result, err = r.AsBool()
		if err != nil {
			return false, err
		}
		if result == false {
			return false, nil
		}
	}
	return result, nil
}
