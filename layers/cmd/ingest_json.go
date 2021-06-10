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
	"os"

	"github.com/spf13/cobra"

	jsoningest "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/repo/fs"
)

func init() {
	ingestCmd.AddCommand(ingestJSONCmd)
	ingestJSONCmd.Flags().String("schema", "", "If repo is given, the schema id. Otherwise schema file.")
	ingestJSONCmd.Flags().String("id", "http://example.org/data", "Base ID to use for ingested nodes")
	ingestJSONCmd.Flags().String("format", "json", "Output format, json, rdf, or dot")
}

var ingestJSONCmd = &cobra.Command{
	Use:   "json",
	Short: "Ingest a JSON document and enrich it with a schema",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var layer *ls.Layer
		repoDir, _ := cmd.Flags().GetString("repo")
		var repo *fs.Repository
		if len(repoDir) != 0 {
			repo = fs.New(repoDir, func(msg string, err error) {
				fmt.Println("%s: %v", msg, err)
			})
			if err := repo.Load(true); err != nil {
				failErr(err)
			}
		}
		schemaName, _ := cmd.Flags().GetString("schema")
		if len(schemaName) > 0 {
			if repo != nil {
				layer, err := repo.GetComposedSchema(schemaName)
				if err != nil {
					failErr(err)
				}
			} else {
				var v interface{}
				err := readJSON(args[0], &v)
				if err != nil {
					failErr(err)
				}
				layer, err = jsonld.UnmarshalLayer(v)
				if err != nil {
					failErr(err)
				}
			}

			compiler := ls.Compiler{Resolver: func(x string) (string, error) {
				if manifest := repo.GetSchemaManifestByObjectType(x); manifest != nil {
					return manifest.ID, nil
				}
				return x, nil
			},
				Loader: repo.LoadAndCompose,
			}
			resolved, err := compiler.Compile(schemaId)
			if err != nil {
				failErr(err)
			}

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
		ingester := json.Ingester{
			Schema:  layer,
			KeyTerm: x,
		}

		target := digraph.New()
		root, err := ingester.Ingest(target, input)

	},
}
