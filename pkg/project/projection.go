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
package project

import (
	"errors"
	"fmt"

	"github.com/cloudprivacylabs/lsa/pkg/gl"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

const PT = ls.LS + "projection#"

// ProjectionTerms defines the terms used to specify projection layers
var ProjectionTerms = struct {
	// If given, the If term specifies a predicate that should be true to project the node
	If string
	// Vars defines a list of expressions that pull values from the
	// source graph and define them as variables
	Vars string
	// Source specifies the source value to be used to generate the target value
	Source string
	// IfEmpty determines whether to project the node even if it has no value
	IfEmpty string
	// JoinMethod determines how to join multiple values to generate a single value
	JoinMethod string
	// JoinDelimiter specifies the join delimiter if there are multiple values to be combined
	JoinDelimiter string
}{
	If:            ls.NewTerm(PT+"if", false, false, ls.OverrideComposition, nil),
	Vars:          ls.NewTerm(PT+"vars", false, true, ls.OverrideComposition, nil),
	Source:        ls.NewTerm(PT+"source", false, false, ls.OverrideComposition, nil),
	IfEmpty:       ls.NewTerm(PT+"ifEmpty", false, false, ls.OverrideComposition, nil),
	JoinMethod:    ls.NewTerm(PT+"joinMethod", false, false, ls.OverrideComposition, nil),
	JoinDelimiter: ls.NewTerm(PT+"joinDelimiter", false, false, ls.OverrideComposition, nil),
}

type Projector struct {
	TargetSchema *ls.Layer

	// If true, adds the references to the target schema
	AddInstanceOfEdges bool

	// GetProjectionProperties will return the projection related
	// properties for the node. This can be set to a function that
	// retrieves properties from an overlay, thus allowing projection
	// computations without layer composition
	GetProjectionProperties func(ls.Node) map[string]*ls.PropertyValue

	// GenerateID will generate a node ID given schema path and target document path up to the new node
	GenerateID func(schemaPath, docPath []ls.Node) string
}

func (projector *Projector) getProperties(node ls.Node) map[string]*ls.PropertyValue {
	if projector.GetProjectionProperties != nil {
		return projector.GetProjectionProperties(node)
	}
	return node.GetProperties()
}

// GenerateID gets the schema path for the new field, and the path to
// the parent container of the generated node
func (projector *Projector) generateID(schemaPath, docPath []ls.Node) string {
	if projector.GenerateID != nil {
		return projector.GenerateID(schemaPath, docPath)
	}
	return schemaPath[len(schemaPath)-1].GetID()
}

// ErrInvalidSchemaNodeType is returned if the schema node type cannot
// be projected (such as a reference, which cannot happen after
// compilation)
type ErrInvalidSchemaNodeType []string

func (e ErrInvalidSchemaNodeType) Error() string {
	return fmt.Sprintf("Invalid schema node type for projection: %v", []string(e))
}

var (
	ErrInvalidSource                = errors.New("Invalid source")
	ErrMultipleSourceNodesForObject = errors.New("Multiple source nodes specified for an object")
	ErrSourceMustBeString           = errors.New("source term value must be a string")
)

type ProjectionContext struct {
	// The expression language interpreter context
	glContext *gl.Context
	// All schema nodes from the root to the current node
	schemaPath []ls.Node
	// Generated document paths from the root to the parent of the current node
	docPath []ls.Node

	// The root node to be used to project
	sourceNode ls.Node
}

func (p *ProjectionContext) CurrentSchemaNode() ls.Node {
	return p.schemaPath[len(p.schemaPath)-1]
}

func (p *ProjectionContext) nestedContext() *ProjectionContext {
	ret := *p
	ret.glContext = p.glContext.NewNestedContext()
	return &ret
}

// Project the graph rooted at the rootNode to the targetSchema, using
// the getProjectionProperties function that will return projection
// properties for given schema nodes
func (projector *Projector) Project(rootNode ls.Node) (ls.Node, error) {
	ctx := ProjectionContext{
		glContext:  gl.NewContext(),
		schemaPath: []ls.Node{projector.TargetSchema.GetSchemaRootNode()},
		docPath:    []ls.Node{},
		sourceNode: rootNode,
	}
	ctx.glContext.Set("source", rootNode)
	return projector.project(&ctx)
}

func (projector *Projector) project(context *ProjectionContext) (ls.Node, error) {
	context = context.nestedContext()
	schemaNode := context.CurrentSchemaNode()
	properties := projector.getProperties(schemaNode)
	// Check conditionals first
	conditionals := properties[ProjectionTerms.If]
	v, err := checkConditionals(context, conditionals)
	if err != nil {
		return nil, err
	}
	if !v {
		return nil, nil
	}
	// Declare the variables
	variables := properties[ProjectionTerms.Vars]
	if err = setupVariables(context, variables); err != nil {
		return nil, err
	}
	switch {
	case schemaNode.HasType(ls.AttributeTypes.Value):
		return projector.value(context)
	case schemaNode.HasType(ls.AttributeTypes.Object):
		return projector.object(context)
	case schemaNode.HasType(ls.AttributeTypes.Array):
	case schemaNode.HasType(ls.AttributeTypes.Polymorphic):
	}
	return nil, ErrInvalidSchemaNodeType(schemaNode.GetTypes())
}

func (projector *Projector) object(context *ProjectionContext) (ls.Node, error) {
	schemaNode := context.CurrentSchemaNode()
	properties := projector.getProperties(schemaNode)
	attributes := ls.SortEdgesItr(schemaNode.GetAllOutgoingEdgesWithLabel(ls.LayerTerms.Attributes)).Targets().All()
	attributes = append(attributes, ls.SortEdgesItr(schemaNode.GetAllOutgoingEdgesWithLabel(ls.LayerTerms.AttributeList)).Targets().All()...)

	// Create a target node for this object node. If the object turns
	// out to be empty, this target node may be thrown away
	targetNode := ls.NewNode(projector.generateID(context.schemaPath, context.docPath), ls.DocumentNodeTerm)
	if projector.AddInstanceOfEdges {
		targetNode.Connect(schemaNode, ls.InstanceOfTerm)
	}
	context.docPath = append(context.docPath, targetNode)

	source, err := getSource(context, properties[ProjectionTerms.Source])
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
		newNode, err := projector.project(context)
		context.schemaPath = context.schemaPath[:len(context.schemaPath)-1]
		if err != nil {
			return nil, err
		}
		if newNode != nil {
			targetNode.Connect(newNode, ls.DataEdgeTerms.ObjectAttributes)
			empty = false
		}
	}
	if empty {
		ifEmpty := properties[ProjectionTerms.IfEmpty]
		if ifEmpty != nil && ifEmpty.IsString() && ifEmpty.AsString() == "true" {
			return targetNode, nil
		}
		return nil, nil
	}
	return targetNode, nil
}

