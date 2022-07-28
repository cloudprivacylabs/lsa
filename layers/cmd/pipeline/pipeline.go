package pipeline

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
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

var Operations = make(map[string]func() Step)

type StepMarshal struct {
	Operation string      `json:"operation" yaml:"operation"`
	Step      interface{} `json:"params" yaml:"params"`
}

func unmarshalStep(operation string, stepData interface{}) (Step, error) {
	op := Operations[operation]
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

func UnmarshalPipeline(stepMarshals []StepMarshal) ([]Step, error) {
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
	var steps []StepMarshal
	err := json.Unmarshal(in, &steps)
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
