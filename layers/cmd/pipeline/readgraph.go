package pipeline

import (
	"fmt"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
)

func init() {
	RegisterPipelineStep("readGraph", func() Step {
		return &ReadGraphStep{Format: "json"}
	})
}

type ReadGraphStep struct {
	Format string `json:"format" yaml:"format"`
}

func (ReadGraphStep) Name() string { return "readGraph" }

func (ReadGraphStep) Flush(*PipelineContext) error { return nil }

func (ReadGraphStep) Help() {
	fmt.Println(`Read graph
Read graph file(s)

operation: readGraph
params:
  format: json`)
}

func (rd ReadGraphStep) Run(pipeline *PipelineContext) error {
	for {
		entryInfo, stream, err := pipeline.NextInput()
		if err != nil {
			pipeline.ErrorLogger(pipeline, fmt.Errorf("Filename %s: %w", entryInfo.GetName(), err))
			return fmt.Errorf("While streaming input %v: %w", stream, err)
		}
		if stream == nil {
			return nil
		}
		pipeline.GetLogger().Debug(map[string]interface{}{"readGraph": entryInfo.GetName()})
		ch, err := cmdutil.ReadGraphFromReader(pipeline.Context, stream, pipeline.Context.GetInterner(), rd.Format)
		if err != nil {
			return err
		}
		for gs := range ch {
			if err != nil {
				pipeline.ErrorLogger(pipeline, fmt.Errorf("Filename %s: %w", entryInfo.GetName(), err))
				return fmt.Errorf("While processing %v: %w", entryInfo.GetName(), err)
			}
			pipeline.SetGraph(gs.G)
			pipeline.Set("input", entryInfo.GetName())
			if err := pipeline.Next(); err != nil {
				return fmt.Errorf("While processing %v: %w", entryInfo.GetName(), err)
			}
		}
	}
	return nil
}
