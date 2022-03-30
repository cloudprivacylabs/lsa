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
	"encoding/csv"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	csvingest "github.com/cloudprivacylabs/lsa/pkg/csv"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

func init() {
	ingestCmd.AddCommand(ingestCSVCmd)
	ingestCSVCmd.Flags().Int("startRow", 1, "Start row 0-based (default 1)")
	ingestCSVCmd.Flags().Int("endRow", -1, "End row 0-based")
	ingestCSVCmd.Flags().Int("headerRow", -1, "Header row 0-based (default: no header)")
	ingestCSVCmd.Flags().String("id", "row_{{.rowIndex}}", "Object ID Go template for ingested data if no ID is declared in the schema")
	ingestCSVCmd.Flags().String("compiledschema", "", "Use the given compiled schema")
	ingestCSVCmd.Flags().String("initialGraph", "", "Load this graph and ingest data onto it")
}

var ingestCSVCmd = &cobra.Command{
	Use:   "csv",
	Short: "Ingest a CSV document and enrich it with a schema",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		initialGraph, _ := cmd.Flags().GetString("initialGraph")
		ctx := getContext()
		layer := loadSchemaCmd(ctx, cmd)
		f, err := os.Open(args[0])
		if err != nil {
			failErr(err)
		}
		valuesets := &Valuesets{}
		loadValuesetsCmd(cmd, valuesets)

		reader := csv.NewReader(f)
		startRow, err := cmd.Flags().GetInt("startRow")
		if err != nil {
			failErr(err)
		}
		endRow, err := cmd.Flags().GetInt("endRow")
		if err != nil {
			failErr(err)
		}
		headerRow, err := cmd.Flags().GetInt("headerRow")
		if err != nil {
			failErr(err)
		}
		if headerRow >= startRow {
			fail("Header row is ahead of start row")
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
		ingester := csvingest.Ingester{
			Ingester: ls.Ingester{
				Schema:               layer,
				EmbedSchemaNodes:     embedSchemaNodes,
				OnlySchemaAttributes: onlySchemaAttributes,
				ValuesetFunc:         valuesets.Lookup,
				Graph:                grph,
			},
		}
		idTemplate, _ := cmd.Flags().GetString("id")
		idTmp, err := template.New("id").Parse(idTemplate)
		if err != nil {
			failErr(err)
		}
		for row := 0; ; row++ {
			rowData, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				failErr(err)
			}
			if headerRow == row {
				ingester.ColumnNames = rowData
			} else if row >= startRow {
				if endRow != -1 && row > endRow {
					break
				}
				templateData := map[string]interface{}{
					"rowIndex":  row,
					"dataIndex": row - startRow,
					"columns":   rowData,
				}
				buf := bytes.Buffer{}
				if err := idTmp.Execute(&buf, templateData); err != nil {
					failErr(err)
				}
				_, err := ingester.Ingest(ctx, rowData, strings.TrimSpace(buf.String()))
				if err != nil {
					failErr(err)
				}
			}
		}
		outFormat, _ := cmd.Flags().GetString("output")
		includeSchema, _ := cmd.Flags().GetBool("includeSchema")
		err = OutputIngestedGraph(cmd, outFormat, ingester.Graph, os.Stdout, includeSchema)
		if err != nil {
			failErr(err)
		}
	},
}
