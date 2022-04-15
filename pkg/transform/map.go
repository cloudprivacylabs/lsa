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
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

// Mapper reads nodes from a source graph and creates a target graph
// based on a schema. This is used for mapping data across
// terminologies.
//
// The nodes are selected based on a given node
// term. The node term in the source graph gives the target schema
// attributes IDs. For example, if the source graph nodes has properties
//
//     term: id
//
// then, the mapper reads the node with `term`, finds the target
// schema attribute using `id`, and copies the value.
type Mapper struct {
	ls.Ingester

	// PropertyName specifies the property to lookup in the source
	// graph. The property values give the target schema attribute IDs.
	// Use PropertyName:ls:schemaNodeId to use the same schema
	PropertyName string
}

type mapContext struct {
	ls.IngestionContext
	sourceGraph graph.Graph
	// contextNode keeps the root node in source graph under which to
	// search for tags. If nil, search is done on the whole graph
	contextNode  graph.Node
	propertyName string
}

func (ctx *mapContext) find(attrID string) ([]graph.Node, error) {
	if ctx.contextNode == nil {
		pattern := graph.Pattern{
			{
				Labels: graph.NewStringSet(ls.DocumentNodeTerm),
				Properties: map[string]interface{}{
					ctx.propertyName: ls.StringPropertyValue(attrID),
				},
			}}
		return pattern.FindNodes(ctx.sourceGraph, nil)
	}
	pattern := graph.Pattern{
		{
			Name: "n",
		},
		{
			Max: -1,
		},
		{
			Labels: graph.NewStringSet(ls.DocumentNodeTerm),
			Properties: map[string]interface{}{
				ctx.propertyName: ls.StringPropertyValue(attrID),
			},
		}}
	nodes := &graph.PatternSymbol{}
	nodes.Add(ctx.contextNode)
	paths, err := pattern.FindPaths(ctx.sourceGraph, map[string]*graph.PatternSymbol{"n": nodes})
	if err != nil {
		return nil, err
	}
	return paths.GetTailNodes(), nil
}

// Map builds a new graph conforming to the target schema. The
// values for the target schema nodes are taken from the source graph
// nodes that match the mapper.PropertyName
func (mapper *Mapper) Map(lsContext *ls.Context, sourceGraph graph.Graph) error {
	ictx := mapper.Start(lsContext, "")
	ctx := mapContext{
		IngestionContext: ictx,
		sourceGraph:      sourceGraph,
		propertyName:     mapper.PropertyName,
	}

	mapper.IngestEmptyValues = true
	roots, err := mapper.buildTarget(&ctx)
	if err != nil {
		return err
	}
	if len(roots) == 1 {
		return mapper.Finish(ctx.IngestionContext, roots[0])
	}
	return mapper.Finish(ctx.IngestionContext, nil)
}

func (mapper *Mapper) buildTarget(ctx *mapContext) ([]graph.Node, error) {
	// Get the schema node to build
	schemaNode := ctx.GetSchemaNode()
	ctx.GetLogger().Debug(map[string]interface{}{"mth": "map.buildTarget", "schemaNode": schemaNode})
	// Find the source nodes
	sourceNodes, err := ctx.find(ls.GetNodeID(schemaNode))
	if err != nil {
		return nil, err
	}
	ctx.GetLogger().Debug(map[string]interface{}{"mth": "map.buildTarget", "sourceNodes": len(sourceNodes)})
	if len(sourceNodes) == 0 {
		return nil, nil
	}
	roots := make([]graph.Node, 0)
	// Create one target node for each instance of the source node
	for _, sourceNode := range sourceNodes {
		// Create the target node
		switch {
		case schemaNode.GetLabels().Has(ls.AttributeTypeValue):
			ctx.GetLogger().Debug(map[string]interface{}{"mth": "map.buildTarget", "sourceValueNode": sourceNode})
			value, err := ls.GetNodeValue(sourceNode)
			if err != nil {
				return nil, err
			}
			_, _, nd, err := mapper.Value(ctx.IngestionContext, "")
			if err != nil {
				return nil, err
			}
			// TODO: Node may not have a value https://github.com/cloudprivacylabs/lsa/issues/15
			ls.SetNodeValue(nd, value)
			roots = append(roots, nd)
		case schemaNode.GetLabels().Has(ls.AttributeTypeArray):
			_, _, arrayNode, err := mapper.Array(ctx.IngestionContext)
			if err != nil {
				return nil, err
			}
			roots = append(roots, arrayNode)
			el := ls.GetArrayElementNode(schemaNode)
			if el != nil {
				ocn := ctx.contextNode
				oc := ctx.IngestionContext
				ctx.contextNode = sourceNode
				ctx.IngestionContext = ctx.NewLevel(arrayNode)
				_, err = mapper.buildTarget(ctx)
				ctx.contextNode = ocn
				ctx.IngestionContext = oc
				if err != nil {
					return nil, err
				}
			}
		case schemaNode.GetLabels().Has(ls.AttributeTypeObject):
			_, _, objectNode, err := mapper.Object(ctx.IngestionContext)
			if err != nil {
				return nil, err
			}
			roots = append(roots, objectNode)
			oc := ctx.IngestionContext
			ic := ctx.NewLevel(objectNode)
			ocn := ctx.contextNode
			ctx.contextNode = sourceNode

			children := ls.GetObjectAttributeNodes(schemaNode)
			for _, child := range children {
				old := ctx.IngestionContext
				an := ls.AsPropertyValue(child.GetProperty(ls.AttributeNameTerm)).AsString()
				ctx.IngestionContext = ic.New(an, child)
				_, err := mapper.buildTarget(ctx)
				ctx.IngestionContext = old
				if err != nil {
					return nil, err
				}
			}
			ctx.contextNode = ocn
			ctx.IngestionContext = oc
		}
	}

	return roots, nil
}
