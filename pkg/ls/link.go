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

	"github.com/cloudprivacylabs/opencypher/graph"
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
	// ReferenceFKFor is used for value nodes that are foreign keys
	ReferenceFKFor = NewTerm(LS+"Reference/", "fkFor", false, false, OverrideComposition, nil)
	// ReferenceFKTerm specifies the foreign key attribute ID
	ReferenceFKTerm = NewTerm(LS+"Reference/", "fk", false, false, OverrideComposition, nil)

	// ReferenceLabelTerm specifies the edge label between the referenced nodes
	ReferenceLabelTerm = NewTerm(LS+"Reference/", "label", false, false, OverrideComposition, nil)

	// ReferenceDirectionTerm specifies the direction of the edge. If
	// "to" or "toTarget", the edge points to the target entity.
	// If "from" or "fromTarget", the edge points
	// to this entity.
	ReferenceDirectionTerm = NewTerm(LS+"Reference/", "dir", false, false, OverrideComposition, nil)

	// ReferenceLinkNodeTerm specifies the node in the current entity
	// that will be linked to the other entity. If the references are
	// defined in a Reference type node, then the node itself if the
	// link. Otherwise, this gives the node that must be linked to the
	// other entity.
	ReferenceLinkNodeTerm = NewTerm(LS+"Reference/", "linkNode", false, false, OverrideComposition, nil)

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

type ErrInvalidForeignKeys struct {
	Spec LinkSpec
	Msg  string
}

func (err ErrInvalidForeignKeys) Error() string {
	return fmt.Sprintf("Invalid foreign keys: %s %s", err.Msg, GetNodeID(err.Spec.SchemaNode))
}

// LinkSpec contains the link field information
type LinkSpec struct {
	SchemaNode graph.Node
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
}

// GetLinkSpec returns a link spec from the node if there is one. The node is a schema node
func GetLinkSpec(schemaNode graph.Node) (*LinkSpec, error) {
	if schemaNode == nil {
		return nil, nil
	}
	ls, ok := schemaNode.GetProperty("$linkSpec")
	if ok {
		return ls.(*LinkSpec), nil
	}

	// A reference to another entity is a reference node
	ref := AsPropertyValue(schemaNode.GetProperty(ReferenceTerm)).AsString()
	if len(ref) == 0 {
		ref = AsPropertyValue(schemaNode.GetProperty(ReferenceFKFor)).AsString()
		if len(ref) == 0 {
			return nil, nil
		}
	}

	link := AsPropertyValue(schemaNode.GetProperty(ReferenceDirectionTerm))
	if link == nil {
		return nil, nil
	}
	ret := LinkSpec{
		SchemaNode:   schemaNode,
		TargetEntity: ref,
		Label:        AsPropertyValue(schemaNode.GetProperty(ReferenceLabelTerm)).AsString(),
		Multi:        AsPropertyValue(schemaNode.GetProperty(ReferenceMultiTerm)).AsString() != "false",
		IngestAs:     GetIngestAs(schemaNode),
	}
	if len(ret.Label) == 0 {
		ret.Label = AsPropertyValue(schemaNode.GetProperty(AttributeNameTerm)).AsString()
	}
	if !schemaNode.GetLabels().Has(AttributeTypeReference) {
		ret.LinkNode = AsPropertyValue(schemaNode.GetProperty(ReferenceLinkNodeTerm)).AsString()
	} else {
		if ret.IngestAs != IngestAsNode && ret.IngestAs != IngestAsEdge {
			return nil, ErrInvalidLinkSpec{ID: GetNodeID(schemaNode), Msg: "Invalid ingestAs for link"}
		}
	}
	switch link.AsString() {
	case "to", "toTarget":
		ret.Forward = true
	case "from", "fromTarget":
		ret.Forward = false
	case "":
		return nil, nil
	default:
		return nil, ErrInvalidLinkSpec{ID: GetNodeID(schemaNode), Msg: "Direction is not one of: `to`, `from`, `toTarget`, `fromTarget`"}
	}

	fk := AsPropertyValue(schemaNode.GetProperty(ReferenceFKTerm))
	if fk.IsString() {
		ret.FK = []string{fk.AsString()}
	}
	if fk.IsStringSlice() {
		ret.FK = fk.AsStringSlice()
	}
	if len(ret.FK) == 0 {
		// If schema node is a value node, then the node is the FK
		if schemaNode.GetLabels().Has(AttributeTypeValue) {
			ret.FK = []string{GetNodeID(schemaNode)}
		}
	}
	schemaNode.SetProperty("$linkSpec", &ret)
	return &ret, nil
}

// FindReference finds the root nodes with entitySchema=spec.Schema, with entityId=fk
func (spec *LinkSpec) FindReference(entityInfo map[graph.Node]EntityInfo, fk []string) ([]graph.Node, error) {
	ret := make([]graph.Node, 0)
	for _, ei := range entityInfo {
		var exists bool
		for _, typeName := range ei.GetValueType() {
			if typeName == spec.TargetEntity {
				exists = true
				break
			}
		}
		if exists || ei.GetEntitySchema() == spec.TargetEntity {
			id := ei.GetID()
			if len(fk) > 0 && len(id) != len(fk) {
				continue
			}
			found := true
			for i := range fk {
				if id[i] != fk[i] {
					found = false
					break
				}
			}
			if found {
				ret = append(ret, ei.GetRoot())
			}
		}
	}

	return ret, nil
}

// GetForeignKeys returns the foreign keys for the link spec given the entity root node
func (spec *LinkSpec) GetForeignKeys(entityRoot graph.Node) ([][]string, error) {
	// There can be multiple instances of a foreign key in an
	// entity. ForeignKeyNdoes[i] keeps all the nodes for spec.FK[i]
	foreignKeyNodes := make([][]graph.Node, len(spec.FK))
	IterateDescendants(entityRoot, func(n graph.Node) bool {
		attrId := AsPropertyValue(n.GetProperty(SchemaNodeIDTerm)).AsString()
		if len(attrId) == 0 {
			return true
		}
		for i := range spec.FK {
			if spec.FK[i] == attrId {
				foreignKeyNodes[i] = append(foreignKeyNodes[i], n)
			}
		}
		return true
	}, OnlyDocumentNodes, false)
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
	fks := make([][]string, numKeys)
	for i := 0; i < numKeys; i++ {
		for key := 0; key < len(spec.FK); key++ {
			v, _ := GetRawNodeValue(foreignKeyNodes[i][key])
			fks[i] = append(fks[i], v)
		}
	}
	return fks, nil
}
