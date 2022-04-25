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
	"fmt"
)

type ErrMultipleValues struct{}

func (e ErrMultipleValues) Error() string { return "Multiple values/result columns" }

type ErrReshape struct {
	Wrapped      error
	SchemaNodeID string
}

func (e ErrReshape) Unwrap() error { return e.Wrapped }
func (e ErrReshape) Error() string {
	return fmt.Sprintf("Reshape error at %s: %s", e.SchemaNodeID, e.Wrapped.Error())
}

func wrapReshapeError(err error, schemaNodeID string) error {
	if err == nil {
		return nil
	}
	if r, ok := err.(ErrReshape); ok {
		if r.SchemaNodeID == schemaNodeID {
			return r
		}
		return ErrReshape{
			Wrapped:      r.Wrapped,
			SchemaNodeID: schemaNodeID + " . " + r.SchemaNodeID,
		}
	}
	return ErrReshape{
		Wrapped:      err,
		SchemaNodeID: schemaNodeID,
	}
}
