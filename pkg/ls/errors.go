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
	"errors"
	"fmt"
)

// ErrInvalidInput is used for errors due to incorrect values,
// unexpected syntax, etc.
type ErrInvalidInput struct {
	ID  string
	Msg string
}

func (e ErrInvalidInput) Error() string {
	if len(e.Msg) > 0 {
		return fmt.Sprintf("Invalid input: %s - %s", e.ID, e.Msg)
	}
	return fmt.Sprintf("Invalid input: %s", e.ID)
}

// MakeErrInvalidInput creates an ErrInvalidInput error. If there is
// only one argument, it is used as the ID field of the error. If
// there are two, then the first is used as the ID, and the second as
// msg. Other arguments are ignored.
func MakeErrInvalidInput(id ...string) error {
	ret := ErrInvalidInput{}
	if len(id) > 0 {
		ret.ID = id[0]
	}
	if len(id) > 1 {
		ret.Msg = id[1]
	}
	return ret
}

// ErrDuplicateAttributeID is used to denote a duplicate attribute in
// a schema
type ErrDuplicateAttributeID string

func (e ErrDuplicateAttributeID) Error() string {
	return fmt.Sprintf("Duplicate attribute id: %s", string(e))
}

// ErrMultipleTypes denotes multiple incompatible types declared for
// an attribute
type ErrMultipleTypes string

func (e ErrMultipleTypes) Error() string {
	return fmt.Sprintf("Multiple types declared for attribute: %s", string(e))
}

// ErrDuplicate is used for duplicate errors
type ErrDuplicate string

func (e ErrDuplicate) Error() string {
	return fmt.Sprintf("Duplicate: %s", string(e))
}

// ErrNotFound is used for all not-found errors.
type ErrNotFound string

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("Not found: %s", string(e))
}

// Error declarations for marshaling and composition
var (
	ErrInvalidJsonLdGraph          = errors.New("Invalid JsonLd graph")
	ErrInvalidJsonGraph            = errors.New("Invalid JSON graph")
	ErrUnexpectedEOF               = errors.New("Unexpected EOF")
	ErrAttributeWithoutID          = errors.New("Attribute without id")
	ErrNotALayer                   = errors.New("Not a layer")
	ErrCompositionSourceNotOverlay = errors.New("Composition source is not an overlay")
	ErrIncompatibleComposition     = errors.New("Incompatible composition of layers")

	ErrInvalidComposition = errors.New("Invalid composition")
)

// ErrTerm is used to denote a term operation error
type ErrTerm struct {
	Term string
	Err  error
}

func (e ErrTerm) Error() string {
	return fmt.Sprintf("Term error '%s': %v", e.Term, e.Err)
}

func (e ErrTerm) Unwrap() error {
	return e.Err
}

// ErrMultipleParentNodes is used to denote multiple parentnodes detected during data ingestion
type ErrMultipleParentNodes struct {
	Of string
}

func (e ErrMultipleParentNodes) Error() string { return "Multiple parent nodes for:" + e.Of }

// ErrNoParentNode is used to denote no parent nodes for an ingested
// node
type ErrNoParentNode struct {
	Of string
}

func (e ErrNoParentNode) Error() string { return "No parent node for:" + e.Of }

type ErrSchemaValidation struct {
	Msg  string
	Path NodePath
}

type ErrCannotDetermineEdgeLabel struct {
	Msg          string
	Path         NodePath
	SchemaNodeID string
}

type ErrCannotDeterminePropertyName struct {
	Path         NodePath
	SchemaNodeID string
}

func (e ErrCannotDeterminePropertyName) Error() string {
	return fmt.Sprintf("Cannot determine property name %s: %s", e.SchemaNodeID, e.Path.String())
}

type ErrCannotFindAncestor struct {
	Path         NodePath
	SchemaNodeID string
}

func (e ErrCannotFindAncestor) Error() string {
	return fmt.Sprintf("Cannot find ancestor %s: %s", e.SchemaNodeID, e.Path.String())
}

type ErrInvalidEntityID struct {
	Path NodePath
}

func (e ErrInvalidEntityID) Error() string {
	return "Invalid entity ID: " + e.Path.String()
}

func (e ErrSchemaValidation) Error() string {
	ret := "Schema validation error: " + e.Msg
	if e.Path != nil {
		ret += " path:" + e.Path.String()
	}
	return ret
}

func (e ErrCannotDetermineEdgeLabel) Error() string {
	ret := fmt.Sprintf("Cannot determine edge label %s: %s", e.SchemaNodeID, e.Msg)
	if e.Path != nil {
		ret += " path:" + e.Path.String()
	}
	return ret
}

type ErrInvalidSchema string

func (e ErrInvalidSchema) Error() string { return "Invalid schema: " + string(e) }

type ErrDataIngestion struct {
	Key string
	Err error
}

func (e ErrDataIngestion) Error() string {
	return fmt.Sprintf("Data ingestion error: Key: %s - %s", e.Key, e.Err)
}

func (e ErrDataIngestion) Unwrap() error { return e.Err }
