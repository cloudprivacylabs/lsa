package ls

import (
	"github.com/cloudprivacylabs/lpg/v2"
)

func ProcessLabeledAs(graph *lpg.Graph) {
	for nodeItr := graph.GetNodes(); nodeItr.Next(); {
		node := nodeItr.Node()
		if node.HasLabel(AttributeNodeTerm.Name) {
			labels := node.GetLabels()
			p := LabeledAsTerm.PropertyValue(node)
			if len(p) > 0 {
				labels.Add(p...)
			}
			node.SetLabels(labels)
			node.RemoveProperty(LabeledAsTerm.Name)
		}
	}
}
