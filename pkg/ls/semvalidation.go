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

	"github.com/cloudprivacylabs/lpg/v2"
)

// A NodeValidator is used to validate document nodes based on their
// schema. The ValidateNode function is called with the document node that
// needs to be validated, and the associated schema node.
type NodeValidator interface {
	ValidateNode(docNode, layerNode *lpg.Node) error
}

// A ValueValidator is used to validate a value based on a schema. The
// ValidateValue function is called with the value that needs to be
// validated, and the associated schema node. If the node does not
// exist, the value is nil
type ValueValidator interface {
	ValidateValue(value *string, layerNode *lpg.Node) error
}

type nopValidator struct{}

func (nopValidator) ValidateNode(_, _ *lpg.Node) error      { return nil }
func (nopValidator) ValidateValue(*string, *lpg.Node) error { return nil }

// GetAttributeValidator returns a validator implementation for the given validation term
func GetAttributeValidator(term string) (NodeValidator, ValueValidator) {
	md := GetTermMetadata(term)
	if md == nil {
		return nopValidator{}, nopValidator{}
	}
	nval, _ := md.(NodeValidator)
	vval, _ := md.(ValueValidator)
	if nval == nil {
		nval = nopValidator{}
	}
	if vval == nil {
		vval = nopValidator{}
	}
	return nval, vval
}

// ValidateDocumentNode runs the validators for the document node
func ValidateDocumentNode(node *lpg.Node) error {
	// Get the schema
	var schemaNode *lpg.Node
	schemaNodes := lpg.NextNodesWith(node, InstanceOfTerm)
	if len(schemaNodes) == 1 {
		schemaNode = schemaNodes[0]
	}
	return ValidateDocumentNodeBySchema(node, schemaNode)
}

// ValidateDocumentNodeBySchema runs the validators for the document node
func ValidateDocumentNodeBySchema(node, schemaNode *lpg.Node) error {
	if schemaNode == nil {
		return nil
	}
	var err error
	var nodeValue *string
	if node != nil {
		v, _ := GetRawNodeValue(node)
		nodeValue = &v
	}
	schemaNode.ForEachProperty(func(key string, value interface{}) bool {
		nval, vval := GetAttributeValidator(key)
		if err = nval.ValidateNode(node, schemaNode); err != nil {
			return false
		}
		if err = vval.ValidateValue(nodeValue, schemaNode); err != nil {
			return false
		}
		return true
	})
	return err
}

// ValidateValueBySchema runs the validators for the value
func ValidateValueBySchema(v *string, schemaNode *lpg.Node) error {
	if schemaNode == nil {
		return nil
	}
	var err error
	schemaNode.ForEachProperty(func(key string, value interface{}) bool {
		_, vval := GetAttributeValidator(key)
		if err = vval.ValidateValue(v, schemaNode); err != nil {
			return false
		}
		return true
	})
	return err
}

// ErrValidatorCompile is returned for validator compilation errors
type ErrValidatorCompile struct {
	Validator string
	Object    interface{}
	Msg       string
	Err       error
}

func (e ErrValidatorCompile) Error() string {
	return fmt.Sprintf("Validator compile error for %s at %s: %s %s", e.Validator, e.Object, e.Msg, e.Err)
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
