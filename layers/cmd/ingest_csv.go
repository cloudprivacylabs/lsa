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
	"runtime/debug"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lpg/v2"

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

func (CSVIngester) Name() string { return "ingest/csv" }

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

func (ci *CSVIngester) Flush(ctx *pipeline.PipelineContext) error {
	return ctx.FlushNext()
}

func (ci *CSVIngester) Run(pipeline *pipeline.PipelineContext) error {
	var layer *ls.Layer
	var err error
	if !ci.initialized {
		layer, err = LoadSchemaFromFile(pipeline.Context, ci.CompiledSchema, ci.Schema, ci.Type, ci.Bundle)
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
	defer ci.Flush(pipeline)

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
		if len(ci.Delimiter) > 0 {
			reader.Comma = rune(ci.Delimiter[0])
		}
		reader.LazyQuotes = true
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
				pipeline.EntryLogger(pipeline, map[string]interface{}{
					"input": entryInfo.GetName(),
					"row":   row,
				})
				r, err := csvingest.ParseIngest(pipeline.Context, ci.ingester, parser, builder, strings.TrimSpace(buf.String()), rowData)
				if err != nil {
					doneErr = err
					return
				}
				r.SetProperty(ls.SourceTerm.Name, ls.NewPropertyValue(ls.SourceTerm.Name, fmt.Sprintf("%s#%d", entryInfo.GetName(), row)))
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
	StartRow      int             `json:"startRow" yaml:"startRow"`
	EndRow        int             `json:"endRow" yaml:"endRow"`
	HeaderRow     int             `json:"headerRow" yaml:"headerRow"`
	ID            string          `json:"id" yaml:"id"`
	Delimiter     string          `json:"delimiter" yaml:"delimiter"`
	Entities      []CSVJoinConfig `json:"entities" yaml:"entities"`
	initialized   bool
	ingester      map[string]*ls.Ingester
	combinedLayer *ls.Layer
}

func (CSVJoinIngester) Name() string { return "ingest/csv/join" }

func (cji *CSVJoinIngester) Flush(pipeline *pipeline.PipelineContext) error {
	return pipeline.FlushNext()
}

type CSVJoinConfig struct {
	VariantID string `json:"variantId" yaml:"variantId"`
	Cols      []int  `json:"cols" yaml:"cols"`
	IDCols    []int  `json:"idCols" yaml:"idCols"`
	// Any one of these columns must be full to assume this is valid entry
	Any []int `json:"any" yaml:"any"`

	layer *ls.Layer
}

func (c CSVJoinConfig) makeRowData(rowData []string) []string {
	if len(c.Any) > 0 {
		full := false
		for _, i := range c.Any {
			if i >= 0 && i < len(rowData) {
				if len(strings.TrimSpace(rowData[i])) > 0 {
					full = true
					break
				}
			}
		}
		if !full {
			return nil
		}
	}
	ret := make([]string, len(c.Cols))
	for index, i := range c.Cols {
		if i >= 0 && i < len(rowData) {
			ret[index] = rowData[i]
		}
	}
	return ret
}

type joinData struct {
	CSVJoinConfig
	data []string
	id   string
}

func (cji *CSVJoinIngester) Run(pipeline *pipeline.PipelineContext) error {
	if !cji.initialized {
		schLoader, err := LoadBundle(pipeline.Context, cji.Bundle)
		if err != nil {
			return err
		}
		cji.ingester = make(map[string]*ls.Ingester)
		cji.initialized = true
		if len(cji.Entities) == 0 {
			return fmt.Errorf("No entities")
		}
		combinedGraph := ls.NewLayerGraph()
		nodeMap := make(map[*lpg.Node]*lpg.Node)
		for i := range cji.Entities {
			compiler := ls.Compiler{
				Loader: schLoader,
			}
			cji.Entities[i].layer, err = compiler.Compile(pipeline.Context, cji.Entities[i].VariantID)
			if err != nil {
				return err
			}
			for k, v := range ls.CopyGraph(combinedGraph, cji.Entities[i].layer.Graph, nil, nil) {
				nodeMap[k] = v
			}
		}
		cji.combinedLayer = ls.NewLayerFromRootNode(nodeMap[cji.Entities[0].layer.GetLayerRootNode()])
	}
	pipeline.Properties["layer"] = cji.combinedLayer
	defer cji.Flush(pipeline)
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
		if cji.Schema != "" {
			return errors.New("Unexpected schema")
		}
		if len(cji.Delimiter) > 0 {
			reader.Comma = rune(cji.Delimiter[0])
		}
		reader.LazyQuotes = true
		parsers := make(map[string]*csvingest.Parser)
		for _, variant := range cji.Entities {
			cji.ingester[variant.VariantID] = &ls.Ingester{Schema: variant.layer}
			parser := csvingest.Parser{
				OnlySchemaAttributes: cji.OnlySchemaAttributes,
				SchemaNode:           variant.layer.GetSchemaRootNode(),
				IngestNullValues:     cji.IngestNullValues,
			}
			parsers[variant.VariantID] = &parser
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
		seenIDs := make(map[string]map[string]struct{})
		joinCtx := make([]joinData, 0)
		done := false
		var doneErr error
		var newGraphStart int = cji.StartRow
		for row := 0; !done; row++ {
			pipeline.EntryLogger(pipeline, map[string]interface{}{
				"input": entryInfo.GetName(),
				"row":   row,
			})
			func() {
				defer func() {
					if err := recover(); err != nil {
						pipeline.ErrorLogger(pipeline, fmt.Errorf("Error in file: %s, row: %d, %v, %v", entryInfo.GetName(), row, err, string(debug.Stack())))
						doneErr = fmt.Errorf("%v", err)
					}
				}()
				buf := bytes.Buffer{}
				builder := ls.NewGraphBuilder(pipeline.Graph, ls.GraphBuilderOptions{
					EmbedSchemaNodes:     cji.EmbedSchemaNodes,
					OnlySchemaAttributes: cji.OnlySchemaAttributes,
				})
				rowData, err := reader.Read()

				ingestAndNext := func() error {
					err = templateExecute(newGraphStart, cji.StartRow, row, rowData, idTmp, &buf)
					if err != nil {
						return err
					}
					pipeline.Context.GetLogger().Debug(map[string]interface{}{"csvingest.row": row - 1, "stage": "Parsing"})
					err = cji.csvParseIngestEntities(pipeline, cji.ingester, joinCtx, parsers, builder, strings.TrimSpace(buf.String()), entryInfo.GetName())
					if err != nil {
						return err
					}
					if err := pipeline.Next(); err != nil {
						return err
					}
					return nil
				}

				if err == io.EOF {
					if err := ingestAndNext(); err != nil {
						doneErr = err
						return
					}
					done = true
					return
				}
				if err != nil {
					doneErr = err
					return
				}
				if row == cji.HeaderRow {
					for _, ent := range cji.Entities {
						parsers[ent.VariantID].ColumnNames = ent.makeRowData(rowData)
					}
					return
				}
				if row < cji.StartRow {
					return
				}
				if cji.EndRow != -1 && row > cji.EndRow {
					if err := ingestAndNext(); err != nil {
						doneErr = err
						return
					}
					done = true
					return
				}
				// compare between joinCtx[0].id and first range of row
				if len(joinCtx) > 0 {
					firstEntityHash := GenerateHashFromIDs(rowData, joinCtx[0].VariantID, joinCtx[0].IDCols)
					if firstEntityHash != joinCtx[0].id {
						if err := ingestAndNext(); err != nil {
							doneErr = err
							return
						}
						// new graph / reset builder, context, buffer
						pipeline.Context.GetLogger().Debug(map[string]interface{}{"csvingest.row": row, "stage": "New Graph"})
						pipeline.Graph = ls.NewDocumentGraph()
						builder = ls.NewGraphBuilder(pipeline.Graph, ls.GraphBuilderOptions{
							EmbedSchemaNodes:     cji.EmbedSchemaNodes,
							OnlySchemaAttributes: cji.OnlySchemaAttributes,
						})
						joinCtx = make([]joinData, 0)
						newGraphStart = row
						seenIDs = make(map[string]map[string]struct{})
					}
				}
				for _, entity := range cji.Entities {
					data := entity.makeRowData(rowData)
					if len(data) > 0 {
						hashID := GenerateHashFromIDs(rowData, entity.VariantID, entity.IDCols)
						if _, seen := seenIDs[entity.VariantID][hashID]; seen {
							continue
						}
						// new map for every entity
						joinCtx = append(joinCtx, joinData{
							CSVJoinConfig: entity,
							data:          data,
							id:            hashID,
						})
						seenIDs[entity.VariantID] = map[string]struct{}{hashID: struct{}{}}
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

func (cji *CSVJoinIngester) csvParseIngestEntities(pipeline *pipeline.PipelineContext, ingesters map[string]*ls.Ingester, joinCtx []joinData, parsers map[string]*csvingest.Parser, builder ls.GraphBuilder, id string, entryName string) error {
	for _, csvEntityData := range joinCtx {
		parser := parsers[csvEntityData.VariantID]
		ingester := ingesters[csvEntityData.VariantID]

		parsed, err := parser.ParseDoc(pipeline.Context, id, csvEntityData.data)
		if err != nil {
			return err
		}
		if parsed == nil {
			return errors.New("Parsed CSV document is nil")
		}
		parsed.GetProperties()[ls.SourceTerm.Name] = ls.NewPropertyValue(ls.SourceTerm.Name, entryName)

		_, err = ingester.Ingest(builder, parsed)
		if err != nil {
			return err
		}
	}

	if err := builder.LinkNodes(pipeline.Context, ingesters[joinCtx[0].VariantID].Schema); err != nil {
		return err
	}
	return nil
}

func templateExecute(newGraphStart, startRow, row int, rowData []string, idTmp *template.Template, buf *bytes.Buffer) error {
	templateData := map[string]interface{}{
		"rowIndex":  newGraphStart,
		"dataIndex": row - startRow,
		"columns":   rowData,
	}
	if err := idTmp.Execute(buf, templateData); err != nil {
		return err
	}
	return nil
}

func GenerateHashFromIDs(rowData []string, variantID string, ids []int) string {
	h := sha256.New()
	h.Write([]byte(variantID))
	// if no ID columns, generate hash from all columns in row
	if len(ids) == 0 {
		for _, x := range rowData {
			h.Write([]byte(x))
		}
		return hex.EncodeToString(h.Sum(nil))
	}
	// generate hash only for given column ids
	for _, ix := range ids {
		if ix < len(rowData) {
			h.Write([]byte(rowData[ix]))
		}
	}
	return hex.EncodeToString(h.Sum(nil))
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
				IngestNullValues: false,
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
				IngestNullValues: false,
			},
			EndRow:    -1,
			HeaderRow: 0,
			StartRow:  1,
			Delimiter: ",",
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
		_, err = runPipeline(p, Environment, initialGraph, args)
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
		pl := []pipeline.Step{
			&ing,
			NewWriteGraphStep(cmd),
		}
		_, err = runPipeline(pl, Environment, initialGraph, args)
		return err
	},
}
