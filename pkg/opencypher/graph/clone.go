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

// Copy a graph
func CopyGraph(source, target Graph, clonePropertyFunc func(string, interface{}) interface{}) map[Node]Node {
	nodeMap := make(map[Node]Node)
	for nodes := source.GetNodes(); nodes.Next(); {
		node := nodes.Node()
		srcLabels := node.GetLabels().Slice()
		labels := make([]string, len(srcLabels))
		copy(labels, srcLabels)
		properties := make(map[string]interface{})
		node.ForEachProperty(func(key string, value interface{}) bool {
			properties[key] = clonePropertyFunc(key, value)
			return true
		})
		nodeMap[node] = target.NewNode(labels, properties)
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
