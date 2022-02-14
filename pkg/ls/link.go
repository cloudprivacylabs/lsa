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
	"fmt"

	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

var (
	// ReferenceFKTerm specifies the foreign key attribute ID
	ReferenceFKTerm = NewTerm(LS+"Reference/", "fk", false, false, OverrideComposition, nil)

	// ReferenceTargetTerm specifies the target entity if the node is not a reference node
	ReferenceTargetTerm = NewTerm(LS+"Reference/", "target", false, false, OverrideComposition, nil)

	// ReferenceTargetIDTerm is the target schema ID field. If not specified, entity ID is used
	ReferenceTargetIDTerm = NewTerm(LS+"Reference/", "targetId", false, false, OverrideComposition, nil)

	// ReferenceLabelTerm specifies the edge label between the referenced nodes
	ReferenceLabelTerm = NewTerm(LS+"Reference/", "label", false, false, OverrideComposition, nil)

	// ReferenceDirectionTerm specifies the direction of the edge. If
	// ->, the edge points to the target entity. If <-, the edge points
	// to this entity.
	ReferenceDirectionTerm = NewTerm(LS+"Reference/", "direction", false, false, OverrideComposition, nil)

	// ReferenceMultiTerm specifies if there can be more than one link targets
	ReferenceMultiTerm = NewTerm(LS+"Reference/", "multi", false, false, OverrideComposition, nil)
)

type ErrInvalidLinkSpec struct {
	ID  string
	Msg string
}

func (err ErrInvalidLinkSpec) Error() string {
	return fmt.Sprintf("Invalid link spec at %s: %s", err.ID, err.Msg)
}

type ErrMultipleTargetsFound struct {
	ID string
}

func (err ErrMultipleTargetsFound) Error() string {
	return fmt.Sprintf("Multiple targets found for %s", err.ID)
}

type ErrCannotResolveLink LinkSpec

func (err ErrCannotResolveLink) Error() string {
	return fmt.Sprintf("Cannot resolve link: %+v", LinkSpec(err))
}

// LinkSpec contains the link field information
type LinkSpec struct {
	// The target schema/entity reference, populated from the
	// `reference` property of the node
	TargetEntity string
	// The foreign key field
	FK string
	// The ID field of the target entity. If not specified, the entity id is used
	TargetID string
	// The label of the link
	Label string
	// If true, the link is from this entity to the target. If false,
	// the link is from the target to this.
	Forward bool
	// If true, the reference can have more than one links
	Multi bool
}

const linkSpecKey = "_linkSpec"

// GetCompiledReferenceLinkSpec returns the compiled reference link
// spec if there is one in the compiled property map of the node
func GetCompiledReferenceLinkSpec(node graph.Node) (LinkSpec, bool) {
	spec, exists := node.GetProperty(linkSpecKey)
	if exists {
		return spec.(LinkSpec), true
	}
	return LinkSpec{}, false
}

// GetEntityRoot tries to find the closes entity containing this
// node. It stops searching when there are more than one incoming
// document node edge
func GetEntityRoot(node graph.Node) graph.Node {
	var movePrev func(graph.Node) graph.Node

	seen := make(map[graph.Node]struct{})
	movePrev = func(start graph.Node) graph.Node {
		var prevNode graph.Node
		for edges := start.GetEdges(graph.IncomingEdge); edges.Next(); {
			edge := edges.Edge()
			if IsDocumentNode(edge.GetFrom()) {
				if prevNode != nil {
					return nil
				}
				prevNode = edge.GetFrom()
			}
		}
		return prevNode
	}

	for {
		trc := movePrev(node)
		if trc == nil {
			return nil
		}
		if _, ok := seen[trc]; ok {
			return nil
		}
		seen[trc] = struct{}{}
		if _, boundary := GetNodeOrSchemaProperty(trc, EntitySchemaTerm); boundary {
			return trc
		}
	}
}

