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
	"encoding/csv"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/layers/cmd/pipeline"
	lscsv "github.com/cloudprivacylabs/lsa/pkg/csv"
)

type CSVExport struct {
	SpecFile string `json:"specFile" yaml:"specFile"`
	lscsv.Writer
	File string `json:"file" yaml:"file"`

	initialized   bool
	writtenHeader bool
	csvWriter     *csv.Writer
}

func (CSVExport) Help() {
	fmt.Println(`Export CSV Data from Graph
Export the graph in the pipeline context as a CSV file

operation: export/csv
params:`)
	fmt.Println(`  specFile: File containing export spec, or
  rowQuery: openCypher query that returns nodes, the roots of CSV rows
            if omitted, all source nodes of the graph are used
  file: output file
  columns:
  - name: column name. This name is written to the output as the column header
    query: column query. If empty, the query is
        match (root)-[]->(n:DocumentNode {attributeName: <attributeName>}) return n
        The query is evauated with 'root' pointing to the current row root node`)
}

func (ecsv *CSVExport) Run(pipeline *pipeline.PipelineContext) error {
	if !ecsv.initialized {
		if ecsv.SpecFile != "" {
			if err := cmdutil.ReadJSONOrYAML(ecsv.SpecFile, &ecsv.Writer); err != nil {
				return err
			}
		}
		if len(ecsv.File) > 0 {
			f, err := os.OpenFile(ecsv.File, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return err
			}
			off, err := f.Seek(0, 2)
			if err != nil {
				return err
			}
			if off > 0 {
				// Assume header written
				ecsv.writtenHeader = true
			}
			ecsv.csvWriter = csv.NewWriter(f)
		} else {
			ecsv.csvWriter = csv.NewWriter(os.Stdout)
		}
		ecsv.initialized = true
	}
	if !ecsv.writtenHeader {
		ecsv.Writer.WriteHeader(ecsv.csvWriter)
		ecsv.writtenHeader = true
	}
	ecsv.Writer.WriteRows(ecsv.csvWriter, pipeline.GetGraphRO())
	ecsv.csvWriter.Flush()
	return nil
}

func init() {
	exportCmd.AddCommand(exportCSVCmd)
	exportCSVCmd.Flags().String("input", "json", "Input graph format (json, jsonld)")
	exportCSVCmd.Flags().String("spec", "", "Export spec")

	pipeline.Operations["export/csv"] = func() pipeline.Step { return &CSVExport{} }
}

var exportCSVCmd = &cobra.Command{
	Use:   "csv",
	Short: "Export a graph as a CSV document",
	Long: `Export a graph as CSV.

If no spec file is given, the output is generated using attributeName properties.
A spec file controls how the output is generated.

{
  "rowQuery": "optional openCypher query that selects row root nodes",
  "columns": [
    "name": "Column name",
    "query": "optional openCypher query that selected column data"
  ]
}

If rowQuery is not specified, all the source nodes of the graph are written as rows. 
The colum queries are evaluated with 'root' predefined to point to the current
row root node. If a column query is not specified, it is assumed to be:

  (root) -[]-> (:DocumentNode {attributeName: <attrName>})
`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		step := &CSVExport{}
		step.SpecFile, _ = cmd.Flags().GetString("spec")
		p := []pipeline.Step{
			NewReadGraphStep(cmd),
			step,
		}
		_, err := runPipeline(p, "", args)
		return err
	},
}
