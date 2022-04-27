package cmd

import (
	"encoding/json"

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

var operations = make(map[string]func() Step)

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
	rootCmd.AddCommand(pipelineCmd)
	pipelineCmd.Flags().String("file", "", "Pipeline build file")
	pipelineCmd.Flags().String("initialGraph", "", "Load this graph and ingest data onto it")
}

var pipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "run pipeline",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file, err := cmd.Flags().GetString("file")
		if err != nil {
			failErr(err)
		}
		type stepMarshal struct {
			Operation string          `json:"operation" yaml:"operation"`
			Step      json.RawMessage `json:"params" yaml:"params"`
		}
		var stepMarshals []stepMarshal
		err = cmdutil.ReadJSONOrYAML(file, &stepMarshals)
		if err != nil {
			failErr(err)
		}
		pipeline := &PipelineContext{Context: ls.DefaultContext(), InputFiles: make([]string, 0)}
		for _, stage := range stepMarshals {
			step := operations[stage.Operation]()
			if err := json.Unmarshal([]byte(stage.Step), &step); err != nil {
				failErr(err)
			}
			initialGraph, _ := cmd.Flags().GetString("initialGraph")
			if initialGraph != "" {
				pipeline.Graph, err = cmdutil.ReadJSONGraph([]string{initialGraph}, nil)
				if err != nil {
					failErr(err)
				}
			}
			if len(args) > 0 {
				for _, arg := range args {
					pipeline.InputFiles = append(pipeline.InputFiles, arg)
				}
			}
			if err := step.Run(pipeline); err != nil {
				failErr(err)
			}
			if err := step.Next(); err != nil {
				failErr(err)
			}
		}
	},
}
