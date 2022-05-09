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
	"os"

	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
)

func init() {
	rootCmd.AddCommand(dotCmd)
	dotCmd.Flags().String("input", "json", "Input graph format (json, jsonld)")
	dotCmd.Flags().String("rankdir", "LR", "rankdir")
	dotCmd.Flags().String("output", "dot", "Output format (dot, json, jsonld, web)")
}

var dotCmd = &cobra.Command{
	Use:   "dot",
	Short: "Convert a graph to dot format",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputFormat, _ := cmd.Flags().GetString("input")
		g, err := cmdutil.ReadGraph(args, nil, inputFormat)
		if err != nil {
			failErr(err)
		}
		output, _ := cmd.Flags().GetString("output")
		cmdutil.WriteGraph(cmd, g, output, os.Stdout)
	},
}
