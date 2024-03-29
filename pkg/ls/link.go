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

	"github.com/cloudprivacylabs/lpg/v2"
)

/*

  The linking definition must include a "reference" and "link" entry.

  Types of links:

  A: {
    entityId: [idField]
  }

  B references to A using a reference field

  B -- aField--> A (ingestAs=edge) or
  B.aField --label-->A (ingestAs=node)

  B: {
    "aField": {
      "reference": "A",  // Reference to an A entity
      "fk": [aIdField]   // This is the attribute ID of a field in B that contains the A ID
      "referenceDir": -> <-  // Edge goes to A, or edge goes to B
      "multi": // Multiple references
      "ingestAs": "edge" or "node"
      "linkNode": "nodeId to create the link if aField is a value field",
      "label": "edgeLabel" if ingestAs=edge
    }
  }

The aField field may itself be a foreign key value. Then, omit fk, or use aField ID as the fk.

*/

var (
	// ReferenceFK specifies the foreign key value
	ReferenceFK = RegisterStringSliceTerm(NewTerm(LS+"Reference/", "fkValue").SetComposition(OverrideComposition).SetTags(SchemaElementTag))
	// ReferenceFKFor is used for value nodes that are foreign keys
	ReferenceFKFor = RegisterStringTerm(NewTerm(LS+"Reference/", "fkFor").SetComposition(OverrideComposition).SetTags(SchemaElementTag))
	// ReferenceFKTerm specifies the foreign key attribute ID
	ReferenceFKTerm = RegisterStringSliceTerm(NewTerm(LS+"Reference/", "fk").SetComposition(OverrideComposition).SetTags(SchemaElementTag))

	// ReferenceLabelTerm specifies the edge label between the referenced nodes
	ReferenceLabelTerm = RegisterStringTerm(NewTerm(LS+"Reference/", "label").SetComposition(OverrideComposition).SetTags(SchemaElementTag))

	// ReferenceDirectionTerm specifies the direction of the edge. If
	// "to" or "toTarget", the edge points to the target entity.
	// If "from" or "fromTarget", the edge points
	// to this entity.
	ReferenceDirectionTerm = RegisterStringTerm(NewTerm(LS+"Reference/", "dir").SetComposition(OverrideComposition).SetTags(SchemaElementTag))

	// ReferenceLinkNodeTerm specifies the node in the current entity
	// that will be linked to the other entity. If the references are
	// defined in a Reference type node, then the node itself if the
	// link. Otherwise, this gives the node that must be linked to the
	// other entity.
	ReferenceLinkNodeTerm = RegisterStringTerm(NewTerm(LS+"Reference/", "linkNode").SetComposition(OverrideComposition).SetTags(SchemaElementTag))

	// ReferenceMultiTerm specifies if there can be more than one link targets
	ReferenceMultiTerm = RegisterStringTerm(NewTerm(LS+"Reference/", "multi").SetComposition(OverrideComposition).SetTags(SchemaElementTag))
)

type ForeignKeyInfo struct {
	DocumentNodes []*lpg.Node
	ForeignKey    []string
}

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

type ErrInvalidForeignKeys struct {
	Spec LinkSpec
	Msg  string
}

func (err ErrInvalidForeignKeys) Error() string {
	return fmt.Sprintf("Invalid foreign keys: %s %s", err.Msg, GetNodeID(err.Spec.SchemaNode))
}

// LinkSpec contains the link field information
type LinkSpec struct {
	SchemaNode *lpg.Node
	// The target schema/entity reference, populated from the
	// `reference` property of the node
	TargetEntity string
	// The foreign key field(s)
	FK []string
	// The label of the link
	Label string
	// If true, the link is from this entity to the target. If false,
	// the link is from the target to this.
	Forward bool
	// If the schema node is not a reference node, then this is the node
	// that should receive the link
	LinkNode string
	// If true, the reference can have more than one links
	Multi bool
	// IngestAs node or edge
	IngestAs string

	// Initialized when linkSpec is initialized
	ParentSchemaNode *lpg.Node
}

