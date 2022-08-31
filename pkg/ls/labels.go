package ls

import (
	"github.com/cloudprivacylabs/lpg"
)

func ProcessLabeledAs(graph *lpg.Graph) {
	for nodeItr := graph.GetNodes(); nodeItr.Next(); {
		node := nodeItr.Node()
		labels := node.GetLabels()
		if AsPropertyValue(node.GetProperty(AttributeNameTerm)).AsString() == "labeledAs" {
			labels.Add(AsPropertyValue(node.GetProperty(NodeValueTerm)).AsString())
			node.RemoveProperty(AttributeNameTerm)
		}
		node.SetLabels(labels)
	}
}
