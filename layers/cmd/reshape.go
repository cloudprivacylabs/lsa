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

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/transform"
	"github.com/spf13/cobra"
)

type ReshapeStep struct {
	BaseIngestParams
	ScriptFile  string `json:"scriptFile" yaml:"scriptFile"`
	layer       *ls.Layer
	script      *transform.TransformScript
	initialized bool
}

func (ReshapeStep) Help() {
	fmt.Println(`Reshape graph to fit into another schema

operation: reshape
params:
  # Specify the schema the input graph should be reshaped to
  scriptFile: transformation script file

The scriptFile contains transformation scripts. This is a YAML or JSON file of the following format:

map:
  - source: sourceSchemaNodeId
    target: targetSchemaNodeId
  - source: sourceSchemaNodeId
    target: targetSchemaNodeId
  ...
targetSchemaNodes:
  targetSchemaNodeId:
    term: value
    term:
     - value
     - value
  targetSchemaNodeId:
    term: value
    term:
     - value
     - value

The map section deals with field mappings. Nodes that are instances of
sourceSchemaNodeId are mapped to the instances of targetSchemaNodeIds.

The targetSchemaNodes section gives openCypher expressions to reshape
the source graph. Each targetSchemaNodeId specifies the rules to build
an instance of the target schema node. The terms are the terms in
https://lschema.org/transform namespace:
  
  https://lschema.org/transform/evaluate: Gives an array of openCypher
  expressions that will be evaluated and the named results will be exported 
  for later use.

  https://lschema.org/transform/valueExpr: Gives an array of openCypher
  expressions that will be evaluated to build the node value. The 
  first expression with a nonempty result will be used.

  https://lschema.org/transform/multi: If true, multiple source values are allowed.

  https://lschema.org/transform/joinWith: If there are multiple source values, how to 
  join them. Can be " ", or ", ", etc.

  https://lschema.org/transform/mapProperty: Gives the names of a property in the 
  source graph that are under the map context. If the map context is not set, all
  the nodes of the source graph are considered. The value of this property gives
  the target schema node Id.

  https://lschema.org/transform/mapContext: This is an openCypher expression that
  evaluates into a node. Any mapProperty lookups will be performed under this node.`)
	fmt.Println(baseIngestParamsHelp)
}

func (rs *ReshapeStep) Run(pipeline *PipelineContext) error {
	var err error
	if !rs.initialized {
		if rs.IsEmptySchema() {
			rs.layer, _ = pipeline.Properties["layer"].(*ls.Layer)
		} else {
			rs.layer, err = LoadSchemaFromFileOrRepo(pipeline.Context, rs.CompiledSchema, rs.Repo, rs.Schema, rs.Type, rs.Bundle)
			if err != nil {
				return err
			}
			pipeline.Properties["layer"] = rs.layer
		}
		if len(rs.ScriptFile) > 0 {
			if err := cmdutil.ReadJSONOrYAML(rs.ScriptFile, &rs.script); err != nil {
				return err
			}
			if err := rs.script.Compile(pipeline.Context); err != nil {
				return err
			}
		}
		rs.initialized = true
	}
	reshaper := transform.Reshaper{
		Script: rs.script,
	}
	reshaper.TargetSchema = rs.layer
	reshaper.Builder = ls.NewGraphBuilder(nil, ls.GraphBuilderOptions{
		EmbedSchemaNodes: true,
	})
	err = reshaper.Reshape(pipeline.Context, pipeline.GetGraphRO())
	if err != nil {
		return err
	}
	pipeline.SetGraph(reshaper.Builder.GetGraph())
	if err := pipeline.Next(); err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(reshapeCmd)
	addSchemaFlags(reshapeCmd.Flags())
	reshapeCmd.Flags().String("compiledschema", "", "Use the given compiled schema")
	reshapeCmd.Flags().String("input", "json", "Input graph format (json, jsonld)")
	reshapeCmd.PersistentFlags().String("output", "json", "Output format, json, jsonld, or dot")
	reshapeCmd.Flags().String("script", "", "Transformation script file")

	operations["reshape"] = func() Step { return &ReshapeStep{} }
}

var reshapeCmd = &cobra.Command{
	Use:   "reshape",
	Short: "Reshape a graph using a target schema",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		step := &ReshapeStep{}
		step.fromCmd(cmd)
		step.ScriptFile, _ = cmd.Flags().GetString("script")
		p := []Step{
			NewReadGraphStep(cmd),
			step,
			NewWriteGraphStep(cmd),
		}
		_, err := runPipeline(p, "", args)
		return err
	},
}
