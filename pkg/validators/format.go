package validators

import (
	"github.com/santhosh-tekuri/jsonschema/v3"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// JsonFormatTerm validates if the value matches one of the json format implementations
const JsonFormatTerm = ls.LS + "validation#json/format"

// JsonFormatValidator checks if the input value matches a given format
type JsonFormatValidator struct{}

func init() {
	ls.RegisterTermMetadata(JsonFormatTerm, struct {
		JsonFormatValidator
	}{
		JsonFormatValidator{},
	})
}

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
func (validator JsonFormatValidator) Validate(docNode ls.DocumentNode, schemaNode *ls.SchemaNode) error {
	if docNode == nil {
		return nil
	}
	value := docNode.GetValue()
	if value == nil {
		return nil
	}
	return validator.ValidateValue(value, schemaNode.Compiled[JsonFormatTerm].(string))
}

func (validator JsonFormatValidator) CompileTerm(_ string, value interface{}) (interface{}, error) {
	str, ok := value.(string)
	if !ok {
		return nil, ls.ErrValidatorCompile{Validator: JsonFormatTerm, Msg: "Invalid format value"}
	}
	if jsonschema.Formats[str] == nil {
		return nil, ls.ErrValidatorCompile{Validator: JsonFormatTerm, Msg: "Invalid format value"}
	}
	return str, nil
}
