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
	"strings"

	"github.com/cloudprivacylabs/lsa/pkg/opencypher"
)

const (
	JoinMethodJoin  = "join"
	JoinMethodError = "error"
)

var ErrJoinFailure = errors.New("Join failure")

func JoinValues(values []opencypher.Value, method, delimiter string) (string, error) {
	strs := make([]string, 0, len(values))
	for _, n := range values {
		strs = append(strs, n.String())
	}
	if len(strs) > 1 && method == JoinMethodError {
		return "", ErrJoinFailure
	}
	if method == JoinMethodJoin {
		return strings.Join(strs, delimiter), nil
	}
	return "", ErrJoinFailure
}