func (projector *Projector) value(context *ProjectionContext) (ls.Node, error) {
	schemaNode := context.CurrentSchemaNode()
	properties := projector.getProperties(schemaNode)
	// Create a target node for this object node. If the object turns
	// out to be empty, this target node may be thrown away
	targetNode := ls.NewNode(projector.generateID(context.schemaPath, context.docPath), ls.DocumentNodeTerm)
	if projector.AddInstanceOfEdges {
		targetNode.Connect(schemaNode, ls.InstanceOfTerm)
	}
	context.docPath = append(context.docPath, targetNode)

	source, err := getSource(context, properties[ProjectionTerms.Source])
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
				prop := properties[ProjectionTerms.JoinMethod]
				if prop != nil && prop.IsString() {
					joinMethod = prop.AsString()
				}
				joinDelimiter := " "
				prop = properties[ProjectionTerms.JoinDelimiter]
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
		ifEmpty := properties[ProjectionTerms.IfEmpty]
		if ifEmpty != nil && ifEmpty.IsString() && ifEmpty.AsString() == "true" {
			return targetNode, nil
		}
		return nil, nil
	}
	return targetNode, nil
}

func getSource(context *ProjectionContext, source *ls.PropertyValue) (gl.Value, error) {
	if source == nil {
		return nil, nil
	}
	if !source.IsString() {
		return nil, ErrSourceMustBeString
	}
	value, err := gl.EvaluateExpression(context.glContext, source.AsString())
	if err != nil {
		return nil, err
	}
	return value, nil
}

func setupVariables(context *ProjectionContext, variables *ls.PropertyValue) error {
	if variables == nil {
		return nil
	}
	if variables.IsString() {
		if _, err := gl.EvaluateExpression(context.glContext, variables.AsString()); err != nil {
			return err
		}
	}
	if variables.IsStringSlice() {
		for _, x := range variables.AsStringSlice() {
			if _, err := gl.EvaluateExpression(context.glContext, x); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkConditionals(context *ProjectionContext, conditionals *ls.PropertyValue) (bool, error) {
	if conditionals == nil {
		return true, nil
	}
	result := true
	if conditionals.IsString() {
		r, err := gl.EvaluateExpression(context.glContext, conditionals.AsString())
		if err != nil {
			return false, err
		}
		result, err = r.AsBool()
		if err != nil {
			return false, err
		}
	}
	if conditionals.IsStringSlice() {
		for _, x := range conditionals.AsStringSlice() {
			r, err := gl.EvaluateExpression(context.glContext, x)
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
	}
	return result, nil
}
