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
package json

import (
	"fmt"

	"github.com/bserdar/digraph"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// LDRenderer renders a graph in JSON-LD flattened format
type LDRenderer struct {
	// If set, generates node identifiers from the given node. It should
	// be able to generate blank node IDs if the node is to be
	// represented as an RDF blank node, or if the node does not have an
	// ID.
	//
	// If not set, the default generator function uses the string node
	// id, or _b:<n> if the node does not have an id.
	NodeIDGeneratorFunc func(digraph.Node) string

	// If set, generates edge labels for the given edge. If it is not set,
	// the default is to use the edge label. If edge does not have a
	// label,
	EdgeLabelGeneratorFunc func(digraph.Edge) string
}

func (rd *LDRenderer) Render(input *digraph.Graph) interface{} {
	type outputNode struct {
		id     string
		ldNode map[string]interface{}
	}
	blankNodeId := 0
	// Assign IDs to all nodes
	nodeIdMap := make(map[digraph.Node]outputNode)
	for nodes := input.AllNodes(); nodes.HasNext(); {
		node := nodes.Next()
		var idstr string
		if rd.NodeIDGeneratorFunc != nil {
			idstr = rd.NodeIDGeneratorFunc(node)
		} else {
			id := node.Label()
			if id == nil {
				idstr = fmt.Sprintf("_b:%d", blankNodeId)
				blankNodeId++
			} else if idstr = fmt.Sprint(id); len(idstr) == 0 {
				idstr = fmt.Sprintf("_b:%d", blankNodeId)
				blankNodeId++
			}
		}
		outNode := outputNode{ldNode: map[string]interface{}{"@id": idstr}, id: idstr}
		nodeIdMap[node] = outNode
	}

	type propertiesSupport interface {
		GetProperties() map[string]*ls.PropertyValue
	}

	// Process the properties
	for gnode, onode := range nodeIdMap {
		switch n := gnode.(type) {
		case ls.LayerNode:
			t := n.GetTypes()
			if len(t) > 0 {
				arr := make([]interface{}, 0, len(t))
				for _, x := range t {
					arr = append(arr, x)
				}
				onode.ldNode["@type"] = arr
			}
		case ls.DocumentNode:
			if v := n.GetValue(); v != nil {
				onode.ldNode[ls.AttributeValueTerm] = v
			}
		}
		if prop, ok := gnode.(propertiesSupport); ok {
			for key, pvalue := range prop.GetProperties() {
				if pvalue.IsString() {
					onode.ldNode[key] = pvalue.AsString()
				} else if pvalue.IsStringSlice() {
					onode.ldNode[key] = pvalue.AsInterfaceSlice()
				}
			}
		}
	}

	// Process outgoing edges
	for gnode, onode := range nodeIdMap {
		for edges := gnode.AllOutgoingEdges(); edges.HasNext(); {
			edge := edges.Next()
			var labelStr string
			if rd.EdgeLabelGeneratorFunc != nil {
				labelStr = rd.EdgeLabelGeneratorFunc(edge)
			} else {
				id := edge.Label()
				if id == nil {
					labelStr = "http://www.w3.org/1999/02/22-rdf-syntax-ns#property"
				} else if labelStr = fmt.Sprint(id); len(labelStr) == 0 {
					labelStr = "http://www.w3.org/1999/02/22-rdf-syntax-ns#property"
				}
			}
			onode.ldNode[labelStr] = map[string]interface{}{"@id": nodeIdMap[edge.To()].id}
		}
	}
	graph := make([]interface{}, 0, len(nodeIdMap))
	for _, v := range nodeIdMap {
		graph = append(graph, v.ldNode)
	}
	return map[string]interface{}{"@graph": graph}
}