// CompileReferenceLinkSpec gets an uncompiled reference node, and
// puts a LinkSpec into compiled data map of the node if the node
// specifies a link. It also returns the LinkSpec and true. If the
// node is not a reference node, or if the node does not specify a
// link, returns false
func CompileReferenceLinkSpec(layer *Layer, node graph.Node) (LinkSpec, bool, error) {
	// Already processed?
	if _, ok := GetCompiledReferenceLinkSpec(node); ok {
		return LinkSpec{}, false, nil
	}
	if !node.GetLabels().Has(AttributeNodeTerm) {
		return LinkSpec{}, false, nil
	}
	var ref, fk string
	if node.GetLabels().Has(AttributeTypeReference) {
		ref = AsPropertyValue(node.GetProperty(ReferenceTerm)).AsString()
		fk = AsPropertyValue(node.GetProperty(ReferenceFKTerm)).AsString()
	} else {
		ref = AsPropertyValue(node.GetProperty(ReferenceTargetTerm)).AsString()
		fk = GetNodeID(node)
	}
	if len(ref) == 0 || len(fk) == 0 {
		return LinkSpec{}, false, nil
	}
	targetID := AsPropertyValue(node.GetProperty(ReferenceTargetIDTerm)).AsString()
	label := AsPropertyValue(node.GetProperty(ReferenceLabelTerm)).AsString()
	dir := AsPropertyValue(node.GetProperty(ReferenceDirectionTerm)).AsString()
	multi := AsPropertyValue(node.GetProperty(ReferenceMultiTerm)).AsString()
	if len(label) == 0 {
		label = HasTerm
	}
	spec := LinkSpec{
		TargetEntity: ref,
		FK:           fk,
		TargetID:     targetID,
		Label:        label,
		Multi:        multi == "true",
	}
	switch dir {
	case "->", "":
		spec.Forward = true
	case "<-":
		spec.Forward = false
	default:
		return LinkSpec{}, true, ErrInvalidLinkSpec{ID: GetNodeID(node), Msg: "Direction is not one of: ->, <-"}
	}
	node.SetProperty(linkSpecKey, spec)
	return spec, true, nil
}

// GetAllLinkSpecs returns all linkspecs under the given root node
func GetAllLinkSpecs(root graph.Node) map[graph.Node]LinkSpec {
	ret := make(map[graph.Node]LinkSpec)
	IterateDescendants(root, func(node graph.Node, _ []graph.Node) bool {
		spec, exists := GetCompiledReferenceLinkSpec(node)
		if exists {
			ret[node] = spec
		}
		return true
	}, SkipSchemaNodes, false)
	return ret
}

// DocumentEntityInfo describes an entity in a document
type DocumentEntityInfo struct {
	Root    graph.Node
	Schema  string
	IDNodes []graph.Node
}

// GetDocumentEntityInfo returns a map containing all entity root
// nodes, with entity info as map values
func GetDocumentEntityInfo(docRoot graph.Node) map[graph.Node]DocumentEntityInfo {
	ret := make(map[graph.Node]DocumentEntityInfo)
	IterateDescendants(docRoot, func(node graph.Node, path []graph.Node) bool {
		prop, root := GetNodeOrSchemaProperty(node, EntitySchemaTerm)
		if root {
			ret[node] = DocumentEntityInfo{Root: node,
				Schema: prop.AsString(),
			}
		}
		_, hasId := GetNodeOrSchemaProperty(node, EntityIDTerm)
		if hasId {
			for i := len(path) - 1; i >= 0; i-- {
				if closest, exists := ret[path[i]]; exists {
					closest.IDNodes = append(closest.IDNodes, node)
					ret[path[i]] = closest
					break
				}
			}
		}
		return true
	}, SkipSchemaNodes, false)
	return ret
}

// Link the given node
func (spec LinkSpec) Link(node graph.Node, entityInfo map[graph.Node]DocumentEntityInfo) error {
	// Get the ID
	// Is this the ID node?
	var fk interface{}
	var err error
	if AsPropertyValue(node.GetProperty(SchemaNodeIDTerm)).AsString() == spec.FK {
		fk, err = GetNodeValue(node)
	} else {
		root := GetEntityRoot(node)
		if root != nil {
			return ErrCannotResolveLink(spec)
		}
		IterateDescendants(root, func(n graph.Node, _ []graph.Node) bool {
			if AsPropertyValue(n.GetProperty(SchemaNodeIDTerm)).AsString() == spec.FK {
				fk, err = GetNodeValue(n)
				return false
			}
			return true
		}, SkipSchemaNodes, false)
	}
	if err != nil {
		return err
	}
	linkTo := make([]graph.Node, 0)
	for _, info := range entityInfo {
		if len(info.IDNodes) != 1 {
			continue
		}
		nv, err := GetNodeValue(info.IDNodes[0])
		if err != nil {
			return err
		}
		if fk == nv {
			linkTo = append(linkTo, info.Root)
		}
	}
	if len(linkTo) == 0 {
		return nil
	}
	if len(linkTo) > 1 && !spec.Multi {
		return ErrMultipleTargetsFound{ID: GetNodeID(node)}
	}

	// Find the parent node of this node
	var parent graph.Node
	for edges := node.GetEdges(graph.IncomingEdge); edges.Next(); {
		edge := edges.Edge()
		if IsDocumentNode(edge.GetFrom()) {
			if parent != nil {
				return ErrInvalidLinkSpec{ID: GetNodeID(node), Msg: "Node has multiple parents"}
			}
			parent = edge.GetFrom()
		}
	}
	if parent == nil {
		return ErrInvalidLinkSpec{ID: GetNodeID(node), Msg: "Node has no parent"}
	}

	for _, target := range linkTo {
		parent.GetGraph().NewEdge(parent, target, spec.Label, nil)
	}
	return nil
}
