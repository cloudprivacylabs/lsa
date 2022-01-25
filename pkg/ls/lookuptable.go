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

type ErrInvalidLookupTable struct {
	ID  string
	Msg string
}

func (e ErrInvalidLookupTable) Error() string {
	return fmt.Sprintf("Invalid lookup table %s: %s", e.ID, e.Msg)
}

// LookupTableElement is the JSON schema representation of a lookup table item. It
// is also the JSON-LD representation under the ls:/lookupTable/element namespace
type LookupTableElement struct {
	// Possible values the data point can take. If empty, this is the default option
	Options []string `json:"options"`
	// If the options are to be compared case sensitive
	CaseSensitive bool `json:"caseSensitive"`
	// This is the value that has to be returned for the field if options match, or if this is the default
	Value string `json:"value"`
	// If set, data ingestion/set/get must fail with this error
	Error string `json:"error"`
}

// LookupTable is the JSON representation of a lookup table containing the ID and elements
type LookupTable struct {
	// Lookup table ID
	ID       string               `json:"id"`
	Ref      string               `json:"ref"`
	Elements []LookupTableElement `json:"elements"`
}

var LookupTableTerm = NewTerm(LS+"lookupTable", false, false, OverrideComposition, struct {
	lookupTableMarshaler
}{
	lookupTableMarshaler{},
})

var LookupTableElementsTerm = NewTerm(LS+"lookupTable/elements", false, true, NoComposition, nil)

// LookupTableReferenceTerm is used as a node type for nodes whose
// lookup table references are not resolved. These nodes should be
// resolved to real lookup tables during compilation
var LookupTableReferenceTerm = NewTerm(LS+"lookupTable/ref", true, false, NoComposition, nil)

var LookupTableElementOptionsTerm = NewTerm(LS+"lookupTable/element/options", false, false, NoComposition, nil)
var LookupTableElementCaseSensitiveTerm = NewTerm(LS+"lookupTable/element/caseSensitive", false, false, NoComposition, nil)
var LookupTableElementValueTerm = NewTerm(LS+"lookupTable/element/value", false, false, NoComposition, nil)
var LookupTableElementErrorTerm = NewTerm(LS+"lookupTable/element/error", false, false, NoComposition, nil)

type lookupTableMarshaler struct{}

func (lookupTableMarshaler) UnmarshalLd(target *Layer, key string, value interface{}, node *LDNode, allNodes map[string]*LDNode, interner Interner) error {
	return lookupTableMarshaler.unmarshalLookupTable(lookupTableMarshaler{}, key, value, node, allNodes)
}

// unmarshalLookupTable flattened jsonld lookup table
//
// {
//   "@id": "fieldId",
//   "@type": ["https://lschema.org/Value"],
//   "https://lschema.org/lookupTable": [
//      {
//        "@id": "_:b1"
//      }
//   ]
// },
// {
//   "@id": "_:b1",
//   "https://lschema.org/lookupTable/elements": [{
//     "@list": [
//        {
//          "@id": "_:b2"
//        }
//     ]
//   }]
// },
// {
//   "@id": "_:b2",
//   "https://lschema.org/lookupTable/element/options": [
//     "a",
//     "b"
//   ],
//   "https://lschema.org/lookupTable/element/value": "a"
// },
func (lookupTableMarshaler) unmarshalLookupTable(key string, value interface{}, node *LDNode, allNodes map[string]*LDNode) error {
	// key-value is:
	//   "https://lschema.org/lookupTable": [{
	//     "@id": "_:b1"
	//   }]
	id := GetNodeID(value)
	if len(id) == 0 {
		return ErrInvalidLookupTable{ID: node.ID, Msg: "Expecting to see @id under `lookupTable` term value"}
	}
	lookupContents := allNodes[id]
	if lookupContents == nil {
		// This is an external lookup table. Create a reference node for it
		referenceNode := NewNode(id, LookupTableReferenceTerm)
		// Put that into allNodes
		newNode := &LDNode{
			ID:        id,
			Types:     []string{LookupTableReferenceTerm},
			GraphNode: referenceNode,
		}
		allNodes[id] = newNode
		// Connect the attribute node to the reference node
		digraph.Connect(node.GraphNode, referenceNode, NewEdge(LookupTableTerm))
		return nil
	}

	// The lookup table contents (or reference) is in this schema If we
	// are referring to an already processed lookup table, simply
	// connect this node to the table
	if lookupContents.GraphNode != nil && lookupContents.GraphNode.GetTypes().Has(LookupTableTerm) {
		digraph.Connect(node.GraphNode, lookupContents.GraphNode, NewEdge(LookupTableTerm))
		return nil
	}

	// This is the first time we are working on this lookup table

	// Create the root node
	lookupTableNode := lookupContents.GraphNode
	if lookupTableNode == nil {
		lookupTableNode := NewNode(id)
		lookupContents.GraphNode = lookupTableNode
	}
	lookupTableNode.GetTypes().Add(LookupTableTerm)
	digraph.Connect(node.GraphNode, lookupTableNode, NewEdge(LookupTableTerm))

	// Create options
	for index, element := range GetLDListElements(lookupContents.Node[LookupTableElementsTerm]) {
		elementNodeID := GetNodeID(element)
		elementLDNode := allNodes[elementNodeID]
		if elementLDNode == nil {
			return ErrInvalidLookupTable{ID: node.ID, Msg: fmt.Sprintf("Cannot find element node with '%s'", elementNodeID)}
		}
		elementNode := NewNode(elementNodeID, LookupTableElementsTerm)
		elementNode.SetIndex(index)
		for k, v := range elementLDNode.Node {
			if k[0] == '@' {
				continue
			}
			arr, ok := v.([]interface{})
			if !ok {
				continue
			}
			if len(arr) == 0 {
				continue
			}
			switch k {
			case LookupTableElementOptionsTerm:
				val := make([]string, 0, len(arr))
				for _, x := range arr {
					val = append(val, GetStringValue("@value", x))
				}
				elementNode.GetProperties()[k] = StringSlicePropertyValue(val)
			default:
				elementNode.GetProperties()[k] = StringPropertyValue(GetStringValue("@value", arr[0]))
			}
		}
		digraph.Connect(lookupTableNode, elementNode, NewEdge(LookupTableElementsTerm))
	}
	return nil
}

