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

package types

import (
	"testing"
)

var booleanTests = []getSetTestCase{
	{
		srcTypes:      []string{JSONBooleanTerm},
		srcValue:      "true",
		targetTypes:   []string{JSONBooleanTerm},
		expectedValue: "true",
	},
	{
		srcTypes:      []string{JSONBooleanTerm},
		srcValue:      "false",
		targetTypes:   []string{JSONBooleanTerm},
		expectedValue: "false",
	},
	{
		srcTypes:       []string{JSONBooleanTerm},
		srcValue:       "False",
		expectGetError: true,
	},
	{
		srcTypes:      []string{XMLBooleanTerm},
		srcValue:      "1",
		targetTypes:   []string{JSONBooleanTerm},
		expectedValue: "true",
	},
	{
		srcTypes:      []string{XMLBooleanTerm},
		srcValue:      "0",
		targetTypes:   []string{JSONBooleanTerm},
		expectedValue: "false",
	},
}

func TestBoolean(t *testing.T) {
	runGetSetTests(t, booleanTests)
}
