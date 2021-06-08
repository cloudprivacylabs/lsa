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
	for _, x := range required {
	}
}

// Compile the required properties array
func (validator RequiredValidator) CompileTerm(_ string, value interface{}) (interface{}, error) {
	arr, ok := value.([]interface{})
	if !ok {
		return nil, ls.ErrValidatorCompile{Validator: RequiredTerm, Msg: "Array of required attributes expected"}
	}
	return arr, nil
}
