package pipeline

import (
	"fmt"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func init() {
	RegisterPipelineStep("unique", func() Step {
		return &UniqueStep{}
	})
}

type UniqueStep struct {
	Labels []string `json:"labels" yaml:"labels"`
}

func (UniqueStep) Name() string { return "unique" }

func (UniqueStep) Help() {
	fmt.Println(`Remove nodes with given labels that share the same id

operation: unique
params:
  labels:
    - lbl1
    - lbl2
`)
}

func (rd UniqueStep) Run(pipeline *PipelineContext) error {
	for _, lbl := range rd.Labels {
		ls.ConsolidateDuplicateEntities(pipeline.Graph, lbl)
	}
	if err := pipeline.Next(); err != nil {
		return err
	}

	return nil
}
