package pipeline

import (
	"errors"
	"fmt"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func init() {
	RegisterPipelineStep("fork", func() Step {
		return &ForkStep{Steps: make(map[string]Pipeline)}
	})
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
		Context:     ls.DefaultContext().SetLogger(ctx.GetLogger()),
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
