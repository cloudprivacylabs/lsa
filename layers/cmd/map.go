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

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/transform"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(mapCmd)
	mapCmd.Flags().String("schema", "", "If repo is given, the schema id. Otherwise schema file.")
	mapCmd.Flags().String("repo", "", "Schema repository directory")
	mapCmd.Flags().String("type", "", "Use if a bundle is given for data types. The type name to ingest.")
	mapCmd.Flags().String("bundle", "", "Schema bundle.")
	mapCmd.Flags().String("compiledschema", "", "Use the given compiled schema")
	mapCmd.Flags().StringSlice("valueset", nil, "Value set file (s)")
	mapCmd.Flags().String("input", "json", "Input graph format (json, jsonld)")
	mapCmd.Flags().String("output", "json", "Output format, json, jsonld, or dot")
	mapCmd.Flags().String("term", "", "The term in the input graph that contains the target schema attribute ids")
}

var mapCmd = &cobra.Command{
	Use:   "map",
	Short: "Map a graph to fit to a target schema by copying the values from the source nodes that has a term property containing target schema node ids",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := getContext()
		input, _ := cmd.Flags().GetString("input")
		g, err := cmdutil.ReadGraph(args, ctx.GetInterner(), input)
		if err != nil {
			failErr(err)
		}
		layer := loadSchemaCmd(ctx, cmd)
		valuesets := &Valuesets{}
		loadValuesetsCmd(cmd, valuesets)

		mapper := transform.Mapper{}
		mapper.Schema = layer
		mapper.EmbedSchemaNodes = true
		mapper.Graph = ls.NewDocumentGraph()
		mapper.PropertyName, _ = cmd.Flags().GetString("term")
		mapper.ValuesetFunc = valuesets.Lookup
		err = mapper.Map(ctx, g)
		if err != nil {
			failErr(err)
		}
		outFormat, _ := cmd.Flags().GetString("output")
		err = OutputIngestedGraph(outFormat, mapper.Graph, os.Stdout, false)
		if err != nil {
			failErr(err)
		}
	},
}
