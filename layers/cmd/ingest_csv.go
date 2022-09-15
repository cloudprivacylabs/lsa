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
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lpg"
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
	StartRow     int    `json:"startRow" yaml:"startRow"`
	EndRow       int    `json:"endRow" yaml:"endRow"`
	HeaderRow    int    `json:"headerRow" yaml:"headerRow"`
	ID           string `json:"id" yaml:"id"`
	IngestByRows bool   `json:"ingestByRows" yaml:"ingestByRows"`
	Delimiter    string `json:"delimiter" yaml:"delimiter"`
	ColumnRange  uint   `json:"columnRange" yaml:"columnRange"`
	initialized  bool
	ingester     *ls.Ingester
	entities     []CSVJoinConfig
}

type CSVJoinConfig struct {
	VariantID        string
	StartCol, EndCol int
	IDCols           []int
}

type joinData struct {
	CSVJoinConfig
	data []string
	id   string
}

func (cji *CSVJoinIngester) Run(pipeline *pipeline.PipelineContext) error {
	var layer *ls.Layer
	var err error
	if !cji.initialized {
		layer, err = LoadSchemaFromFileOrRepo(pipeline.Context, cji.CompiledSchema, cji.Repo, cji.Schema, cji.Type, cji.Bundle)
		if err != nil {
			return err
		}
		pipeline.Properties["layer"] = layer
		cji.initialized = true
		cji.ingester = &ls.Ingester{Schema: layer}
	}

	parser := csvingest.Parser{
		OnlySchemaAttributes: cji.OnlySchemaAttributes,
		SchemaNode:           layer.GetSchemaRootNode(),
		IngestNullValues:     cji.IngestNullValues,
	}
	idTemplate := cji.ID
	if idTemplate == "" {
		idTemplate = "row_{{.rowIndex}}"
	}
	idTmp, err := template.New("id").Parse(idTemplate)
	if err != nil {
		return err
	}
	if cji.HeaderRow >= cji.StartRow {
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
		if !cji.IngestByRows {
			pipeline.SetGraph(cmdutil.NewDocumentGraph())
		}
		//
		seenIds := make(map[string][]string)
		joinCtx := make([]joinData, 0)
		done := false
		var doneErr error
		for row := 0; !done; row++ {
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
				if row == cji.HeaderRow {
					parser.ColumnNames = rowData
					return
				}
				if row < cji.StartRow {
					return
				}
				if cji.EndRow != -1 && row > cji.EndRow {
					done = true
					return
				}
				builder := ls.NewGraphBuilder(pipeline.Graph, ls.GraphBuilderOptions{
					EmbedSchemaNodes:     cji.EmbedSchemaNodes,
					OnlySchemaAttributes: cji.OnlySchemaAttributes,
				})
				templateData := map[string]interface{}{
					"rowIndex":  row,
					"dataIndex": row - cji.StartRow,
					"columns":   rowData,
				}
				buf := bytes.Buffer{}
				if err := idTmp.Execute(&buf, templateData); err != nil {
					doneErr = err
					return
				}

				// compare between joinCtx[0].id and first range of row
				if len(joinCtx) > 0 {
					firstEntityHash := GenerateHashFromIDs(rowData, joinCtx[0].IDCols)
					if firstEntityHash != joinCtx[0].id {
						// ingest previous graph
						pipeline.Context.GetLogger().Debug(map[string]interface{}{"csvingest.row": row, "stage": "Parsing"})
						//////////////////// not joinCtx[0].data, need all data
						csvGraphData := make([]string, 0, len(joinCtx))
						for _, val := range joinCtx {
							csvGraphData = append(csvGraphData, val.data...)
						}
						parsed, err := cji.parseCSV(pipeline, parser, csvGraphData, buf)
						if err != nil {
							doneErr = err
							return
						}
						r, err := cji.ingestCSV(builder, parsed)
						if err != nil {
							doneErr = err
							return
						}
						if cmdutil.GetConfig().SourceProperty == "" {
							r.SetProperty("source", entryInfo.GetName())
						} else {
							r.SetProperty(cmdutil.GetConfig().SourceProperty, entryInfo.GetName())
						}

						// new graph, output previous
						builder = ls.NewGraphBuilder(pipeline.Graph, ls.GraphBuilderOptions{
							EmbedSchemaNodes:     cji.EmbedSchemaNodes,
							OnlySchemaAttributes: cji.OnlySchemaAttributes,
						})
						joinCtx = make([]joinData, 0)
						seenIds = make(map[string][]string)
						fmt.Println("new graph")
					}
				}
				for _, schema := range cji.entities {
					if schema.EndCol < len(rowData) {
						// seenIds = make(map[string][]string)
						hashID := GenerateHashFromIDs(rowData, schema.IDCols)
						if _, seen := seenIds[hashID]; seen {
							continue
						}
						data := rowData[schema.StartCol : schema.EndCol+1]
						joinCtx = append(joinCtx, joinData{
							CSVJoinConfig: schema,
							data:          data,
							id:            hashID,
						})
						seenIds[hashID] = data
					}
				}
			}()
			if doneErr != nil {
				return doneErr
			}
		}
	}
	return nil
}

