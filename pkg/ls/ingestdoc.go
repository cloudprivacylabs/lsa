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
	"strconv"
	"strings"
	"sync"

	"github.com/cloudprivacylabs/lpg"
)

// IngestAs constants
const (
	IngestAsNode     = "node"
	IngestAsEdge     = "edge"
	IngestAsProperty = "property"
)

// NodePath contains the name components identifying a node. For JSON,
// this is the components of a JSON pointer
type NodePath []string

// String returns '.' combined path components
func (n NodePath) String() string {
	return strings.Join([]string(n), ".")
}

// Create a deep-copy of the nodepath
func (n NodePath) Copy() NodePath {
	ret := make(NodePath, len(n))
	copy(ret, n)
	return ret
}

func (n NodePath) AppendString(s string) NodePath {
	return append(n, s)
}

func (n NodePath) AppendInt(i int) NodePath {
	return append(n, strconv.Itoa(i))
}

func (n NodePath) Append(i interface{}) NodePath {
	return n.AppendString(fmt.Sprint(i))
}

type ParsedDocNode interface {
	GetSchemaNode() *lpg.Node

	// Returns value, object, array
	GetTypeTerm() string

	GetID() string

	GetValue() string
	GetValueTypes() []string

	GetAttributeIndex() int
	GetAttributeName() string

	GetChildren() []ParsedDocNode

	GetProperties() map[string]interface{}
}

// Returns the entity schema if this is a root
func getParsedDocNodeEntitySchema(node ParsedDocNode) string {
	sch := node.GetSchemaNode()
	if sch == nil {
		return ""
	}
	s := AsPropertyValue(sch.GetProperty(EntitySchemaTerm)).AsString()
	return s
}

// Collect entity ID for the entity from the given root
func collectEntityID(entityRoot ParsedDocNode) []string {
	var iterateEntity func(ParsedDocNode)
	idFields := GetEntityIDFields(entityRoot.GetSchemaNode()).MustStringSlice()
	if len(idFields) == 0 {
		return nil
	}
	idValues := make([]string, len(idFields))

	iterateEntity = func(root ParsedDocNode) {
		sch := root.GetSchemaNode()
		if sch != nil {
			schNodeID := GetNodeID(sch)
			for i := range idFields {
				if schNodeID == idFields[i] {
					if native, ok := root.(HasNativeValue); ok {
						if v, ok := native.GetNativeValue(); ok {
							idValues[i] = fmt.Sprint(v)
						} else {
							idValues[i] = root.GetValue()
						}
					} else {
						idValues[i] = root.GetValue()
					}
				}
			}
		}
		for _, child := range root.GetChildren() {
			if len(getParsedDocNodeEntitySchema(child)) != 0 {
				continue
			}
			iterateEntity(child)
		}
	}
	iterateEntity(entityRoot)
	for _, x := range idValues {
		if len(x) == 0 {
			return nil
		}
	}
	return idValues
}

type ingestedEntityInfo struct {
	schemaName string
	id         []string
	node       *lpg.Node
}

func (iei ingestedEntityInfo) check(schemaName string, id []string) bool {
	if iei.schemaName != schemaName {
		return false
	}
	if len(iei.id) != len(id) {
		return false
	}
	for i := range id {
		if id[i] != iei.id[i] {
			return false
		}
	}
	return true
}

// HasNativeValue is implemented by parsed doc nodes if the node knows its native value
type HasNativeValue interface {
	GetNativeValue() (interface{}, bool)
}

type HasLabels interface {
	GetLabels() []string
}

type ingestCursor struct {
	input      []ParsedDocNode
	output     []*lpg.Node
	entityInfo map[string][]*ingestedEntityInfo
}

func (i ingestCursor) getInput() ParsedDocNode {
	return i.input[len(i.input)-1]
}

func (i ingestCursor) getOutput() *lpg.Node {
	if len(i.output) == 0 {
		return nil
	}
	return i.output[len(i.output)-1]
}

func (i ingestCursor) findInIngestedEntityInfo(schemaName string, id []string) *ingestedEntityInfo {
	inf := i.entityInfo[schemaName]
	if len(inf) == 0 {
		return nil
	}
	for i := range inf {
		if inf[i].check(schemaName, id) {
			return inf[i]
		}
	}
	return nil
}

func (i ingestCursor) addEntityInfo(schemaName string, entityId []string, node *lpg.Node) {
	i.entityInfo[schemaName] = append(i.entityInfo[schemaName], &ingestedEntityInfo{
		schemaName: schemaName,
		id:         entityId,
		node:       node,
	})
}

type Ingester struct {
	mu                    sync.RWMutex
	Schema                *Layer
	postIngestSchemaNodes []*lpg.Node
}

