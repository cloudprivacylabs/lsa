package types

import (
	"fmt"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type ErrNotAStringValue struct {
	NodeID string
}

func (e ErrNotAStringValue) Error() string {
	return fmt.Sprintf("%s: Not a string value", e.NodeID)
}

// getStringValue tries to get a string value from the node. If the
// node value is nil, returns "", false, nil
func getStringNodeValue(node ls.Node) (string, bool, error) {
	v := node.GetValue()
	if v == nil {
		return "", false, nil
	}
	str, ok := v.(string)
	if !ok {
		return "", true, ErrNotAStringValue{node.GetID()}
	}
	return str, true, nil
}
