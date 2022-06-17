package pipeline

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	_ "unsafe"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type PipelineContext struct {
	*ls.Context
	Graph       graph.Graph
	Roots       []graph.Node
	NextInput   func() (io.ReadCloser, error)
	CurrentStep int
	Steps       []Step
	Properties  map[string]interface{}
	GraphOwner  *PipelineContext
}

type Step interface {
	Run(*PipelineContext) error
}

type Pipeline []Step

func (ctx *PipelineContext) GetGraphRO() graph.Graph {
	return ctx.Graph
}

func (ctx *PipelineContext) GetGraphRW() graph.Graph {
	if ctx != ctx.GraphOwner {
		newTarget := graph.NewOCGraph()
		nodeMap := ls.CopyGraph(newTarget, ctx.Graph, nil, nil)
		ctx.Graph = newTarget
		for _, root := range ctx.GraphOwner.Roots {
			ctx.Roots = append(ctx.Roots, nodeMap[root])
		}
		ctx.GraphOwner = ctx
	}
	return ctx.Graph
}

func (ctx *PipelineContext) SetGraph(g graph.Graph) *PipelineContext {
	ctx.Graph = g
	return ctx
}

func (ctx *PipelineContext) Next() error {
	ctx.CurrentStep++
	if ctx.CurrentStep >= len(ctx.Steps) {
		ctx.CurrentStep--
		return nil
	}
	err := ctx.Steps[ctx.CurrentStep].Run(ctx)
	var perr PipelineError
	if err != nil && !errors.As(err, &perr) {
		err = PipelineError{Wrapped: err, Step: ctx.CurrentStep}
	}
	ctx.CurrentStep--
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
	for idx, pipe := range fork.Steps {
		if err := forkPipeline(pipe, ctx, idx); err != nil {
			return err
		}
	}
	return nil
}

func forkPipeline(pipe Pipeline, ctx *PipelineContext, name string) error {
	pctx := &PipelineContext{
		Context:     ls.DefaultContext().SetLogger(ls.NewDefaultLogger()),
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
	var perr PipelineError
	if err != nil {
		if !errors.As(err, &perr) {
			err = PipelineError{Wrapped: fmt.Errorf("fork: %s, %w", name, err), Step: pctx.CurrentStep}
		}
		return err
	}
	return nil
}

var registeredSteps = make(map[string]func() Step)

func RegisterPipelineStep(op string, step func() Step) {
	registeredSteps[op] = step
}

func ListPipelineSteps() []func() Step {
	steps := make([]func() Step, 0, len(registeredSteps))
	for _, op := range registeredSteps {
		steps = append(steps, op)
	}
	return steps
}

type stepMarshal struct {
	Operation string      `json:"operation" yaml:"operation"`
	Step      interface{} `json:"params" yaml:"params"`
}

func unmarshalStep(operation string, stepData interface{}) (Step, error) {
	op := registeredSteps[operation]
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

func UnmarshalPipeline(stepMarshals []stepMarshal) ([]Step, error) {
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
	*p, err = UnmarshalPipeline(steps)
	return err
}

func (p *Pipeline) UnmarshalYAML(in []byte) error {
	var steps []stepMarshal
	err := yaml.Unmarshal(in, &steps)
	if err != nil {
		return err
	}
	*p, err = UnmarshalPipeline(steps)
	return err
}

type PipelineError struct {
	Wrapped error
	Step    int
}

func (e PipelineError) Error() string {
	return fmt.Sprintf("Pipeline error at step %d: %v", e.Step, e.Wrapped)
}

func (e PipelineError) Unwrap() error {
	return e.Wrapped
}

func ReadPipeline(file string) ([]Step, error) {
	var stepMarshals []stepMarshal
	err := cmdutil.ReadJSONOrYAML(file, &stepMarshals)
	if err != nil {
		return nil, err
	}
	return UnmarshalPipeline(stepMarshals)
}

type StepFunc func(*PipelineContext) error

func (f StepFunc) Run(ctx *PipelineContext) error { return f(ctx) }

type ReadGraphStep struct {
	Format string
}

func NewReadGraphStep(cmd *cobra.Command) ReadGraphStep {
	rd := ReadGraphStep{}
	rd.Format, _ = cmd.Flags().GetString("input")
	return rd
}

func (ReadGraphStep) Help() {
	fmt.Println(`Read graph
Read graph file(s)

operation: readGraph
params:`)
}

func (rd ReadGraphStep) Run(pipeline *PipelineContext) error {

	gs, err := cmdutil.StreamGraph(pipeline, nil, pipeline.Context.GetInterner(), rd.Format)
	if err != nil {
		return err
	}
	for g := range gs {
		if g.Err != nil {
			return err
		}
		pipeline.SetGraph(g.G)
		pipeline.Set("input", "stdin")
		if err := pipeline.Next(); err != nil {
			return err
		}

	}
	for {
		stream, err := pipeline.NextInput()
		if err != nil {
			return fmt.Errorf("While streaming input %v: %w", stream, err)
		}
		pipeline.GetLogger().Debug(map[string]interface{}{"readGraph": stream})
		gs, err = cmdutil.StreamGraph(pipeline, []string{}, pipeline.Context.GetInterner(), rd.Format)
		if err != nil {
			return fmt.Errorf("While reading %s: %w", stream, err)
		}
		for g := range gs {
			if g.Err != nil {
				return fmt.Errorf("While reading %v: %w", stream, err)
			}
			pipeline.SetGraph(g.G)
			pipeline.Set("input", stream)
			if err := pipeline.Next(); err != nil {
				return fmt.Errorf("While processing %v: %w", stream, err)
			}
		}
	}
	return nil
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

func OutputIngestedGraph(cmd *cobra.Command, outFormat string, target graph.Graph, wr io.Writer, includeSchema bool) error {
	if !includeSchema {
		schemaNodes := make(map[graph.Node]struct{})
		for nodes := target.GetNodes(); nodes.Next(); {
			node := nodes.Node()
			if ls.IsDocumentNode(node) {
				for _, edge := range graph.EdgeSlice(node.GetEdgesWithLabel(graph.OutgoingEdge, ls.InstanceOfTerm)) {
					schemaNodes[edge.GetTo()] = struct{}{}
				}
			}
		}
		newTarget := graph.NewOCGraph()
		ls.CopyGraph(newTarget, target, func(n graph.Node) bool {
			if !ls.IsAttributeNode(n) {
				return true
			}
			if _, ok := schemaNodes[n]; ok {
				return true
			}
			return false
		},
			func(edge graph.Edge) bool {
				return !ls.IsAttributeTreeEdge(edge)
			})
		target = newTarget
	}
	return cmdutil.WriteGraph(cmd, target, outFormat, wr)
}
