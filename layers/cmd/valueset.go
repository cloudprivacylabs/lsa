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

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

type ValueSets struct {
	ValueSets map[string]ls.LookupTable
}

// Lookup can be used as the external lookup func of LookupProcessor
func (v ValueSets) Lookup(lookupTableID string, dataNode graph.Node) (ls.LookupResult, error) {
	table, ok := v.ValueSets[lookupTableID]
	if !ok {
		return ls.LookupResult{}, ls.ErrNotFound(fmt.Sprintf("Lookup table %s", lookupTableID))
	}
	value, _ := ls.GetRawNodeValue(dataNode)
	results := table.FindValue(value)
	if len(results) == 0 {
		return ls.LookupResult{}, nil
	}
	if len(results) == 1 {
		return results[1], nil
	}
	if len(results) > 2 {
		return ls.LookupResult{}, fmt.Errorf("Lookup table error in %s: Multiple values match, value was: %s", lookupTableID, value)
	}
	if results[0].DefaultValue {
		return results[1], nil
	}
	if results[1].DefaultValue {
		return results[0], nil
	}
	return ls.LookupResult{}, fmt.Errorf("Lookup table error in %s: Multiple values match, value was: %s", lookupTableID, value)
}

type ValueSetMarshal struct {
	ValueSets []ls.LookupTable `json:"valuesets"`
	ls.LookupTable
}

func LoadValuesetFiles(valueSets *ValueSets, files []string) error {
	if valueSets.ValueSets == nil {
		valueSets.ValueSets = make(map[string]ls.LookupTable)
	}
	for _, file := range files {
		var vs ValueSetMarshal
		err := cmdutil.ReadJSON(file, &vs)
		if err != nil {
			return err
		}
		for _, v := range vs.ValueSets {
			if _, exists := valueSets.ValueSets[v.ID]; exists {
				return fmt.Errorf("Value set %s already defined", v.ID)
			}
			valueSets.ValueSets[v.ID] = v
		}
		if len(vs.ID) > 0 {
			if _, exists := valueSets.ValueSets[vs.ID]; exists {
				return fmt.Errorf("Value set %s already defined", vs.ID)
			}
			valueSets.ValueSets[vs.ID] = vs.LookupTable
		}
	}
	return nil
}

func loadValuesetsCmd(cmd *cobra.Command, valueSets *ValueSets) {
	vsf, _ := cmd.Flags().GetStringSlice("valueset")
	if len(vsf) > 0 {
		err := LoadValuesetFiles(valueSets, vsf)
		if err != nil {
			failErr(err)
		}
	}
}
