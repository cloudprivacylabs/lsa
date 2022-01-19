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
)

var (
	// ReferenceFieldsTerm specifies the id matching vector based on the
	// values in this entity. For instance ["$field"] will create a
	// vector using the contents of the field `field`. A vector
	// `["$field","value"]` will use the value of `field` and the string
	// "value" as the id vector
	ReferenceFieldsTerm = NewTerm(LS+"Reference/fields", false, false, OverrideComposition, nil)

	// ReferenceTargetFieldsTerm is a string of slice whose elements
	// match the elements of ReferenceFieldsTerm. The referenced object
	// is found by finding an instance whose id vector is obtained by
	// the fields in this property. The field values are access by
	// `$field` notation
	ReferenceTargetFieldsTerm = NewTerm(LS+"Reference/targetFields", false, false, OverrideComposition, nil)

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

// LinkSpec contains the link field information
type LinkSpec struct {
	// The target schema/entity reference, populated from the
	// `reference` property of the node
	TargetEntity string
	// The field IDs, or values in this entity to create an ID vector
	IDFields []string
	// The field IDs in the remote entity that when all values are equal
	// to the ID vector, a link will be created
	TargetFields []string
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
// spec if there is one
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
func CompileReferenceLinkSpec(node Node) (LinkSpec, bool, error) {
	if !node.GetTypes().HasAll(AttributeTypes.Attribute, AttributeTypes.Reference) {
		return LinkSpec{}, false, nil
	}
	properties := node.GetProperties()
	ref := properties[LayerTerms.Reference].AsString()
	if len(ref) == 0 {
		return LinkSpec{}, false, nil
	}
	fields := properties[ReferenceFieldsTerm].MustStringSlice()
	targetFields := properties[ReferenceTargetFieldsTerm].MustStringSlice()
	label := properties[ReferenceLabelTerm].AsString()
	dir := properties[ReferenceDirectionTerm].AsString()
	multi := properties[ReferenceMultiTerm].AsString()
	if len(fields) == 0 && len(targetFields) == 0 {
		return LinkSpec{}, false, nil
	}
	if len(fields) != len(targetFields) {
		return LinkSpec{}, false, ErrInvalidLinkSpec{ID: node.GetID(), Msg: "fields and targetFields have different lengths"}
	}
	if len(label) == 0 {
		label = HasTerm
	}
	spec := LinkSpec{
		TargetEntity: ref,
		IDFields:     fields,
		TargetFields: targetFields,
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

// IDVectorElement contains either a node or a value
type IDVectorElement struct {
	Node  Node
	Value string
}

// GetIDVectorNodes collects the schema nodes specified in the
// vector. The value of a schema node is specified as $attrId, where
// attrId is the schema node ID. This function fills the vector by
// finding document nodes that are an instance of that schema
// node. The returned array contains either Node or string values
func GetIDVectorNodes(schemaRoot Node, idVector []string) ([]IDVectorElement, error) {
	ret := make([]IDVectorElement, len(idVector))
	searchFieldSet := make(map[string]int)
	nFound := 0
	// Fill values first
	for i, val := range idVector {
		if len(val) > 0 && val[0] != '$' {
			ret[i] = IDVectorElement{Value: val}
			nFound++
		} else {
			searchFieldSet[val[1:]] = i
		}
	}
	if nFound == len(idVector) {
		return ret, nil
	}
	zero := IDVectorElement{}
	var err error
	IterateDescendants(schemaRoot, func(node Node, _ []Node) bool {
		// If this node is an instance of one of the idVector values, get its value
		inst := InstanceOfID(node)
		for _, t := range inst {
			if ix, has := searchFieldSet[t]; has {
				if ret[ix] != zero {
					err = ErrInvalidLinkSpec{ID: t, Msg: "Node has multiple id vector values"}
					return false
				}
				ret[ix] = IDVectorElement{Node: node}
				nFound++
				break
			}
		}
		return nFound < len(idVector)
	}, func(_ Edge, _ []Node) EdgeFuncResult { return FollowEdgeResult }, false)

	return ret, err
}

// GetIDVector finds the ID vector for the document at the given root,
// based on the ID vector elements
func GetIDVector(docRoot Node, elements []IDVectorElement) ([][]string, error) {
	ret := make([][]string, 0)

	IterateDescendants(docRoot, func(node Node, _ []Node) bool {
		if !node.GetTypes().Has(DocumentNodeTerm) {
			return true
		}

		return true
	}, func(edge Edge, _ []Node) EdgeFuncResult {
		if edge.GetLabel() == HasTerm {
			return FollowEdgeResult
		}
		return SkipEdgeResult
	}, false)
	return ret, nil
}
