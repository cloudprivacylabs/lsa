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

	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/layers/cmd/pipeline"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	xmlingest "github.com/cloudprivacylabs/lsa/pkg/xml"
)

type XMLIngester struct {
	BaseIngestParams
	ID          string
	initialized bool
	parser      xmlingest.Parser
}

func (XMLIngester) Help() {
	fmt.Println(`Ingest XML data
Ingest an XML file based on a schema variant and output a graph

operation: ingest/xml
params:`)
	fmt.Println(baseIngestParamsHelp)
	fmt.Println(`  id:""   # Base ID for the root node`)
}

func (xml *XMLIngester) Run(pipeline *pipeline.PipelineContext) error {
	var layer *ls.Layer
	var err error
	if !xml.initialized {
		layer, err = LoadSchemaFromFileOrRepo(pipeline.Context, xml.CompiledSchema, xml.Repo, xml.Schema, xml.Type, xml.Bundle)
		if err != nil {
			return err
		}
		pipeline.Properties["layer"] = layer
		xml.parser = xmlingest.Parser{
			OnlySchemaAttributes: xml.OnlySchemaAttributes,
			IngestEmptyValues:    xml.IngestNullValues,
			Layer:                layer,
		}
		xml.initialized = true
	}

	for {
		_, stream, err := pipeline.NextInput()
		if err != nil {
			return err
		}
		if stream == nil {
			break
		}
		var doneErr error
		func() {
			defer func() {
				if err := recover(); err != nil {
					if !pipeline.ErrorLogger(pipeline, fmt.Errorf("Error in file: %s, %v", xml.Schema, err)) {
						doneErr = fmt.Errorf("%v", err)
					}
				}
			}()
			pipeline.SetGraph(cmdutil.NewDocumentGraph())
			builder := ls.NewGraphBuilder(pipeline.Graph, ls.GraphBuilderOptions{
				EmbedSchemaNodes:     xml.EmbedSchemaNodes,
				OnlySchemaAttributes: xml.OnlySchemaAttributes,
			})

			baseID := xml.ID

			parsed, err := xml.parser.ParseStream(pipeline.Context, baseID, stream)
			if err != nil {
				doneErr = err
				return
			}
			_, err = ls.Ingest(builder, parsed)
			if err != nil {
				doneErr = err
				return
			}
			if err := builder.LinkNodes(pipeline.Context, xml.parser.Layer, ls.GetEntityInfo(builder.GetGraph())); err != nil {
				doneErr = err
				return
			}
			if err := pipeline.Next(); err != nil {
				doneErr = err
				return
			}
		}()
		if doneErr != nil {
			return doneErr
		}
	}
	return nil
}

func init() {
	ingestCmd.AddCommand(ingestXMLCmd)
	ingestXMLCmd.Flags().String("id", "http://example.org/root", "Base ID to use for ingested nodes")

	pipeline.RegisterPipelineStep("ingest/xml", func() pipeline.Step {
		return &XMLIngester{
			BaseIngestParams: BaseIngestParams{
				EmbedSchemaNodes: true,
			},
		}
	})
}

var ingestXMLCmd = &cobra.Command{
	Use:   "xml",
	Short: "Ingest an XML document and enrich it with a schema",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		initialGraph, _ := cmd.Flags().GetString("initialGraph")
		ing := XMLIngester{}
		ing.fromCmd(cmd)
		ing.ID, _ = cmd.Flags().GetString("id")
		p := []pipeline.Step{
			&ing,
			NewWriteGraphStep(cmd),
		}
		_, err := runPipeline(p, initialGraph, args)
		return err
	},
}
