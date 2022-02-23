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

package graph

func copyLabels(in StringSet) []string {
	srcLabels := in.Slice()
	labels := make([]string, len(srcLabels))
	copy(labels, srcLabels)
	return labels
}

type withProperties interface {
	ForEachProperty(func(string, interface{}) bool) bool
}

func copyProperties(in withProperties, cloneProperty func(string, interface{}) interface{}) map[string]interface{} {
	properties := make(map[string]interface{})
	in.ForEachProperty(func(key string, value interface{}) bool {
		properties[key] = cloneProperty(key, value)
		return true
	})
	return properties
}

// Copy a graph
func CopyGraph(source, target Graph, clonePropertyFunc func(string, interface{}) interface{}) map[Node]Node {
	nodeMap := make(map[Node]Node)
	for nodes := source.GetNodes(); nodes.Next(); {
		node := nodes.Node()
		nodeMap[node] = target.NewNode(copyLabels(node.GetLabels()), copyProperties(node, clonePropertyFunc))
	}
	for edges := source.GetEdges(); edges.Next(); {
		edge := edges.Edge()
		properties := make(map[string]interface{})
		edge.ForEachProperty(func(key string, value interface{}) bool {
			properties[key] = clonePropertyFunc(key, value)
			return true
		})
		target.NewEdge(nodeMap[edge.GetFrom()], nodeMap[edge.GetTo()], edge.GetLabel(), properties)
	}
	return nodeMap
}

// CopySubgraph copies all nodes that are accessible from sourceNode to the target graph
func CopySubgraph(sourceNode Node, target Graph, clonePropertyFunc func(string, interface{}) interface{}, nodeMap map[Node]Node) {
	if _, ok := nodeMap[sourceNode]; ok {
		return
	}
	nodeMap[sourceNode] = CopyNode(sourceNode, target, clonePropertyFunc)
	for edges := sourceNode.GetEdges(OutgoingEdge); edges.Next(); {
		edge := edges.Edge()
		CopySubgraph(edge.GetTo(), target, clonePropertyFunc, nodeMap)
		CopyEdge(edge, target, clonePropertyFunc, nodeMap)
	}
}

// CopyNode copies the sourceNode into target graph
func CopyNode(sourceNode Node, target Graph, clonePropertyFunc func(string, interface{}) interface{}) Node {
	return target.NewNode(copyLabels(sourceNode.GetLabels()), copyProperties(sourceNode, clonePropertyFunc))
}

// CopyEdge copies the edge into graph
func CopyEdge(edge Edge, target Graph, clonePropertyFunc func(string, interface{}) interface{}, nodeMap map[Node]Node) Edge {
	return target.NewEdge(nodeMap[edge.GetFrom()], nodeMap[edge.GetTo()], edge.GetLabel(), copyProperties(edge, clonePropertyFunc))
}
