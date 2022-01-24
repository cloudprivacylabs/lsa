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

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	xmlingest "github.com/cloudprivacylabs/lsa/pkg/xml"
)

func init() {
	ingestCmd.AddCommand(ingestXMLCmd)
	ingestXMLCmd.Flags().String("schema", "", "If repo is given, the schema id. Otherwise schema file.")
	ingestXMLCmd.Flags().String("id", "http://example.org/root", "Base ID to use for ingested nodes")
	ingestXMLCmd.Flags().String("compiledschema", "", "Use the given compiled schema")
}

var ingestXMLCmd = &cobra.Command{
	Use:   "xml",
	Short: "Ingest an XML document and enrich it with a schema",
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
			input, err = cmdutil.StreamFileOrStdin(args, enc)
			if err != nil {
				failErr(err)
			}
		} else {
			input, err = cmdutil.StreamFileOrStdin(args)
			if err != nil {
				failErr(err)
			}
		}
		embedSchemaNodes, _ := cmd.Flags().GetBool("embedSchemaNodes")
		onlySchemaAttributes, _ := cmd.Flags().GetBool("onlySchemaAttributes")
		ingester := xmlingest.Ingester{
			Ingester: ls.Ingester{
				Schema:               layer,
				EmbedSchemaNodes:     embedSchemaNodes,
				OnlySchemaAttributes: onlySchemaAttributes,
			},
		}

		baseID, _ := cmd.Flags().GetString("id")
		root, err := xmlingest.IngestStream(&ingester, baseID, input)
		if err != nil {
			failErr(err)
		}
		target := digraph.New()
		target.AddNode(root)
		outFormat, _ := cmd.Flags().GetString("output")
		includeSchema, _ := cmd.Flags().GetBool("includeSchema")
		err = OutputIngestedGraph(outFormat, target, os.Stdout, includeSchema)
		if err != nil {
			failErr(err)
		}
	},
}