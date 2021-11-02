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
	"io"
	"os"

	"github.com/bserdar/digraph"
	"github.com/spf13/cobra"

	jsoningest "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func init() {
	ingestCmd.AddCommand(ingestJSONCmd)
	ingestJSONCmd.Flags().String("schema", "", "If repo is given, the schema id. Otherwise schema file.")
	ingestJSONCmd.Flags().String("id", "http://example.org/root", "Base ID to use for ingested nodes")
	ingestJSONCmd.Flags().String("compiledschema", "", "Use the given compiled schema")
	ingestJSONCmd.Flags().String("output", "graph", "Output json, or graph")
}

var ingestJSONCmd = &cobra.Command{
	Use:   "json",
	Short: "Ingest a JSON document and enrich it with a schema",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		interner := ls.NewInterner()
		compiledSchema, _ := cmd.Flags().GetString("compiledschema")
		repoDir, _ := cmd.Flags().GetString("repo")
		schemaName, _ := cmd.Flags().GetString("schema")
		layer, err := LoadSchemaFromFileOrRepo(compiledSchema, repoDir, schemaName, interner)
		if err != nil {
			failErr(err)
		}
		var input io.Reader
		if layer != nil {
			enc, err := layer.GetEncoding()
			if err != nil {
				failErr(err)
			}
			input, err = streamJSONFileOrStdin(args, enc)
			if err != nil {
				failErr(err)
			}
		} else {
			input, err = streamJSONFileOrStdin(args)
			if err != nil {
				failErr(err)
			}
		}
		ingester := jsoningest.Ingester{
			Ingester: ls.Ingester{
				Schema: layer,
			},
		}

		baseID, _ := cmd.Flags().GetString("id")
		root, err := jsoningest.IngestStream(&ingester, baseID, input)
		if err != nil {
			failErr(err)
		}
		target := digraph.New()
		target.AddNode(root)
		outFormat, _ := cmd.Flags().GetString("format")
		includeSchema, _ := cmd.Flags().GetBool("includeSchema")
		output, _ := cmd.Flags().GetString("output")
		if output == "graph" {
			err = OutputIngestedGraph(outFormat, target, os.Stdout, includeSchema)
			if err != nil {
				failErr(err)
			}
			return
		}
		if output == "json" {
			exportOptions := jsoningest.ExportOptions{
				BuildNodeKeyFunc: jsoningest.GetBuildNodeKeyBySchemaNodeFunc(func(schemaNode, docNode ls.Node) (string, bool, error) {
					return schemaNode.GetID(), true, nil
				}),
			}
			data, err := jsoningest.Export(root, exportOptions)
			if err != nil {
				failErr(err)
			}
			data.Encode(os.Stdout)
		}
	},
}
