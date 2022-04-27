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
	xmlingest "github.com/cloudprivacylabs/lsa/pkg/xml"
	"github.com/cloudprivacylabs/opencypher/graph"
)

type XMLIngester struct {
	BaseIngestParams
	ID string
}

func (XMLIngester) Next() error { return nil }

func (xml XMLIngester) Run(pipeline *PipelineContext) error {
	layer, err := LoadSchemaFromFileOrRepo(pipeline.Context, xml.CompiledSchema, xml.Repo, xml.Schema, xml.Type, xml.Bundle)
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

	embedSchemaNodes := xml.EmbedSchemaNodes
	onlySchemaAttributes := xml.OnlySchemaAttributes
	parser := xmlingest.Parser{
		OnlySchemaAttributes: onlySchemaAttributes,
	}
	if layer != nil {
		parser.SchemaNode = layer.GetSchemaRootNode()
	}
	builder := ls.NewGraphBuilder(grph, ls.GraphBuilderOptions{
		EmbedSchemaNodes:     embedSchemaNodes,
		OnlySchemaAttributes: onlySchemaAttributes,
	})

	baseID := xml.ID

	parsed, err := parser.ParseStream(pipeline.Context, baseID, input)
	if err != nil {
		failErr(err)
	}
	_, err = ls.Ingest(builder, parsed)
	if err != nil {
		failErr(err)
	}
	if err := pipeline.Next(); err != nil {
		return err
	}
	return nil
}

func init() {
	ingestCmd.AddCommand(ingestXMLCmd)
	ingestXMLCmd.Flags().String("id", "http://example.org/root", "Base ID to use for ingested nodes")

	operations["xmlingest"] = func() Step { return &XMLIngester{} }
}

var ingestXMLCmd = &cobra.Command{
	Use:   "xml",
	Short: "Ingest an XML document and enrich it with a schema",
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
		embedSchemaNodes, _ := cmd.Flags().GetBool("embedSchemaNodes")
		onlySchemaAttributes, _ := cmd.Flags().GetBool("onlySchemaAttributes")
		parser := xmlingest.Parser{
			OnlySchemaAttributes: onlySchemaAttributes,
		}
		if layer != nil {
			parser.SchemaNode = layer.GetSchemaRootNode()
		}
		builder := ls.NewGraphBuilder(grph, ls.GraphBuilderOptions{
			EmbedSchemaNodes:     embedSchemaNodes,
			OnlySchemaAttributes: onlySchemaAttributes,
		})

		baseID, _ := cmd.Flags().GetString("id")

		parsed, err := parser.ParseStream(ctx, baseID, input)
		if err != nil {
			failErr(err)
		}
		_, err = ls.Ingest(builder, parsed)
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