func (cji *CSVJoinIngester) ingestCSVJoin(entities []CSVJoinConfig, reader *csv.Reader) error {
	// reader := csv.NewReader(stream)
	// reader.Comma = rune(cji.Delimiter[0])
	seenIds := make(map[string][]string)
	joinCtx := make([]joinData, 0)
	for row := 0; ; row++ {
		rowData, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if row == cji.HeaderRow {
			continue
		}

		// compare between joinCtx[0].id and first range of row
		if len(joinCtx) > 0 {
			checkRange := GenerateHashFromIDs(rowData, joinCtx[0].IDCols)
			if checkRange != joinCtx[0].id {
				// new graph, output previous
				joinCtx = make([]joinData, 0)
				seenIds = make(map[string][]string)
				fmt.Println("new graph")
			}
		}
		for _, schema := range entities {
			if schema.EndCol < len(rowData) {
				hashID := GenerateHashFromIDs(rowData, schema.IDCols)
				if _, seen := seenIds[hashID]; seen {
					continue
				}
				data := rowData[schema.StartCol : schema.EndCol+1]
				joinCtx = append(joinCtx, joinData{
					CSVJoinConfig: schema,
					data:          data,
					id:            hashID,
				})
				seenIds[hashID] = data
			}
		}
	}

	return nil
}

func (cji *CSVJoinIngester) parseCSV(pipeline *pipeline.PipelineContext, parser csvingest.Parser, data []string, buf bytes.Buffer) (ls.ParsedDocNode, error) {
	parsed, err := parser.ParseDoc(pipeline.Context, strings.TrimSpace(buf.String()), data)
	if err != nil {
		return nil, err
	}
	if parsed == nil {
		return nil, err
	}
	return parsed, nil
}

func (cji *CSVJoinIngester) ingestCSV(builder ls.GraphBuilder, parsed ls.ParsedDocNode) (*lpg.Node, error) {
	r, err := cji.ingester.Ingest(builder, parsed)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func GenerateHashFromIDs(rowData []string, ids []int) string {
	buf := bytes.Buffer{}
	// if no ID columns, generate hash from all columns in row
	if len(ids) == 0 {
		buf.WriteString(strings.Join(rowData, ""))
		hash := sha256.Sum256([]byte(buf.String()))
		return hex.EncodeToString(hash[:])
	}
	// generate hash only for given column ids
	buf.WriteString(strings.Join(rowData[ids[0]:ids[len(ids)-1]], ""))
	hash := sha256.Sum256([]byte(buf.String()))
	return hex.EncodeToString(hash[:])
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
