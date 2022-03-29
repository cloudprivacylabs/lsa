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
         or
      "entitySchema": "A"
      "fk": [aIdField]   // This is the attribute ID of a field in B that contains the A ID
      "link": -> <-  // Edge goes to A, or edge goes to B
      "multi": // Multiple references
      "ingestAs": "edge" or "node"
      "label": "edgeLabel" if ingestAs=edge
    }
  }


*/

var (
	// ReferenceFKTerm specifies the foreign key attribute ID
	ReferenceFKTerm = NewTerm(LS+"Reference/", "fk", false, false, OverrideComposition, nil)

	// ReferenceLabelTerm specifies the edge label between the referenced nodes
	ReferenceLabelTerm = NewTerm(LS+"Reference/", "label", false, false, OverrideComposition, nil)

	// ReferenceLinkTerm specifies the direction of the edge. If
	// ->, the edge points to the target entity. If <-, the edge points
	// to this entity.
	ReferenceLinkTerm = NewTerm(LS+"Reference/", "link", false, false, OverrideComposition, nil)

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

	// A reference to another entity is either a reference node, or a node that has Entity schema reference in it
	ref := AsPropertyValue(schemaNode.GetProperty(ReferenceTerm)).AsString()
	if len(ref) == 0 {
		ref = AsPropertyValue(schemaNode.GetProperty(EntitySchemaTerm)).AsString()
		if len(ref) == 0 {
			return nil, nil
		}
	}

	link := AsPropertyValue(schemaNode.GetProperty(ReferenceLinkTerm))
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
	switch link.AsString() {
	case "to":
		ret.Forward = true
	case "from":
		ret.Forward = false
	case "":
		return nil, nil
	default:
		return nil, ErrInvalidLinkSpec{ID: GetNodeID(schemaNode), Msg: "Direction is not one of: `to`, `from`"}
	}

	if ret.IngestAs != IngestAsNode && ret.IngestAs != IngestAsEdge {
		return nil, ErrInvalidLinkSpec{ID: GetNodeID(schemaNode), Msg: "Invalid ingestAs for link"}
	}
	fk := AsPropertyValue(schemaNode.GetProperty(ReferenceFKTerm))
	if fk.IsString() {
		ret.FK = []string{fk.AsString()}
	}
	if fk.IsStringSlice() {
		ret.FK = fk.AsStringSlice()
	}
	if len(ret.FK) == 0 {
		return nil, ErrInvalidLinkSpec{ID: GetNodeID(schemaNode), Msg: "Empty foreign key"}
	}
	schemaNode.SetProperty("$linkSpec", &ret)
	return &ret, nil
}

// Link the given node, or create a link from the parent node as the
// docNode may not exist.
func (ingester *Ingester) Link(ictx IngestionContext, spec *LinkSpec, docNode, parentNode graph.Node, entityInfo map[graph.Node]EntityInfo) error {
	entityRoot := GetEntityRoot(parentNode)
	if entityRoot == nil {
		return ErrCannotResolveLink(*spec)
	}
	foreignKeyNodes := make([][]graph.Node, len(spec.FK))
	IterateDescendants(entityRoot, func(n graph.Node, _ []graph.Node) bool {
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
				return ErrInvalidForeignKeys{Spec: *spec, Msg: "Inconsistent foreign keys"}
			}
		}
	}
	if numKeys == 0 {
		// Nothing to link
		return nil
	}
	if numKeys > 1 && !spec.Multi {
		return ErrInvalidForeignKeys{Spec: *spec, Msg: "Multiple foreign key values not allowed"}
	}

	g := parentNode.GetGraph()
	var nodeProperties map[string]interface{}
	if spec.IngestAs == IngestAsEdge && docNode != nil {
		// This document node is removed and a link from the parent to the target is created
		nodeProperties = CloneProperties(docNode)
		docNode.DetachAndRemove()
	}

	// Find remote references
	for i := 0; i < numKeys; i++ {
		fk := make([]string, len(foreignKeyNodes))
		for k, v := range foreignKeyNodes {
			fk[k], _ = GetRawNodeValue(v[i])
		}
		ref, err := spec.FindReference(entityInfo, fk)
		if err != nil {
			return err
		}
		if len(ref) == 0 {
			continue
		}
		for _, linkRef := range ref {
			if spec.IngestAs == IngestAsEdge {
				// Node is already removed. Make an edge
				if spec.Forward {
					g.NewEdge(parentNode, linkRef, spec.Label, nodeProperties)
				} else {
					g.NewEdge(linkRef, parentNode, spec.Label, nodeProperties)
				}
			} else {
				if docNode == nil {
					ctx := ictx
					ctx.SourcePath = nil
					ctx.SchemaPath = []graph.Node{spec.SchemaNode}
					ctx.GraphPath = []graph.Node{parentNode}
					docNode = ingester.NewNode(ctx)
					g.NewEdge(parentNode, docNode, HasTerm, nil)
				}
				// A link from this document node to target is created
				if spec.Forward {
					g.NewEdge(docNode, linkRef, spec.Label, nil)
				} else {
					g.NewEdge(linkRef, docNode, spec.Label, nil)
				}
			}
		}
	}

	return nil
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
		if exists {
			id := ei.GetID()
			if len(id) != len(fk) {
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
