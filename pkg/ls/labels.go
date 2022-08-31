package ls

import (
	"github.com/cloudprivacylabs/lpg"
)

func ProcessLabeledAs(graph *lpg.Graph) {
	for nodeItr := graph.GetNodes(); nodeItr.Next(); {
		node := nodeItr.Node()
		if node.HasLabel(AttributeNodeTerm) {
			labels := node.GetLabels()
			labels.Add(AsPropertyValue(node.GetProperty(LabeledAsTerm)).MustStringSlice()...)
			node.SetLabels(labels)
			node.RemoveProperty(LabeledAsTerm)
		}
	}
}
