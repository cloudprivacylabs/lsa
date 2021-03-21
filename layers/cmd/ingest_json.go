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
		repo := SchemaRepository{}
		repoDir, _ := cmd.Flags().GetString("repo")
		if len(repoDir) == 0 {
			fail("Specify a repository directory using --repo")
		}
		if err := repo.LoadDir(repoDir); err != nil {
			failErr(err)
		}

		var input map[string]interface{}
		if err := readJSONFileOrStdin(args, &input); err != nil {
			failErr(err)
		}

		schemaId, _ := cmd.Flags().GetString("schema")
		schema := repo.GetSchemaByID(schemaId)
		if schema == nil {
			fail(fmt.Sprintf("Schema not found: %s", schemaId))
		}
		resolved, err := repo.ResolveSchemaForID(schemaId)
		if err != nil {
			failErr(err)
		}

		ingested, err := jsoningest.Ingest(schema.ObjectType, input, resolved)
		if err != nil {
			failErr(err)
		}
		out, _ := json.MarshalIndent(ingested.ToMap(), "", "  ")
		fmt.Println(string(out))
	},
}
