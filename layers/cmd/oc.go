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
	Expr string
}

func (OCStep) Help() {
	fmt.Println(`Run Opencypher expression on the graph

operation: oc
params:
  expr: opencypherExpression`)
}

func (oc *OCStep) Run(pipeline *PipelineContext) error {
	ctx := opencypher.NewEvalContext(pipeline.GetGraphRW())
	output, err := opencypher.ParseAndEvaluate(oc.Expr, ctx)
	if err != nil {
		return err
	}
	pipeline.Properties["ocResult"] = output
	if err := pipeline.Next(); err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(ocCmd)
	ocCmd.Flags().String("input", "json", "Input graph format (json, jsonld)")
	ocCmd.Flags().String("expr", "", "Opencypher expression to run")
	ocCmd.MarkFlagRequired("expr")

	operations["oc"] = func() Step { return &OCStep{} }
}

var ocCmd = &cobra.Command{
	Use:   "oc",
	Short: "Run an opencypher expression on a graph",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		step := &OCStep{}
		step.Expr, _ = cmd.Flags().GetString("expr")

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
