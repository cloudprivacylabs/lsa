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
	"encoding/json"
	"fmt"

	"github.com/piprate/json-gold/ld"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(compactCmd)
	compactCmd.Flags().StringSlice("context", nil, "Use the given context files")
}

var compactCmd = &cobra.Command{
	Use:   "compact",
	Short: "Compact a json-ld document",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var input interface{}
		if err := readJSONFileOrStdin(args, &input); err != nil {
			failErr(err)
		}
		contexts, _ := cmd.Flags().GetStringSlice("context")
		output, err := compact(input, contexts)
		if err != nil {
			failErr(err)
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
	},
}

func compact(base interface{}, contexts []string) (interface{}, error) {
	processor := ld.NewJsonLdProcessor()
	localContext := map[string]interface{}{}
	for _, c := range contexts {
		var m map[string]interface{}
		if err := readJSON(c, &m); err != nil {
			return nil, err
		}
		ctx := m["@context"]
		if mp, ok := ctx.(map[string]interface{}); ok {
			for k, v := range mp {
				localContext[k] = v
			}
		}
	}
	options := ld.NewJsonLdOptions("")
	options.CompactArrays = true
	output, err := processor.Compact(base, map[string]interface{}{"@context": localContext}, options)
	if err != nil {
		return nil, err
	}
	if contexts != nil {
		output["@context"] = contexts
	}
	return output, nil
}
