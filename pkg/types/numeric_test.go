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

var numericTests = []getSetTestCase{
	{
		srcTypes:      []string{XSDByte},
		srcValue:      "10",
		targetTypes:   []string{XSDByte},
		expectedValue: "10",
	},
	{
		srcTypes:      []string{XSDByte},
		srcValue:      "-10",
		targetTypes:   []string{XSDByte},
		expectedValue: "-10",
	},
	{
		srcTypes:      []string{XSDByte},
		srcValue:      "10",
		targetTypes:   []string{XSDInt},
		expectedValue: "10",
	},
	{
		srcTypes:      []string{XSDByte},
		srcValue:      "-10",
		targetTypes:   []string{XSDInt},
		expectedValue: "-10",
	},
	{
		srcTypes:      []string{XSDByte},
		srcValue:      "10",
		targetTypes:   []string{XSDLong},
		expectedValue: "10",
	},
	{
		srcTypes:      []string{XSDByte},
		srcValue:      "-10",
		targetTypes:   []string{XSDLong},
		expectedValue: "-10",
	},
	{
		srcTypes:      []string{XSDByte},
		srcValue:      "10",
		targetTypes:   []string{XSDShort},
		expectedValue: "10",
	},
	{
		srcTypes:      []string{XSDByte},
		srcValue:      "-10",
		targetTypes:   []string{XSDShort},
		expectedValue: "-10",
	},
	{
		srcTypes:      []string{XSDByte},
		srcValue:      "10",
		expectedValue: "10",
	},
	{
		srcTypes:      []string{XSDByte},
		srcValue:      "-10",
		expectedValue: "-10",
	},
	{
		srcValue:      "10",
		targetTypes:   []string{XSDShort},
		expectedValue: "10",
	},
	{
		srcValue:      "-10",
		targetTypes:   []string{XSDShort},
		expectedValue: "-10",
	},
	{
		srcValue:       "-1000",
		targetTypes:    []string{XSDByte},
		expectSetError: true,
	},
}

func TestNumeric(t *testing.T) {
	runGetSetTests(t, numericTests)
}
