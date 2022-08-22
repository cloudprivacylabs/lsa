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
	"github.com/cloudprivacylabs/opencypher/graph"
	"golang.org/x/text/encoding"
)

// A Layer is either a schema or an overlay. It keeps the definition
// of a layer as a directed labeled property graph.
//
// The root node of the layer keeps layer identifying information. The
// root node is connected to the schema node which contains the actual
// object defined by the layer.
type Layer struct {
	Graph     graph.Graph
	layerInfo graph.Node
}

// NewLayerGraph creates a new graph indexes to store layers
func NewLayerGraph() graph.Graph {
	g := graph.NewOCGraph()
	g.AddNodePropertyIndex(NodeIDTerm)
	for _, f := range newLayerGraphHooks {
		f(g)
	}
	return g
}

var newLayerGraphHooks = []func(*graph.OCGraph){}

func RegisterNewLayerGraphHook(f func(*graph.OCGraph)) {
	newLayerGraphHooks = append(newLayerGraphHooks, f)
}

// NewLayer returns a new empty layer
func NewLayer() *Layer {
	g := NewLayerGraph()
	ret := &Layer{Graph: g}
	ret.layerInfo = ret.Graph.NewNode(nil, nil)
	return ret
}

// NewLayerInGraph creates a new layer in the given graph by creating
// a layerinfo root node for the layer. The graph may contain many
// other layers
func NewLayerInGraph(g graph.Graph) *Layer {
	ret := &Layer{Graph: g}
	ret.layerInfo = g.NewNode(nil, nil)
	return ret
}

// LayersFromGraph returns the layers from an existing graph. All
// Schema and Overlay nodes are returned as layers.
func LayersFromGraph(g graph.Graph) []*Layer {
	ret := make([]*Layer, 0)
	set := graph.NewStringSet(SchemaTerm)
	for nodes := g.GetNodesWithAllLabels(set); nodes.Next(); {
		node := nodes.Node()
		l := Layer{Graph: g, layerInfo: node}
		ret = append(ret, &l)
	}
	set = graph.NewStringSet(OverlayTerm)
	for nodes := g.GetNodesWithAllLabels(set); nodes.Next(); {
		node := nodes.Node()
		l := Layer{Graph: g, layerInfo: node}
		ret = append(ret, &l)
	}
	return ret
}

// Clone returns a copy of the layer in a new graph. If the graph
// contains other layers, they are not copied.
func (l *Layer) Clone() *Layer {
	targetGraph := NewLayerGraph()
	newLayer, _ := l.CloneInto(targetGraph)
	return newLayer
}

// CloneInto clones the layer into the targetgraph. If the source
// graph contains other layers, they are not copied.
func (l *Layer) CloneInto(targetGraph graph.Graph) (*Layer, map[graph.Node]graph.Node) {
	ret := &Layer{Graph: targetGraph}
	nodeMap := make(map[graph.Node]graph.Node)
	graph.CopySubgraph(l.layerInfo, targetGraph, func(key string, value interface{}) interface{} {
		if p, ok := value.(*PropertyValue); ok {
			return p.Clone()
		}
		return value
	}, nodeMap)
	ret.layerInfo = nodeMap[l.layerInfo]
	return ret, nodeMap
}

// GetLayerRootNode returns the root node of the schema
func (l *Layer) GetLayerRootNode() graph.Node { return l.layerInfo }

// Returns the overlay attribute nodes if there are any
func (l *Layer) GetOverlayAttributes() []graph.Node {
	return graph.TargetNodes(l.layerInfo.GetEdgesWithLabel(graph.OutgoingEdge, AttributeOverlaysTerm))
}

// GetSchemaRootNode returns the root node of the object defined by the schema
func (l *Layer) GetSchemaRootNode() graph.Node {
	if l == nil {
		return nil
	}
	x := graph.TargetNodes(l.layerInfo.GetEdgesWithLabel(graph.OutgoingEdge, LayerRootTerm))
	if len(x) != 1 {
		return nil
	}
	return x[0]
}

// GetID returns the ID of the layer
func (l *Layer) GetID() string {
	return GetNodeID(l.layerInfo)
}

// SetID sets the ID of the layer
func (l *Layer) SetID(ID string) {
	SetNodeID(l.layerInfo, ID)
}

