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

	"github.com/cloudprivacylabs/opencypher"

	"github.com/spf13/cobra"
)

type OCStep struct {
	Expr []string
}

func (OCStep) Help() {
	fmt.Println(`Run Opencypher expression(s) on the graph

operation: oc
params:
  expr: 
   - opencypherExpression
   - opencypherExpression

The expressions share the same evaluation context. That means, symbols
defined in an expression are avaliable to subsequent expressions.
The output of the operations is the modified graph, and the final result is
available in pipeline property "ocResult".
`)
}

func (oc *OCStep) Run(pipeline *PipelineContext) error {
	ctx := opencypher.NewEvalContext(pipeline.GetGraphRW())
	for _, expr := range oc.Expr {
		output, err := opencypher.ParseAndEvaluate(expr, ctx)
		if err != nil {
			return err
		}
		pipeline.Properties["ocResult"] = output
	}
	if err := pipeline.Next(); err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(ocCmd)
	rootCmd.AddCommand(ocqCmd)
	ocCmd.Flags().String("input", "json", "Input graph format (json, jsonld)")
	ocCmd.Flags().String("output", "json", "Output format, json, jsonld, or dot")
	ocCmd.Flags().String("expr", "", "Opencypher expression to run")
	ocCmd.MarkFlagRequired("expr")
	ocqCmd.Flags().String("input", "json", "Input graph format (json, jsonld)")
	ocqCmd.Flags().String("expr", "", "Opencypher expression to run")
	ocqCmd.MarkFlagRequired("expr")

	operations["oc"] = func() Step { return &OCStep{} }
}

var ocCmd = &cobra.Command{
	Use:   "oc",
	Short: "Run an opencypher expression on a graph and return the graph",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		step := &OCStep{}
		e, _ := cmd.Flags().GetString("expr")
		step.Expr = []string{e}

		p := []Step{
			NewReadGraphStep(cmd),
			step,
			NewWriteGraphStep(cmd),
		}
		_, err := runPipeline(p, "", args)
		if err != nil {
			return err
		}
		return nil
	},
}

var ocqCmd = &cobra.Command{
	Use:   "ocq",
	Short: "Run an opencypher query on a graph and return the results",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		step := &OCStep{}
		e, _ := cmd.Flags().GetString("expr")
		step.Expr = []string{e}

		p := []Step{
			NewReadGraphStep(cmd),
			step,
		}
		ctx, err := runPipeline(p, "", args)
		if err != nil {
			return err
		}
		fmt.Println(ctx.Properties["ocResult"])
		return nil
	},
}
