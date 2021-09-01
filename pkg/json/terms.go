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

const JSON = "https://json.org#"

var StringTypeTerm = ls.NewTerm(JSON+"string", false, false, ls.OverrideComposition, nil)
var NumberTypeTerm = ls.NewTerm(JSON+"number", false, false, ls.OverrideComposition, nil)
var IntegerTypeTerm = ls.NewTerm(JSON+"integer", false, false, ls.OverrideComposition, nil)
var BooleanTypeTerm = ls.NewTerm(JSON+"boolean", false, false, ls.OverrideComposition, nil)
var ObjectTypeTerm = ls.NewTerm(JSON+"object", false, false, ls.OverrideComposition, nil)
var ArrayTypeTerm = ls.NewTerm(JSON+"array", false, false, ls.OverrideComposition, nil)
