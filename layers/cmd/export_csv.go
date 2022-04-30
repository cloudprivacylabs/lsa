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
	"os"

	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	lscsv "github.com/cloudprivacylabs/lsa/pkg/csv"
)

type CSVExport struct {
	SpecFile string
	lscsv.Writer
	initialized   bool
	writtenHeader bool
}

func (ecsv *CSVExport) Run(pipeline *PipelineContext) error {
	var spec string
	if !ecsv.initialized {
		if ecsv.SpecFile != "" {
			if err := cmdutil.ReadJSONOrYAML(spec, &ecsv.Writer); err != nil {
				failErr(err)
			}
		}
		ecsv.initialized = true
	}
	csvExporter := ecsv.Writer

	wr := csv.NewWriter(os.Stdout)
	if !ecsv.writtenHeader {
		csvExporter.WriteHeader(wr)
		ecsv.writtenHeader = true
	}
	csvExporter.WriteRows(wr, pipeline.Graph)
	wr.Flush()
	return nil
}

func init() {
	exportCmd.AddCommand(exportCSVCmd)
	exportCSVCmd.Flags().String("input", "json", "Input graph format (json, jsonld)")
	exportCSVCmd.Flags().String("spec", "", "Export spec")

	operations["csvexport"] = func() Step { return &CSVExport{} }
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
	Run: func(cmd *cobra.Command, args []string) {
		input, _ := cmd.Flags().GetString("input")
		g, err := cmdutil.ReadGraph(args, nil, input)
		if err != nil {
			failErr(err)
		}
		csvExporter := lscsv.Writer{}
		spec, _ := cmd.Flags().GetString("spec")
		if len(spec) > 0 {
			if err := cmdutil.ReadJSONOrYAML(spec, &csvExporter); err != nil {
				failErr(err)
			}
		}

		wr := csv.NewWriter(ExportTarget)
		csvExporter.WriteHeader(wr)
		csvExporter.WriteRows(wr, g)
		wr.Flush()
	},
}
