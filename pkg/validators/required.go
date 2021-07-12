package validators

import (
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// RequiredTerm validates if a required properties exist
var RequiredTerm = ls.NewTerm(ls.LS+"validation#required", false, false, ls.OverrideComposition, struct {
	RequiredValidator
}{
	RequiredValidator{},
})

// RequiredValidator validates if a required value exists
type RequiredValidator struct{}

// Validate checks if value is nil. If value is nil and it is required, returns an error
func (validator RequiredValidator) Validate(docNode ls.DocumentNode, schemaNode ls.LayerNode) error {
	if docNode == nil {
		return nil
	}
	required := schemaNode.GetPropertyMap()[RequiredTerm].AsStringSlice()
	if len(required) > 0 {
		names := make(map[string]struct{})
		for nodes := docNode.AllOutgoingEdgesWithLabel(ls.DataEdgeTerms.ObjectAttributes).Targets(); nodes.HasNext(); {
			node := nodes.Next().(ls.DocumentNode)
			name, _ := node.GetProperty(ls.AttributeNameTerm)
			if name.IsString() {
				names[name.AsString()] = struct{}{}
			}
		}
		for _, str := range required {
			if _, ok := names[str]; !ok {
				return ls.ErrValidation{Validator: RequiredTerm, Msg: "Missing required attribute: %s" + str}
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
