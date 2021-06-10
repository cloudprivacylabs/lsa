package validators

import (
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// RequiredTerm validates if a required properties exist
const RequiredTerm = ls.LS + "validation#required"

// RequiredValidator validates if a required value exists
type RequiredValidator struct{}

func init() {
	ls.RegisterTermMetadata(RequiredTerm, struct {
		RequiredValidator
	}{
		RequiredValidator{},
	})
}

// Validate checks if value is nil. If value is nil and it is required, returns an error
func (validator RequiredValidator) Validate(docNode ls.DocumentNode, schemaNode *ls.SchemaNode) error {
	if docNode == nil {
		return nil
	}
	required := schemaNode.Properties[RequiredTerm].([]interface{})
	if len(required) > 0 {
		names := make(map[string]struct{})
		for nodes := docNode.AllOutgoingEdgesWithLabel(ls.DataEdgeTerms.ObjectAttributes).Targets(); nodes.HasNext(); {
			node := nodes.Next().(ls.DocumentNode)
			name, _ := node.GetProperty(ls.AttributeNameTerm)
			if str, ok := name.(string); ok {
				names[str] = struct{}{}
			}
		}
		for _, x := range required {
			if str, ok := x.(string); ok {
				if _, ok := names[str]; !ok {
					return ls.ErrValidation{Validator: RequiredTerm, Msg: "Missing required attribute: %s" + str}
				}
			}
		}
	}
	return nil
}

// Compile the required properties array
func (validator RequiredValidator) CompileTerm(_ string, value interface{}) (interface{}, error) {
	arr, ok := value.([]interface{})
	if !ok {
		return nil, ls.ErrValidatorCompile{Validator: RequiredTerm, Msg: "Array of required attributes expected"}
	}
	return arr, nil
}
