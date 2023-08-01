package validators

import (
	"fmt"

	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// EnumTerm is used for enumeration validator
//
// Enumeration is declared as a string slice:
//
//	{
//	   @id: attrId,
//	   @type: Value,
//	   validation/enumeration: ["a","b","c"]
//	}
var EnumTerm = ls.NewTerm(ls.LS, "validation/enumeration").SetComposition(ls.OverrideComposition).SetTags(ls.ValidationTag, ls.SchemaElementTag).SetMetadata(struct {
	EnumValidator
}{
	EnumValidator{},
}).Register()

// ConstTerm is used for constant value validator
//
// Const is declared as a string value:
//
//	{
//	   @id: attrId,
//	   @type: Value,
//	   validation/const: "a"
//	}
//
// Const is syntactic sugar for enum with a single value
var ConstTerm = ls.NewTerm(ls.LS, "validation/const").SetComposition(ls.OverrideComposition).SetTags(ls.ValidationTag, ls.SchemaElementTag).SetMetadata(struct {
	EnumValidator
}{
	EnumValidator{},
}).Register

// EnumValidator checks if a value is equal to one of the given options.
type EnumValidator struct{}

// validateValue checks if the value is the same as one of the
// options.
func (validator EnumValidator) validateValue(value *string, options []string) error {
	if value != nil {
		// fmt.Println("Validator", *value, options)
		// Check for trivial match
		for _, option := range options {
			if option == *value {
				return nil
			}
		}
	}
	return ls.ErrValidation{Validator: "EnumTerm", Msg: "None of the options match", Value: fmt.Sprint(value)}
}

func (validator EnumValidator) ValidateValue(value *string, schemaNode *lpg.Node) error {
	optionspv, _ := schemaNode.GetProperty(EnumTerm.Name)
	options, ok := optionspv.(ls.PropertyValue)
	if !ok {
		return ls.ErrInvalidValidator{Validator: EnumTerm.Name, Msg: "Invalid enum options"}
	}
	if str, ok := options.Value().(string); ok {
		return validator.validateValue(value, []string{str})
	}
	slice, _ := ls.StringSliceType{}.Coerce(options.Value())
	if slice != nil {
		return validator.validateValue(value, slice.([]string))
	}
	return nil
}

// ValidateNode validates the node value if it is non-nil
func (validator EnumValidator) ValidateNode(docNode, schemaNode *lpg.Node) error {
	if docNode == nil {
		return nil
	}

	value, _ := ls.GetRawNodeValue(docNode)
	return validator.ValidateValue(&value, schemaNode)
}
