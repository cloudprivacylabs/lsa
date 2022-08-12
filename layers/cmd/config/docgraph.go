package config

import (
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
)

func NewDocumentGraph() graph.Graph {
	grph := ls.NewDocumentGraph()
	// add indexes defined in config
	cfg := GetConfig()
	for nodeItr := grph.GetNodes(); nodeItr.Next(); {
		node := nodeItr.Node()
		node.ForEachProperty(func(s string, i interface{}) bool {
			if _, exists := cfg.IndexMappings[s]; exists {
				node.SetProperty(s, cfg.IndexMappings[s])
			}
			return true
		})
	}
	return grph
}
