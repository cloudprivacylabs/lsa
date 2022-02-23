package types

import (
	"fmt"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

type ErrNotAStringValue struct {
	NodeID string
}

func (e ErrNotAStringValue) Error() string {
	return fmt.Sprintf("%s: Not a string value", e.NodeID)
}

// getStringValue tries to get a string value from the node. If the
// node value is nil, returns "", false, nil
func getStringNodeValue(node graph.Node) (string, bool, error) {
	v, ok := ls.GetRawNodeValue(node)
	if !ok {
		return "", false, nil
	}
	return v, true, nil
}
