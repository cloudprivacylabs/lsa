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
	"os"

	"github.com/bserdar/digraph"
	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/pkg/dot"
	jsoningest "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/jsonld"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/repo/fs"
)

func init() {
	ingestCmd.AddCommand(ingestJSONCmd)
	ingestJSONCmd.Flags().String("schema", "", "If repo is given, the schema id. Otherwise schema file.")
	ingestJSONCmd.Flags().String("id", "http://example.org/root", "Base ID to use for ingested nodes")
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
		if len(repoDir) > 0 {
			var err error
			repo, err = getRepo(repoDir)
			if err != nil {
				failErr(err)
			}
		}
		schemaName, _ := cmd.Flags().GetString("schema")
		if len(schemaName) > 0 {
			if repo != nil {
				var err error
				layer, err = repo.GetComposedSchema(schemaName)
				if err != nil {
					failErr(err)
				}
				compiler := ls.Compiler{Resolver: func(x string) (string, error) {
					if manifest := repo.GetSchemaManifestByObjectType(x); manifest != nil {
						return manifest.ID, nil
					}
					return x, nil
				},
					Loader: func(x string) (*ls.Layer, error) {
						return repo.LoadAndCompose(x)
					},
				}
				layer, err = compiler.Compile(schemaName)
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
				compiler := ls.Compiler{Resolver: func(x string) (string, error) {
					if x == schemaName {
						return x, nil
					}
					if x == layer.GetID() {
						return x, nil
					}
					return "", fmt.Errorf("Not found")
				},
					Loader: func(x string) (*ls.Layer, error) {
						if x == schemaName || x == layer.GetID() {
							return layer, nil
						}
						return nil, fmt.Errorf("Not found")
					},
				}
				layer, err = compiler.Compile(schemaName)
				if err != nil {
					failErr(err)
				}
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
		ingester := jsoningest.Ingester{
			Schema:  layer,
			KeyTerm: ls.AttributeNameTerm,
		}

		baseID, _ := cmd.Flags().GetString("id")
		target := digraph.New()
		_, err := ingester.Ingest(target, baseID, input)
		if err != nil {
			failErr(err)
		}
		renderer := dot.Renderer{Options: dot.DefaultOptions()}
		renderer.Render(target, "g", os.Stdout)
	},
}
