package ls

import (
	"github.com/cloudprivacylabs/lpg"
)

func ProcessLabeledAs(graph *lpg.Graph) {
	for nodeItr := graph.GetNodes(); nodeItr.Next(); {
		node := nodeItr.Node()
		labels := node.GetLabels()
		labels.Add(AsPropertyValue(node.GetProperty(LabeledAsTerm)).AsString())
		node.SetLabels(labels)
		node.RemoveProperty(LabeledAsTerm)
	}
}
