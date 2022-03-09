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
	"strings"

	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

type ErrInvalidLookupTable struct {
	ID  string
	Msg string
}

func (e ErrInvalidLookupTable) Error() string {
	return fmt.Sprintf("Invalid lookup table %s: %s", e.ID, e.Msg)
}

type ErrInputValueNotFoundInLookup struct {
	Msg string
}

func (e ErrInputValueNotFoundInLookup) Error() string {
	return "Input value not found in lookup in node: " + e.Msg
}

type ErrAmbiguousLookup struct {
	Msg string
}

func (e ErrAmbiguousLookup) Error() string {
	return "Ambiguous lookup value in node: " + e.Msg
}

type ErrLookupTableError struct {
	Errors []string
}

func (e ErrLookupTableError) Error() string {
	return "Lookup table error: " + strings.Join(e.Errors, ", ")
}

type LookupResult struct {
	// Matched is true if the lookup value matched something in the lookup table
	Matched bool
	// If true, the value did not match anything and the returned value
	// is the default value
	DefaultValue bool
	// This is the value returned if Matched==true
	Value string
	// This is nonempty if an error must be returned
	Error string
}

// LookupProcessor keeps the lookup configuration and provides the
// methods to evaulate lookup annotations
type LookupProcessor struct {
	Graph graph.Graph

	ExternalLookup func(lookupTableID string, dataNode graph.Node) (LookupResult, error)
}

// ProcessLookup will process the lookup annotations on the node. If the node has none, this will not do anything
func (prc *LookupProcessor) ProcessLookup(node graph.Node) error {
	processed, err := prc.processLookup(node, node)
	if processed {
		if err != nil {
			return err
		}
		return nil
	}
	for _, sch := range InstanceOf(node) {
		processed, err := prc.processLookup(node, sch)
		if err != nil {
			return err
		}
		if processed {
			return nil
		}
	}
	return nil
}

// processLookup traverses all the edges with LookupTableTerm, and
// processes all the nodes. If the node has a reference, then it tries
// to find the lookup table with that ID in the graph. If the node has
// LookupTableElementsTerm, then it tries to process the input using
// those element definitions.
func (prc *LookupProcessor) processLookup(dataNode, schemaNode graph.Node) (bool, error) {
	// Is there a lookup table defined in this node?
	// Lookup for a lookup table edge
	lookupTableNodes := graph.NextNodesWith(schemaNode, LookupTableTerm)
	if len(lookupTableNodes) == 0 {
		return false, nil
	}
	value, _ := GetRawNodeValue(dataNode)
	results := make([]LookupResult, 0)
	for _, ltNode := range lookupTableNodes {
		if ltNode.GetLabels().Has(LookupTableReferenceTerm) {
			// An external reference
			id := GetNodeID(ltNode)
			result, err := prc.ExternalLookup(id, dataNode)
			if err != nil {
				return true, err
			}
			if result.Matched {
				results = append(results, result)
			}
		} else {
			// An internal reference
			elements := graph.NextNodesWith(ltNode, LookupTableElementsTerm)
			for _, e := range elements {
				result, err := prc.processLookupNode(dataNode, e, value)
				if err != nil {
					return true, err
				}
				if result.Matched {
					results = append(results, result)
				}
			}
		}
	}
	if len(results) == 0 {
		return true, ErrInputValueNotFoundInLookup{Msg: fmt.Sprint(dataNode)}
	}
	setNewValue := func(newValue string) {
		SetRawNodeValue(dataNode, newValue)
		dataNode.SetProperty(RawInputValueTerm, StringPropertyValue(value))
	}
	if len(results) > 1 {
		// How many non-default and default values?
		nonDef := 0
		def := 0
		e := 0
		for _, x := range results {
			if len(x.Error) > 0 {
				e++
			} else {
				if x.DefaultValue {
					def++
				} else {
					nonDef++
				}
			}
		}

		if nonDef > 1 || (nonDef == 0 && def > 1) {
			return true, ErrAmbiguousLookup{Msg: fmt.Sprint(dataNode)}
		}
		if nonDef == 0 && def == 0 && e > 0 {
			errors := make([]string, 0)
			for _, x := range results {
				if len(x.Error) > 0 {
					errors = append(errors, x.Error)
				}
			}
			return true, ErrLookupTableError{Errors: errors}
		}
		if nonDef == 1 && def == 0 && e == 0 {
			for _, x := range results {
				if len(x.Error) == 0 && !x.DefaultValue {
					setNewValue(x.Value)
					break
				}
			}
		}
		if nonDef == 0 && def == 1 && e == 0 {
			for _, x := range results {
				if len(x.Error) == 0 && x.DefaultValue {
					setNewValue(x.Value)
					break
				}
			}
		}
		return true, ErrAmbiguousLookup{Msg: fmt.Sprint(dataNode)}
	}
	if len(results[0].Error) > 0 {
		return true, ErrLookupTableError{Errors: []string{results[0].Error}}
	}
	setNewValue(results[0].Value)
	return true, nil
}

