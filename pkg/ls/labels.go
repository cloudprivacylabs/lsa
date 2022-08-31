package ls

import (
	"github.com/cloudprivacylabs/lpg"
)

func ProcessLabeledAs(graph *lpg.Graph) {
	for nodeItr := graph.GetNodes(); nodeItr.Next(); {
		node := nodeItr.Node()
		if node.HasLabel(AttributeNodeTerm) {
			labels := node.GetLabels()
			for _, label := range AsPropertyValue(node.GetProperty(LabeledAsTerm)).MustStringSlice() {
				labels.Add(label)
				node.RemoveProperty(LabeledAsTerm)
			}
			node.SetLabels(labels)
		}
	}
}