// GetLinkSpec returns a link spec from the node if there is one. The node is a schema node
func GetLinkSpec(schemaNode *lpg.Node) (*LinkSpec, error) {
	if schemaNode == nil {
		return nil, nil
	}
	ls, ok := schemaNode.GetProperty("$linkSpec")
	if ok {
		return ls.(*LinkSpec), nil
	}

	// A reference to another entity is a reference node
	ref := ReferenceTerm.PropertyValue(schemaNode)
	if len(ref) == 0 {
		ref = ReferenceFKFor.PropertyValue(schemaNode)
		if len(ref) == 0 {
			return nil, nil
		}
	}

	link := ReferenceDirectionTerm.PropertyValue(schemaNode)
	if len(link) == 0 {
		return nil, nil
	}
	// schemaNode.SetProperty(ReferenceFK, "test_fk_val")
	ret := LinkSpec{
		SchemaNode:   schemaNode,
		TargetEntity: ref,
		IngestAs:     GetIngestAs(schemaNode),
	}
	ret.Label = ReferenceLabelTerm.PropertyValue(schemaNode)
	s := ReferenceMultiTerm.PropertyValue(schemaNode)
	ret.Multi = s != "false"
	if len(ret.Label) == 0 {
		ret.Label = AttributeNameTerm.PropertyValue(schemaNode)
	}
	if !schemaNode.GetLabels().Has(AttributeTypeReference.Name) {
		ret.LinkNode = ReferenceLinkNodeTerm.PropertyValue(schemaNode)
	} else {
		if ret.IngestAs != IngestAsNode && ret.IngestAs != IngestAsEdge {
			return nil, ErrInvalidLinkSpec{ID: GetNodeID(schemaNode), Msg: "Invalid ingestAs for link"}
		}
	}
	switch link {
	case "to", "toTarget":
		ret.Forward = true
	case "from", "fromTarget":
		ret.Forward = false
	case "":
		return nil, nil
	default:
		return nil, ErrInvalidLinkSpec{ID: GetNodeID(schemaNode), Msg: "Direction is not one of: `to`, `from`, `toTarget`, `fromTarget`"}
	}

	ret.FK = ReferenceFKTerm.PropertyValue(schemaNode)
	if len(ret.FK) == 0 {
		// If schema node is a value node, then the node is the FK
		if schemaNode.GetLabels().Has(AttributeTypeValue.Name) {
			ret.FK = []string{GetNodeID(schemaNode)}
		}
	}
	// Found a link spec. Find corresponding parent nodes in the document
	ret.ParentSchemaNode = GetParentAttribute(schemaNode)
	schemaNode.SetProperty("$linkSpec", &ret)
	return &ret, nil
}

// FindReference finds the root nodes with entitySchema=spec.Schema, with entityId=fk
func (spec *LinkSpec) FindReference(entityInfo EntityInfoIndex, fk []string) ([]*lpg.Node, error) {
	return entityInfo.Find(spec.TargetEntity, fk), nil
}

// GetForeignKeys returns the foreign keys for the link spec given the entity root node
func (spec *LinkSpec) GetForeignKeys(entityRoot *lpg.Node) ([]ForeignKeyInfo, error) {
	// There can be multiple instances of a foreign key in an
	// entity. ForeignKeyNdoes[i] keeps all the nodes for spec.FK[i]
	foreignKeyNodes := make([][]*lpg.Node, len(spec.FK))
	IterateDescendants(entityRoot, func(n *lpg.Node) bool {
		attrId := SchemaNodeIDTerm.PropertyValue(n)
		if len(attrId) == 0 {
			return true
		}
		for i := range spec.FK {
			if spec.FK[i] == attrId {
				foreignKeyNodes[i] = append(foreignKeyNodes[i], n)
			}
		}
		return true
	}, FollowEdgesInEntity, false)
	// All foreign key elements must have the same number of elements, and no index must be skipped
	var numKeys int
	for index := 0; index < len(foreignKeyNodes); index++ {
		if index == 0 {
			numKeys = len(foreignKeyNodes[index])
		} else {
			if len(foreignKeyNodes[index]) != numKeys {
				return nil, ErrInvalidForeignKeys{Spec: *spec, Msg: "Inconsistent foreign keys"}
			}
		}
	}
	// foreignKeyNodes is organized as:
	//
	//   0          1         2
	// fk0_key0  fk0_key1  fk0_key2  --> foreign key 1
	// fk1_key0  fk1_key1  fk1_key2  --> foreign key 2
	foreignKeyInfo := make([]ForeignKeyInfo, numKeys)
	for i := 0; i < numKeys; i++ {
		for key := 0; key < len(spec.FK); key++ {
			v, _ := GetRawNodeValue(foreignKeyNodes[i][key])
			foreignKeyInfo[i] = ForeignKeyInfo{
				DocumentNodes: []*lpg.Node{foreignKeyNodes[i][key]},
				ForeignKey:    []string{v},
			}
		}
	}
	return foreignKeyInfo, nil
}