// Input is the data node and the lookup element node. This determines
// if the node value matches lookup element node options
func (prc *LookupProcessor) processLookupNode(dataNode, lookupElementNode graph.Node, value string) (LookupResult, error) {
	options := AsPropertyValue(lookupElementNode.GetProperty(LookupTableElementOptionsTerm)).MustStringSlice()
	if len(options) == 0 {
		// this is the default option
		return LookupResult{
			Matched:      true,
			DefaultValue: true,
			Value:        AsPropertyValue(lookupElementNode.GetProperty(LookupTableElementValueTerm)).AsString(),
			Error:        AsPropertyValue(lookupElementNode.GetProperty(LookupTableElementErrorTerm)).AsString(),
		}, nil
	}
	filter := func(s string) string { return s }
	if AsPropertyValue(lookupElementNode.GetProperty(LookupTableElementCaseSensitiveTerm)).AsString() != "true" {
		filter = func(s string) string { return strings.ToLower(s) }
	}
	value = filter(value)
	for _, x := range options {
		if filter(x) == value {
			return LookupResult{
				Matched: true,
				Value:   AsPropertyValue(lookupElementNode.GetProperty(LookupTableElementValueTerm)).AsString(),
				Error:   AsPropertyValue(lookupElementNode.GetProperty(LookupTableElementErrorTerm)).AsString(),
			}, nil
		}
	}
	return LookupResult{Matched: false}, nil
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

var LookupTableTerm = NewTerm(LS, "lookupTable", false, false, OverrideComposition, struct {
	lookupTableMarshaler
}{
	lookupTableMarshaler{},
})

// RawInputValueTerm keeps the raw input value if the value is processed using a lookup table
var RawInputValueTerm = NewTerm(LS, "rawValue", false, false, NoComposition, nil)

var LookupTableElementsTerm = NewTerm(LS, "lookupTable/elements", false, true, NoComposition, nil)

// LookupTableReferenceTerm is used as a node type for nodes whose
// lookup table references are not resolved. These nodes should be
// resolved to real lookup tables during compilation
var LookupTableReferenceTerm = NewTerm(LS, "lookupTable/ref", true, false, NoComposition, nil)

var LookupTableElementOptionsTerm = NewTerm(LS, "lookupTable/element/options", false, false, NoComposition, nil)
var LookupTableElementCaseSensitiveTerm = NewTerm(LS, "lookupTable/element/caseSensitive", false, false, NoComposition, nil)
var LookupTableElementValueTerm = NewTerm(LS, "lookupTable/element/value", false, false, NoComposition, nil)
var LookupTableElementErrorTerm = NewTerm(LS, "lookupTable/element/error", false, false, NoComposition, nil)

type lookupTableMarshaler struct{}

func (lookupTableMarshaler) UnmarshalLd(target *Layer, key string, value interface{}, node *LDNode, allNodes map[string]*LDNode, interner Interner) error {
	return lookupTableMarshaler.unmarshalLookupTable(lookupTableMarshaler{}, target, key, value, node, allNodes)
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
//
// Graph representation:
//
//   (valueNode) --lookupTableTerm--> (lookupTableTerm) --lookupTableElementsTerm-->(lookupTableElementsTerm)
//                                                      --lookupTableElementsTerm-->(lookupTableElementsTerm)
//
// External Reference:
//
//   (valueNode) --lookupTableTerm-->(lookupTableReference, id: ref)
//
// Internal reference:
//
//   (valueNode) --lookupTableTerm-->(lookupTableTerm)
//
func (lookupTableMarshaler) unmarshalLookupTable(target *Layer, key string, value interface{}, node *LDNode, allNodes map[string]*LDNode) error {
	// key-value is:
	//   "https://lschema.org/lookupTable": [{
	//     "@id": "_:b1"
	//   }]
	id := LDGetNodeID(value)
	if len(id) == 0 {
		return ErrInvalidLookupTable{ID: node.ID, Msg: "Expecting to see @id under `lookupTable` term value"}
	}
	lookupContents := allNodes[id]
	if lookupContents == nil {
		// This is an external lookup table. Create a reference node for it
		referenceNode := target.Graph.NewNode([]string{LookupTableReferenceTerm}, nil)
		SetNodeID(referenceNode, id)
		// Put that into allNodes
		newNode := &LDNode{
			ID:        id,
			Types:     []string{LookupTableReferenceTerm},
			GraphNode: referenceNode,
		}
		allNodes[id] = newNode
		// Connect the attribute node to the reference node
		target.Graph.NewEdge(node.GraphNode, referenceNode, LookupTableTerm, nil)
		return nil
	}

	// The lookup table contents (or reference) is in this schema If we
	// are referring to an already processed lookup table, simply
	// connect this node to the table
	if lookupContents.GraphNode != nil && lookupContents.GraphNode.GetLabels().Has(LookupTableTerm) {
		target.Graph.NewEdge(node.GraphNode, lookupContents.GraphNode, LookupTableTerm, nil)
		return nil
	}

	// This is the first time we are working on this lookup table

	// Create the root node
	lookupTableNode := lookupContents.GraphNode
	if lookupTableNode == nil {
		lookupTableNode := target.Graph.NewNode(nil, nil)
		SetNodeID(lookupTableNode, id)
		lookupContents.GraphNode = lookupTableNode
	}
	types := lookupTableNode.GetLabels()
	types.Add(LookupTableTerm)
	lookupTableNode.SetLabels(types)
	target.Graph.NewEdge(node.GraphNode, lookupTableNode, LookupTableTerm, nil)

	// Create options
	for index, element := range LDGetListElements(lookupContents.Node[LookupTableElementsTerm]) {
		elementNodeID := LDGetNodeID(element)
		elementLDNode := allNodes[elementNodeID]
		if elementLDNode == nil {
			return ErrInvalidLookupTable{ID: node.ID, Msg: fmt.Sprintf("Cannot find element node with '%s'", elementNodeID)}
		}
		elementNode := target.Graph.NewNode([]string{LookupTableElementsTerm}, nil)
		SetNodeID(elementNode, elementNodeID)
		SetNodeIndex(elementNode, index)
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
					val = append(val, LDGetStringValue("@value", x))
				}
				elementNode.SetProperty(k, StringSlicePropertyValue(val))
			default:
				elementNode.SetProperty(k, StringPropertyValue(LDGetStringValue("@value", arr[0])))
			}
		}
		target.Graph.NewEdge(lookupTableNode, elementNode, LookupTableElementsTerm, nil)
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

func (lookupTableMarshaler) UnmarshalJSON(target *Layer, key string, value interface{}, node graph.Node, interner Interner) error {
	table, err := unmarshalLookupTable(value)
	if err != nil {
		return err
	}
	if len(table.Ref) > 0 {
		if len(table.ID) > 0 || len(table.Elements) > 0 {
			return ErrInvalidLookupTable{ID: table.ID, Msg: "ref specified along with other attributes"}
		}
		referenceNode := target.Graph.NewNode([]string{LookupTableReferenceTerm}, nil)
		SetNodeID(referenceNode, table.ID)
		// Connect the attribute node to the reference node
		target.Graph.NewEdge(node, referenceNode, LookupTableTerm, nil)
		return nil
	}
	// Non-reference table
	rootNode := target.Graph.NewNode([]string{LookupTableTerm}, nil)
	SetNodeID(rootNode, table.ID)
	target.Graph.NewEdge(node, rootNode, LookupTableTerm, nil)
	for index, element := range table.Elements {
		elementNode := target.Graph.NewNode([]string{LookupTableElementsTerm}, nil)
		SetNodeIndex(elementNode, index)
		if len(element.Options) > 0 {
			elementNode.SetProperty(LookupTableElementOptionsTerm, StringSlicePropertyValue(element.Options))
		}
		if element.CaseSensitive {
			elementNode.SetProperty(LookupTableElementCaseSensitiveTerm, StringPropertyValue("true"))
		}
		if len(element.Value) > 0 {
			elementNode.SetProperty(LookupTableElementValueTerm, StringPropertyValue(element.Value))
		}
		if len(element.Error) > 0 {
			elementNode.SetProperty(LookupTableElementErrorTerm, StringPropertyValue(element.Error))
		}
		target.Graph.NewEdge(rootNode, elementNode, LookupTableElementsTerm, nil)
	}
	return nil
}

func (lookupTableMarshaler) MarshalLd(source *Layer, sourceNode graph.Node, key string) (interface{}, error) {
	return nil, nil
}
