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
	"github.com/cloudprivacylabs/opencypher"

	//	"github.com/cloudprivacylabs/opencypher/graph"
	"github.com/spf13/cobra"
)

type OCpipeline struct {
	Step
	Expr string
}

func (oc OCpipeline) Run(pipeline *PipelineContext) error {
	ctx := opencypher.NewEvalContext(pipeline.Graph)
	output, err := opencypher.ParseAndEvaluate(oc.Expr, ctx)
	if err != nil {
		failErr(err)
	}
	fmt.Println(output)
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

	operations["oc"] = func() Step { return &OCpipeline{} }
}

var ocCmd = &cobra.Command{
	Use:   "oc",
	Short: "Run an opencypher expression on a graph",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		interner := ls.NewInterner()
		input, _ := cmd.Flags().GetString("input")
		g, err := cmdutil.ReadGraph(args, interner, input)
		if err != nil {
			failErr(err)
		}
		expr, _ := cmd.Flags().GetString("expr")
		ctx := opencypher.NewEvalContext(g)
		output, err := opencypher.ParseAndEvaluate(expr, ctx)
		if err != nil {
			failErr(err)
		}

		fmt.Println(output)
	},
}
