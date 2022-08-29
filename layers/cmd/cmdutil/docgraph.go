package cmdutil

import (
	"github.com/cloudprivacylabs/lpg"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func NewDocumentGraph() *lpg.Graph {
	// add indexes defined in config
	cfg := GetConfig()
	grph := ls.NewDocumentGraph()
	for _, p := range cfg.IndexedProperties {
		grph.AddNodePropertyIndex(p)
	}
	return grph
}
