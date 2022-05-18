package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
	"github.com/spf13/cobra"
)

type PipelineContext struct {
	*ls.Context
	graph       graph.Graph
	roots       []graph.Node
	InputFiles  []string
	currentStep int
	steps       []Step
	Properties  map[string]interface{}
	mu          sync.RWMutex
	graphOwner  *PipelineContext
	concurrent  bool
}

type Step interface {
	Run(*PipelineContext) error
}

type Pipeline []Step

func unmarshalStep(operation string, stepData interface{}) (Step, error) {
	op := operations[operation]
	if op == nil {
		return nil, fmt.Errorf("Unknown pipeline operation: %s", operation)
	}
	step := op()
	if step == nil {
		return nil, fmt.Errorf("Invalid step: %s", operation)
	}
	stepData = cmdutil.YAMLToMap(stepData)
	d, err := json.Marshal(stepData)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(d, step); err != nil {
		return nil, err
	}
	return step, nil
}

type stepMarshal struct {
	Operation string      `json:"operation" yaml:"operation"`
	Step      interface{} `json:"params" yaml:"params"`
}

func unmarshalPipeline(stepMarshals []stepMarshal) ([]Step, error) {
	steps := make([]Step, 0, len(stepMarshals))
	for _, stage := range stepMarshals {
		step, err := unmarshalStep(stage.Operation, stage.Step)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}
	return steps, nil
}

func (p *Pipeline) UnmarshalJSON(in []byte) error {
	var steps []stepMarshal
	err := json.Unmarshal(in, &steps)
	if err != nil {
		return err
	}
	*p, err = unmarshalPipeline(steps)
	return err
}

type ForkStep struct {
	Steps map[string]Pipeline `json:"pipelines" yaml:"pipelines"`
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

func (fork ForkStep) Run(ctx *PipelineContext) error {
	var wg sync.WaitGroup
	wg.Add(len(fork.Steps))
	errs := make(map[string]error)
	for k := range fork.Steps {
		errs[k] = nil
	}
	for idx, pipe := range fork.Steps {
		if ctx.concurrent {
			concurrentPipelineStep(pipe, ctx, idx, errs, &wg)
		} else {
			sequentialPipelineStep(pipe, ctx, idx, errs)
		}
	}
	if ctx.concurrent {
		wg.Wait()
	}
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func concurrentPipelineStep(pipe Pipeline, ctx *PipelineContext, name string, errs map[string]error, wg *sync.WaitGroup) {
	go func(steps []Step, currCtx *PipelineContext, grname string) {
		defer wg.Done()
		pctx := &PipelineContext{
			graph:       currCtx.graph,
			roots:       currCtx.roots,
			Context:     getContext(),
			InputFiles:  make([]string, 0),
			steps:       steps,
			currentStep: -1,
			Properties:  make(map[string]interface{}),
			graphOwner:  currCtx.graphOwner,
			mu:          sync.RWMutex{},
		}
		pctx.Context.GetLogger().Debug(map[string]interface{}{"Starting new fork": grname})
		err := pctx.Next()
		var perr pipelineError
		if err != nil && !errors.As(err, &perr) {
			err = pipelineError{wrapped: fmt.Errorf("fork: %s, %w", grname, err), step: pctx.currentStep}
			errs[grname] = err
			return
		}
	}(pipe, ctx, name)
}

func sequentialPipelineStep(pipe Pipeline, ctx *PipelineContext, name string, errs map[string]error) {
	pctx := &PipelineContext{
		Context:     getContext(),
		graph:       ctx.graph,
		roots:       ctx.roots,
		InputFiles:  ctx.InputFiles,
		steps:       pipe,
		currentStep: -1,
		graphOwner:  ctx.graphOwner,
		mu:          sync.RWMutex{},
	}
	cpMap := make(map[string]interface{})
	for k, prop := range ctx.Properties {
		cpMap[k] = ls.ClonePropertyValueFunc(k, prop)
	}
	pctx.Properties = cpMap
	pctx.Context.GetLogger().Debug(map[string]interface{}{"Starting new fork": name})
	err := pctx.Next()
	var perr pipelineError
	if err != nil && !errors.As(err, &perr) {
		err = pipelineError{wrapped: fmt.Errorf("fork: %s, %w", name, err), step: pctx.currentStep}
		errs[name] = err
		return
	}
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

func (wr WriteGraphStep) Run(pipeline *PipelineContext) error {
	if len(wr.Format) == 0 {
		wr.Format = "json"
	}
	grph := pipeline.GetGraphRO()
	return OutputIngestedGraph(wr.Cmd, wr.Format, grph, os.Stdout, wr.IncludeSchema)
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

func (ctx *PipelineContext) GetGraphRO() graph.Graph {
	// ctx.mu.RLock()
	// defer ctx.mu.RUnlock()
	return ctx.graph
}

func (ctx *PipelineContext) GetGraphRW() graph.Graph {
	// ctx.mu.Lock()
	// defer ctx.mu.Unlock()
	if ctx != ctx.graphOwner {
		// ctx.graphOwner.mu.Lock()
		newTarget := graph.NewOCGraph()
		nodeMap := ls.CopyGraph(newTarget, ctx.graph, nil, nil)
		ctx.graph = newTarget
		for _, root := range ctx.graphOwner.roots {
			ctx.roots = append(ctx.roots, nodeMap[root])
		}
		// ctx.graphOwner.mu.Unlock()
		ctx.graphOwner = ctx
	}
	return ctx.graph
}

func (ctx *PipelineContext) SetGraph(g graph.Graph) *PipelineContext {
	// ctx.mu.Lock()
	// defer ctx.mu.Unlock()
	ctx.graph = g
	return ctx
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
	pipelineCmd.Flags().String("concurrent", "", "Run pipeline steps concurrently")

	operations["writeGraph"] = func() Step { return &WriteGraphStep{} }
	operations["fork"] = func() Step { return &ForkStep{} }

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
	var stepMarshals []stepMarshal
	err := cmdutil.ReadJSONOrYAML(file, &stepMarshals)
	if err != nil {
		return nil, err
	}
	return unmarshalPipeline(stepMarshals)
}

func runPipeline(steps []Step, initialGraph string, inputs []string, concurrent bool) (*PipelineContext, error) {
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
		graph:       g,
		Context:     getContext(),
		InputFiles:  inputs,
		steps:       steps,
		currentStep: -1,
		Properties:  make(map[string]interface{}),
		concurrent:  concurrent,
	}
	pipeline.graphOwner = pipeline
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
		concurrent, _ := cmd.Flags().GetBool("concurrent")
		initialGraph, _ := cmd.Flags().GetString("initialGraph")
		_, err = runPipeline(steps, initialGraph, args, concurrent)
		return err
	},
}