// GetLayerType returns the layer type, SchemaTerm or OverlayTerm.
func (l *Layer) GetLayerType() string {
	labels := l.layerInfo.GetLabels()
	if labels.Has(SchemaTerm) {
		return SchemaTerm
	}
	if labels.Has(OverlayTerm) {
		return OverlayTerm
	}
	return ""
}

// SetLayerType sets if the layer is a schema or an overlay
func (l *Layer) SetLayerType(t string) {
	if t != SchemaTerm && t != OverlayTerm {
		panic("Invalid layer type:" + t)
	}
	labels := l.layerInfo.GetLabels()
	labels.Remove(SchemaTerm, OverlayTerm)
	labels.Add(t)
	l.layerInfo.SetLabels(labels)
}

// GetEncoding returns the encoding that should be used to
// ingest/export data using this layer. The encoding information is
// taken from the schema root node characterEncoding annotation. If missing,
// the default encoding is used, which does not perform any character
// translation
func (l *Layer) GetEncoding() (encoding.Encoding, error) {
	oi := l.GetSchemaRootNode()
	var enc string
	if oi != nil {
		enc = AsPropertyValue(oi.GetProperty(CharacterEncodingTerm)).AsString()
		if len(enc) == 0 {
			return encoding.Nop, nil
		}
	}
	return UnknownEncodingIndex.Encoding(enc)
}

// GetValueType returns the value of the valueType field from the
// layer information node. This is the type of the entity defined by
// the schema
func (l *Layer) GetValueType() string {
	return AsPropertyValue(l.layerInfo.GetProperty(ValueTypeTerm)).AsString()
}

// SetValueType sets the value types of the layer
func (l *Layer) SetValueType(t string) {
	if oldT := l.GetValueType(); len(oldT) > 0 {
		if oin := l.GetSchemaRootNode(); oin != nil {
			labels := oin.GetLabels()
			labels.Remove(oldT)
			oin.SetLabels(labels)
		}
	}
	if len(t) > 0 {
		l.layerInfo.SetProperty(ValueTypeTerm, StringPropertyValue(GetTermInfo(ValueTypeTerm).Term, t))
		if oin := l.GetSchemaRootNode(); oin != nil {
			labels := oin.GetLabels()
			labels.Add(t)
			oin.SetLabels(labels)
		}
	}
}

// GetArrayElementNode returns the array element node from an array node
func GetArrayElementNode(arraySchemaNode graph.Node) graph.Node {
	if arraySchemaNode == nil {
		return nil
	}
	n := graph.TargetNodes(arraySchemaNode.GetEdgesWithLabel(graph.OutgoingEdge, ArrayItemsTerm))
	if len(n) == 1 {
		return n[0]
	}
	return nil
}

// GetObjectAttributeNodesBy returns the schema attribute nodes under a
// schema object. The returned map is keyed by the keyTerm
func GetObjectAttributeNodesBy(objectSchemaNode graph.Node, keyTerm string) (map[string][]graph.Node, error) {
	nextNodes := make(map[string][]graph.Node)
	addNextNode := func(node graph.Node) error {
		key := AsPropertyValue(node.GetProperty(keyTerm)).AsString()
		if len(key) == 0 {
			return nil
		}
		nextNodes[key] = append(nextNodes[key], node)
		return nil
	}
	if objectSchemaNode != nil {
		for _, node := range graph.TargetNodes(objectSchemaNode.GetEdgesWithLabel(graph.OutgoingEdge, ObjectAttributesTerm)) {
			if err := addNextNode(node); err != nil {
				return nil, err
			}
		}
		for _, node := range graph.TargetNodes(objectSchemaNode.GetEdgesWithLabel(graph.OutgoingEdge, ObjectAttributeListTerm)) {
			if err := addNextNode(node); err != nil {
				return nil, err
			}
		}
	}
	return nextNodes, nil
}

// GetObjectAttributeNodes returns the schema attribute nodes under a
// schema object.
func GetObjectAttributeNodes(objectSchemaNode graph.Node) []graph.Node {
	nextNodes := make([]graph.Node, 0)
	if objectSchemaNode != nil {
		for _, node := range graph.TargetNodes(objectSchemaNode.GetEdgesWithLabel(graph.OutgoingEdge, ObjectAttributesTerm)) {
			nextNodes = append(nextNodes, node)
		}
		for _, node := range graph.TargetNodes(objectSchemaNode.GetEdgesWithLabel(graph.OutgoingEdge, ObjectAttributeListTerm)) {
			nextNodes = append(nextNodes, node)
		}
	}
	return nextNodes
}

