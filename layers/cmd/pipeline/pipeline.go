package pipeline

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
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
	ErrorLogger func(*PipelineContext, error) bool
}

type Step interface {
	Run(*PipelineContext) error
}

type Pipeline []Step

// InputsFromFiles returns a function that will read input files sequentially for pipeline Run inputs func
func InputsFromFiles(files []string) func() (io.ReadCloser, error) {
	i := 0
	return func() (io.ReadCloser, error) {
		if len(files) == 0 {
			if i > 0 {
				return nil, nil
			}
			i++
			inp, err := cmdutil.StreamFileOrStdin(nil)
			if err != nil {
				return nil, err
			}
			return io.NopCloser(inp), nil
		}
		if i >= len(files) {
			return nil, nil
		}
		stream, err := cmdutil.StreamFileOrStdin([]string{files[i]})
		i++
		if err != nil {
			return nil, err
		}
		return io.NopCloser(stream), nil
	}
}

// create new pipeline context with an optional initial graph and inputs func
func NewContext(lsctx *ls.Context, pipeline Pipeline, initialGraph graph.Graph, inputs func() (io.ReadCloser, error)) *PipelineContext {
	var g graph.Graph
	if initialGraph != nil {
		g = initialGraph
	} else {
		g = cmdutil.NewDocumentGraph()
	}
	ctx := &PipelineContext{
		Graph:       g,
		Context:     lsctx,
		NextInput:   inputs,
		Steps:       pipeline,
		CurrentStep: -1,
		Properties:  make(map[string]interface{}),
		ErrorLogger: func(pctx *PipelineContext, err error) bool {
			fmt.Println(fmt.Errorf("pipeline error: %w", err))
			return err != nil
		},
	}
	ctx.GraphOwner = ctx
	return ctx
}

// Run pipeline
func Run(ctx *PipelineContext) error {
	return ctx.Next()
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
