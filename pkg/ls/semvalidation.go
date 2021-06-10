package ls

import (
	"fmt"
)

// A Validator is used to validate document nodes based on their
// schema. The Validate function is called with the document node that
// needs to be validated, and the associated schema node.
type Validator interface {
	Validate(DocumentNode, *SchemaNode) error
}

type nopValidator struct{}

func (nopValidator) Validate(DocumentNode, *SchemaNode) error { return nil }

// GetAttributeValidator returns a validator implementation for the given validation term
func GetAttributeValidator(term string) Validator {
	md := GetTermMetadata(term)
	if md == nil {
		return nopValidator{}
	}
	val, ok := md.(Validator)
	if ok {
		return val
	}
	return nopValidator{}
}

// ValidateDocumentNode runs the validators for the document node
func ValidateDocumentNode(node DocumentNode) error {
	// Get the schema
	schemaNode, _ := node.NextNode(InstanceOfTerm).(*SchemaNode)
	if schemaNode == nil {
		return nil
	}
	return schemaNode.Validate(node)
}

// Validate the document node based on the validators of the schema
func (node *SchemaNode) Validate(docNode DocumentNode) error {
	for key := range node.Properties {
		if err := GetAttributeValidator(key).Validate(docNode, node); err != nil {
			return err
		}
	}
	return nil
}

// ErrValidatorCompile is returned for validator compilation errors
type ErrValidatorCompile struct {
	Validator string
	NodeID    string
	Msg       string
	Err       error
}

func (e ErrValidatorCompile) Error() string {
	return fmt.Sprintf("Validator compile error for %s at %s: %s %w", e.Validator, e.NodeID, e.Msg, e.Err)
}

func (e ErrValidatorCompile) Unwrap() error { return e.Err }

// ErrValidation is used to return validator errors
type ErrValidation struct {
	Validator string
	Msg       string
	Err       error
}

func (e ErrValidation) Error() string {
	return fmt.Sprintf("Validation error: %s %s", e.Validator, e.Msg)
}

func (e ErrValidation) Unwrap() error {
	return e.Err
}

// ErrInvalidValidator is used to return validator compilation errors
type ErrInvalidValidator struct {
	Validator string
	Msg       string
	Err       error
}

func (e ErrInvalidValidator) Error() string {
	return fmt.Sprintf("Validator error: %s %s", e.Validator, e.Msg)
}

func (e ErrInvalidValidator) Unwrap() error {
	return e.Err
}
