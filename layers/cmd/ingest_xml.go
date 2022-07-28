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
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	xmlingest "github.com/cloudprivacylabs/lsa/pkg/xml"
)

type XMLIngester struct {
	BaseIngestParams
	ID          string
	initialized bool
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
		xml.initialized = true
	}

	// enc := encoding.Nop
	// if layer != nil {
	// 	enc, err = layer.GetEncoding()
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// inputIndex := 0
	// var inputName string
	// nextInput := func() (io.Reader, error) {
	// 	if len(pipeline.InputFiles) == 0 {
	// 		if inputIndex > 0 {
	// 			return nil, nil
	// 		}
	// 		inputIndex++
	// 		inp, err := cmdutil.StreamFileOrStdin(nil, enc)
	// 		inputName = "stdin"
	// 		return inp, err
	// 	}
	// 	if inputIndex >= len(pipeline.InputFiles) {
	// 		return nil, nil
	// 	}
	// 	inputName = pipeline.InputFiles[inputIndex]
	// 	data, err := cmdutil.ReadURL(inputName, enc)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	inputIndex++
	// 	return bytes.NewReader(data), nil
	// }
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
		if rc == nil {
			break
		}
		pipeline.SetGraph(ls.NewDocumentGraph())
		parser := xmlingest.Parser{
			OnlySchemaAttributes: xml.OnlySchemaAttributes,
			IngestEmptyValues:    xml.IngestNullValues,
		}
		if layer != nil {
			parser.Layer = layer
		}
		builder := ls.NewGraphBuilder(pipeline.GetGraphRW(), ls.GraphBuilderOptions{
			EmbedSchemaNodes:     xml.EmbedSchemaNodes,
			OnlySchemaAttributes: xml.OnlySchemaAttributes,
		})

		baseID := xml.ID

		parsed, err := parser.ParseStream(pipeline.Context, baseID, rc)
		if err != nil {
			return fmt.Errorf("While reading input %s: %w", "stdin", err)
		}
		_, err = ls.Ingest(builder, parsed)
		if err != nil {
			return fmt.Errorf("While reading input %s: %w", "stdin", err)
		}
		if err := builder.LinkNodes(pipeline.Context, parser.Layer, ls.GetEntityInfo(builder.GetGraph())); err != nil {
			return fmt.Errorf("While reading input %s: %w", "stdin", err)
		}

		if err := pipeline.Next(); err != nil {
			return fmt.Errorf("Input was %s: %w", "stdin", err)
		}
	}
	return nil
}

func init() {
	ingestCmd.AddCommand(ingestXMLCmd)
	ingestXMLCmd.Flags().String("id", "http://example.org/root", "Base ID to use for ingested nodes")

	pipeline.Operations["ingest/xml"] = func() pipeline.Step {
		return &XMLIngester{
			BaseIngestParams: BaseIngestParams{
				EmbedSchemaNodes: true,
			},
		}
	}
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
