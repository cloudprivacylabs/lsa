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
	"github.com/spf13/pflag"

	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func addSchemaFlags(flags *pflag.FlagSet) {
	flags.String("schema", "", "Schema file.")
	flags.String("type", "", "Use if a bundle is given for data types. The type name to ingest.")
	flags.StringSlice("bundle", nil, "Schema bundle(s).")
}

func init() {
	rootCmd.AddCommand(ingestCmd)
	addSchemaFlags(ingestCmd.PersistentFlags())
	ingestCmd.PersistentFlags().String("compiledschema", "", "Use the given compiled schema")
	ingestCmd.PersistentFlags().String("output", "json", "Output format, json, jsonld, or dot")
	ingestCmd.PersistentFlags().Bool("includeSchema", false, "Include schema in the output")
	ingestCmd.PersistentFlags().Bool("embedSchemaNodes", true, "Embed schema nodes into document nodes")
	ingestCmd.PersistentFlags().Bool("onlySchemaAttributes", false, "Only ingest nodes that have an associated schema attribute")
	ingestCmd.PersistentFlags().Bool("ingestNullValues", false, "Ingest values even if they are empty")
}

type BaseIngestParams struct {
	Schema               string   `json:"schema" yaml:"schema"`
	Type                 string   `json:"type" yaml:"type"`
	Bundle               []string `json:"bundle" yaml:"bundle"`
	CompiledSchema       string   `json:"compiledSchema" yaml:"compiledSchema"`
	EmbedSchemaNodes     bool     `json:"embedSchemaNodes" yaml:"embedSchemaNodes"`
	OnlySchemaAttributes bool     `json:"onlySchemaAttributes" yaml:"onlySchemaAttributes"`
	IngestNullValues     bool     `json:"ingestNullValues" yaml:"ingestNullValues"`
}

// IsEmptySchema returns true if none of the schema properties are set
func (b BaseIngestParams) IsEmptySchema() bool {
	return len(b.Schema) == 0 && len(b.Type) == 0 && len(b.Bundle) == 0 && len(b.CompiledSchema) == 0
}

func (b *BaseIngestParams) fromCmd(cmd *cobra.Command) {
	b.CompiledSchema, _ = cmd.Flags().GetString("compiledschema")
	b.Schema, _ = cmd.Flags().GetString("schema")
	b.Bundle, _ = cmd.Flags().GetStringSlice("bundle")
	b.Type, _ = cmd.Flags().GetString("type")
	b.EmbedSchemaNodes, _ = cmd.Flags().GetBool("embedSchemaNodes")
	b.OnlySchemaAttributes, _ = cmd.Flags().GetBool("onlySchemaAttributes")
	b.IngestNullValues, _ = cmd.Flags().GetBool("ingestNullValues")
}

const baseIngestParamsHelp = `  
  # Schema loading parameters
  # One of:
  #   bundle/type
  #   schema (schema file)
  #   compiledSchema

  bundle: bundleFileName
  type: typeName in bundle
  schema: The schema file
  compiledSchema: compiled schema graph file
  ingestNullValues: false # Whether to ingest empty/null values

  # Ingestion control

  embedSchemaNodes: false
  onlySchemaAttributes: false`

var ingestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "Ingest and enrich data with a schema",
	Long: `Ingest data using a schema variant. 

There are several ways a schema variant can be specified:

  layers ingest json --schema <schemaFile>

This form ingests data using the given schema file.
  
  layers ingest csv --bundle <bundlefile> --type <typeName>

This form uses a bundle that specifies type names and corresponding
schema variants. The input will be ingested using the variant for the
type. The bundle is a JSON file:

{
  "types": {
    "typeName": {
       "schema": "schemaFile",
       "overlays": [
          "overlayFile","overlayFile"
       ]
    },
    "typeName": {
       "schema": "schemaFile",
       "overlays": [
          "overlayFile","overlayFile"
       ]
    },
    ...
}


  layers ingest csv --compiledSchema <schemaGraphFile> --schema <schemaId>

This form will use a previously compiled schema.
`,
}

