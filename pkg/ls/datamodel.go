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

// DocumentNode represents the common interface implemented by all
// ingested document nodes
type DocumentNode interface {
	// The ID of the node. Initialized when the node is created. Unique
	// in document
	GetID() string
	// Name of the node. Initialized from the ingested data element when
	// created
	GetName() string
	SetName(string)
	// Returns the associated schema node. May be nil, if the ingested
	// field has no attribute in the schema
	GetSchemaNode() *Attribute

	// Iterate every node depth first until function returns false
	Iterate(func(DocumentNode) bool) bool
}

// BasicDocumentNode is the trivial implementation of DocumentNode
type BasicDocumentNode struct {
	ID         string
	Name       string
	SchemaNode *Attribute
}

func (node *BasicDocumentNode) GetID() string             { return node.ID }
func (node *BasicDocumentNode) GetName() string           { return node.Name }
func (node *BasicDocumentNode) SetName(name string)       { node.Name = name }
func (node *BasicDocumentNode) GetSchemaNode() *Attribute { return node.SchemaNode }

// ValueNode represent simple values in a document
type ValueNode struct {
	BasicDocumentNode
	Value interface{}
}

// Calls f with v
func (v *ValueNode) Iterate(f func(DocumentNode) bool) bool {
	return f(v)
}

func NewValueNode(ID, name string, schemaNode *Attribute, value interface{}) *ValueNode {
	return &ValueNode{BasicDocumentNode: BasicDocumentNode{
		ID:         ID,
		Name:       name,
		SchemaNode: schemaNode,
	},
		Value: value,
	}
}

// ObjectNode represents objects containing key-value pairs
type ObjectNode struct {
	BasicDocumentNode
	Children []DocumentNode
}

// Iterate calls f for object and its children until f returns false
func (object *ObjectNode) Iterate(f func(DocumentNode) bool) bool {
	if !f(object) {
		return false
	}
	for i := range object.Children {
		if !object.Children[i].Iterate(f) {
			return false
		}
	}
	return true
}

func NewObjectNode(ID, name string, schemaNode *Attribute) *ObjectNode {
	return &ObjectNode{BasicDocumentNode: BasicDocumentNode{
		ID:         ID,
		Name:       name,
		SchemaNode: schemaNode}}
}

// ArrayNode represents objects containing an ordered list of element
type ArrayNode struct {
	BasicDocumentNode
	Elements []DocumentNode
}

// Iterate calls f for object and its children until f returns false
func (array *ArrayNode) Iterate(f func(DocumentNode) bool) bool {
	if !f(array) {
		return false
	}
	for i := range array.Elements {
		if !array.Elements[i].Iterate(f) {
			return false
		}
	}
	return true
}

func NewArrayNode(ID, name string, schemaNode *Attribute) *ArrayNode {
	return &ArrayNode{BasicDocumentNode: BasicDocumentNode{
		ID:         ID,
		Name:       name,
		SchemaNode: schemaNode}}
}

// NullNode is a node whose value is null. It can be an object, array, value, etc.
type NullNode struct {
	BasicDocumentNode
}

// Iterate calls f for node
func (node *NullNode) Iterate(f func(DocumentNode) bool) bool {
	return f(node)
}

func NewNullNode(ID, name string, schemaNode *Attribute) *NullNode {
	return &NullNode{BasicDocumentNode: BasicDocumentNode{
		ID:         ID,
		Name:       name,
		SchemaNode: schemaNode,
	}}
}

// DataModelToMap marshals the datamodel to a
// map[string]interface{}. If embedSchema is non-nil, schema
// attributes will be embedded
func DataModelToMap(root DocumentNode, embedSchema bool) interface{} {
	var result map[string]interface{}
	embed := func() {
		if result == nil {
			return
		}
		sch := root.GetSchemaNode()
		if sch == nil {
			if len(root.GetName()) > 0 {
				result[AttributeAnnotations.Name.ID] = []interface{}{map[string]interface{}{"@value": root.GetName()}}
			}
			return
		}
		if len(sch.ID) > 0 {
			result[DocTerms.SchemaAttributeID.ID] = []interface{}{map[string]interface{}{"@id": sch.ID}}
		}
		if embedSchema {
			for k, v := range sch.Values {
				result[k] = v
			}
		}
	}
	switch node := root.(type) {
	case *NullNode:
		result = map[string]interface{}{DocTerms.Value.ID: []interface{}{map[string]interface{}{"@value": nil}}}
		embed()

	case *ValueNode:
		result = map[string]interface{}{DocTerms.Value.ID: []interface{}{map[string]interface{}{"@value": node.Value}}}
		embed()

	case *ObjectNode:
		children := make(map[string]interface{})
		for _, ch := range node.Children {
			children[ch.GetID()] = DataModelToMap(ch, embedSchema)
		}
		result = map[string]interface{}{
			DocTerms.Attributes.ID: []interface{}{children},
		}
		embed()

	case *ArrayNode:
		el := make([]interface{}, 0, len(node.Elements))
		for _, x := range node.Elements {
			el = append(el, DataModelToMap(x, embedSchema))
		}
		result = map[string]interface{}{
			DocTerms.ArrayElements.ID: []interface{}{map[string]interface{}{"@list": el}},
		}
		embed()
	}
	return result
}
