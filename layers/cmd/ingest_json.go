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

// import (
// 	"encoding/json"
// 	"fmt"
// 	"os"

// 	"github.com/piprate/json-gold/ld"
// 	"github.com/spf13/cobra"

// 	jsoningest "github.com/cloudprivacylabs/lsa/pkg/json"
// 	"github.com/cloudprivacylabs/lsa/pkg/ls"
// 	"github.com/cloudprivacylabs/lsa/pkg/rdf"
// 	"github.com/cloudprivacylabs/lsa/pkg/rdf/mrdf"
// 	"github.com/cloudprivacylabs/lsa/pkg/repo/fs"
// )

// func init() {
// 	ingestCmd.AddCommand(ingestJSONCmd)
// 	ingestJSONCmd.Flags().String("schema", "", "Schema id to use")
// 	ingestJSONCmd.MarkFlagRequired("schema")
// 	ingestJSONCmd.Flags().String("id", "http://example.org/data", "Base ID to use for ingested nodes")
// 	ingestJSONCmd.Flags().String("format", "json", "Output format, json, rdf, or dot")
// }

// var ingestJSONCmd = &cobra.Command{
// 	Use:   "json",
// 	Short: "Ingest a JSON document and enrich it with a schema",
// 	Args:  cobra.MaximumNArgs(1),
// 	Run: func(cmd *cobra.Command, args []string) {
// 		repoDir, _ := cmd.Flags().GetString("repo")
// 		if len(repoDir) == 0 {
// 			fail("Specify a repository directory using --repo")
// 		}
// 		repo := fs.New(repoDir, ls.Terms, func(fname string, err error) {
// 			fmt.Printf("%s: %s\n", fname, err)
// 		})
// 		if err := repo.Load(true); err != nil {
// 			failErr(err)
// 		}

// 		var input map[string]interface{}
// 		if err := readJSONFileOrStdin(args, &input); err != nil {
// 			failErr(err)
// 		}

// 		ID, _ := cmd.Flags().GetString("id")
// 		schemaId, _ := cmd.Flags().GetString("schema")
// 		compiler := ls.Compiler{Resolver: func(x string) (string, error) {
// 			if manifest := repo.GetSchemaManifestByObjectType(x); manifest != nil {
// 				return manifest.ID, nil
// 			}
// 			return x, nil
// 		},
// 			Loader: repo.LoadAndCompose,
// 		}
// 		resolved, err := compiler.Compile(schemaId)
// 		if err != nil {
// 			failErr(err)
// 		}

// 		format, _ := cmd.Flags().GetString("format")
// 		switch format {
// 		case "json":
// 			ingested, err := jsoningest.Ingest(ID, input, resolved)
// 			if err != nil {
// 				failErr(err)
// 			}
// 			out, _ := json.MarshalIndent(ls.DataModelToMap(ingested, true), "", "  ")
// 			fmt.Println(string(out))

// 		case "rdf", "dot":
// 			ingester := jsoningest.NewGraphIngester(resolved)
// 			output := mrdf.NewGraph()
// 			if err := ingester.Ingest(output, ID, input); err != nil {
// 				failErr(err)
// 			}
// 			if format == "rdf" {
// 				ds := rdf.ToRDFDataset(output.GetTriples())
// 				ser := ld.NQuadRDFSerializer{}
// 				v, _ := ser.Serialize(ds)
// 				fmt.Println(v)
// 			} else {
// 				nodes, edges := output.ToDOT()
// 				rdf.ToDOT("g", nodes, edges, os.Stdout)
// 			}
// 		}
// 	},
// }
