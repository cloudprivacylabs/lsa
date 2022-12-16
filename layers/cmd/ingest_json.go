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

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/layers/cmd/pipeline"
	jsoningest "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type JSONIngester struct {
	BaseIngestParams
	ID          string
	initialized bool
	parser      jsoningest.Parser
	ingester    *ls.Ingester
}

func (JSONIngester) Help() {
	fmt.Println(`Ingest JSON data
Ingest a JSON file using a schema variant and output a graph

operation: ingest/json
params:`)
	fmt.Println(baseIngestParamsHelp)
	fmt.Println(`  id:""   # Base ID for the root node`)
}

func (ji *JSONIngester) Run(pipeline *pipeline.PipelineContext) error {
	var layer *ls.Layer
	var err error
	if !ji.initialized {
		layer, err = LoadSchemaFromFileOrRepo(pipeline.Context, ji.CompiledSchema, ji.Repo, ji.Schema, ji.Type, ji.Bundle)
		if err != nil {
			return err
		}
		pipeline.Properties["layer"] = layer
		ji.parser = jsoningest.Parser{
			OnlySchemaAttributes: ji.OnlySchemaAttributes,
			IngestNullValues:     ji.IngestNullValues,
			Layer:                layer,
		}
		ji.initialized = true
		ji.ingester = &ls.Ingester{Schema: layer}
	}

	for {
		entryInfo, stream, err := pipeline.NextInput()
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
					pipeline.ErrorLogger(pipeline, fmt.Errorf("Error in file: %s, %v", entryInfo.GetName(), err))
					doneErr = fmt.Errorf("%v", err)
				}
			}()
			pipeline.SetGraph(cmdutil.NewDocumentGraph())
			builder := ls.NewGraphBuilder(pipeline.Graph, ls.GraphBuilderOptions{
				EmbedSchemaNodes:     ji.EmbedSchemaNodes,
				OnlySchemaAttributes: ji.OnlySchemaAttributes,
			})
			baseID := ji.ID

			_, err := jsoningest.IngestStream(pipeline.Context, baseID, stream, ji.parser, builder, ji.ingester)
			if err != nil {
				doneErr = err
				return
			}
			entities := ls.GetEntityInfo(pipeline.Graph)
			for e := range entities {
				if cmdutil.GetConfig().SourceProperty == "" {
					e.SetProperty("source", entryInfo.GetName())
				} else {
					e.SetProperty(cmdutil.GetConfig().SourceProperty, entryInfo.GetName())
				}
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
	ingestCmd.AddCommand(ingestJSONCmd)
	ingestJSONCmd.Flags().String("id", "http://example.org/root", "Base ID to use for ingested nodes")

	pipeline.RegisterPipelineStep("ingest/json", func() pipeline.Step {
		return &JSONIngester{
			BaseIngestParams: BaseIngestParams{
				EmbedSchemaNodes: true,
			},
		}
	})
}

var ingestJSONCmd = &cobra.Command{
	Use:   "json",
	Short: "Ingest a JSON document and enrich it with a schema",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		initialGraph, _ := cmd.Flags().GetString("initialGraph")
		ing := JSONIngester{}
		ing.fromCmd(cmd)
		ing.ID, _ = cmd.Flags().GetString("id")
		p := []pipeline.Step{
			&ing,
			NewWriteGraphStep(cmd),
		}
		env, err := godotenv.Read(".env")
		if err != nil {
			return err
		}
		_, err = runPipeline(p, env, initialGraph, args)
		return err
	},
}