func loadSchemaCmd(ctx *ls.Context, cmd *cobra.Command) *ls.Layer {
	compiledSchema, _ := cmd.Flags().GetString("compiledschema")
	schemaName, _ := cmd.Flags().GetString("schema")
	bundleNames, _ := cmd.Flags().GetStringSlice("bundle")
	typeName, _ := cmd.Flags().GetString("type")
	layer, err := LoadSchemaFromFile(ctx, compiledSchema, schemaName, typeName, bundleNames)
	if err != nil {
		failErr(err)
	}
	return layer
}

func loadCompiledSchema(ctx *ls.Context, compiledSchema, schemaName string) (*ls.Layer, []*ls.Layer, error) {
	sch, err := cmdutil.ReadURL(compiledSchema)
	if err != nil {
		return nil, nil, err
	}
	m := ls.NewJSONMarshaler(ctx.GetInterner())
	g := ls.NewLayerGraph()
	err = m.Unmarshal(sch, g)
	if err != nil {
		return nil, nil, err
	}
	layers := ls.LayersFromGraph(g)
	var layer *ls.Layer
	for i := range layers {
		if schemaName == layers[i].GetID() {
			layer = layers[i]
			break
		}
	}
	if layer == nil {
		return nil, nil, fmt.Errorf("Not found: %s", schemaName)
	}
	compiler := ls.Compiler{}
	layer, err = compiler.CompileSchema(ctx, layer)
	if err != nil {
		return nil, nil, err
	}
	return layer, layers, nil
}

func LoadSchemaFromFile(ctx *ls.Context, compiledSchema, schemaName, typeName string, bundleNames []string) (*ls.Layer, error) {
	if len(compiledSchema) > 0 {
		l, _, err := loadCompiledSchema(ctx, compiledSchema, schemaName)
		return l, err
	}
	if len(bundleNames) > 0 {
		schLoader, err := LoadBundle(ctx, bundleNames)
		if err != nil {
			return nil, err
		}
		compiler := ls.Compiler{
			Loader: schLoader,
		}
		name := schemaName
		if len(typeName) > 0 {
			name = typeName
		}
		layer, err := compiler.Compile(ctx, name)
		if err != nil {
			return nil, err
		}
		return layer, nil
	}
	if len(schemaName) == 0 {
		return nil, fmt.Errorf("Empty schema name")
	}
	data, err := cmdutil.ReadURL(schemaName)
	if err != nil {
		return nil, err
	}
	layers, err := ReadLayers(data, ctx.GetInterner())
	if err != nil {
		return nil, err
	}
	if len(layers) > 1 {
		return nil, fmt.Errorf("Multiple layers in schema input")
	}
	compiler := ls.Compiler{
		Loader: ls.SchemaLoaderFunc(func(x string) (*ls.Layer, error) {
			if x == schemaName || x == layers[0].GetID() {
				return layers[0], nil
			}
			return nil, fmt.Errorf("Not found")
		}),
	}
	return compiler.Compile(ctx, schemaName)
}

func OutputIngestedGraph(cmd *cobra.Command, outFormat string, target *lpg.Graph, wr io.Writer, includeSchema bool) error {
	if !includeSchema {
		schemaNodes := make(map[*lpg.Node]struct{})
		for nodes := target.GetNodes(); nodes.Next(); {
			node := nodes.Node()
			if ls.IsDocumentNode(node) {
				for _, edge := range lpg.EdgeSlice(node.GetEdgesWithLabel(lpg.OutgoingEdge, ls.InstanceOfTerm.Name)) {
					schemaNodes[edge.GetTo()] = struct{}{}
				}
			}
		}
		newTarget := lpg.NewGraph()
		ls.CopyGraph(newTarget, target, func(n *lpg.Node) bool {
			if !ls.IsAttributeNode(n) {
				return true
			}
			if _, ok := schemaNodes[n]; ok {
				return true
			}
			return false
		},
			func(edge *lpg.Edge) bool {
				return !ls.IsAttributeTreeEdge(edge)
			})
		target = newTarget
	}
	return cmdutil.WriteGraph(cmd, target, outFormat, wr)
}
