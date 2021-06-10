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
)

// Compose schema layers. Directly modifies the target. The source
// must be an overlay.
func (layer *Layer) Compose(source *Layer) error {
	if source.GetLayerType() != OverlayTerm {
		return ErrCompositionSourceNotOverlay
	}
	// Check if target types are compatible. If they are non-empty, then
	// intersection must be non-empty
	layerTypes := layer.GetTargetTypes()
	sourceTypes := source.GetTargetTypes()
	if len(layerTypes) > 0 && len(sourceTypes) > 0 {
		intersection := StringSetIntersection(layerTypes, sourceTypes)
		if len(intersection) == 0 {
			return ErrIncompatibleComposition
		}
	}

	var err error
	// Process attributes of the source layer depth-first
	// Compose the source attribute nodes with the target attribute nodes, ignoring any nodes attached to them
	processedSourceNodes := make(map[*SchemaNode]struct{})
	source.ForEachAttribute(func(sourceNode *SchemaNode) bool {
		if _, processed := processedSourceNodes[sourceNode]; processed {
			return true
		}
		// If node exists in target, merge
		// If this is the root node, match directly
		if sourceNode == source.GetRoot() {
			if err = mergeNodes(layer, layer.GetRoot(), sourceNode, processedSourceNodes); err != nil {
				return false
			}
		} else {
			targetNodes := layer.AllNodesWithLabel(sourceNode.Label()).All()
			switch len(targetNodes) {
			case 1:
				// Target node exists. Merge if paths match
				if pathsMatch(targetNodes[0].(*SchemaNode), sourceNode, sourceNode) {
					if err = mergeNodes(layer, targetNodes[0].(*SchemaNode), sourceNode, processedSourceNodes); err != nil {
						return false
					}
				}

			case 0:
				// Target node does not exist.
				// Parent node must exist, because this is a depth-first algorithm
				parent, edge := sourceNode.GetParentAttribute()
				if parent == nil {
					err = ErrInvalidComposition
					return false
				}
				// Add the same node to this layer
				newNode := sourceNode.Clone()
				layer.AddEdge(parent, newNode, edge.Clone())

			default:
				err = ErrDuplicateAttributeID(fmt.Sprint(sourceNode.Label()))
				return false
			}
		}
		processedSourceNodes[sourceNode] = struct{}{}
		return true
	})
	if err != nil {
		return err
	}
	// Copy all non-attribute nodes of source to target
	nodeMap := make(map[*SchemaNode]*SchemaNode)
	seen := make(map[*SchemaNode]struct{})
	source.ForEachAttribute(func(sourceNode *SchemaNode) bool {
		if sourceNode == source.GetRoot() {
			mergeNonattributeGraph(layer, layer.GetRoot(), sourceNode, seen, nodeMap)
		} else {
			targetNodes := layer.AllNodesWithLabel(sourceNode.Label()).All()
			if len(targetNodes) != 1 {
				// This should not really happen
				panic("Cannot find node even after adding")
			}
			mergeNonattributeGraph(layer, targetNodes[0].(*SchemaNode), sourceNode, seen, nodeMap)
		}
		return true
	})
	if err != nil {
		return err
	}
	return nil
}

// Merge source into target.
func mergeNodes(targetLayer *Layer, target, source *SchemaNode, processedSourceNodes map[*SchemaNode]struct{}) error {
	if _, processed := processedSourceNodes[source]; processed {
		return nil
	}
	processedSourceNodes[source] = struct{}{}
	if source == nil || target == nil {
		return nil
	}

	if err := ComposeProperties(target.Properties, source.Properties); err != nil {
		return err
	}
	return nil
}

func mergeNonattributeGraph(targetLayer *Layer, targetNode, sourceNode *SchemaNode, seen map[*SchemaNode]struct{}, nodeMap map[*SchemaNode]*SchemaNode) {
	// If the source node is already seen, return
	if _, processed := seen[sourceNode]; processed {
		return
	}
	for edges := sourceNode.AllOutgoingEdges(); edges.HasNext(); {
		edge := edges.Next().(*SchemaEdge)
		// Skip all attribute nodes
		if edge.IsAttributeTreeEdge() {
			continue
		}

		toNode := edge.To().(*SchemaNode)
		seen[toNode] = struct{}{}
		// If edge.to is not in target, add it
		targetTo, exists := nodeMap[toNode]
		if !exists {
			targetTo = toNode.Clone()
			nodeMap[toNode] = targetTo
		}
		// Connect the nodes
		targetLayer.AddEdge(targetNode, targetTo, edge.Clone())
		mergeNonattributeGraph(targetLayer, targetTo, toNode, seen, nodeMap)
	}
}

// ComposeProperties will combine the properties in source to
// target. The target properties will be modified directly
func ComposeProperties(target, source map[string]interface{}) error {
	for k, v := range source {
		newValue, _ := target[k]
		newValue, err := GetComposerForTerm(k).Compose(v, newValue)
		if err != nil {
			return ErrTerm{Term: k, Err: err}
		}
		target[k] = newValue
	}
	return nil
}

// pathsMatch returns true if the attribute predecessors of source matches target's
func pathsMatch(target, source, initialSource *SchemaNode) bool {
	if source.GetID() != target.GetID() {
		return false
	}
	sourceParent, _ := source.GetParentAttribute()
	// If sourceParents reached the top level, stop
	if sourceParent == nil {
		return true
	}
	if sourceParent.HasType(SchemaTerm) || sourceParent.HasType(OverlayTerm) {
		return true
	}
	targetParent, _ := target.GetParentAttribute()
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
