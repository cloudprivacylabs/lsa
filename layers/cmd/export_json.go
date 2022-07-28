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

	"github.com/cloudprivacylabs/lsa/layers/cmd/pipeline"
	jsoningest "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/opencypher/graph"
)

type JSONExport struct{}

func (JSONExport) Help() {
	fmt.Println(`Export JSON Data from Graph
Export the graph in the pipeline context as a JSON file.
The output is constructed using "attributeName" annotations.

operation: export/json
params:`)
}

func (*JSONExport) Run(pipeline *pipeline.PipelineContext) error {
	for _, node := range graph.Sources(pipeline.GetGraphRO()) {
		exportOptions := jsoningest.ExportOptions{}
		data, err := jsoningest.Export(node, exportOptions)
		if err != nil {
			failErr(err)
		}
		data.Encode(ExportTarget)
	}
	return nil
}

func init() {
	exportCmd.AddCommand(exportJSONCmd)
	exportJSONCmd.Flags().String("input", "json", "Input graph format (json, jsonld)")

	pipeline.Operations["export/json"] = func() pipeline.Step { return &JSONExport{} }
}

var exportJSONCmd = &cobra.Command{
	Use:   "json",
	Short: "Export a graph as a JSON document",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		step := &JSONExport{}
		p := []pipeline.Step{
			NewReadGraphStep(cmd),
			step,
		}
		_, err := runPipeline(p, "", args)
		return err
	},
}
