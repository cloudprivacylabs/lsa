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
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

type ValuesetProcessor struct {
	ValuesetFunc func(ValuesetLookupRequest) (ValuesetLookupResponse, error)
}

var (
	ValuesetSourceIDTerm      = NewTerm(LS, "vs/source", false, false, OverrideComposition, nil)
	ValuesetTablesTerm        = NewTerm(LS, "vs/valuesets", false, false, OverrideComposition, nil)
	ValuesetRequestKeysTerm   = NewTerm(LS, "vs/requestKeys", false, false, OverrideComposition, nil)
	ValuesetRequestValuesTerm = NewTerm(LS, "vs/requestValues", false, false, OverrideComposition, nil)
	ValuesetResultKeysTerm    = NewTerm(LS, "vs/resultKeys", false, false, OverrideComposition, nil)
	ValuesetResultValuesTerm  = NewTerm(LS, "vs/resultValues", false, false, OverrideComposition, nil)
)

type ValuesetInfo struct {
	// If the valueset lookup requires a single value, the attribute id
	// of the source node. Otherwise, the root node containing all the
	// required values.
	SourceID string

	// Optional lookup table IDs. If ommitted, all compatible tables are
	// looked up
	TableIDs []string

	// Ordered list of valueset keys. The request to the valueset
	// function will use these as the request keys
	RequestKeys []string

	// Ordered list of attribute ids containing valueset request
	// values. The elements of this match the keys array
	RequestValues []string

	// The keys of the valueset result
	ResultKeys []string

	// The attribute ids of the nodes under this node to receive values
	ResultValues []string
}

// ValueSetInfoFromNode parses the valuese information from a
// node. Returns nil if the node does not have valueset info
func ValuesetInfoFromNode(node graph.Node) *ValuesetInfo {
	v, ok := node.GetProperty(ValuesetSourceIDTerm)
	if !ok {
		return nil
	}
	pv, _ := v.(*PropertyValue)
	if pv == nil {
		return nil
	}
	ret := &ValuesetInfo{
		SourceID:      pv.AsString(),
		TableIDs:      AsPropertyValue(node.GetProperty(ValuesetTablesTerm)).MustStringSlice(),
		RequestKeys:   AsPropertyValue(node.GetProperty(ValuesetRequestKeysTerm)).MustStringSlice(),
		RequestValues: AsPropertyValue(node.GetProperty(ValuesetRequestValuesTerm)).MustStringSlice(),
		ResultKeys:    AsPropertyValue(node.GetProperty(ValuesetResultKeysTerm)).MustStringSlice(),
		ResultValues:  AsPropertyValue(node.GetProperty(ValuesetResultValuesTerm)).MustStringSlice(),
	}
	if len(ret.SourceID) == 0 {
		return nil
	}
	return ret
}

// Process attempts to process the valueset lookup for the schema
// node. If the node value is to be set using a valueset lookup, it
// performs the lookup and returns the result in valuesetResult. It is
// up to the caller to set the ingested values from ValuesetResult
func (prc ValuesetProcessor) Process(ctx *Context, g graph.Graph, schemaNode graph.Node) (*ValuesetResult, error) {
	vsi := ValuesetInfoFromNode(schemaNode)
	if vsi == nil {
		return nil, nil
	}
	// Node has valueset info
	ctx.GetLogger().Debug(map[string]interface{}{"mth": "valueset.process", "schemaNodeId": GetNodeID(schemaNode)})

}