// GetPolymorphicOptions returns the polymorphic options of a schema node
func GetPolymorphicOptions(polymorphicSchemaNode graph.Node) []graph.Node {
	return graph.TargetNodes(polymorphicSchemaNode.GetEdgesWithLabel(graph.OutgoingEdge, OneOfTerm))
}

// GetNodesWithValidators returns all nodes under root that has validators
func GetNodesWithValidators(root graph.Node) map[graph.Node]struct{} {
	ret := make(map[graph.Node]struct{})
	ForEachAttributeNode(root, func(node graph.Node, _ []graph.Node) bool {
		node.ForEachProperty(func(k string, in interface{}) bool {
			pv, ok := in.(*PropertyValue)
			if ok {
				md := pv.GetSem()
				if md == nil {
					return true
				}
				if _, ok := md.Metadata.(NodeValidator); ok {
					ret[node] = struct{}{}
				}
				if _, ok := md.Metadata.(ValueValidator); ok {
					ret[node] = struct{}{}
				}
				return true
			}
			return false
		})
		return true
	})
	return ret
}

// GetEntityIDNodes returns the entity id attribute IDs from the layer
// root node
func (l *Layer) GetEntityIDNodes() []string {
	root := l.GetSchemaRootNode()
	if root == nil {
		return nil
	}
	return AsPropertyValue(root.GetProperty(EntityIDFieldsTerm)).MustStringSlice()
}

// GetAttributesByID returns attribute nodes by ID
func (l *Layer) GetAttributesByID(ids []string) []graph.Node {
	ret := make([]graph.Node, len(ids))
	for x := range ids {
		ret[x] = l.GetAttributeByID(ids[x])
	}
	return ret
}

// ForEachAttribute calls f with each attribute node, depth first. If
// f returns false, iteration stops
func (l *Layer) ForEachAttribute(f func(graph.Node, []graph.Node) bool) bool {
	oi := l.GetSchemaRootNode()
	if oi != nil {
		return ForEachAttributeNode(oi, f)
	}
	return true
}

// ForEachAttributeOrdered calls f with each attribute node, depth
// first and in order. If f returns false, iteration stops
func (l *Layer) ForEachAttributeOrdered(f func(graph.Node, []graph.Node) bool) bool {
	oi := l.GetSchemaRootNode()
	if oi != nil {
		return ForEachAttributeNodeOrdered(oi, f)
	}
	return true
}

// GetParentAttribute returns the parent attribute of the given node
func GetParentAttribute(node graph.Node) graph.Node {
	for edges := node.GetEdges(graph.IncomingEdge); edges.Next(); {
		edge := edges.Edge()
		if IsAttributeTreeEdge(edge) && IsAttributeNode(edge.GetFrom()) && !IsCompilationArtifact(edge) {
			return edge.GetFrom()
		}
	}
	return nil
}

// GetAttributePath returns the path to the given attribute node
func (l *Layer) GetAttributePath(node graph.Node) []graph.Node {
	root := l.GetSchemaRootNode()
	return GetAttributePath(root, node)
}

// GetAttributePath returns the path from root to node. There must
// exist exactly one path. If not, returns nil
func GetAttributePath(root, node graph.Node) []graph.Node {
	ret := make([]graph.Node, 0)
	ret = append(ret, node)
	for node != root {
		hasEdges := false
		for edges := node.GetEdges(graph.IncomingEdge); edges.Next(); {
			hasEdges = true
			edge := edges.Edge()
			if IsAttributeTreeEdge(edge) && IsAttributeNode(edge.GetFrom()) {
				ret = append(ret, edge.GetFrom())
				node = edge.GetFrom()
			}
		}
		if !hasEdges {
			return nil
		}
	}
	for i := 0; i < len(ret)/2; i++ {
		ret[i], ret[len(ret)-i-1] = ret[len(ret)-i-1], ret[i]
	}
	return ret
}

// GetAttributeByID returns the attribute node by its ID.
func (l *Layer) GetAttributeByID(id string) graph.Node {
	getAttributeByIDPattern := graph.Pattern{{
		Labels:     graph.NewStringSet(AttributeNodeTerm),
		Properties: map[string]interface{}{NodeIDTerm: id}}}
	nodes, _ := getAttributeByIDPattern.FindNodes(l.Graph, nil)
	if len(nodes) == 1 {
		return nodes[0]
	}
	return nil
}

