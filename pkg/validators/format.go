package validators

import (
	"github.com/santhosh-tekuri/jsonschema/v3"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// JsonFormatTerm validates if the value matches one of the json format implementations
var JsonFormatTerm = ls.NewTerm(ls.LS+"validation#json/format", false, false, ls.OverrideComposition, struct {
	JsonFormatValidator
}{
	JsonFormatValidator{},
})

// JsonFormatValidator checks if the input value matches a given format
type JsonFormatValidator struct{}

// ValidateValue checks if the value matches the format
func (validator JsonFormatValidator) ValidateValue(value interface{}, format string) error {
	f := jsonschema.Formats[format]
	if f == nil {
		return ls.ErrValidation{Validator: JsonFormatTerm, Msg: "Unknown format: " + format}
	}
	if !f(value) {
		return ls.ErrValidation{Validator: JsonFormatTerm, Msg: "Invalid value for " + format}
	}
	return nil
}

// Validate validates the node value if it is non-nil
func (validator JsonFormatValidator) Validate(docNode, schemaNode ls.Node) error {
	if docNode == nil {
		return nil
	}
	value := docNode.GetValue()
	if value == nil {
		return nil
	}
	return validator.ValidateValue(value, schemaNode.GetCompiledDataMap()[JsonFormatTerm].(string))
}

func (validator JsonFormatValidator) CompileTerm(_ string, value *ls.PropertyValue) (interface{}, error) {
	if !value.IsString() {
		return nil, ls.ErrValidatorCompile{Validator: JsonFormatTerm, Msg: "Invalid format value"}
	}
	if jsonschema.Formats[value.AsString()] == nil {
		return nil, ls.ErrValidatorCompile{Validator: JsonFormatTerm, Msg: "Invalid format value"}
	}
	return value.AsString(), nil
}
