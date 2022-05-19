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
	"bytes"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"golang.org/x/text/encoding"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
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

func (ji *JSONIngester) Run(pipeline *PipelineContext) error {
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

	enc := encoding.Nop
	if layer != nil {
		enc, err = layer.GetEncoding()
		if err != nil {
			return err
		}
	}

	inputIndex := 0
	var inputName string
	nextInput := func() (io.Reader, error) {
		if len(pipeline.InputFiles) == 0 {
			if inputIndex > 0 {
				return nil, nil
			}
			inputIndex++
			inp, err := cmdutil.StreamFileOrStdin(nil, enc)
			inputName = "stdin"
			return inp, err
		}
		if inputIndex >= len(pipeline.InputFiles) {
			return nil, nil
		}
		inputName = pipeline.InputFiles[inputIndex]
		data, err := cmdutil.ReadURL(inputName, enc)
		if err != nil {
			return nil, err
		}
		inputIndex++
		return bytes.NewReader(data), nil
	}
	for {
		input, err := nextInput()
		if err != nil {
			return err
		}
		if input == nil {
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

		_, err = jsoningest.IngestStream(pipeline.Context, baseID, input, parser, builder)
		if err != nil {
			return fmt.Errorf("While reading input %s: %w", inputName, err)
		}

		if err := pipeline.Next(); err != nil {
			return fmt.Errorf("Input was %s: %w", inputName, err)
		}
	}
	return nil
}

func init() {
	ingestCmd.AddCommand(ingestJSONCmd)
	ingestJSONCmd.Flags().String("id", "http://example.org/root", "Base ID to use for ingested nodes")

	operations["ingest/json"] = func() Step {
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
		p := []Step{
			&ing,
			NewWriteGraphStep(cmd),
		}
		_, err := runPipeline(p, initialGraph, args)
		return err
	},
}
