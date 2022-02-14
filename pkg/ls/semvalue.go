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

	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

// ErrInconsistentTypes is returned if a node has multiple types that
// can interpret the value
type ErrInconsistentTypes struct {
	ID        string
	TypeNames []string
}

func (e ErrInconsistentTypes) Error() string {
	return fmt.Sprintf("Inconsistent types at '%s': %v", e.ID, e.TypeNames)
}

type ErrInvalidValue struct {
	ID    string
	Type  string
	Value interface{}
	Msg   string
}

func (e ErrInvalidValue) Error() string {
	return fmt.Sprintf("Invalid value at '%s': Type: %s, Value: %v, %s", e.ID, e.Type, e.Value, e.Msg)
}

// A ValueAccessor gets node values in native type, and sets node values
type ValueAccessor interface {
	GetNodeValue(graph.Node) (interface{}, error)
	SetNodeValue(graph.Node, interface{}) error
}

// GetValueAccessor returns the value accessor for the term. If the term has none, returns nil
func GetValueAccessor(term string) ValueAccessor {
	acc, _ := GetTermMetadata(term).(ValueAccessor)
	return acc
}
