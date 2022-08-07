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
	"strings"

	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
)

func init() {
	rootCmd.AddCommand(listPropsCmd)
	listPropsCmd.Flags().String("input", "json", "Input graph format (json, jsonld)")
	listPropsCmd.Flags().StringSlice("property", []string{"https://lschema.org/nodeId"}, "Properties to output")
}

var listPropsCmd = &cobra.Command{
	Use:   "lp",
	Short: "List properties",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputFormat, _ := cmd.Flags().GetString("input")
		g, err := cmdutil.ReadGraph(args, nil, inputFormat)
		if err != nil {
			failErr(err)
		}
		props, _ := cmd.Flags().GetStringSlice("property")
		for nodes := g.GetNodes(); nodes.Next(); {
			node := nodes.Node()
			strs := make([]string, 0, len(props))
			for _, p := range props {
				v, o := node.GetProperty(p)
				if o {
					strs = append(strs, fmt.Sprint(v))
				}
			}
			if len(strs) > 0 {
				fmt.Println(strings.Join(strs, ","))
			}
		}
	},
}