// FindAttributeByID returns the attribute and the path to it
func (l *Layer) FindAttributeByID(id string) (graph.Node, []graph.Node) {
	node := l.GetAttributeByID(id)
	if node == nil {
		return nil, nil
	}
	return node, l.GetAttributePath(node)
}

// FindFirstAttribute returns the first attribute for which the predicate holds
func (l *Layer) FindFirstAttribute(predicate func(graph.Node) bool) (graph.Node, []graph.Node) {
	var node graph.Node
	var path []graph.Node
	ForEachAttributeNode(l.GetSchemaRootNode(), func(n graph.Node, p []graph.Node) bool {
		if predicate(n) {
			node = n
			path = p
			return false
		}
		return true
	})
	return node, path
}

// ForEachAttributeNode calls f with each attribute node, depth
// first. Path contains all the nodes from root to the current
// node. If f returns false, iteration stops. This function visits
// each node only once
func ForEachAttributeNode(root graph.Node, f func(node graph.Node, path []graph.Node) bool) bool {
	return forEachAttributeNode(root, make([]graph.Node, 0, 32), f, map[graph.Node]struct{}{}, false)
}

func forEachAttributeNode(root graph.Node, path []graph.Node, f func(graph.Node, []graph.Node) bool, loop map[graph.Node]struct{}, ordered bool) bool {
	if _, exists := loop[root]; exists {
		return true
	}
	loop[root] = struct{}{}

	path = append(path, root)
	if IsAttributeNode(root) {
		if !f(root, path) {
			return false
		}
	}

	outgoing := root.GetEdges(graph.OutgoingEdge)
	if ordered {
		outgoing = SortEdgesItr(outgoing)
	}

	for outgoing.Next() {
		edge := outgoing.Edge()
		if !IsAttributeTreeEdge(edge) {
			continue
		}
		next := edge.GetTo()
		if next.GetLabels().Has(AttributeNodeTerm) {
			if !forEachAttributeNode(next, path, f, loop, ordered) {
				return false
			}
		}
	}
	return true
}

// ForEachAttributeNodeOrdered calls f with each attribute node, depth
// first, preserving order. Path contains all the nodes from root to the current
// node. If f returns false, iteration stops. This function visits
// each node only once
func ForEachAttributeNodeOrdered(root graph.Node, f func(node graph.Node, path []graph.Node) bool) bool {
	return forEachAttributeNode(root, make([]graph.Node, 0, 32), f, map[graph.Node]struct{}{}, true)
}

// GetAttributeID returns the attributeID
func GetAttributeID(node graph.Node) string {
	v, _ := node.GetProperty(NodeIDTerm)
	s, _ := v.(string)
	return s
}

// SetAttrributeID sets the attribute ID
func SetAttributeID(node graph.Node, ID string) {
	node.SetProperty(NodeIDTerm, ID)
}

// CopySchemaNodeIntoGraph copies a schema node and the subtree under
// it that does not belong the schema into the target graph
func CopySchemaNodeIntoGraph(target graph.Graph, schemaNode graph.Node) graph.Node {
	nodeMap := make(map[graph.Node]graph.Node)

	newNode := graph.CopyNode(schemaNode, target, ClonePropertyValueFunc)
	nodeMap[schemaNode] = newNode

	for edges := schemaNode.GetEdges(graph.OutgoingEdge); edges.Next(); {
		edge := edges.Edge()
		if IsAttributeTreeEdge(edge) {
			continue
		}
		graph.CopySubgraph(edge.GetTo(), target, ClonePropertyValueFunc, nodeMap)
		graph.CopyEdge(edge, target, ClonePropertyValueFunc, nodeMap)
	}
	return newNode
}

// GetLayerEntityRoot returns the layer entity root node containing the given schema node
func GetLayerEntityRoot(node graph.Node) graph.Node {
	var find func(graph.Node) graph.Node
	seen := make(map[graph.Node]struct{})
	find = func(root graph.Node) graph.Node {
		if _, ok := root.GetProperty(EntitySchemaTerm); ok {
			return root
		}
		if _, ok := seen[root]; ok {
			return nil
		}
		seen[root] = struct{}{}
		var ret graph.Node
		for edges := root.GetEdges(graph.IncomingEdge); edges.Next(); {
			edge := edges.Edge()
			ancestor := edge.GetFrom()
			if !ancestor.GetLabels().Has(AttributeNodeTerm) {
				continue
			}
			ret = find(ancestor)
		}
		return ret
	}
	return find(node)
}
