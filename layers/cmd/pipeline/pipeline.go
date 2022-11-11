package pipeline

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/cloudprivacylabs/lpg"
	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"gopkg.in/yaml.v2"
)

type PipelineContext struct {
	*ls.Context
	Graph       *lpg.Graph
	Roots       []*lpg.Node
	NextInput   func() (PipelineEntryInfo, io.ReadCloser, error)
	CurrentStep int
	Steps       []Step
	Properties  map[string]interface{}
	GraphOwner  *PipelineContext
	ErrorLogger func(*PipelineContext, error)
	EntryLogger func(*PipelineContext, map[string]interface{})
	// If any goroutines are started with this pipeline, waitgroup is used to wait for them
	Wait sync.WaitGroup
	Err  chan error
}

type PipelineEntryInfo interface {
	GetName() string
}

type DefaultPipelineEntryInfo struct {
	Name string
}

func (pinfo DefaultPipelineEntryInfo) GetName() string {
	return pinfo.Name
}

type Step interface {
	Run(*PipelineContext) error
}

type Pipeline []Step

// InputsFromFiles returns a function that will read input files sequentially for pipeline Run inputs func
func InputsFromFiles(files []string) func() (PipelineEntryInfo, io.ReadCloser, error) {
	i := 0
	return func() (PipelineEntryInfo, io.ReadCloser, error) {
		if len(files) == 0 {
			if i > 0 {
				return nil, nil, nil
			}
			i++
			inp, err := cmdutil.StreamFileOrStdin(nil)
			if err != nil {
				return DefaultPipelineEntryInfo{}, nil, err
			}
			return DefaultPipelineEntryInfo{}, io.NopCloser(inp), nil
		}
		if i >= len(files) {
			return nil, nil, nil
		}
		stream, err := cmdutil.StreamFileOrStdin([]string{files[i]})
		i++
		if err != nil {
			return nil, nil, err
		}
		return DefaultPipelineEntryInfo{Name: files[i-1]}, io.NopCloser(stream), nil
	}
}

// create new pipeline context with an optional initial graph and inputs func
func NewContext(lsctx *ls.Context, pipeline Pipeline, initialGraph *lpg.Graph, inputs func() (PipelineEntryInfo, io.ReadCloser, error)) *PipelineContext {
	var g *lpg.Graph
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
		ErrorLogger: func(pctx *PipelineContext, err error) {
			fmt.Println(fmt.Errorf("pipeline error: %w", err))
		},
		EntryLogger: func(_ *PipelineContext, _ map[string]interface{}) {},
	}
	ctx.GraphOwner = ctx
	return ctx
}

// Run pipeline
func Run(ctx *PipelineContext) error {
	ctx.Err = make(chan error)
	errors := make([]error, 0)
	go func() {
		for e := range ctx.Err {
			errors = append(errors, e)
		}
	}()
	err := ctx.Next()
	ctx.Wait.Wait()
	close(ctx.Err)
	if err == nil && len(errors) == 0 {
		return nil
	}
	if len(errors) > 0 && err != nil {
		return fmt.Errorf("%v %w", errors, err)
	}
	if len(errors) == 0 {
		return err
	}
	return nil
}

func (ctx *PipelineContext) SetGraph(g *lpg.Graph) *PipelineContext {
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

// HasNext returns if there is a next step in the pipeline
func (ctx *PipelineContext) HasNext() bool {
	return ctx.CurrentStep+1 < len(ctx.Steps)
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
