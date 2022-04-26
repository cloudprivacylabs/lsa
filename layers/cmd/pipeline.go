package cmd

import (
	"io"
	"path/filepath"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
	"github.com/spf13/cobra"
)

func init() {
	ingestCmd.AddCommand(pipelineCmd)
	pipelineCmd.Flags().String("file", "", "Pipeline build file")
	pipelineCmd.Flags().String("inputFile", "", "User provided input file")
}

var pipelineCmd = &cobra.Command{
	Use:   "pipeline",
	Short: "build pipeline JSON",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputFile, err := cmd.Flags().GetString("inputFile")
		if err != nil {
			failErr(err)
		}
		file, err := cmd.Flags().GetString("file")
		if err != nil {
			failErr(err)
		}
		format := filepath.Ext(file)
		switch format {
		case "csv":
		case "json":
		case "xml":

		}
	},
}

type Step interface {
	Run(*PipelineContext)
}

type PipelineContext struct {
	Context *ls.Context
	Graph   graph.Graph
	Roots   []graph.Node
	Input   io.Reader
}
