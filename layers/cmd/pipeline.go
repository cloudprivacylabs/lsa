package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
	"github.com/spf13/cobra"
)

type PipelineContext struct {
	*ls.Context
	Graph       graph.Graph
	Roots       []graph.Node
	InputFiles  []string
	currentStep int
	steps       []Step
	Properties  map[string]interface{}
}

type Step interface {
	Run(*PipelineContext) error
}

type StepFunc func(*PipelineContext) error

func (f StepFunc) Run(ctx *PipelineContext) error { return f(ctx) }

var operations = make(map[string]func() Step)

type ReadGraphStep struct {
	Format string
}

func NewReadGraphStep(cmd *cobra.Command) ReadGraphStep {
	rd := ReadGraphStep{}
	rd.Format, _ = cmd.Flags().GetString("input")
	return rd
}

func (rd ReadGraphStep) Run(pipeline *PipelineContext) error {
	g, err := cmdutil.ReadGraph(pipeline.InputFiles, pipeline.Context.GetInterner(), rd.Format)
	if err != nil {
		return err
	}
	pipeline.Graph = g
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
  format: json, jsonld, dot. Json is the default
  includeSchema: If false, filter out schema nodes`)
}

func NewWriteGraphStep(cmd *cobra.Command) WriteGraphStep {
	wr := WriteGraphStep{Cmd: cmd}
	wr.Format, _ = cmd.Flags().GetString("output")
	wr.IncludeSchema, _ = cmd.Flags().GetBool("includeSchema")
	return wr
}

func (wr WriteGraphStep) Run(pipeline *PipelineContext) error {
	if len(wr.Format) == 0 {
		wr.Format = "json"
	}
	return OutputIngestedGraph(wr.Cmd, wr.Format, pipeline.Graph, os.Stdout, wr.IncludeSchema)
}

type pipelineError struct {
	wrapped error
	step    int
}

func (e pipelineError) Error() string {
	return fmt.Sprintf("Pipeline error at step %d: %v", e.step, e.wrapped)
}

func (e pipelineError) Unwrap() error {
	return e.wrapped
}

func (ctx *PipelineContext) Next() error {
	ctx.currentStep++
	if ctx.currentStep >= len(ctx.steps) {
		return nil
	}
	err := ctx.steps[ctx.currentStep].Run(ctx)
	var perr pipelineError
	if err != nil && !errors.As(err, &perr) {
		err = pipelineError{wrapped: err, step: ctx.currentStep}
	}
	ctx.currentStep--
	return err
}

func init() {
	rootCmd.AddCommand(pipelineCmd)
	pipelineCmd.Flags().String("file", "", "Pipeline build file")
	pipelineCmd.Flags().String("initialGraph", "", "Load this graph and ingest data onto it")

	operations["writeGraph"] = func() Step { return &WriteGraphStep{} }

	oldHelp := pipelineCmd.HelpFunc()
	pipelineCmd.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		oldHelp(cmd, []string{})
		type helper interface{ Help() }
		for _, x := range operations {
			w := x()
			if h, ok := w.(helper); ok {
				fmt.Println("------------------------")
				h.Help()
			}
		}
	})
}

func readPipeline(file string) ([]Step, error) {
	type stepMarshal struct {
		Operation string      `json:"operation" yaml:"operation"`
		Step      interface{} `json:"params" yaml:"params"`
	}
	var stepMarshals []stepMarshal
	err := cmdutil.ReadJSONOrYAML(file, &stepMarshals)
	if err != nil {
		return nil, err
	}
	steps := make([]Step, 0, len(stepMarshals))
	for _, stage := range stepMarshals {
		op := operations[stage.Operation]
		if op == nil {
			return nil, fmt.Errorf("Unknown pipeline operation: %s", stage.Operation)
		}
		step := op()
		if step == nil {
			return nil, fmt.Errorf("Invalid step: %s", stage.Operation)
		}
		if stage.Step != nil {
			stage.Step = cmdutil.YAMLToMap(stage.Step)
			d, err := json.Marshal(stage.Step)
			if err != nil {
				panic(err)
			}
			if err := json.Unmarshal(d, step); err != nil {
				return nil, err
			}
		}
		steps = append(steps, step)
	}
	return steps, nil
}

func runPipeline(steps []Step, initialGraph string, inputs []string) (*PipelineContext, error) {
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
	pipeline := &PipelineContext{
		Graph:       g,
		Context:     getContext(),
		InputFiles:  inputs,
		steps:       steps,
		currentStep: -1,
		Properties:  make(map[string]interface{}),
	}
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
