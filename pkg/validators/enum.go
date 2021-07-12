package validators

import (
	"reflect"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// EnumTerm is used for enumeration validator
var EnumTerm = ls.NewTerm(ls.LS+"validation#enumeration", false, false, ls.OverrideComposition, struct {
	EnumValidator
}{
	EnumValidator{},
})

// EnumValidator checks if a value is equal to one of the given options.
type EnumValidator struct{}

// ValidateValue checks if the value is the same as one of the
// options.
func (validator EnumValidator) ValidateValue(value interface{}, options []interface{}) error {
	// Check for trivial match
	for _, option := range options {
		if option == value {
			return nil
		}
	}
	for _, option := range options {
		if reflect.DeepEqual(option, value) {
			return nil
		}
	}
	return ls.ErrValidation{Validator: "EnumTerm", Msg: "None of the options match"}
}

// Validate validates the node value if it is non-nil
func (validator EnumValidator) Validate(docNode ls.DocumentNode, schemaNode ls.LayerNode) error {
	if docNode == nil {
		return nil
	}
	value := docNode.GetValue()
	if value == nil {
		return nil
	}
	options := schemaNode.GetPropertyMap()[EnumTerm]
	if options == nil {
		return ls.ErrInvalidValidator{Validator: EnumTerm, Msg: "Invalid enum options"}
	}
	if options.IsString() {
		return validator.ValidateValue(value, []interface{}{options.AsString()})
	}
	return validator.ValidateValue(value, options.AsInterfaceSlice())
}
