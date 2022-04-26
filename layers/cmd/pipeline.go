package cmd

import (
	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
	"github.com/spf13/cobra"
)

type PipelineContext struct {
	Context     *ls.Context
	Graph       graph.Graph
	Roots       []graph.Node
	InputFiles  []string
	currentStep int
	steps       []Step
}

type Step interface {
	Run(*PipelineContext) error
	Next() error
}

func (ctx *PipelineContext) Next() error {
	ctx.currentStep++
	if ctx.currentStep >= len(ctx.steps) {
		return nil
	}
	err := ctx.steps[ctx.currentStep].Run(ctx)
	ctx.currentStep--
	return err
}

func init() {
	ingestCmd.AddCommand(pipelineCmd)
	pipelineCmd.Flags().String("file", "", "Pipeline build file")
	pipelineCmd.Flags().String("initialGraph", "", "Load this graph and ingest data onto it")
	// pipelineCmd.Flags().String("inputFile", "", "User provided input file")
}

var pipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "build pipeline JSON",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file, err := cmd.Flags().GetString("file")
		if err != nil {
			failErr(err)
		}
		// var steps []Step
		var stages []map[string][]Step
		err = cmdutil.ReadJSONOrYAML(file, &stages)
		if err != nil {
			failErr(err)
		}
		pipeline := &PipelineContext{Context: ls.DefaultContext()}

		for _, mStep := range stages {
			//operation := mStep["operation"]
			for _, steps := range mStep {
				initialGraph, _ := cmd.Flags().GetString("initialGraph")
				if initialGraph != "" {
					pipeline.Graph, err = cmdutil.ReadJSONGraph([]string{initialGraph}, nil)
					if err != nil {
						failErr(err)
					}
				} else {
					pipeline.Graph = ls.NewDocumentGraph()
				}
				err = cmdutil.ReadJSONFileOrStdin(args, &pipeline.InputFiles)
				if err != nil {

					failErr(err)
				}

				steps[pipeline.currentStep].Run(pipeline)
				steps[pipeline.currentStep].Next()
				// if step.Next().Error() != "" {
				// 	failErr(errors.New(step.Next().Error()))
				// }

			}
		}
	},
}
