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
	"strings"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// JoinWithTerm specifies how to join multiple values. JoinWith: " " will join them using a space.
var JoinWithTerm = ls.NewTerm(TRANSFORM, "joinWith", false, false, ls.OverrideComposition, nil)

func JoinValues(values []string, delimiter string) string {
	if len(values) == 0 {
		return ""
	}
	if len(values) == 1 {
		return values[0]
	}
	return strings.Join(values, delimiter)
}
