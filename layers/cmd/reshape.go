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
	rootCmd.AddCommand(reshapeCmd)
	reshapeCmd.Flags().String("schema", "", "If repo is given, the schema id. Otherwise schema file.")
	reshapeCmd.Flags().String("repo", "", "Schema repository directory")
	reshapeCmd.Flags().String("type", "", "Use if a bundle is given for data types. The type name to ingest.")
	reshapeCmd.Flags().String("bundle", "", "Schema bundle.")
	reshapeCmd.Flags().String("compiledschema", "", "Use the given compiled schema")
	reshapeCmd.Flags().String("input", "json", "Input graph format (json, jsonld)")
	reshapeCmd.PersistentFlags().String("output", "json", "Output format, json, jsonld, or dot")
}

var reshapeCmd = &cobra.Command{
	Use:   "reshape",
	Short: "Reshape a graph using a target schema",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := getContext()
		input, _ := cmd.Flags().GetString("input")
		g, err := cmdutil.ReadGraph(args, ctx.GetInterner(), input)
		if err != nil {
			failErr(err)
		}
		layer := loadSchemaCmd(ctx, cmd)

		reshaper := transform.Reshaper{}
		reshaper.TargetSchema = layer
		reshaper.Builder = ls.NewGraphBuilder(nil, ls.GraphBuilderOptions{
			EmbedSchemaNodes: true,
		})
		err = reshaper.Reshape(ctx, g)
		if err != nil {
			failErr(err)
		}
		outFormat, _ := cmd.Flags().GetString("output")
		err = OutputIngestedGraph(cmd, outFormat, reshaper.Builder.GetGraph(), os.Stdout, false)
		if err != nil {
			failErr(err)
		}
	},
}
