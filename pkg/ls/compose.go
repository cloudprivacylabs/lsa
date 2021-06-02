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
package ls

import (
	"fmt"

	"github.com/cloudprivacylabs/lsa/pkg/term"
)

type ComposeOptions struct {
	// While composing an object with layer1 and layer2, if layer2 has
	// attributes missing in layer1, then add those attributes to the
	// result. By default, the result will only have attributes of
	// layer1
	Union bool
}

// Compose schema layers. Directly modifies the target. The source
// must be an overlay.
func (layer *Layer) Compose(options ComposeOptions, source *Layer) error {
	if source.GetLayerType() != OverlayTerm {
		return ErrCompositionSourceNotOverlay
	}
	layerTypes := layer.GetTargetTypes()
	sourceTypes := source.GetTargetTypes()
	if len(layerTypes) > 0 && len(sourceTypes) > 0 {
		intersection := StringSetIntersection(layerTypes, sourceTypes)
		if len(intersection) == 0 {
			return ErrIncompatibleComposition
		}
	}

	var err error
	processedSourceNodes := make(map[*SchemaNode]struct{})
	// Process attributes depth-first
	source.ForEachAttribute(func(sourceNode *SchemaNode) bool {
		if _, processed := processedSourceNodes[sourceNode]; processed {
			return true
		}
		// If node exists in target, merge
		targetNodes := layer.AllNodesWithLabel(sourceNode.Label()).All()
		switch len(targetNodes) {
		case 1:
			// Target node exists. Merge if paths match
			if pathsMatch(targetNodes[0].(*SchemaNode), sourceNode, sourceNode) {
				if err = mergeNodes(layer, targetNodes[0].(*SchemaNode), sourceNode, options, processedSourceNodes); err != nil {
					return false
				}
			}
		case 0:
			// Target node does not exist. Add if options==Union
			if options.Union {

			}
		default:
			err = ErrDuplicateAttributeID(fmt.Sprint(sourceNode.Label()))
			return false
		}
		processedSourceNodes[sourceNode] = struct{}{}
		return true
	})
	return nil
}

// Merge source into target.
func mergeNodes(targetLayer *Layer, target, source *SchemaNode, options ComposeOptions, processedSourceNodes map[*SchemaNode]struct{}) error {
	if _, processed := processedSourceNodes[source]; processed {
		return nil
	}
	processedSourceNodes[source] = struct{}{}
	if source == nil || target == nil {
		return nil
	}
	// Merge properties
	for k, v := range source.Properties {
		var err error
		target.Properties[k], err = mergeProperty(k, target.Properties[k], v, options)
		if err != nil {
			return err
		}
	}
	// Merge graphs

	// Map of source node -> target node, so that we know the target node created for each source node
	nodeMap := map[*SchemaNode]*SchemaNode{}
	return mergeGraphs(targetLayer, target, source, nodeMap)
}

func mergeGraphs(targetLayer *Layer, targetNode, sourceNode *SchemaNode, nodeMap map[*SchemaNode]*SchemaNode) error {
	// If the source node is already seen, return
	if _, processed := nodeMap[sourceNode]; processed {
		return nil
	}
	for edges := sourceNode.AllOutgoingEdges(); edges.HasNext(); {
		edge := edges.Next().(*SchemaEdge)
		// Skip all attribute nodes, as they will be processed later
		if edge.IsAttributeTreeEdge() {
			continue
		}
		var targetNodes []digraph.Node
		if edge.To().Label() != nil {
			targetNodes = targetGraph.AllNodesWithLabel(edge.To().Label()).All()
			if len(targetNodes) > 1 {
				return ErrDuplicateNodeID(fmt.Sprint(edge.To().Label()))
			}
			if len(targetNodes) == 0 {
				newNode := edge.To().(*SchemaNode).Clone()
				targetGraph.AddNode(newNode)
				targetNodes = []digraph.Node{newNode}
			}
		} else {
			newNode := edge.To().(*SchemaNode).Clone()
			newNode.SetLabel(nil)
			targetGraph.AddNode(newNode)
			targetNodes = []digraph.Node{newNode}
		}
		nodeMap[edge.To().(*SchemaNode)] = targetNodes[0].(*SchemaNode)
		targetGraph.AddEdge(targetNode, targetNodes[0], edge.Clone())
		if err := mergeGraphs(targetGraph, targetNodes[0].(*SchemaNode), edge.To().(*SchemaNode), nodeMap); err != nil {
			return err
		}
	}
	return nil
}

func mergeProperty(property string, existingValue, newValue interface{}, options ComposeOptions) (interface{}, error) {
	return term.GetComposer(term.GetTermMeta(property)).Compose(existingValue, newValue)
}

// pathsMatch returns true if the attribute predecessors of source matches target's
func pathsMatch(target, source, initialSource *SchemaNode) bool {
	if source.GetID() != target.GetID() {
		return false
	}
	sourceParent := source.GetParentAttribute()
	// If sourceParents reached the top level, stop
	if sourceParent == nil {
		return true
	}
	if sourceParent.HasType(SchemaTerm) || sourceParent.HasType(OverlayTerm) {
		return true
	}
	targetParent := target.GetParentAttribute()
	if targetParent == nil {
		return false
	}
	if targetParent.HasType(SchemaTerm) || targetParent.HasType(OverlayTerm) {
		return false
	}
	// Loop?
	if sourceParent == initialSource {
		return false
	}
	return pathsMatch(targetParent, sourceParent, initialSource)
}
