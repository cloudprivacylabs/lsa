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

	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
	xmlingest "github.com/cloudprivacylabs/lsa/pkg/xml"
)

func init() {
	ingestCmd.AddCommand(ingestXMLCmd)
	ingestXMLCmd.Flags().String("id", "http://example.org/root", "Base ID to use for ingested nodes")
}

var ingestXMLCmd = &cobra.Command{
	Use:   "xml",
	Short: "Ingest an XML document and enrich it with a schema",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		initialGraph, _ := cmd.Flags().GetString("initialGraph")
		ctx := getContext()
		layer := loadSchemaCmd(ctx, cmd)
		valuesets := &Valuesets{}
		loadValuesetsCmd(cmd, valuesets)
		var input io.Reader
		var err error
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
		var grph graph.Graph
		if layer != nil && initialGraph != "" {
			grph, err = cmdutil.ReadJSONGraph([]string{initialGraph}, nil)
			if err != nil {
				failErr(err)
			}
		} else {
			grph = ls.NewDocumentGraph()
		}
		embedSchemaNodes, _ := cmd.Flags().GetBool("embedSchemaNodes")
		onlySchemaAttributes, _ := cmd.Flags().GetBool("onlySchemaAttributes")
		ingester := xmlingest.Ingester{
			Ingester: ls.Ingester{
				Schema:               layer,
				EmbedSchemaNodes:     embedSchemaNodes,
				OnlySchemaAttributes: onlySchemaAttributes,
				ValuesetFunc:         valuesets.Lookup,
				Graph:                grph,
			},
		}

		baseID, _ := cmd.Flags().GetString("id")
		_, err = xmlingest.IngestStream(ctx, &ingester, baseID, input)
		if err != nil {
			failErr(err)
		}
		outFormat, _ := cmd.Flags().GetString("output")
		includeSchema, _ := cmd.Flags().GetBool("includeSchema")
		err = OutputIngestedGraph(cmd, outFormat, ingester.Graph, os.Stdout, includeSchema)
		if err != nil {
			failErr(err)
		}
	},
}
