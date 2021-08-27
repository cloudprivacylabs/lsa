// Copyright 2021 Cloud Privacy Labs, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package ls

import (
	"fmt"
)

// A Validator is used to validate document nodes based on their
// schema. The Validate function is called with the document node that
// needs to be validated, and the associated schema node.
type Validator interface {
	Validate(docNode, layerNode Node) error
}

type nopValidator struct{}

func (nopValidator) Validate(docNode, layerNode Node) error { return nil }

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
func ValidateDocumentNode(node Node) error {
	// Get the schema
	schemaNode, _ := node.Next(InstanceOfTerm).(Node)
	return ValidateDocumentNodeBySchema(node, schemaNode)
}

// ValidateDocumentNodeBySchema runs the validators for the document node
func ValidateDocumentNodeBySchema(node, schemaNode Node) error {
	if schemaNode == nil {
		return nil
	}
	for key := range schemaNode.GetProperties() {
		if err := GetAttributeValidator(key).Validate(node, schemaNode); err != nil {
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
	return fmt.Sprintf("Validator compile error for %s at %s: %s %s", e.Validator, e.NodeID, e.Msg, e.Err)
}

func (e ErrValidatorCompile) Unwrap() error { return e.Err }

// ErrValidation is used to return validator errors
type ErrValidation struct {
	Validator string
	Msg       string
	Value     string
	Err       error
}

func (e ErrValidation) Error() string {
	return fmt.Sprintf("Validation error: %s %s %s", e.Validator, e.Msg, e.Value)
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
