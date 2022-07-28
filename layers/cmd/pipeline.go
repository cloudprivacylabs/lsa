package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/layers/cmd/pipeline"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
	"github.com/spf13/cobra"
)

type ForkStep struct {
	Steps map[string]pipeline.Pipeline `json:"pipelines" yaml:"pipelines"`
}

func (ForkStep) Help() {
	fmt.Println(`Create multiple parallel pipelines.

operation: fork
params: 
  pipelines:
    pipelineName:
      -
      -
    pipelineName:
      -
      -`)
}

func (fork ForkStep) Run(ctx *pipeline.PipelineContext) error {
	for idx, pipe := range fork.Steps {
		if err := forkPipeline(pipe, ctx, idx); err != nil {
			return err
		}
	}
	return nil
}

func forkPipeline(pipe pipeline.Pipeline, ctx *pipeline.PipelineContext, name string) error {
	pctx := &pipeline.PipelineContext{
		Context:     getContext(),
		Graph:       ctx.Graph,
		Roots:       ctx.Roots,
		NextInput:   ctx.NextInput,
		Steps:       pipe,
		CurrentStep: -1,
		GraphOwner:  ctx.GraphOwner,
	}
	cpMap := make(map[string]interface{})
	for k, prop := range ctx.Properties {
		cpMap[k] = prop
	}
	pctx.Properties = cpMap
	pctx.Context.GetLogger().Debug(map[string]interface{}{"Starting new fork": name})
	err := pctx.Next()
	var perr pipeline.PipelineError
	if err != nil {
		if !errors.As(err, &perr) {
			err = pipeline.PipelineError{Wrapped: fmt.Errorf("fork: %s, %w", name, err), Step: pctx.CurrentStep}
		}
		return err
	}
	return nil
}

type StepFunc func(*pipeline.PipelineContext) error

func (f StepFunc) Run(ctx *pipeline.PipelineContext) error { return f(ctx) }

type ReadGraphStep struct {
	Format string
}

func NewReadGraphStep(cmd *cobra.Command) ReadGraphStep {
	rd := ReadGraphStep{}
	rd.Format, _ = cmd.Flags().GetString("input")
	return rd
}

func (rd ReadGraphStep) Run(pipeline *pipeline.PipelineContext) error {
	g, err := cmdutil.ReadGraph([]string{}, pipeline.Context.GetInterner(), rd.Format)
	if err != nil {
		return err
	}
	pipeline.SetGraph(g)
	return pipeline.Next()
}

type WriteGraphStep struct {
	Format        string
	IncludeSchema bool
	Cmd           *cobra.Command
}

func (WriteGraphStep) Help() {
	fmt.Println(`Write graph as a JSON file

operation: writeGraph
params:
  format: json, jsonld, dot, web. Json is the default
  includeSchema: If false, filter out schema nodes`)
}

func NewWriteGraphStep(cmd *cobra.Command) WriteGraphStep {
	wr := WriteGraphStep{Cmd: cmd}
	wr.Format, _ = cmd.Flags().GetString("output")
	wr.IncludeSchema, _ = cmd.Flags().GetBool("includeSchema")
	return wr
}

func (wr WriteGraphStep) Run(pipeline *pipeline.PipelineContext) error {
	if len(wr.Format) == 0 {
		wr.Format = "json"
	}
	grph := pipeline.GetGraphRO()
	return OutputIngestedGraph(wr.Cmd, wr.Format, grph, os.Stdout, wr.IncludeSchema)
}

func init() {
	rootCmd.AddCommand(pipelineCmd)
	pipelineCmd.Flags().String("file", "", "Pipeline build file")
	pipelineCmd.Flags().String("initialGraph", "", "Load this graph and ingest data onto it")

	pipeline.Operations["writeGraph"] = func() pipeline.Step { return &WriteGraphStep{} }
	pipeline.Operations["fork"] = func() pipeline.Step { return &ForkStep{} }

	oldHelp := pipelineCmd.HelpFunc()
	pipelineCmd.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		oldHelp(cmd, []string{})
		type helper interface{ Help() }
		for _, x := range pipeline.Operations {
			w := x()
			if h, ok := w.(helper); ok {
				fmt.Println("------------------------")
				h.Help()
			}
		}
	})
}

func readPipeline(file string) ([]pipeline.Step, error) {
	var stepMarshals []pipeline.StepMarshal
	err := cmdutil.ReadJSONOrYAML(file, &stepMarshals)
	if err != nil {
		return nil, err
	}
	return pipeline.UnmarshalPipeline(stepMarshals)
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
		steps, err := readPipeline(file)
		if err != nil {
			failErr(err)
		}
		initialGraph, _ := cmd.Flags().GetString("initialGraph")
		_, err = runPipeline(steps, initialGraph, args)
		return err
	},
}
