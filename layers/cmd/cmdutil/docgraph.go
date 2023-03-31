package cmdutil

import (
	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func NewDocumentGraph() *lpg.Graph {
	// add indexes defined in config
	cfg := GetConfig()
	grph := ls.NewDocumentGraph()
	for _, p := range cfg.IndexedProperties {
		grph.AddNodePropertyIndex(p, lpg.BtreeIndex)
	}
	return grph
}
