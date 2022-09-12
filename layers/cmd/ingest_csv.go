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
	"reflect"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/layers/cmd/pipeline"
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
	Delimiter    string `json:"delimiter" yaml:"delimiter"`
	initialized  bool
	ingester     *ls.Ingester
}

func (CSVIngester) Help() {
	fmt.Println(`Ingest CSV Data
Ingest CSV data from files using a schema variant and output a graph

operation: ingest/csv
params:`)
	fmt.Println(baseIngestParamsHelp)
	fmt.Println(`  # CSV Specifics
  startRow: 1   # Data starts at this row. 0-based
  endRow: -1    # Data ends at this row. 0-based
  headerRow: 0 # The row containing CSV header. 0-based
  delimiter: ,  # separator character
  id:"row_{{.rowIndex}}"   # Go template for node ID generation 
  # The template is evaluated with these variables:
  #  .rowIndex: The index of the current row in file
  #  .dataIndex: The index of the current data row
  #  .columns: The current row data
  ingestByRows: false  # If true, ingest row by row. Otherwise, ingest one file at a time.`)
}

func (ci *CSVIngester) Run(pipeline *pipeline.PipelineContext) error {
	var layer *ls.Layer
	var err error
	if !ci.initialized {
		layer, err = LoadSchemaFromFileOrRepo(pipeline.Context, ci.CompiledSchema, ci.Repo, ci.Schema, ci.Type, ci.Bundle)
		if err != nil {
			return err
		}
		pipeline.Properties["layer"] = layer
		ci.initialized = true
		ci.ingester = &ls.Ingester{Schema: layer}
	}

	parser := csvingest.Parser{
		OnlySchemaAttributes: ci.OnlySchemaAttributes,
		SchemaNode:           layer.GetSchemaRootNode(),
		IngestNullValues:     ci.IngestNullValues,
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

	for {
		entryInfo, stream, err := pipeline.NextInput()
		if err != nil {
			return err
		}
		if stream == nil {
			break
		}
		pipeline.Context.GetLogger().Debug(map[string]interface{}{"csvingest": "start new stream"})
		reader := csv.NewReader(stream)
		if !ci.IngestByRows {
			pipeline.SetGraph(cmdutil.NewDocumentGraph())
		}
		reader.Comma = rune(ci.Delimiter[0])
		var doneErr error
		done := false
		for row := 0; !done; row++ {
			pipeline.Context.GetLogger().Debug(map[string]interface{}{"csvingest.row": row})
			func() {
				defer func() {
					if err := recover(); err != nil {
						pipeline.ErrorLogger(pipeline, fmt.Errorf("Error in file: %s, row: %d %v", entryInfo.GetName(), row, err))
						doneErr = fmt.Errorf("%v", err)
					}
				}()
				rowData, err := reader.Read()
				if err == io.EOF {
					done = true
					return
				}
				if err != nil {
					doneErr = err
					return
				}
				if ci.HeaderRow == row {
					parser.ColumnNames = rowData
					return
				}
				if row < ci.StartRow {
					return
				}
				if ci.EndRow != -1 && row > ci.EndRow {
					done = true
					return
				}
				if ci.IngestByRows {
					pipeline.SetGraph(cmdutil.NewDocumentGraph())
				}
				builder := ls.NewGraphBuilder(pipeline.Graph, ls.GraphBuilderOptions{
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
					doneErr = err
					return
				}
				pipeline.Context.GetLogger().Debug(map[string]interface{}{"csvingest.row": row, "stage": "Parsing"})
				parsed, err := parser.ParseDoc(pipeline.Context, strings.TrimSpace(buf.String()), rowData)
				if err != nil {
					doneErr = err
					return
				}
				if parsed == nil {
					return
				}

				r, err := ci.ingester.Ingest(builder, parsed)
				if err != nil {
					doneErr = err
					return
				}
				if cmdutil.GetConfig().SourceProperty == "" {
					r.SetProperty("source", entryInfo.GetName())
				} else {
					r.SetProperty(cmdutil.GetConfig().SourceProperty, entryInfo.GetName())
				}
				if ci.IngestByRows {
					if err := pipeline.Next(); err != nil {
						doneErr = err
						return
					}
				}
			}()
			if doneErr != nil {
				return doneErr
			}
			if !ci.IngestByRows {
				if err := pipeline.Next(); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

type CSVJoinIngester struct {
	BaseIngestParams
	Schemas      []string `json:"schemas" yaml:"schemas"`
	StartRow     int      `json:"startRow" yaml:"startRow"`
	EndRow       int      `json:"endRow" yaml:"endRow"`
	HeaderRow    int      `json:"headerRow" yaml:"headerRow"`
	ID           string   `json:"id" yaml:"id"`
	IngestByRows bool     `json:"ingestByRows" yaml:"ingestByRows"`
	Delimiter    string   `json:"delimiter" yaml:"delimiter"`
	ColumnRange  uint     `json:"columnRange" yaml:"columnRange"`
	initialized  bool
	ingester     *ls.Ingester
}

/*
// vertical scan
	for i := 0; i < len(strs[0]); i++ {
		currentChar := strs[0][i]
		for j := 1; j < len(strs); j++ {
			if i == len(strs[j]) || currentChar != strs[j][i] {
				return strs[0][0:i]
			}
		}
	}
	return strs[0]
*/
func (cji *CSVJoinIngester) ingestCSVJoin(context *ls.Context, layer *ls.Layer, bundle string, schemas []string, columnRange uint, stream io.ReadCloser) error {
	// TODO
	// var graphBoundaryByRow int
	findSchemaDeviation := func(row1, row2 []string, position int) int {
		for x := position; x < int(columnRange); x++ {
			if row1[x] != row2[x] {
				return x
			}
		}
		return 0
	}
	reader := csv.NewReader(stream)
	var schemaRange int
	var nextSchemaRange int
	var firstSchemaRangeValues []string
	data, err := reader.ReadAll()
	if err == io.EOF {
		return err
	}
	// establish the first schema boundary
	schemaRange = findSchemaDeviation(data[0], data[1], 0)
	firstSchemaRangeValues = data[0][0:schemaRange]
	prevRangeValues := firstSchemaRangeValues
	var nextRangeValues []string
	parser := csvingest.Parser{
		OnlySchemaAttributes: cji.OnlySchemaAttributes,
		SchemaNode:           layer.GetSchemaRootNode(),
		IngestNullValues:     cji.IngestNullValues,
	}
	builder := ls.NewGraphBuilder(ls.NewDocumentGraph(), ls.GraphBuilderOptions{
		EmbedSchemaNodes:     cji.EmbedSchemaNodes,
		OnlySchemaAttributes: cji.OnlySchemaAttributes,
	})
	parsed, err := parser.ParseDoc(context, "", firstSchemaRangeValues)
	if err != nil {
		return err
	}
	if parsed == nil {
		return fmt.Errorf("Parsed is nil")
	}

	r, err := cji.ingester.Ingest(builder, parsed)
	if err != nil {
		return err
	}
	parent := r
	// assert schema ranges are same length
	for rowIdx, row := range data {
		for colIdx := schemaRange; colIdx < int(columnRange); colIdx += schemaRange {
			// create nodes on a per row by schema range basis
			if colIdx == schemaRange {
				// graph boundary is found when current row does having the matching schema values for the schema range
				// establish new range and first schema values for that range
				if !reflect.DeepEqual(firstSchemaRangeValues, row[0:colIdx]) {
					// graphBoundaryByRow = rowIdx
					builder = ls.NewGraphBuilder(ls.NewDocumentGraph(), ls.GraphBuilderOptions{
						EmbedSchemaNodes:     cji.EmbedSchemaNodes,
						OnlySchemaAttributes: cji.OnlySchemaAttributes,
					})
					firstSchemaRangeValues = data[rowIdx][0:schemaRange]
				}
			}
			if rowIdx+1 < len(data) && colIdx+schemaRange < len(row) {
				nextSchemaRange = findSchemaDeviation(row[colIdx:schemaRange], data[rowIdx+1][colIdx:schemaRange], colIdx)
				nextRangeValues = row[colIdx:schemaRange]
			}
			if !reflect.DeepEqual(prevRangeValues, nextRangeValues) {
				n := builder.NewNode(parent)
				builder.GetGraph().NewEdge(r, n, ls.HasTerm, nil)
			}
		}
	}
	// // vertical scan to find graph boundary
	// for rowIdx, row := range data {
	// 	// graph boundary found
	// 	if !reflect.DeepEqual(row[0:schemaRange], firstSchemaRangeValues) {

	// 	}
	// }
	return nil
}

func (cji *CSVJoinIngester) Run(pipeline *pipeline.PipelineContext) error {
	// TODO
	layer, err := LoadSchemaFromFileOrRepo(pipeline.Context, cji.CompiledSchema, cji.Repo, cji.Schema, cji.Type, cji.Bundle)
	if err != nil {
		return err
	}
	cji.initialized = true
	cji.ingester = &ls.Ingester{Schema: layer}
	return nil
}

func init() {
	ingestCmd.AddCommand(ingestCSVCmd)
	ingestCmd.AddCommand(ingestCSVJoinCmd)
	ingestCSVCmd.Flags().Int("startRow", 1, "Start row 0-based")
	ingestCSVCmd.Flags().Int("endRow", -1, "End row 0-based")
	ingestCSVCmd.Flags().Int("headerRow", 0, "Header row 0-based (default: 0) ")
	ingestCSVCmd.Flags().String("id", "row_{{.rowIndex}}", "Object ID Go template for ingested data if no ID is declared in the schema")
	ingestCSVCmd.Flags().String("compiledschema", "", "Use the given compiled schema")
	ingestCSVCmd.Flags().String("delimiter", ",", "Delimiter char")
	ingestCSVCmd.Flags().String("initialGraph", "", "Load this graph and ingest data onto it")
	ingestCSVCmd.Flags().Bool("byFile", false, "Ingest one file at a time. Default is row at a time.")

	pipeline.RegisterPipelineStep("ingest/csv", func() pipeline.Step {
		return &CSVIngester{
			BaseIngestParams: BaseIngestParams{
				EmbedSchemaNodes: true,
			},
			EndRow:       -1,
			HeaderRow:    0,
			StartRow:     1,
			Delimiter:    ",",
			IngestByRows: true,
		}
	})

	pipeline.RegisterPipelineStep("ingest/csv/join", func() pipeline.Step {
		return &CSVJoinIngester{
			BaseIngestParams: BaseIngestParams{
				EmbedSchemaNodes: true,
			},
			EndRow:       -1,
			HeaderRow:    0,
			StartRow:     1,
			Delimiter:    ",",
			IngestByRows: true,
		}
	})
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
		ing.Delimiter, err = cmd.Flags().GetString("delimiter")
		if err != nil {
			return err
		}
		byFile, err := cmd.Flags().GetBool("byFile")
		if err != nil {
			return err
		}
		ing.IngestByRows = !byFile
		p := []pipeline.Step{
			&ing,
			NewWriteGraphStep(cmd),
		}
		_, err = runPipeline(p, initialGraph, args)
		return err
	},
}

var ingestCSVJoinCmd = &cobra.Command{
	Use:   "csvjoin",
	Short: "Ingest a CSV document whose content is the result of SQL joins and enrich it with a schema",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		initialGraph, _ := cmd.Flags().GetString("initialGraph")
		ing := CSVJoinIngester{}
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
		ing.Delimiter, err = cmd.Flags().GetString("delimiter")
		if err != nil {
			return err
		}
		byFile, err := cmd.Flags().GetBool("byFile")
		if err != nil {
			return err
		}
		ing.IngestByRows = !byFile
		pl := []pipeline.Step{
			&ing,
			NewWriteGraphStep(cmd),
		}
		_, err = runPipeline(pl, initialGraph, args)
		return err
	},
}
