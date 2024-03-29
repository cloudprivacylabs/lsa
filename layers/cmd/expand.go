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

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
)

func init() {
	rootCmd.AddCommand(expandCmd)
}

var expandCmd = &cobra.Command{
	Use:   "expand",
	Short: "Expand a json-ld document",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		processor := ld.NewJsonLdProcessor()
		var input interface{}
		if err := cmdutil.ReadJSONFileOrStdin(args, &input); err != nil {
			failErr(err)
		}
		output, err := processor.Expand(input, nil)
		if err != nil {
			failErr(err)
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
	},
}
