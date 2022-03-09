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
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

// Compose schema layers. Directly modifies the source and the
// target. The source must be an overlay.
func (layer *Layer) Compose(context *Context, source *Layer) error {
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
	nodeMap := make(map[graph.Node]graph.Node)
	var err error
	// Process attributes of the source layer depth-first
	// Compose the source attribute nodes with the target attribute nodes, ignoring any nodes attached to them
	processedSourceNodes := make(map[graph.Node]struct{})
	source.ForEachAttribute(func(sourceNode graph.Node, sourcePath []graph.Node) bool {
		if _, processed := processedSourceNodes[sourceNode]; processed {
			return true
		}
		sourceID := GetAttributeID(sourceNode)
		if len(sourceID) == 0 {
			return true
		}
		// If node exists in target, merge
		targetNode, targetPath := layer.FindAttributeByID(sourceID)
		if targetNode != nil {
			// Target node exists. Merge if paths match
			if pathsMatch(targetPath, sourcePath) {
				if err = mergeNodes(context, layer, targetNode, sourceNode, processedSourceNodes); err != nil {
					return false
				}
				// Add any annotation subtrees
				nodeMap[sourceNode] = targetNode
				for edges := sourceNode.GetEdges(graph.OutgoingEdge); edges.Next(); {
					edge := edges.Edge()
					if IsAttributeTreeEdge(edge) {
						continue
					}
					graph.CopySubgraph(edge.GetTo(), layer.Graph, ClonePropertyValueFunc, nodeMap)
					graph.CopyEdge(edge, layer.Graph, ClonePropertyValueFunc, nodeMap)
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
			parentInLayer, _ := layer.FindAttributeByID(GetAttributeID(parent))
			if parentInLayer == nil {
				err = ErrInvalidComposition
				return false
			}

			newNode := CopySchemaNodeIntoGraph(layer.Graph, sourceNode)
			for edges := sourceNode.GetEdges(graph.IncomingEdge); edges.Next(); {
				edge := edges.Edge()
				if edge.GetFrom() == parent {
					layer.Graph.NewEdge(parentInLayer, newNode, edge.GetLabel(), CloneProperties(edge))
				}
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
func mergeNodes(context *Context, targetLayer *Layer, target, source graph.Node, processedSourceNodes map[graph.Node]struct{}) error {
	if _, processed := processedSourceNodes[source]; processed {
		return nil
	}
	processedSourceNodes[source] = struct{}{}
	if source == nil || target == nil {
		return nil
	}

	if err := ComposeProperties(context, target, source); err != nil {
		return err
	}
	return nil
}

// ComposeProperty composes targetValue and sourceValue for key
func ComposeProperty(context *Context, key string, targetValue, sourceValue *PropertyValue) (*PropertyValue, error) {
	newValue := targetValue
	newValue, err := GetComposerForTerm(key).Compose(newValue, sourceValue)
	if err != nil {
		return nil, ErrTerm{Term: key, Err: err}
	}
	return newValue, nil
}

// ComposeProperties will combine the properties in source to
// target. The target properties will be modified directly
func ComposeProperties(context *Context, target, source graph.Node) error {
	var retErr error
	source.ForEachProperty(func(key string, value interface{}) bool {
		if p, ok := value.(*PropertyValue); ok {
			tp, _ := target.GetProperty(key)
			targetProperty, _ := tp.(*PropertyValue)
			newValue, err := ComposeProperty(context, key, targetProperty, p)
			if err != nil {
				retErr = err
				return false
			}
			target.SetProperty(key, newValue)
		}
		return true
	})
	return retErr
}

// pathsMatch returns true if the attribute predecessors of source matches target's
func pathsMatch(targetPath, sourcePath []graph.Node) bool {
	tn := len(targetPath)
	sn := len(sourcePath)
	for {
		if tn == 0 {
			return true
		}
		if sn == 0 {
			return false
		}
		if sourcePath[sn-1].GetLabels().Has(SchemaTerm) || sourcePath[sn-1].GetLabels().Has(OverlayTerm) {
			return true
		}
		if targetPath[tn-1].GetLabels().Has(SchemaTerm) || targetPath[tn-1].GetLabels().Has(OverlayTerm) {
			return false
		}
		if GetAttributeID(targetPath[tn-1]) != GetAttributeID(sourcePath[sn-1]) {
			return false
		}
		tn--
		sn--
	}
}
