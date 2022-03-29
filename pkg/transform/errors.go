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

package transform

import (
	"errors"
	"fmt"
)

// ErrInvalidSchemaNodeType is returned if the schema node type cannot
// be projected (such as a reference, which cannot happen after
// compilation)
type ErrInvalidSchemaNodeType []string

func (e ErrInvalidSchemaNodeType) Error() string {
	return fmt.Sprintf("Invalid schema node type for reshaping: %v", []string(e))
}

var (
	ErrInvalidSource      = errors.New("Invalid source")
	ErrInvalidSourceValue = errors.New("Invalid source value")
	ErrSourceMustBeString = errors.New("source term value must be a string")
	ErrMultipleValues     = errors.New("Multiple values/result columns found")
)
