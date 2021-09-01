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
func (validator RequiredValidator) Validate(docNode, schemaNode ls.Node) error {
	if docNode == nil {
		return nil
	}
	required := schemaNode.GetProperties()[RequiredTerm].AsStringSlice()
	if len(required) > 0 {
		names := make(map[string]struct{})
		for nodes := docNode.OutWith(ls.DataEdgeTerms.ObjectAttributes).Targets(); nodes.HasNext(); {
			node := nodes.Next().(ls.Node)
			name, _ := node.GetProperties()[ls.AttributeNameTerm]
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
func (validator RequiredValidator) CompileTerm(_ string, value *ls.PropertyValue) (interface{}, error) {
	if !value.IsStringSlice() {
		return nil, ls.ErrValidatorCompile{Validator: RequiredTerm, Msg: "Array of required attributes expected"}
	}
	return value.AsStringSlice(), nil
}
