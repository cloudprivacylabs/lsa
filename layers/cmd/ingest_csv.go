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
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	csvingest "github.com/cloudprivacylabs/lsa/pkg/csv"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type CSVIngester struct {
	BaseIngestParams
	StartRow     int    `json:"startRow" yaml:"startRow"`
	EndRow       int    `json:"endRow" yaml:"endRow"`
	HeaderRow    int    `json:"headerRow" yaml:"headerRow"`
	ID           string `json:"id" yaml:"id"`
	IngestByRows bool   `json:"ingestByRows" yaml:"ingestByRows"`
	initialized  bool
}

func (CSVIngester) Help() {
	fmt.Println(`Ingest CSV Data
Ingest CSV data from files using a schema variant and output a graph

operation: ingest/csv
params:`)
	fmt.Println(baseIngestParamsHelp)
	fmt.Println(`  # CSV Specifics
  startRow: 0   # Data starts at this row. 0-based
  endRow: -1    # Data ends at this row. 0-based
  headerRow: -1 # The row containing CSV header. 0-based
  id:"row_{{.rowIndex}}"   # Go template for node ID generation 
  # The template is evaluated with these variables:
  #  .rowIndex: The index of the current row in file
  #  .dataIndex: The index of the current data row
  #  .columns: The current row data
  ingestByRows: false  # If true, ingest row by row. Otherwise, ingest one file at a time.`)
}

func (ci *CSVIngester) Run(pipeline *PipelineContext) error {
	var layer *ls.Layer
	var err error
	if !ci.initialized {
		layer, err = LoadSchemaFromFileOrRepo(pipeline.Context, ci.CompiledSchema, ci.Repo, ci.Schema, ci.Type, ci.Bundle)
		if err != nil {
			return err
		}
		pipeline.Properties["layer"] = layer
		ci.initialized = true
	}

	parser := csvingest.Parser{
		OnlySchemaAttributes: ci.OnlySchemaAttributes,
		SchemaNode:           layer.GetSchemaRootNode(),
	}
	idTemplate := ci.ID
	if idTemplate == "" {
		idTemplate = "row_{{.rowIndex}}"
	}
	idTmp, err := template.New("id").Parse(idTemplate)
	if err != nil {
		return err
	}
	if ci.HeaderRow >= ci.StartRow {
		return errors.New("Header row is ahead of start row")
	}

	for _, inputFile := range pipeline.InputFiles {
		file, err := os.Open(inputFile)
		if err != nil {
			return err
		}
		reader := csv.NewReader(file)
		grph := pipeline.Graph

		for row := 0; ; row++ {
			rowData, err := reader.Read()
			if err == io.EOF {
				file.Close()
				break
			}
			if err != nil {
				file.Close()
				return err
			}
			if ci.HeaderRow == row {
				parser.ColumnNames = rowData
				continue
			}
			if row < ci.StartRow {
				continue
			}
			if ci.EndRow != -1 && row > ci.EndRow {
				file.Close()
				break
			}
			builder := ls.NewGraphBuilder(grph, ls.GraphBuilderOptions{
				EmbedSchemaNodes:     ci.EmbedSchemaNodes,
				OnlySchemaAttributes: ci.OnlySchemaAttributes,
			})
			templateData := map[string]interface{}{
				"rowIndex":  row,
				"dataIndex": row - ci.StartRow,
				"columns":   rowData,
			}
			buf := bytes.Buffer{}
			if err := idTmp.Execute(&buf, templateData); err != nil {
				file.Close()
				return err
			}
			parsed, err := parser.ParseDoc(pipeline.Context, strings.TrimSpace(buf.String()), rowData)
			if err != nil {
				file.Close()
				return err
			}
			_, err = ls.Ingest(builder, parsed)
			if err != nil {
				file.Close()
				return err
			}
			if ci.IngestByRows {
				if err := pipeline.Next(); err != nil {
					file.Close()
					return err
				}
				// New graph here
				pipeline.Graph = ls.NewDocumentGraph()
			}
		}
		if !ci.IngestByRows {
			if err := pipeline.Next(); err != nil {
				return err
			}
			pipeline.Graph = ls.NewDocumentGraph()
		}
	}
	return nil
}

func init() {
	ingestCmd.AddCommand(ingestCSVCmd)
	ingestCSVCmd.Flags().Int("startRow", 1, "Start row 0-based (default 1)")
	ingestCSVCmd.Flags().Int("endRow", -1, "End row 0-based")
	ingestCSVCmd.Flags().Int("headerRow", -1, "Header row 0-based (default: no header)")
	ingestCSVCmd.Flags().String("id", "row_{{.rowIndex}}", "Object ID Go template for ingested data if no ID is declared in the schema")
	ingestCSVCmd.Flags().String("compiledschema", "", "Use the given compiled schema")
	ingestCSVCmd.Flags().String("initialGraph", "", "Load this graph and ingest data onto it")
	ingestCSVCmd.Flags().Bool("byFile", false, "Ingest one file at a time. Default is row at a time.")

	operations["ingest/csv"] = func() Step {
		return &CSVIngester{
			EndRow:    -1,
			HeaderRow: -1,
			StartRow:  0,
		}
	}
}

var ingestCSVCmd = &cobra.Command{
	Use:   "csv",
	Short: "Ingest a CSV document and enrich it with a schema",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		initialGraph, _ := cmd.Flags().GetString("initialGraph")
		ing := CSVIngester{}
		ing.fromCmd(cmd)
		var err error
		ing.StartRow, err = cmd.Flags().GetInt("startRow")
		if err != nil {
			return err
		}
		ing.EndRow, err = cmd.Flags().GetInt("endRow")
		if err != nil {
			return err
		}
		ing.HeaderRow, err = cmd.Flags().GetInt("headerRow")
		if err != nil {
			return err
		}
		if ing.HeaderRow >= ing.StartRow {
			return fmt.Errorf("Header row is ahead of start row")
		}
		ing.ID, err = cmd.Flags().GetString("id")
		if err != nil {
			return err
		}
		byFile, err := cmd.Flags().GetBool("byFile")
		if err != nil {
			return err
		}
		ing.IngestByRows = !byFile
		p := []Step{
			&ing,
			NewWriteGraphStep(cmd),
		}
		_, err = runPipeline(p, initialGraph, args)
		return err
	},
}
