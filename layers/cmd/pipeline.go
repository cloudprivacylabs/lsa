package cmd

import (
	"fmt"

	"github.com/cloudprivacylabs/lpg"
	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/layers/cmd/pipeline"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(pipelineCmd)
	pipelineCmd.Flags().String("file", "", "Pipeline build file")
	pipelineCmd.Flags().String("initialGraph", "", "Load this graph and ingest data onto it")

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

func NewReadGraphStep(cmd *cobra.Command) pipeline.ReadGraphStep {
	rd := pipeline.ReadGraphStep{}
	rd.Format, _ = cmd.Flags().GetString("input")
	return rd
}

func runPipeline(steps []pipeline.Step, env map[string]string, initialGraph string, inputs []string) (*pipeline.PipelineContext, error) {
	var g *lpg.Graph
	var err error
	if initialGraph != "" {
		g, err = cmdutil.ReadJSONGraph([]string{initialGraph}, nil)
		if err != nil {
			return nil, err
		}
	}
	pctx := pipeline.NewContext(getContext(), env, steps, g, pipeline.InputsFromFiles(inputs))
	return pctx, pipeline.Run(pctx)
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
		// load env from .env file here
		if err = godotenv.Load(); err != nil {
			return err
		}
		env, err := godotenv.Unmarshal("KEY=value")
		if err != nil {
			return err
		}
		initialGraph, _ := cmd.Flags().GetString("initialGraph")
		_, err = runPipeline(steps, env, initialGraph, args)
		return err
	},
}
