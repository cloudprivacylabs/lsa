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

package json

import (
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// JSON namespace
const JSON = "https://json.org#"

// JSON related vocabulry
var (
	StringTypeTerm  = ls.NewTerm(JSON, "string", false, false, ls.OverrideComposition, nil)
	NumberTypeTerm  = ls.NewTerm(JSON, "number", false, false, ls.OverrideComposition, nil)
	IntegerTypeTerm = ls.NewTerm(JSON, "integer", false, false, ls.OverrideComposition, nil)
	BooleanTypeTerm = ls.NewTerm(JSON, "boolean", false, false, ls.OverrideComposition, nil)
	ObjectTypeTerm  = ls.NewTerm(JSON, "object", false, false, ls.OverrideComposition, nil)
	ArrayTypeTerm   = ls.NewTerm(JSON, "array", false, false, ls.OverrideComposition, nil)
)
