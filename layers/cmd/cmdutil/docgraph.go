package cmdutil

import (
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
)

func NewDocumentGraph() graph.Graph {
	// add indexes defined in config
	cfg := GetConfig()
	grph := ls.NewDocumentGraph().(*graph.OCGraph)
	for _, p := range cfg.IndexedProperties {
		grph.AddNodePropertyIndex(p)
	}
	return grph
}
