package pipeline

import (
	"fmt"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
)

func init() {
	RegisterPipelineStep("readGraph", func() Step {
		return &ReadGraphStep{}
	})
}

type ReadGraphStep struct {
	Format string
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
