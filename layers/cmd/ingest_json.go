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

	"github.com/spf13/cobra"

	jsoningest "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/repo/fs"
)

func init() {
	ingestCmd.AddCommand(ingestJSONCmd)
	ingestJSONCmd.Flags().String("schema", "", "Schema id to use")
	ingestJSONCmd.MarkFlagRequired("schema")
}

var ingestJSONCmd = &cobra.Command{
	Use:   "json",
	Short: "Ingest a JSON document and enrich it with a schema",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoDir, _ := cmd.Flags().GetString("repo")
		if len(repoDir) == 0 {
			fail("Specify a repository directory using --repo")
		}
		repo := fs.New(repoDir, ls.Terms, func(fname string, err error) {
			fmt.Printf("%s: %s\n", fname, err)
		})
		if err := repo.Load(true); err != nil {
			failErr(err)
		}

		var input map[string]interface{}
		if err := readJSONFileOrStdin(args, &input); err != nil {
			failErr(err)
		}

		schemaId, _ := cmd.Flags().GetString("schema")
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

		ingested, err := jsoningest.Ingest(resolved.ObjectType, input, resolved)
		if err != nil {
			failErr(err)
		}
		out, _ := json.MarshalIndent(ls.DataModelToMap(ingested, true), "", "  ")
		fmt.Println(string(out))
	},
}
