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

type ErrDuplicateAttributeID string

func (e ErrDuplicateAttributeID) Error() string {
	return fmt.Sprintf("Duplicate attribute id: %s", string(e))
}

type ErrDuplicateNodeID string

func (e ErrDuplicateNodeID) Error() string {
	return fmt.Sprintf("Duplicate node id: %v", string(e))
}

type ErrMultipleTypes string

func (e ErrMultipleTypes) Error() string {
	return fmt.Sprintf("Multiple types declared for attribute: %s", string(e))
}

type ErrNotFound string

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("Not found: %s", string(e))
}

var ErrInvalidJsonLdGraph = errors.New("Invalid JsonLd graph")
var ErrAttributeWithoutID = errors.New("Attribute without id")
var ErrNotALayer = errors.New("Not a layer")
var ErrCompositionSourceNotOverlay = errors.New("Composition source is not an overlay")
var ErrIncompatibleComposition = errors.New("Incompatible composition of layers")

var ErrInvalidComposition = errors.New("Invalid composition")

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