func (ing *Ingester) Ingest(builder GraphBuilder, root ParsedDocNode) (*lpg.Node, error) {
	cursor := ingestCursor{
		input:      []ParsedDocNode{root},
		entityInfo: make(map[string][]*ingestedEntityInfo),
	}
	_, n, err := ingestWithCursor(builder, cursor)
	if err != nil {
		return n, err
	}
	snodes := ing.GetPostIngestSchemaNodes(root.GetSchemaNode())
	var nodeIDMap map[string][]*lpg.Node
	if len(snodes) > 0 {
		nodeIDMap = GetSchemaNodeIDMap(n)
	}
	for _, schemaNode := range snodes {
		builder.PostIngestSchemaNode(root.GetSchemaNode(), schemaNode, n, nodeIDMap)
	}
	builder.AddDefaults(root.GetSchemaNode(), n)
	return n, err
}

func (ing *Ingester) GetPostIngestSchemaNodes(schemaRootNode *lpg.Node) []*lpg.Node {
	if schemaRootNode == nil {
		return nil
	}
	ing.mu.RLock()
	if ing.postIngestSchemaNodes != nil {
		ing.mu.RUnlock()
		return ing.postIngestSchemaNodes
	}
	ing.mu.RUnlock()
	ing.mu.Lock()
	defer ing.mu.Unlock()
	if ing.postIngestSchemaNodes != nil {
		return ing.postIngestSchemaNodes
	}
	ForEachAttributeNode(schemaRootNode, func(schemaNode *lpg.Node, _ []*lpg.Node) bool {
		schemaNode.ForEachProperty(func(key string, value interface{}) bool {
			pv := AsPropertyValue(value, true)
			if pv == nil {
				return true
			}
			_, ok := pv.GetSem().Metadata.(PostIngest)
			if !ok {
				return true
			}
			ing.postIngestSchemaNodes = append(ing.postIngestSchemaNodes, schemaNode)
			return true
		})
		return true
	})
	return ing.postIngestSchemaNodes
}

// GetIngestAs returns "node", "edge", "property", or "none" based on IngestAsTerm
func GetIngestAs(schemaNode *lpg.Node) string {
	if schemaNode == nil {
		return "node"
	}
	p, ok := schemaNode.GetProperty(IngestAsTerm)
	if !ok {
		return "node"
	}
	s := AsPropertyValue(p, ok).AsString()
	if s == "edge" || s == "property" || s == "none" {
		return s
	}
	return "node"
}

func GetIngestAsProperty(schemaNode *lpg.Node) (asPropertyOf, propertyName string) {
	asPropertyOf = AsPropertyValue(schemaNode.GetProperty(AsPropertyOfTerm)).AsString()
	propertyName = AsPropertyValue(schemaNode.GetProperty(PropertyNameTerm)).AsString()
	if len(propertyName) == 0 {
		propertyName = AsPropertyValue(schemaNode.GetProperty(AttributeNameTerm)).AsString()
	}
	if len(propertyName) == 0 {
		propertyName = GetNodeID(schemaNode)
	}
	return
}