func unmarshalLookupTableElement(in map[string]interface{}) (LookupTableElement, error) {
	ret := LookupTableElement{}
	for k, v := range in {
		switch k {
		case "options", LookupTableElementOptionsTerm:
			if s, ok := v.(string); ok {
				ret.Options = []string{s}
			} else if arr, ok := v.([]interface{}); ok {
				for _, val := range arr {
					if s, ok := val.(string); ok {
						ret.Options = append(ret.Options, s)
					} else {
						return LookupTableElement{}, ErrInvalidLookupTable{Msg: "Non-string lookup table element option"}
					}
				}
			}
		case "caseSensitive", LookupTableElementCaseSensitiveTerm:
			if s, ok := v.(bool); ok {
				ret.CaseSensitive = s
			} else {
				return LookupTableElement{}, ErrInvalidLookupTable{Msg: "Non-boolean lookup table element case sensitive flag"}
			}
		case "value", LookupTableElementValueTerm:
			if s, ok := v.(string); ok {
				ret.Value = s
			} else {
				return LookupTableElement{}, ErrInvalidLookupTable{Msg: "Non-string lookup table element value"}
			}

		case "error", LookupTableElementErrorTerm:
			if s, ok := v.(string); ok {
				ret.Error = s
			} else {
				return LookupTableElement{}, ErrInvalidLookupTable{Msg: "Non-string lookup table element error"}
			}
		}
	}
	return ret, nil
}

func unmarshalLookupTableElements(in []interface{}) ([]LookupTableElement, error) {
	ret := make([]LookupTableElement, 0, len(in))
	for _, x := range in {
		m, ok := x.(map[string]interface{})
		if !ok {
			return nil, ErrInvalidLookupTable{Msg: "Expecting an object for lookup table element"}
		}
		el, err := unmarshalLookupTableElement(m)
		if err != nil {
			return nil, err
		}
		ret = append(ret, el)
	}
	return ret, nil
}

func unmarshalLookupTable(in interface{}) (LookupTable, error) {
	m, ok := in.(map[string]interface{})
	if !ok {
		return LookupTable{}, ErrInvalidLookupTable{Msg: "JSON object expected for lookup table"}
	}

	ret := LookupTable{}
	for k, v := range m {
		var ok bool
		switch k {
		case "id", "@id":
			ret.ID, ok = v.(string)
			if !ok {
				return LookupTable{}, ErrInvalidLookupTable{Msg: "String expected for id"}
			}
		case "ref", LookupTableReferenceTerm:
			ret.Ref, ok = v.(string)
			if !ok {
				return LookupTable{}, ErrInvalidLookupTable{Msg: "String expected for ref"}
			}
		case "elements", LookupTableElementsTerm:
			arr, ok := v.([]interface{})
			if !ok {
				return LookupTable{}, ErrInvalidLookupTable{Msg: "Array expected for elements"}
			}
			var err error
			ret.Elements, err = unmarshalLookupTableElements(arr)
			if err != nil {
				return LookupTable{}, err
			}
		}
	}
	return ret, nil
}

func (lookupTableMarshaler) UnmarshalJSON(target *Layer, key string, value interface{}, node Node, interner Interner) error {
	table, err := unmarshalLookupTable(value)
	if err != nil {
		return err
	}
	if len(table.Ref) > 0 {
		if len(table.ID) > 0 || len(table.Elements) > 0 {
			return ErrInvalidLookupTable{ID: table.ID, Msg: "ref specified along with other attributes"}
		}
		referenceNode := NewNode(table.ID, LookupTableReferenceTerm)
		// Connect the attribute node to the reference node
		digraph.Connect(node, referenceNode, NewEdge(LookupTableTerm))
		return nil
	}
	// Non-reference table
	rootNode := NewNode(table.ID, LookupTableTerm)
	digraph.Connect(node, rootNode, NewEdge(LookupTableTerm))
	for index, element := range table.Elements {
		elementNode := NewNode("", LookupTableElementsTerm)
		elementNode.SetIndex(index)
		if len(element.Options) > 0 {
			elementNode.GetProperties()[LookupTableElementOptionsTerm] = StringSlicePropertyValue(element.Options)
		}
		if element.CaseSensitive {
			elementNode.GetProperties()[LookupTableElementCaseSensitiveTerm] = StringPropertyValue("true")
		}
		if len(element.Value) > 0 {
			elementNode.GetProperties()[LookupTableElementValueTerm] = StringPropertyValue(element.Value)
		}
		if len(element.Error) > 0 {
			elementNode.GetProperties()[LookupTableElementErrorTerm] = StringPropertyValue(element.Error)
		}
		digraph.Connect(rootNode, elementNode, NewEdge(LookupTableElementsTerm))
	}
	return nil
}

func (lookupTableMarshaler) MarshalLd(source *Layer, sourceNode Node, key string) (interface{}, error) {
	return nil, nil
}
