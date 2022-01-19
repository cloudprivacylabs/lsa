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

// Parser is a type-specific string-to-native parser
type Parser interface {
	// ParseValue parses a node value, optionally by looking at the
	// other annotations at the node, or at the schema nodes attached to
	// the node
	ParseValue(ls.Node) (interface{}, error)
}

type Formatter interface {
	// Set the value of target node using the Go native value, but based
	// on the annotations of the target node
	SetValue(ls.Node, interface{}) error
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
