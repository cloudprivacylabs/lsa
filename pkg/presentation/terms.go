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
package presentation

// IDBase is https://lschema.org/presentation/"
const IDBase = ls.LS + "presentation/"

// Presentation related terms
var (
	LabelTerm       = ls.NewTerm(IDBase+"label", false, false, ls.OverrideComposition, nil)
	ChoiceTerm      = ls.NewTerm(IDBase+"choice", false, true, ls.OverrideComposition, nil)
	ChoiceKeyTerm   = ls.NewTerm(IDBase+"choice#key", false, false, ls.OverrideComposition, nil)
	ChoiceLabelTerm = ls.NewTerm(IDBase+"choice#label", false, false, ls.OverrideComposition, nil)
	HelpTerm        = ls.NewTerm(IDBase+"help", false, false, ls.OverrideComposition, nil)
)