func ingestWithCursor(builder GraphBuilder, cursor ingestCursor) (bool, *lpg.Node, error) {
	root := cursor.getInput()
	schemaNode := root.GetSchemaNode()
	typeTerm := root.GetTypeTerm()
	setID := func(node *lpg.Node) {
		if node != nil {
			if id := root.GetID(); len(id) > 0 {
				SetNodeID(node, id)
			}
		}
	}
	setLabels := func(node *lpg.Node) {
		lbl, ok := root.(HasLabels)
		if !ok {
			return
		}
		labels := node.GetLabels()
		labels.Add(lbl.GetLabels()...)
		node.SetLabels(labels)
	}
	setProp := func(node *lpg.Node) {
		node.SetProperty(AttributeIndexTerm, StringPropertyValue(AttributeIndexTerm, strconv.Itoa(root.GetAttributeIndex())))
		if s := root.GetAttributeName(); len(s) > 0 {
			node.SetProperty(AttributeNameTerm, StringPropertyValue(AttributeNameTerm, s))
		}
		for k, v := range root.GetProperties() {
			node.SetProperty(k, v)
		}
	}
	hasData := false
	if typeTerm == AttributeTypeValue {
		setValue := func(node *lpg.Node) error {
			SetRawNodeValue(node, root.GetValue())
			return nil
		}
		hasNativeValue := false
		var nativeValue interface{}
		nvi, hn := root.(HasNativeValue)
		if hn {
			nativeValue, hasNativeValue = nvi.GetNativeValue()
			if hasNativeValue {
				setValue = func(node *lpg.Node) error {
					err := SetNodeValue(node, nativeValue)
					return err
				}
			}
		}
		switch GetIngestAs(schemaNode) {
		case "node":

			var entitySchema string
			var entityId []string

			if entitySchema = getParsedDocNodeEntitySchema(root); len(entitySchema) > 0 {
				entityId = collectEntityID(root)
				if len(entityId) > 0 {
					ei := cursor.findInIngestedEntityInfo(entitySchema, entityId)
					if ei != nil {
						builder.GetGraph().NewEdge(cursor.getOutput(), ei.node, HasTerm, nil)
						return true, ei.node, nil
					}
				}
			}
			_, node, err := builder.ValueAsNode(schemaNode, cursor.getOutput(), setValue)
			if err != nil {
				return false, nil, err
			}
			if node != nil {
				setID(node)
				setProp(node)
				setLabels(node)
				hasData = true
			}
			if err := builder.PostNodeIngest(schemaNode, node); err != nil {
				return hasData, node, err
			}
			if len(entitySchema) > 0 && len(entityId) > 0 {
				// If here, we created a new entity root node
				cursor.addEntityInfo(entitySchema, entityId, node)
			}
			return hasData, node, nil
		case "edge":
			edge, err := builder.ValueAsEdge(schemaNode, cursor.getOutput(), setValue)
			if err != nil {
				return false, nil, err
			}
			if edge == nil {
				return false, nil, nil
			}
			setID(edge.GetTo())
			setProp(edge.GetTo())
			if err := builder.PostNodeIngest(schemaNode, edge.GetTo()); err != nil {
				return true, edge.GetTo(), err
			}
			return true, edge.GetTo(), nil
		case "property":
			var err error
			if hasNativeValue {
				err = builder.NativeValueAsProperty(schemaNode, cursor.output, nativeValue)
			} else {
				err = builder.RawValueAsProperty(schemaNode, cursor.output, root.GetValue())
			}
			if err != nil {
				return false, nil, err
			}
			return true, nil, nil
		case "none":
			return false, nil, nil
		}
		return false, nil, nil
	}
	newCursor := cursor
	switch GetIngestAs(schemaNode) {
	case "node":
		var entitySchema string
		var entityId []string

		if entitySchema = getParsedDocNodeEntitySchema(root); len(entitySchema) > 0 {
			entityId = collectEntityID(root)
			if len(entityId) > 0 {
				ei := cursor.findInIngestedEntityInfo(entitySchema, entityId)
				if ei != nil {
					// connect
					builder.GetGraph().NewEdge(cursor.getOutput(), ei.node, HasTerm, nil)
					return true, ei.node, nil
				}
			}
		}
		_, node, err := builder.CollectionAsNode(schemaNode, cursor.getOutput(), typeTerm)
		if err != nil {
			return false, nil, err
		}
		setID(node)
		setProp(node)
		setLabels(node)
		newCursor.output = append(newCursor.output, node)
		hasData = true
		if len(entitySchema) > 0 && len(entityId) > 0 {
			// If here, we created a new entity root node
			cursor.addEntityInfo(entitySchema, entityId, node)
		}
	case "edge":
		edge, err := builder.CollectionAsEdge(schemaNode, cursor.getOutput(), typeTerm)
		if err != nil {
			return false, nil, err
		}
		setID(edge.GetTo())
		setProp(edge.GetTo())
		newCursor.output = append(newCursor.output, edge.GetTo())
		hasData = true
	case "none":
	}
	newCursor.input = append(newCursor.input, nil)
	hasChildren := false
	for _, child := range root.GetChildren() {
		newCursor.input[len(newCursor.input)-1] = child
		ch, node, err := ingestWithCursor(builder, newCursor)
		if ch {
			hasChildren = true
		}
		if err != nil {
			return hasData, nil, err
		}
		if node != nil {
			n := child.GetAttributeIndex()
			if newCursor.getOutput() != nil {
				n = newCursor.getOutput().GetEdges(lpg.OutgoingEdge).MaxSize() - 1
			}
			if n == -1 {
				n = child.GetAttributeIndex()
			}
			node.SetProperty(AttributeIndexTerm, IntPropertyValue(AttributeIndexTerm, n))
		}
	}
	if err := builder.PostNodeIngest(schemaNode, newCursor.getOutput()); err != nil {
		return hasData, newCursor.getOutput(), err
	}
	if schemaNode != nil && hasData {
		switch AsPropertyValue(schemaNode.GetProperty(ConditionalTerm)).AsString() {
		case "mustHaveChildren":
			if !hasChildren {
				newCursor.getOutput().DetachAndRemove()
				return false, nil, nil
			}
		}
	}
	return hasData, newCursor.getOutput(), nil
}
