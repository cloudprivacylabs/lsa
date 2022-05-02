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

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/transform"
	"github.com/spf13/cobra"
)

type ReshapeStep struct {
	BaseIngestParams
	layer       *ls.Layer
	initialized bool
}

func (ReshapeStep) Help() {
	fmt.Println(`Reshape graph to fit into another schema

operation: reshape
params:
  # Specify the schema the input graph should be reshaped to`)
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
		}
		rs.initialized = true
	}
	reshaper := transform.Reshaper{}
	reshaper.TargetSchema = rs.layer
	reshaper.Builder = ls.NewGraphBuilder(nil, ls.GraphBuilderOptions{
		EmbedSchemaNodes: true,
	})
	err = reshaper.Reshape(pipeline.Context, pipeline.Graph)
	if err != nil {
		return err
	}
	pipeline.Graph = reshaper.Builder.GetGraph()
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

	operations["reshape"] = func() Step { return &ReshapeStep{} }
}

var reshapeCmd = &cobra.Command{
	Use:   "reshape",
	Short: "Reshape a graph using a target schema",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		step := &ReshapeStep{}
		step.fromCmd(cmd)
		p := []Step{
			NewReadGraphStep(cmd),
			step,
			NewWriteGraphStep(cmd),
		}
		_, err := runPipeline(p, "", args)
		return err
	},
}
