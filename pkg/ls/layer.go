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
	"sync"

	"github.com/cloudprivacylabs/lpg/v2"
	"golang.org/x/text/encoding"
)

// A Layer is either a schema or an overlay. It keeps the definition
// of a layer as a directed labeled property graph.
//
// The root node of the layer keeps layer identifying information. The
// root node is connected to the schema node which contains the actual
// object defined by the layer.
type Layer struct {
	Graph     *lpg.Graph
	layerInfo *lpg.Node

	linkSpecsOnce sync.Once
	linkSpecs     []*LinkSpec
}

// NewLayerGraph creates a new graph indexes to store layers
func NewLayerGraph() *lpg.Graph {
	g := lpg.NewGraph()
	g.AddNodePropertyIndex(NodeIDTerm.Name, lpg.HashIndex)
	for _, f := range newLayerGraphHooks {
		f(g)
	}
	return g
}

var newLayerGraphHooks = []func(*lpg.Graph){}

func RegisterNewLayerGraphHook(f func(*lpg.Graph)) {
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
func NewLayerInGraph(g *lpg.Graph) *Layer {
	ret := &Layer{Graph: g}
	ret.layerInfo = g.NewNode(nil, nil)
	return ret
}

func NewLayerFromRootNode(layerInfoNode *lpg.Node) *Layer {
	ret := Layer{
		Graph:     layerInfoNode.GetGraph(),
		layerInfo: layerInfoNode,
	}
	return &ret
}

// LayersFromGraph returns the layers from an existing graph. All
// Schema and Overlay nodes are returned as layers.
func LayersFromGraph(g *lpg.Graph) []*Layer {
	ret := make([]*Layer, 0)
	set := lpg.NewStringSet(SchemaTerm.Name)
	for nodes := g.GetNodesWithAllLabels(set); nodes.Next(); {
		node := nodes.Node()
		l := Layer{Graph: g, layerInfo: node}
		ret = append(ret, &l)
	}
	set = lpg.NewStringSet(OverlayTerm.Name)
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
func (l *Layer) CloneInto(targetGraph *lpg.Graph) (*Layer, map[*lpg.Node]*lpg.Node) {
	ret := &Layer{Graph: targetGraph}
	nodeMap := make(map[*lpg.Node]*lpg.Node)
	lpg.CopySubgraph(l.layerInfo, targetGraph, func(key string, value interface{}) interface{} {
		return value
	}, nodeMap)
	ret.layerInfo = nodeMap[l.layerInfo]
	return ret, nodeMap
}

// GetLayerRootNode returns the root node of the schema
func (l *Layer) GetLayerRootNode() *lpg.Node { return l.layerInfo }

// Returns the overlay attribute nodes if there are any
func (l *Layer) GetOverlayAttributes() []*lpg.Node {
	return lpg.TargetNodes(l.layerInfo.GetEdgesWithLabel(lpg.OutgoingEdge, AttributeOverlaysTerm.Name))
}

// GetSchemaRootNode returns the root node of the object defined by the schema
func (l *Layer) GetSchemaRootNode() *lpg.Node {
	if l == nil {
		return nil
	}
	edges := l.layerInfo.GetEdgesWithLabel(lpg.OutgoingEdge, LayerRootTerm.Name)
	if edges.MaxSize() != 1 {
		return nil
	}
	edges.Next()
	return edges.Edge().GetTo()
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
	if labels.Has(SchemaTerm.Name) {
		return SchemaTerm.Name
	}
	if labels.Has(OverlayTerm.Name) {
		return OverlayTerm.Name
	}
	return ""
}

// SetLayerType sets if the layer is a schema or an overlay
func (l *Layer) SetLayerType(t string) {
	if t != SchemaTerm.Name && t != OverlayTerm.Name {
		panic("Invalid layer type:" + t)
	}
	labels := l.layerInfo.GetLabels()
	labels.Remove(SchemaTerm.Name, OverlayTerm.Name)
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
		enc = CharacterEncodingTerm.PropertyValue(oi)
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
	s, _ := GetPropertyValueAs[string](l.layerInfo, ValueTypeTerm.Name)
	return s
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
		l.layerInfo.SetProperty(ValueTypeTerm.Name, NewPropertyValue(ValueTypeTerm.Name, t))
		if oin := l.GetSchemaRootNode(); oin != nil {
			labels := oin.GetLabels()
			labels.Add(t)
			oin.SetLabels(labels)
		}
	}
}

// GetArrayElementNode returns the array element node from an array node
func GetArrayElementNode(arraySchemaNode *lpg.Node) *lpg.Node {
	if arraySchemaNode == nil {
		return nil
	}
	n := lpg.TargetNodes(arraySchemaNode.GetEdgesWithLabel(lpg.OutgoingEdge, ArrayItemsTerm.Name))
	if len(n) == 1 {
		return n[0]
	}
	return nil
}

// GetObjectAttributeNodesBy returns the schema attribute nodes under a
// schema object. The returned map is keyed by the keyTerm
func GetObjectAttributeNodesBy(objectSchemaNode *lpg.Node, keyTerm string) (map[string][]*lpg.Node, error) {
	nextNodes := make(map[string][]*lpg.Node)
	addNextNode := func(node *lpg.Node) error {
		key, _ := GetPropertyValueAs[string](node, keyTerm)
		if len(key) == 0 {
			return nil
		}
		nextNodes[key] = append(nextNodes[key], node)
		return nil
	}
	if objectSchemaNode != nil {
		for _, node := range lpg.TargetNodes(objectSchemaNode.GetEdgesWithLabel(lpg.OutgoingEdge, ObjectAttributeListTerm.Name)) {
			if err := addNextNode(node); err != nil {
				return nil, err
			}
		}
		for _, node := range lpg.TargetNodes(objectSchemaNode.GetEdgesWithLabel(lpg.OutgoingEdge, ObjectAttributesTerm.Name)) {
			if err := addNextNode(node); err != nil {
				return nil, err
			}
		}
	}
	return nextNodes, nil
}

// GetObjectAttributeNodes returns the schema attribute nodes under a
// schema object.
func GetObjectAttributeNodes(objectSchemaNode *lpg.Node) []*lpg.Node {
	nextNodes := make([]*lpg.Node, 0)
	if objectSchemaNode != nil {
		for _, node := range lpg.TargetNodes(objectSchemaNode.GetEdgesWithLabel(lpg.OutgoingEdge, ObjectAttributesTerm.Name)) {
			nextNodes = append(nextNodes, node)
		}
		for _, node := range lpg.TargetNodes(objectSchemaNode.GetEdgesWithLabel(lpg.OutgoingEdge, ObjectAttributeListTerm.Name)) {
			nextNodes = append(nextNodes, node)
		}
	}
	return nextNodes
}

// GetPolymorphicOptions returns the polymorphic options of a schema node
func GetPolymorphicOptions(polymorphicSchemaNode *lpg.Node) []*lpg.Node {
	return lpg.TargetNodes(polymorphicSchemaNode.GetEdgesWithLabel(lpg.OutgoingEdge, OneOfTerm.Name))
}

// GetNodesWithValidators returns all nodes under root that has validators
func GetNodesWithValidators(root *lpg.Node) map[*lpg.Node]struct{} {
	ret := make(map[*lpg.Node]struct{})
	ForEachAttributeNode(root, func(node *lpg.Node, _ []*lpg.Node) bool {
		node.ForEachProperty(func(k string, in interface{}) bool {
			pv, ok := in.(PropertyValue)
			if !ok {
				return true
			}
			md := pv.Sem()
			if _, ok := md.Metadata.(NodeValidator); ok {
				ret[node] = struct{}{}
			}
			if _, ok := md.Metadata.(ValueValidator); ok {
				ret[node] = struct{}{}
			}
			return true
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
	return EntityIDFieldsTerm.PropertyValue(root)
}

// GetAttributesByID returns attribute nodes by ID
func (l *Layer) GetAttributesByID(ids []string) []*lpg.Node {
	ret := make([]*lpg.Node, len(ids))
	for x := range ids {
		ret[x] = l.GetAttributeByID(ids[x])
	}
	return ret
}

// ForEachAttribute calls f with each attribute node, depth first. If
// f returns false, iteration stops
func (l *Layer) ForEachAttribute(f func(*lpg.Node, []*lpg.Node) bool) bool {
	oi := l.GetSchemaRootNode()
	if oi != nil {
		return ForEachAttributeNode(oi, f)
	}
	return true
}

// ForEachAttributeOrdered calls f with each attribute node, depth
// first and in order. If f returns false, iteration stops
func (l *Layer) ForEachAttributeOrdered(f func(*lpg.Node, []*lpg.Node) bool) bool {
	oi := l.GetSchemaRootNode()
	if oi != nil {
		return ForEachAttributeNodeOrdered(oi, f)
	}
	return true
}

// GetParentAttribute returns the parent attribute of the given node
func GetParentAttribute(node *lpg.Node) *lpg.Node {
	for edges := node.GetEdges(lpg.IncomingEdge); edges.Next(); {
		edge := edges.Edge()
		if IsAttributeTreeEdge(edge) && IsAttributeNode(edge.GetFrom()) && !IsCompilationArtifact(edge) {
			return edge.GetFrom()
		}
	}
	return nil
}

// GetAttributePath returns the path to the given attribute node
func (l *Layer) GetAttributePath(node *lpg.Node) []*lpg.Node {
	root := l.GetSchemaRootNode()
	return GetAttributePath(root, node)
}

// GetAttributePath returns the path from root to node. There must
// exist exactly one path. If not, returns nil
func GetAttributePath(root, node *lpg.Node) []*lpg.Node {
	ret := make([]*lpg.Node, 0)
	ret = append(ret, node)
	seen := make(map[*lpg.Node]struct{})
	for node != root {
		hasEdges := false
		for edges := node.GetEdges(lpg.IncomingEdge); edges.Next(); {
			edge := edges.Edge()
			if _, ok := seen[edge.GetFrom()]; ok {
				continue
			}
			seen[edge.GetFrom()] = struct{}{}
			hasEdges = true
			if edge.GetFrom() == root {
				ret = append(ret, edge.GetFrom())
				node = edge.GetFrom()
				break
			}
			if IsCompilationArtifact(edge) {
				continue
			}
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
func (l *Layer) GetAttributeByID(id string) *lpg.Node {
	getAttributeByIDPattern := lpg.Pattern{{
		Labels:     lpg.NewStringSet(AttributeNodeTerm.Name),
		Properties: map[string]any{NodeIDTerm.Name: NodeIDTerm.MustPropertyValue(id)}}}
	nodes, _ := getAttributeByIDPattern.FindNodes(l.Graph, nil)
	if len(nodes) == 1 {
		return nodes[0]
	}
	return nil
}

// FindAttributeByID returns the attribute and the path to it
func (l *Layer) FindAttributeByID(id string) (*lpg.Node, []*lpg.Node) {
	node := l.GetAttributeByID(id)
	if node == nil {
		return nil, nil
	}
	return node, l.GetAttributePath(node)
}

func (l *Layer) NodeSlice() []*lpg.Node {
	var forEach func(*lpg.Node)
	seen := make(map[*lpg.Node]struct{})
	ret := make([]*lpg.Node, 0)
	forEach = func(root *lpg.Node) {
		if _, exists := seen[root]; exists {
			return
		}
		seen[root] = struct{}{}
		if !IsAttributeNode(root) {
			return
		}
		ret = append(ret, root)
		for outgoing := root.GetEdges(lpg.OutgoingEdge); outgoing.Next(); {
			forEach(outgoing.Edge().GetTo())
		}
	}
	forEach(l.GetSchemaRootNode())
	return ret
}

// FindFirstAttribute returns the first attribute for which the predicate holds
func (l *Layer) FindFirstAttribute(predicate func(*lpg.Node) bool) (*lpg.Node, []*lpg.Node) {
	var node *lpg.Node
	var path []*lpg.Node
	ForEachAttributeNode(l.GetSchemaRootNode(), func(n *lpg.Node, p []*lpg.Node) bool {
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
func ForEachAttributeNode(root *lpg.Node, f func(node *lpg.Node, path []*lpg.Node) bool) bool {
	return forEachAttributeNode(root, make([]*lpg.Node, 0, 32), f, map[*lpg.Node]struct{}{}, false)
}

func forEachAttributeNode(root *lpg.Node, path []*lpg.Node, f func(*lpg.Node, []*lpg.Node) bool, loop map[*lpg.Node]struct{}, ordered bool) bool {
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

	outgoing := root.GetEdges(lpg.OutgoingEdge)
	if ordered {
		outgoing = SortEdgesItr(outgoing)
	}

	for outgoing.Next() {
		edge := outgoing.Edge()
		if !IsAttributeTreeEdge(edge) {
			continue
		}
		next := edge.GetTo()
		if next.GetLabels().Has(AttributeNodeTerm.Name) {
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
func ForEachAttributeNodeOrdered(root *lpg.Node, f func(node *lpg.Node, path []*lpg.Node) bool) bool {
	return forEachAttributeNode(root, make([]*lpg.Node, 0, 32), f, map[*lpg.Node]struct{}{}, true)
}

// GetAttributeID returns the attributeID
func GetAttributeID(node *lpg.Node) string {
	return NodeIDTerm.PropertyValue(node)
}

// SetAttrributeID sets the attribute ID
func SetAttributeID(node *lpg.Node, ID string) {
	node.SetProperty(NodeIDTerm.Name, NodeIDTerm.MustPropertyValue(ID))
}

// CopySchemaNodeIntoGraph copies a schema node and the subtree under
// it that does not belong the schema into the target graph
func CopySchemaNodeIntoGraph(target *lpg.Graph, schemaNode *lpg.Node) *lpg.Node {
	nodeMap := make(map[*lpg.Node]*lpg.Node)

	newNode := lpg.CopyNode(schemaNode, target, ClonePropertyValueFunc)
	nodeMap[schemaNode] = newNode

	for edges := schemaNode.GetEdges(lpg.OutgoingEdge); edges.Next(); {
		edge := edges.Edge()
		if IsAttributeTreeEdge(edge) {
			continue
		}
		lpg.CopySubgraph(edge.GetTo(), target, ClonePropertyValueFunc, nodeMap)
		lpg.CopyEdge(edge, target, ClonePropertyValueFunc, nodeMap)
	}
	return newNode
}

// GetLayerEntityRoot returns the layer entity root node containing the given schema node
func GetLayerEntityRoot(node *lpg.Node) *lpg.Node {
	var find func(*lpg.Node) *lpg.Node
	seen := make(map[*lpg.Node]struct{})
	find = func(root *lpg.Node) *lpg.Node {
		if _, ok := root.GetProperty(EntitySchemaTerm.Name); ok {
			return root
		}
		if _, ok := seen[root]; ok {
			return nil
		}
		seen[root] = struct{}{}
		var ret *lpg.Node
		for edges := root.GetEdges(lpg.IncomingEdge); edges.Next(); {
			edge := edges.Edge()
			ancestor := edge.GetFrom()
			if !ancestor.GetLabels().Has(AttributeNodeTerm.Name) {
				continue
			}
			ret = find(ancestor)
		}
		return ret
	}
	return find(node)
}

// GetPathFromEntityRoot returns the path from the entity root node of the schema
func GetPathFromRoot(schemaNode *lpg.Node) []*lpg.Node {
	path := make([]*lpg.Node, 0)
	var find func(*lpg.Node) *lpg.Node
	seen := make(map[*lpg.Node]struct{})
	find = func(root *lpg.Node) *lpg.Node {
		if _, ok := root.GetProperty(EntitySchemaTerm.Name); ok {
			return root
		}
		if _, ok := seen[root]; ok {
			return nil
		}
		seen[root] = struct{}{}
		for edges := root.GetEdges(lpg.IncomingEdge); edges.Next(); {
			edge := edges.Edge()
			ancestor := edge.GetFrom()
			if !ancestor.GetLabels().Has(AttributeNodeTerm.Name) {
				continue
			}
			ret := find(ancestor)
			if ret != nil {
				path = append(path, ret)
				return root
			}
		}
		return nil
	}
	node := find(schemaNode)
	if node != nil {
		path = append(path, node)
	}
	return path
}

// GetLinkSpecs retrieves link specs for the layer
func (l *Layer) GetLinkSpecs() ([]*LinkSpec, error) {
	var err error
	l.linkSpecsOnce.Do(func() {
		specs := make([]*LinkSpec, 0)
		for nodes := l.Graph.GetNodes(); nodes.Next(); {
			attrNode := nodes.Node()
			var ls *LinkSpec
			ls, err = GetLinkSpec(attrNode)
			if err != nil {
				return
			}
			if ls == nil {
				continue
			}
			specs = append(specs, ls)
		}
		l.linkSpecs = specs
	})
	return l.linkSpecs, err
}

// GetOutputEdgeLabel returns the label that should be used to connect
// child nodes. This is determined by the outputEdgeLabel
// term. ParentNode is either a schema node or a doc node
func GetOutputEdgeLabel(parentNode *lpg.Node) string {
	if parentNode == nil {
		return HasTerm.Name
	}
	pv, exists := GetNodeOrSchemaProperty(parentNode, OutputEdgeLabelTerm.Name)
	if !exists {
		return HasTerm.Name
	}
	if s, _ := pv.Value().(string); len(s) > 0 {
		return s
	}
	return HasTerm.Name
}
