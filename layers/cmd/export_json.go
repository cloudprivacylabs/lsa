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

	"github.com/bserdar/digraph"
	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	jsoningest "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func init() {
	exportCmd.AddCommand(exportJSONCmd)
	exportCmd.Flags().String("input", "json", "Input graph format (json, jsonld)")
}

var exportJSONCmd = &cobra.Command{
	Use:   "json",
	Short: "Export a graph as a JSON document",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input, _ := cmd.Flags().GetString("input")
		graph, err := cmdutil.ReadGraph(args, nil, input)
		if err != nil {
			failErr(err)
		}
		ix := graph.GetIndex()
		for _, node := range digraph.Sources(ix) {
			exportOptions := jsoningest.ExportOptions{}
			data, err := jsoningest.Export(node.(ls.Node), exportOptions)
			if err != nil {
				failErr(err)
			}
			data.Encode(os.Stdout)
		}
	},
}