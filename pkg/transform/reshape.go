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
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

type Reshaper struct {
	TargetSchema *ls.Layer

	// If true, adds the references to the target schema
	AddInstanceOfEdges bool

	// GenerateID will generate a node ID given schema path and target document path up to the new node
	GenerateID func(schemaPath, docPath []graph.Node) string
}

// GenerateID gets the schema path for the new field, and the path to
// the parent container of the generated node
func (respaher *Reshaper) generateID(schemaPath, docPath []graph.Node) string {
	if respaher.GenerateID != nil {
		return respaher.GenerateID(schemaPath, docPath)
	}
	return ls.GetNodeID(schemaPath[len(schemaPath)-1])
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
	schemaPath []graph.Node
	// Generated document paths from the root to the parent of the current node
	docPath []graph.Node

	// The root node to be used to reshape
	sourceNode graph.Node

	targetGraph graph.Graph
}

func (p *ReshapeContext) CurrentSchemaNode() graph.Node {
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
func (respaher *Reshaper) Reshape(rootNode graph.Node, targetGraph graph.Graph) (graph.Node, error) {
	ctx := ReshapeContext{
		glContext:   gl.NewScope(),
		schemaPath:  []graph.Node{respaher.TargetSchema.GetSchemaRootNode()},
		docPath:     []graph.Node{},
		sourceNode:  rootNode,
		targetGraph: targetGraph,
	}
	ctx.glContext.Set("source", rootNode)
	return respaher.reshape(&ctx)
}

func (reshaper *Reshaper) reshape(context *ReshapeContext) (graph.Node, error) {
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
	case schemaNode.GetLabels().Has(ls.AttributeTypeValue):
		return reshaper.value(context)
	case schemaNode.GetLabels().Has(ls.AttributeTypeObject):
		return reshaper.object(context)
	case schemaNode.GetLabels().Has(ls.AttributeTypeArray):
	case schemaNode.GetLabels().Has(ls.AttributeTypePolymorphic):
	}
	return nil, ErrInvalidSchemaNodeType(schemaNode.GetLabels().Slice())
}

func (reshaper *Reshaper) object(context *ReshapeContext) (graph.Node, error) {
	schemaNode := context.CurrentSchemaNode()
	attributes := graph.TargetNodes(ls.SortEdgesItr(schemaNode.GetEdgesWithLabel(graph.OutgoingEdge, ls.ObjectAttributesTerm)))
	attributes = append(attributes, graph.TargetNodes(ls.SortEdgesItr(schemaNode.GetEdgesWithLabel(graph.OutgoingEdge, ls.ObjectAttributeListTerm)))...)

	// Create a target node for this object node. If the object turns
	// out to be empty, this target node may be thrown away
	targetNode := context.targetGraph.NewNode([]string{ls.DocumentNodeTerm, ls.AttributeTypeObject}, nil)
	ls.SetNodeID(targetNode, reshaper.generateID(context.schemaPath, context.docPath))
	if reshaper.AddInstanceOfEdges {
		context.targetGraph.NewEdge(targetNode, schemaNode, ls.InstanceOfTerm, nil)
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
		schemaAttribute := a.(graph.Node)
		context.schemaPath = append(context.schemaPath, schemaAttribute)
		newNode, err := reshaper.reshape(context)
		context.schemaPath = context.schemaPath[:len(context.schemaPath)-1]
		if err != nil {
			return nil, err
		}
		if newNode != nil {
			context.targetGraph.NewEdge(targetNode, newNode, ls.HasTerm, nil)
			empty = false
		}
	}
	if empty {
		ifEmpty := ls.AsPropertyValue(schemaNode.GetProperty(ReshapeTerms.IfEmpty))
		if ifEmpty != nil && ifEmpty.IsString() && ifEmpty.AsString() == "true" {
			return targetNode, nil
		}
		return nil, nil
	}
	return targetNode, nil
}

func (reshaper *Reshaper) value(context *ReshapeContext) (graph.Node, error) {
	schemaNode := context.CurrentSchemaNode()
	// Create a target node for this object node. If the object turns
	// out to be empty, this target node may be thrown away
	targetNode := context.targetGraph.NewNode([]string{ls.DocumentNodeTerm, ls.AttributeTypeValue}, nil)
	ls.SetNodeID(targetNode, reshaper.generateID(context.schemaPath, context.docPath))
	if reshaper.AddInstanceOfEdges {
		context.targetGraph.NewEdge(targetNode, schemaNode, ls.InstanceOfTerm, nil)
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
				v, _ := ls.GetNodeValue(sourceValue.Nodes.Slice()[0])
				ls.SetNodeValue(targetNode, v)
				empty = false
			case sourceValue.Nodes.Len() > 1:
				joinMethod := "join"
				prop := ls.AsPropertyValue(schemaNode.GetProperty(ReshapeTerms.JoinMethod))
				if prop != nil && prop.IsString() {
					joinMethod = prop.AsString()
				}
				joinDelimiter := " "
				prop = ls.AsPropertyValue(schemaNode.GetProperty(ReshapeTerms.JoinDelimiter))
				if prop != nil && prop.IsString() {
					joinDelimiter = prop.AsString()
				}
				result, err := JoinValues(sourceValue.Nodes.Slice(), joinMethod, joinDelimiter)
				if err != nil {
					return nil, err
				}
				ls.SetNodeValue(targetNode, result)
				empty = false
			}
		case gl.BoolValue, gl.NumberValue, gl.StringValue:
			str, err := source.AsString()
			if err != nil {
				return nil, err
			}
			ls.SetNodeValue(targetNode, str)
			empty = false
		}
	}

	if empty {
		ifEmpty := ls.AsPropertyValue(schemaNode.GetProperty(ReshapeTerms.IfEmpty))
		if ifEmpty != nil && ifEmpty.IsString() && ifEmpty.AsString() == "true" {
			return targetNode, nil
		}
		return nil, nil
	}
	return targetNode, nil
}

func getSource(context *ReshapeContext, schemaNode graph.Node) (gl.Value, error) {
	data, _ := schemaNode.GetProperty(ReshapeTerms.Source)
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

func (reshaper *Reshaper) setupVariables(context *ReshapeContext, schemaNode graph.Node) error {
	data, _ := schemaNode.GetProperty(ReshapeTerms.Vars)
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

func (reshaper *Reshaper) checkConditionals(context *ReshapeContext, schemaNode graph.Node) (bool, error) {
	data, _ := schemaNode.GetProperty(ReshapeTerms.If)
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
