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
	"github.com/bserdar/digraph"
)

// Compose schema layers. Directly modifies the source and the
// target. The source must be an overlay.
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
	processedSourceNodes := make(map[Node]struct{})
	source.ForEachAttribute(func(sourceNode Node, sourcePath []Node) bool {
		if _, processed := processedSourceNodes[sourceNode]; processed {
			return true
		}
		// If node exists in target, merge
		targetNode, targetPath := layer.FindAttributeByID(sourceNode.GetID())
		if targetNode != nil {
			// Target node exists. Merge if paths match
			if pathsMatch(targetPath, sourcePath) {
				if err = mergeNodes(layer, targetNode, sourceNode, processedSourceNodes); err != nil {
					return false
				}
			}
		} else {
			// Target node does not exist.
			// Parent node must exist, because this is a depth-first algorithm
			if len(sourcePath) <= 1 {
				err = ErrInvalidComposition
				return false
			}
			parent := sourcePath[len(sourcePath)-2]
			parentInLayer, _ := layer.FindAttributeByID(parent.GetID())
			if parentInLayer == nil {
				err = ErrInvalidComposition
				return false
			}
			edge := GetLayerEdgeBetweenNodes(parent, sourceNode)
			if edge != nil {
				digraph.Connect(parentInLayer, sourceNode, edge.Clone())
			}
		}
		processedSourceNodes[sourceNode] = struct{}{}
		return true
	})
	if err != nil {
		return err
	}
	return nil
}

// Merge source into target.
func mergeNodes(targetLayer *Layer, target, source Node, processedSourceNodes map[Node]struct{}) error {
	if _, processed := processedSourceNodes[source]; processed {
		return nil
	}
	processedSourceNodes[source] = struct{}{}
	if source == nil || target == nil {
		return nil
	}

	if err := ComposeProperties(target.GetProperties(), source.GetProperties()); err != nil {
		return err
	}
	return nil
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
func pathsMatch(targetPath, sourcePath []Node) bool {
	tn := len(targetPath)
	sn := len(sourcePath)
	for {
		if tn == 0 {
			return true
		}
		if sn == 0 {
			return false
		}
		if sourcePath[sn-1].GetTypes().Has(SchemaTerm) || sourcePath[sn-1].GetTypes().Has(OverlayTerm) {
			return true
		}
		if targetPath[tn-1].GetTypes().Has(SchemaTerm) || targetPath[tn-1].GetTypes().Has(OverlayTerm) {
			return false
		}
		if targetPath[tn-1].GetID() != sourcePath[sn-1].GetID() {
			return false
		}
		tn--
		sn--
	}
}
