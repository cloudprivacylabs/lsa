package cmd

import (
	"fmt"
	"os"

	"github.com/cloudprivacylabs/lsa/layers/cmd/pipeline"
	"github.com/spf13/cobra"
)

type WriteGraphStep struct {
	Format        string
	IncludeSchema bool
	Cmd           *cobra.Command
}

func (WriteGraphStep) Name() string { return "writeGraph" }

func init() {
	pipeline.RegisterPipelineStep("writeGraph", func() pipeline.Step {
		return &WriteGraphStep{
			Format: "json",
		}
	})
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

func (wr WriteGraphStep) Run(pipeline *pipeline.PipelineContext) error {
	if len(wr.Format) == 0 {
		wr.Format = "json"
	}
	grph := pipeline.Graph
	return OutputIngestedGraph(wr.Cmd, wr.Format, grph, os.Stdout, wr.IncludeSchema)
}
