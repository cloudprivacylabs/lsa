package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/layers/cmd/pipeline"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
	"github.com/spf13/cobra"
	"golang.org/x/text/encoding"
)

func init() {
	rootCmd.AddCommand(pipelineCmd)
	pipelineCmd.Flags().String("file", "", "Pipeline build file")
	pipelineCmd.Flags().String("initialGraph", "", "Load this graph and ingest data onto it")

	pipeline.RegisterPipelineStep("writeGraph", func() pipeline.Step { return &pipeline.WriteGraphStep{} })
	pipeline.RegisterPipelineStep("fork", func() pipeline.Step { return &pipeline.ForkStep{} })

	oldHelp := pipelineCmd.HelpFunc()
	pipelineCmd.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		oldHelp(cmd, []string{})
		type helper interface{ Help() }
		for _, x := range pipeline.ListPipelineSteps() {
			w := x()
			if h, ok := w.(helper); ok {
				fmt.Println("------------------------")
				h.Help()
			}
		}
	})
}

func runPipeline(steps []pipeline.Step, initialGraph string, inputs []string) (*pipeline.PipelineContext, error) {
	var g graph.Graph
	var err error
	if initialGraph != "" {
		g, err = cmdutil.ReadJSONGraph([]string{initialGraph}, nil)
		if err != nil {
			return nil, err
		}
	} else {
		g = ls.NewDocumentGraph()
	}
	pipeline := &pipeline.PipelineContext{
		Graph:   g,
		Context: getContext(),
		NextInput: func() (io.ReadCloser, error) {
			enc := encoding.Nop
			if len(inputs) == 0 {
				inp, err := cmdutil.StreamFileOrStdin(nil, enc)
				return io.NopCloser(inp), err
			}
			return io.NopCloser(strings.NewReader(inputs[0])), nil
		},
		Steps:       steps,
		CurrentStep: -1,
		Properties:  make(map[string]interface{}),
	}
	pipeline.GraphOwner = pipeline
	return pipeline, pipeline.Next()
}

var pipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "run pipeline",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		file, err := cmd.Flags().GetString("file")
		if err != nil {
			failErr(err)
		}
		steps, err := pipeline.ReadPipeline(file)
		if err != nil {
			failErr(err)
		}
		initialGraph, _ := cmd.Flags().GetString("initialGraph")
		_, err = runPipeline(steps, initialGraph, args)
		return err
	},
}
