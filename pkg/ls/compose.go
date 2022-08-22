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
	"strings"

	"github.com/cloudprivacylabs/opencypher/graph"
)

// Compose schema layers. Directly modifies the source and the
// target. The source must be an overlay.
func (layer *Layer) Compose(context *Context, source *Layer) error {
	if source.GetLayerType() != OverlayTerm {
		return ErrCompositionSourceNotOverlay
	}
	// Check if target types are compatible. If they are non-empty, they must be the same
	layerType := layer.GetValueType()
	sourceType := source.GetValueType()
	if len(layerType) > 0 && len(sourceType) > 0 {
		if layerType != sourceType {
			return ErrIncompatibleComposition
		}
	}
	sourceCompose := AsPropertyValue(source.GetLayerRootNode().GetProperty(ComposeTerm)).AsString()
	nodeMap := make(map[graph.Node]graph.Node)
	nsMap, err := GetNSMap(AsPropertyValue(source.GetLayerRootNode().GetProperty(NSMapTerm)).MustStringSlice())
	if err != nil {
		return err
	}
	if len(nsMap) > 0 {
		layer.ForEachAttribute(func(node graph.Node, _ []graph.Node) bool {
			id := GetNodeID(node)
			for _, m := range nsMap {
				if strings.HasPrefix(id, m[0]) {
					SetNodeID(node, m[1]+id[len(m[0]):])
				}
			}
			node.ForEachProperty(func(key string, value interface{}) bool {
				if key == NodeValueTerm {
					return true
				}
				pv, ok := value.(*PropertyValue)
				if !ok {
					return true
				}

				if pv.IsString() {
					s := pv.AsString()
					for _, m := range nsMap {
						if strings.HasPrefix(s, m[0]) {
							node.SetProperty(key, StringPropertyValue(key, m[1]+s[len(m[0]):]))
						}
					}
				}
				if pv.IsStringSlice() {
					slice := pv.AsStringSlice()
					newSlice := make([]string, len(slice))
					changed := false
					for i := range slice {
						s := slice[i]
						for _, m := range nsMap {
							if strings.HasPrefix(s, m[0]) {
								newSlice[i] = m[1] + s[len(m[0]):]
								changed = true
							} else {
								newSlice[i] = s
							}
						}
					}
					if changed {
						node.SetProperty(key, StringSlicePropertyValue(key, newSlice))
					}
				}
				return true
			})
			return true
		})
	}

	copySubtree := func(targetNode, sourceNode graph.Node) {
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

	processedSourceNodes := make(map[graph.Node]struct{})
	// Process overlay attributes first
	targetOverlayAttrs := make(map[string]graph.Node)
	sourceOverlayAttrs := make(map[string]graph.Node)
	for _, x := range layer.GetOverlayAttributes() {
		targetOverlayAttrs[GetNodeID(x)] = x
	}
	for _, x := range source.GetOverlayAttributes() {
		sourceOverlayAttrs[GetNodeID(x)] = x
	}
	for srcId, srcAttr := range sourceOverlayAttrs {
		if tgt, ok := targetOverlayAttrs[srcId]; ok {
			// Compose target
			if err = mergeNodes(context, layer, tgt, srcAttr, sourceCompose, processedSourceNodes); err != nil {
				return err
			}
			copySubtree(tgt, srcAttr)
		}
		targetNode, _ := layer.FindAttributeByID(srcId)
		if targetNode == nil {
			continue
		}
		if err = mergeNodes(context, layer, targetNode, srcAttr, sourceCompose, processedSourceNodes); err != nil {
			return err
		}
		copySubtree(targetNode, srcAttr)
	}

	// Process attributes of the source layer depth-first
	// Compose the source attribute nodes with the target attribute nodes, ignoring any nodes attached to them
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
				if err = mergeNodes(context, layer, targetNode, sourceNode, sourceCompose, processedSourceNodes); err != nil {
					return false
				}
				// Add any annotation subtrees
				copySubtree(targetNode, sourceNode)
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
func mergeNodes(context *Context, targetLayer *Layer, target, source graph.Node, sourceCompose string, processedSourceNodes map[graph.Node]struct{}) error {
	if _, processed := processedSourceNodes[source]; processed {
		return nil
	}
	processedSourceNodes[source] = struct{}{}
	if source == nil || target == nil {
		return nil
	}
	// Apply labels
	s := target.GetLabels()
	s.AddSet(source.GetLabels())
	target.SetLabels(s)

	if len(sourceCompose) > 0 {
		cType := CompositionType(sourceCompose)
		var retErr error
		source.ForEachProperty(func(key string, value interface{}) bool {
			if p, ok := value.(*PropertyValue); ok {
				tp, _ := target.GetProperty(key)
				targetProperty, _ := tp.(*PropertyValue)
				newValue, err := cType.Compose(targetProperty, p)
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

type ErrInvalidNSMapExpression string

func (e ErrInvalidNSMapExpression) Error() string { return "Invalid nsMap: " + string(e) }

// ParseNSMap parses a string pair of the form
//
//    string1 -> string2
//
// This is used in nsMap expression to specify namespace (prefix)
// mapping for node ids.
func ParseNSMap(in string) (string, string, error) {
	items := strings.Split(in, "->")
	if len(items) != 2 {
		return "", "", ErrInvalidNSMapExpression(in)
	}
	return strings.TrimSpace(items[0]), strings.TrimSpace(items[1]), nil
}

// GetNSMap parses the namespace map and returns the mapping
func GetNSMap(in []string) ([][]string, error) {
	ret := make([][]string, 0, len(in))
	for _, x := range in {
		f, t, err := ParseNSMap(x)
		if err != nil {
			return nil, err
		}
		ret = append(ret, []string{f, t})
	}
	return ret, nil
}
