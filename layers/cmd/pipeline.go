package cmd

import (
	"io"

	"github.com/cloudprivacylabs/opencypher/graph"
)

type Step interface {
	Run(*PipelineContext, string) graph.Graph
}

type PipelineContext struct {
	Graph graph.Graph
	Roots []graph.Node
	Input io.Reader
}
