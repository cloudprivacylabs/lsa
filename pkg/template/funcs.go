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

package template

// import (
// 	"strings"

// 	"github.com/Masterminds/sprig"
// 	"github.com/bserdar/digraph"

// 	"github.com/cloudprivacylabs/lsa/pkg/ls"
// )

// // Functions contain the graph functions for template evaluation
// var Functions = map[string]interface{}{
// 	"gexpr":       graphExprFunc,
// 	"gnode":       graphNodeFunc,
// 	"glinkedWith": graphLinkedWithFunc,
// 	"ginstanceOf": graphInstanceOfFunc,
// 	"gpath":       graphPath,
// }

// func init() {
// 	for k, v := range sprig.FuncMap() {
// 		Functions[k] = v
// 	}
// }

// // Evaluate a graph expression, written in JSON with ' instead of "
// //
// //  {{gexpr .graph "{'op': { expr } }"}}
// func graphExprFunc(graph *digraph.Graph, expression string) (interface{}, error) {
// 	expr, err := ls.UnmarshalExpression([]byte(strings.ReplaceAll(expression, "'", "\"")))
// 	if err != nil {
// 		return nil, err
// 	}
// 	return expr.EvaluateExpression(graph)
// }

// // Return the first node with the given id
// func graphNodeFunc(graph *digraph.Graph, id string) interface{} {
// 	ix := graph.GetNodeIndex()
// 	nodes := ix.NodesByLabel(id)
// 	if nodes.HasNext() {
// 		return nodes.Next()
// 	}
// 	return nil
// }

// // Return the nodes that are directly linked to this node with the given label
// func graphLinkedWithFunc(graph *digraph.Graph, targetID, linkLabel string) interface{} {
// 	var s *string
// 	if len(linkLabel) > 0 {
// 		s = &linkLabel
// 	}
// 	nodes, _ := ls.SelectNodes(graph, &ls.NodeLinkedPredicate{TargetPredicate: ls.NewNodeIDPredicate(targetID), Label: s})
// 	return nodes
// }

// // Return the nodes that are instance of the target node
// func graphInstanceOfFunc(graph *digraph.Graph, t string) interface{} {
// 	s := ls.InstanceOfTerm
// 	nodes, _ := ls.SelectNodes(graph, &ls.NodeLinkedPredicate{TargetPredicate: ls.NewNodeTypePredicate(t), Label: &s})
// 	return nodes
// }

// // Follow the path from the given node using edgelabel, property=value, edgelabel, ...
// func graphPath(node ls.Node, path ...string) []ls.Node {
// 	return ls.EvaluatePathExpression([]ls.Node{node}, path...)
// }
