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
	jsoningest "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
)

type JSONIngester struct {
	BaseIngestParams
	ID string
}

func (ji JSONIngester) Run(pipeline *PipelineContext) error {
	ctx := pipeline.Context
	layer, err := LoadSchemaFromFileOrRepo(pipeline.Context, ji.CompiledSchema, ji.Repo, ji.Schema, ji.Type, ji.Bundle)
	if err != nil {
		return err
	}
	var input io.Reader
	if layer != nil {
		enc, err := layer.GetEncoding()
		if err != nil {
			failErr(err)
		}
		input, err = cmdutil.StreamFileOrStdin(pipeline.InputFiles, enc)
		if err != nil {
			failErr(err)
		}
	} else {
		input, err = cmdutil.StreamFileOrStdin(pipeline.InputFiles)
		if err != nil {
			failErr(err)
		}
	}
	grph := pipeline.Graph

	parser := jsoningest.Parser{}

	parser.OnlySchemaAttributes = ji.OnlySchemaAttributes
	parser.SchemaNode = layer.GetSchemaRootNode()
	embedSchemaNodes := ji.EmbedSchemaNodes

	builder := ls.NewGraphBuilder(grph, ls.GraphBuilderOptions{
		EmbedSchemaNodes:     embedSchemaNodes,
		OnlySchemaAttributes: parser.OnlySchemaAttributes,
	})
	baseID := ji.ID
	_, err = jsoningest.IngestStream(ctx, baseID, input, parser, builder)
	if err != nil {
		failErr(err)
	}

	pipeline.Graph = grph
	pipeline.Next()
	return nil
}

func init() {
	ingestCmd.AddCommand(ingestJSONCmd)
	ingestJSONCmd.Flags().String("id", "http://example.org/root", "Base ID to use for ingested nodes")
}

var ingestJSONCmd = &cobra.Command{
	Use:   "json",
	Short: "Ingest a JSON document and enrich it with a schema",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		initialGraph, _ := cmd.Flags().GetString("initialGraph")
		ctx := getContext()
		layer := loadSchemaCmd(ctx, cmd)
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

		parser := jsoningest.Parser{}

		parser.OnlySchemaAttributes, _ = cmd.Flags().GetBool("onlySchemaAttributes")
		parser.SchemaNode = layer.GetSchemaRootNode()
		embedSchemaNodes, _ := cmd.Flags().GetBool("embedSchemaNodes")

		builder := ls.NewGraphBuilder(grph, ls.GraphBuilderOptions{
			EmbedSchemaNodes:     embedSchemaNodes,
			OnlySchemaAttributes: parser.OnlySchemaAttributes,
		})
		baseID, _ := cmd.Flags().GetString("id")
		_, err = jsoningest.IngestStream(ctx, baseID, input, parser, builder)

		if err != nil {
			failErr(err)
		}
		outFormat, _ := cmd.Flags().GetString("output")
		includeSchema, _ := cmd.Flags().GetBool("includeSchema")
		err = OutputIngestedGraph(cmd, outFormat, grph, os.Stdout, includeSchema)
		if err != nil {
			failErr(err)
		}
	},
}
