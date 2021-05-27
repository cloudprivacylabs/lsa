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
package layers

import (
	"fmt"

	"github.com/bserdar/digraph"
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
		compatible := false
		for _, t := range sourceTypes {
			found := false
			for _, x := range layerTypes {
				if x == t {
					found = true
					break
				}
			}
			if found {
				compatible = true
				break
			}
		}
		if !compatible {
			return ErrIncompatibleComposition
		}
	}

	var err error
	processedSourceNodes := make(map[*digraph.Node]struct{})
	// Process attributes depth-first
	source.ForEachAttribute(func(sourceNode *digraph.Node) bool {
		if _, processed := processedSourceNodes[sourceNode]; processed {
			return true
		}
		// If node exists in target, merge
		targetNodes := layer.AllNodesWithLabel(sourceNode.Label()).All()
		switch len(targetNodes) {
		case 1:
			// Target node exists. Merge if paths match
			if pathsMatch(targetNodes[0], sourceNode, sourceNode) {
				if err = mergeNodes(layer.Graph, targetNodes[0], sourceNode, options, processedSourceNodes); err != nil {
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
func mergeNodes(targetGraph *digraph.Graph, target, source *digraph.Node, options ComposeOptions, processedSourceNodes map[*digraph.Node]struct{}) error {
	if _, processed := processedSourceNodes[source]; processed {
		return nil
	}
	processedSourceNodes[source] = struct{}{}
	sourcePayload, _ := source.Payload.(*SchemaNode)
	targetPayload, _ := target.Payload.(*SchemaNode)
	if sourcePayload == nil || targetPayload == nil {
		return nil
	}
	// Merge properties
	for k, v := range sourcePayload.Properties {
		var err error
		targetPayload.Properties[k], err = mergeProperty(k, targetPayload.Properties[k], v, options)
		if err != nil {
			return err
		}
	}
	// Merge graphs

	// Map of source node -> target node, so that we know the target node created for each source node
	nodeMap := map[*digraph.Node]*digraph.Node{}
	return mergeGraphs(targetGraph, target, source, nodeMap)
}

func mergeGraphs(targetGraph *digraph.Graph, targetNode, sourceNode *digraph.Node, nodeMap map[*digraph.Node]*digraph.Node) error {
	// If the source node is already seen, return
	if _, processed := nodeMap[sourceNode]; processed {
		return nil
	}
	for edges := sourceNode.AllOutgoingEdges(); edges.HasNext(); {
		edge := edges.Next()
		// Skip all attribute nodes, as they will be processed later
		if IsAttributeTreeEdge(edge) {
			continue
		}
		var targetNodes []*digraph.Node
		if edge.To().Label() != nil {
			targetNodes = targetGraph.AllNodesWithLabel(edge.To().Label()).All()
			if len(targetNodes) > 1 {
				return ErrDuplicateNodeID(fmt.Sprint(edge.To().Label()))
			}
			if len(targetNodes) == 0 {
				targetNodes = []*digraph.Node{targetGraph.NewNode(edge.To().Label(), edge.To().Payload)}
			}
		} else {
			targetNodes = []*digraph.Node{targetGraph.NewNode(nil, edge.To().Payload)}
		}
		nodeMap[edge.To()] = targetNodes[0]
		targetGraph.NewEdge(targetNode, targetNodes[0], edge.Label(), edge.Payload)
		if err := mergeGraphs(targetGraph, targetNodes[0], edge.To(), nodeMap); err != nil {
			return err
		}
	}
	return nil
}

func mergeProperty(property string, existingValue, newValue interface{}, options ComposeOptions) (interface{}, error) {
	return SetUnion(existingValue, newValue), nil
}

func SetUnion(v1, v2 interface{}) interface{} {
	if v1 == nil {
		return v2
	}
	if v2 == nil {
		return v1
	}
	switch e := v1.(type) {
	case []interface{}:
		values := make(map[interface{}]struct{})
		for _, k := range e {
			values[k] = struct{}{}
		}
		ret := e
		if n, ok := v2.([]interface{}); ok {
			for _, item := range n {
				if _, exists := values[item]; !exists {
					values[item] = struct{}{}
					ret = append(ret, item)
				}
			}
			return ret
		}
		if _, exists := values[v2]; !exists {
			return append(e, v2)
		}
		return e
	default:
		ret := []interface{}{e}
		if n, ok := v2.([]interface{}); ok {
			for _, item := range n {
				if item != e {
					ret = append(ret, item)
				}
			}
			if len(ret) == 1 {
				return ret[0]
			}
			return ret
		}
		if e != v2 {
			return []interface{}{e, v2}
		}
		return e
	}
}

// pathsMatch returns true if the attribute predecessors of source matches target's
func pathsMatch(target, source, initialSource *digraph.Node) bool {
	if source.Label() != target.Label() {
		return false
	}
	sourceParent := GetParentAttribute(source)
	// If sourceParents reached the top level, stop
	if sourceParent == nil {
		return true
	}
	payload := sourceParent.Payload.(*SchemaNode)
	if payload.HasType(SchemaTerm) || payload.HasType(OverlayTerm) {
		return true
	}
	targetParent := GetParentAttribute(target)
	if targetParent == nil {
		return false
	}
	payload = targetParent.Payload.(*SchemaNode)
	if payload.HasType(SchemaTerm) || payload.HasType(OverlayTerm) {
		return false
	}
	// Loop?
	if sourceParent == initialSource {
		return false
	}
	return pathsMatch(targetParent, sourceParent, initialSource)
}
