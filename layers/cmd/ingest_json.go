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
	"io"

	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/layers/cmd/pipeline"
	jsoningest "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type JSONIngester struct {
	BaseIngestParams
	ID          string
	initialized bool
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
		ji.initialized = true
	}

	for {
		rc, _ := pipeline.NextInput()
		buf := make([]byte, 1024)
		_, err := rc.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return fmt.Errorf("While streaming from %v: %w", rc, err)
			}
		}
		defer rc.Close()
		if err != nil {
			return err
		}
		if rc == nil {
			break
		}

		parser := jsoningest.Parser{
			OnlySchemaAttributes: ji.OnlySchemaAttributes,
		}
		if layer != nil {
			parser.SchemaNode = layer.GetSchemaRootNode()
		}
		pipeline.SetGraph(ls.NewDocumentGraph())
		builder := ls.NewGraphBuilder(pipeline.GetGraphRW(), ls.GraphBuilderOptions{
			EmbedSchemaNodes:     ji.EmbedSchemaNodes,
			OnlySchemaAttributes: ji.OnlySchemaAttributes,
		})
		baseID := ji.ID

		_, err = jsoningest.IngestStream(pipeline.Context, baseID, rc, parser, builder)
		if err != nil {
			return fmt.Errorf("While reading input %s: %w", "stdin", err)
		}

		if err := pipeline.Next(); err != nil {
			return fmt.Errorf("Input was %s: %w", "stdin", err)
		}
	}
	return nil
}

func init() {
	ingestCmd.AddCommand(ingestJSONCmd)
	ingestJSONCmd.Flags().String("id", "http://example.org/root", "Base ID to use for ingested nodes")

	pipeline.Operations["ingest/json"] = func() pipeline.Step {
		return &JSONIngester{
			BaseIngestParams: BaseIngestParams{
				EmbedSchemaNodes: true,
			},
		}
	}
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
		_, err := runPipeline(p, initialGraph, args)
		return err
	},
}
