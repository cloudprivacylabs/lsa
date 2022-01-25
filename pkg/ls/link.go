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

	"github.com/bserdar/digraph"
)

var (
	// ReferenceFKTerm specifies the foreign key attribute ID
	ReferenceFKTerm = NewTerm(LS+"Reference/fk", false, false, OverrideComposition, nil)

	// ReferenceTargetTerm specifies the target entity if the node is not a reference node
	ReferenceTargetTerm = NewTerm(LS+"Reference/target", false, false, OverrideComposition, nil)

	// ReferenceTargetIDTerm is the target schema ID field. If not specified, entity ID is used
	ReferenceTargetIDTerm = NewTerm(LS+"Reference/targetId", false, false, OverrideComposition, nil)

	// ReferenceLabelTerm specifies the edge label between the referenced nodes
	ReferenceLabelTerm = NewTerm(LS+"Reference/label", false, false, OverrideComposition, nil)

	// ReferenceDirectionTerm specifies the direction of the edge. If
	// ->, the edge points to the target entity. If <-, the edge points
	// to this entity.
	ReferenceDirectionTerm = NewTerm(LS+"Reference/direction", false, false, OverrideComposition, nil)

	// ReferenceMultiTerm specifies if there can be more than one link targets
	ReferenceMultiTerm = NewTerm(LS+"Reference/multi", false, false, OverrideComposition, nil)
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

type linkSpecKeyType struct{}

var linkSpecKey linkSpecKeyType

// GetCompiledReferenceLinkSpec returns the compiled reference link
// spec if there is one in the compiled property map of the node
func GetCompiledReferenceLinkSpec(node Node) (LinkSpec, bool) {
	spec, exists := node.GetCompiledProperties().GetCompiledProperty(linkSpecKey)
	if exists {
		return spec.(LinkSpec), true
	}
	return LinkSpec{}, false
}

// CompileReferenceLinkSpec gets an uncompiled reference node, and
// puts a LinkSpec into compiled data map of the node if the node
// specifies a link. It also returns the LinkSpec and true. If the
// node is not a reference node, or if the node does not specify a
// link, returns false
func CompileReferenceLinkSpec(layer *Layer, node Node) (LinkSpec, bool, error) {
	if !node.GetTypes().Has(AttributeTypes.Attribute) {
		return LinkSpec{}, false, nil
	}
	var ref string
	if node.GetTypes().Has(AttributeTypes.Reference) {
		ref = node.GetProperties()[LayerTerms.Reference].AsString()
	} else {
		ref = node.GetProperties()[ReferenceTargetTerm].AsString()
	}
	if len(ref) == 0 {
		return LinkSpec{}, false, nil
	}
	properties := node.GetProperties()
	fk := properties[ReferenceFKTerm].AsString()
	targetID := properties[ReferenceTargetIDTerm].AsString()
	label := properties[ReferenceLabelTerm].AsString()
	dir := properties[ReferenceDirectionTerm].AsString()
	multi := properties[ReferenceMultiTerm].AsString()
	if len(fk) == 0 {
		return LinkSpec{}, false, nil
	}
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
		return LinkSpec{}, true, ErrInvalidLinkSpec{ID: node.GetID(), Msg: "Direction is not one of: ->, <-"}
	}
	node.GetCompiledProperties().SetCompiledProperty(linkSpecKey, spec)
	return spec, true, nil
}

// GetAllLinkSpecs returns all linkspecs under the given root node
func GetAllLinksSpecs(root Node) map[Node]LinkSpec {
	ret := make(map[Node]LinkSpec)
	IterateDescendants(root, func(node Node, _ []Node) bool {
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
	Root    Node
	Schema  string
	IDNodes []Node
}

// GetDocumentEntityInfo returns a map containing all entity root
// nodes, with entity info as map values
func GetDocumentEntityInfo(docRoot Node) map[Node]DocumentEntityInfo {
	ret := make(map[Node]DocumentEntityInfo)
	IterateDescendants(docRoot, func(node Node, path []Node) bool {
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
func (spec LinkSpec) Link(entityRoot, node Node, entityInfo map[Node]DocumentEntityInfo) error {
	// Get the ID
	// Is this the ID node?
	var fk interface{}
	if node.GetProperties()[SchemaNodeIDTerm].AsString() == spec.FK {
		fk = GetNodeValue()
	} else {
		IterateDescendants(entityRoot, func(n Node, _ []Node) bool {
			if n.GetProperties()[SchemaNodeIDTerm].AsString() == spec.FK {
				fk = GetNodeValue(n)
				return false
			}
			return true
		}, SkipSchemaNodes, false)
	}
	linkTo := make([]Node, 0)
	for _, info := range entityInfo {
		if len(info.IDNodes) != 1 {
			continue
		}
		if fk == GetNodeValue(info.IDNodes[0]) {
			linkTo = append(linkTo, info.Root)
		}
	}
	if len(linkTo) == 0 {
		return nil
	}
	if len(linkTo) > 1 && !spec.Multi {
		return ErrMultipleTargetsFound{ID: node.GetID()}
	}
	for _, target := range linkTo {
		digraph.Connect(node, target, NewEdge(spec.Label))
	}
	return nil
}
