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
	// Check if target types are compatible. If they are non-empty, they must be the same
	layerType := layer.GetTargetType()
	sourceType := source.GetTargetType()
	if len(layerType) > 0 && len(sourceType) > 0 {
		if layerType != sourceType {
			return ErrIncompatibleComposition
		}
	}

	var err error
	// Process attributes of the source layer depth-first
	// Compose the source attribute nodes with the target attribute nodes, ignoring any nodes attached to them
	processedSourceNodes := make(map[LayerNode]struct{})
	source.ForEachAttribute(func(sourceNode LayerNode) bool {
		if _, processed := processedSourceNodes[sourceNode]; processed {
			return true
		}
		// If node exists in target, merge
		targetNodes := layer.AllNodesWithLabel(sourceNode.Label()).All()
		switch len(targetNodes) {
		case 1:
			// Target node exists. Merge if paths match
			if pathsMatch(targetNodes[0].(LayerNode), sourceNode, sourceNode) {
				if err = mergeNodes(layer, targetNodes[0].(LayerNode), sourceNode, processedSourceNodes); err != nil {
					return false
				}
			}

		case 0:
			// Target node does not exist.
			// Parent node must exist, because this is a depth-first algorithm
			parent, edge := GetParentAttribute(sourceNode)
			if parent == nil {
				err = ErrInvalidComposition
				return false
			}
			parentInLayer := layer.AllNodesWithLabel(parent.Label()).All()
			switch len(parentInLayer) {
			case 0:
				err = ErrInvalidComposition
				return false
			case 1:
				// Add the same node to this layer
				newNode := sourceNode.Clone()
				layer.AddNode(newNode)
				layer.AddEdge(parentInLayer[0], newNode, edge.Clone())
			default:
				err = ErrDuplicateAttributeID(fmt.Sprint(sourceNode.Label()))
				return false
			}
		default:
			err = ErrDuplicateAttributeID(fmt.Sprint(sourceNode.Label()))
			return false
		}
		processedSourceNodes[sourceNode] = struct{}{}
		return true
	})
	if err != nil {
		return err
	}
	// Copy all non-attribute nodes of source to target
	nodeMap := make(map[LayerNode]LayerNode)
	seen := make(map[LayerNode]struct{})
	source.ForEachAttribute(func(sourceNode LayerNode) bool {
		targetNodes := layer.AllNodesWithLabel(sourceNode.Label()).All()
		if len(targetNodes) != 1 {
			// This should not really happen
			panic("Cannot find node even after adding")
		}
		mergeNonattributeGraph(layer, targetNodes[0].(LayerNode), sourceNode, seen, nodeMap)
		return true
	})
	if err != nil {
		return err
	}
	return nil
}

// Merge source into target.
func mergeNodes(targetLayer *Layer, target, source LayerNode, processedSourceNodes map[LayerNode]struct{}) error {
	if _, processed := processedSourceNodes[source]; processed {
		return nil
	}
	processedSourceNodes[source] = struct{}{}
	if source == nil || target == nil {
		return nil
	}

	if err := ComposeProperties(target.GetPropertyMap(), source.GetPropertyMap()); err != nil {
		return err
	}
	return nil
}

func mergeNonattributeGraph(targetLayer *Layer, targetNode, sourceNode LayerNode, seen map[LayerNode]struct{}, nodeMap map[LayerNode]LayerNode) {
	// If the source node is already seen, return
	if _, processed := seen[sourceNode]; processed {
		return
	}
	for edges := sourceNode.AllOutgoingEdges(); edges.HasNext(); {
		edge := edges.Next().(LayerEdge)
		// Skip all attribute nodes
		if edge.IsAttributeTreeEdge() {
			continue
		}

		toNode := edge.To().(LayerNode)
		seen[toNode] = struct{}{}
		// If edge.to is not in target, add it
		targetTo, exists := nodeMap[toNode]
		if !exists {
			targetTo = toNode.Clone()
			targetLayer.AddNode(targetTo)
			nodeMap[toNode] = targetTo
		}
		// Connect the nodes
		targetLayer.AddEdge(targetNode, targetTo, edge.Clone())
		mergeNonattributeGraph(targetLayer, targetTo, toNode, seen, nodeMap)
	}
}

// ComposeProperties will combine the properties in source to
// target. The target properties will be modified directly
func ComposeProperties(target, source map[string]*PropertyValue) error {
	for k, v := range source {
		newValue, _ := target[k]
		newValue, err := GetComposerForTerm(k).Compose(newValue, v)
		if err != nil {
			return ErrTerm{Term: k, Err: err}
		}
		target[k] = newValue
	}
	return nil
}

// pathsMatch returns true if the attribute predecessors of source matches target's
func pathsMatch(target, source, initialSource LayerNode) bool {
	if source.GetID() != target.GetID() {
		return false
	}
	sourceParent, _ := GetParentAttribute(source)
	// If sourceParents reached the top level, stop
	if sourceParent == nil {
		return true
	}
	if sourceParent.HasType(SchemaTerm) || sourceParent.HasType(OverlayTerm) {
		return true
	}
	targetParent, _ := GetParentAttribute(target)
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
