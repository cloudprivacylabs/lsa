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

	jsoningest "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func init() {
	ingestCmd.AddCommand(ingestJSONCmd)
	ingestJSONCmd.Flags().String("schema", "", "If repo is given, the schema id. Otherwise schema file.")
	ingestJSONCmd.Flags().String("id", "http://example.org/root", "Base ID to use for ingested nodes")
	ingestJSONCmd.Flags().String("format", "json", "Output format, json(ld), rdf, or dot")
	ingestJSONCmd.Flags().String("compiledschema", "", "Use the given compiled schema")
}

var ingestJSONCmd = &cobra.Command{
	Use:   "json",
	Short: "Ingest a JSON document and enrich it with a schema",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		compiledSchema, _ := cmd.Flags().GetString("compiledschema")
		repoDir, _ := cmd.Flags().GetString("repo")
		schemaName, _ := cmd.Flags().GetString("schema")
		layer, err := LoadSchemaFromFileOrRepo(compiledSchema, repoDir, schemaName)
		if err != nil {
			failErr(err)
		}
		var input map[string]interface{}
		if layer != nil {
			enc, err := layer.GetEncoding()
			if err != nil {
				failErr(err)
			}
			if err := readJSONFileOrStdin(args, &input, enc); err != nil {
				failErr(err)
			}
		}
		if err := readJSONFileOrStdin(args, &input); err != nil {
			failErr(err)
		}
		ingester := jsoningest.Ingester{
			Schema:  layer,
			KeyTerm: ls.AttributeNameTerm,
		}

		baseID, _ := cmd.Flags().GetString("id")
		target := digraph.New()
		_, err = ingester.Ingest(target, baseID, input)
		if err != nil {
			failErr(err)
		}
		outFormat, _ := cmd.Flags().GetString("format")
		err = OutputIngestedGraph(outFormat, target, os.Stdout)
		if err != nil {
			failErr(err)
		}
	},
}
