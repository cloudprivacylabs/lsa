// Copyright 2021 Cloud Privacy Labs, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package types

import (
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

var dateTests = []getSetTestCase{
	// XSDDate: "2006-01-2"
	// JSONDate: "2006-01-02"
	// XSDDateTime: "2006-01-02T15:04:05Z"
	// JSONDateTime: "2006-01-02T00:00:00Z"
	{
		name:          "source: XSDDate, target: XSDDateTime",
		srcTypes:      []string{XSDDateTerm.Name},
		srcValue:      "2006-01-02",
		targetTypes:   []string{XSDDateTimeTerm.Name},
		expectedValue: "2006-01-02T00:00:00Z", // "2006-01-02T15:04:05Z"
	},
	{
		name:          "source: JSONDate, target: JSONDateTime",
		srcTypes:      []string{JSONDateTerm.Name},
		srcValue:      "2006-01-02",
		targetTypes:   []string{JSONDateTimeTerm.Name},
		expectedValue: "2006-01-02T00:00:00Z", //// "2006-01-02T15:04:05Z"
	},
	{
		name:          "source: JSONDate, target: XSDDate",
		srcTypes:      []string{JSONDateTerm.Name},
		srcValue:      "2006-01-02",
		targetTypes:   []string{XSDDateTerm.Name},
		expectedValue: "2006-01-02",
	},
	{
		name:          "source: XSDDate, target: JSONDate",
		srcTypes:      []string{XSDDateTerm.Name},
		srcValue:      "2006-01-2",
		targetTypes:   []string{JSONDateTerm.Name},
		expectedValue: "2006-01-02",
	},
	{
		name:          "source: JSONDateTime, target: XSDDate",
		srcTypes:      []string{JSONDateTimeTerm.Name},
		srcValue:      "2006-01-02T15:04:05Z",
		targetTypes:   []string{XSDDateTerm.Name},
		expectedValue: "2006-01-02",
	},
	{
		name:          "source: XSDDate, target: JSONDateTime",
		srcTypes:      []string{XSDDateTerm.Name},
		srcValue:      "2006-01-2",
		targetTypes:   []string{JSONDateTimeTerm.Name},
		expectedValue: "2006-01-02T00:00:00Z",
	},
	{
		name:          "source: JSONDateTime, target: XSDDateTime",
		srcTypes:      []string{JSONDateTimeTerm.Name},
		srcValue:      "2006-01-02T15:04:05Z",
		targetTypes:   []string{XSDDateTimeTerm.Name},
		expectedValue: "2006-01-02T15:04:05Z",
	},
	{
		name:          "source: XSDDate, target: XSDGDay",
		srcTypes:      []string{XSDDateTerm.Name},
		srcValue:      "2006-01-2",
		targetTypes:   []string{XSDGDayTerm.Name},
		expectedValue: "2",
	},
	{
		name:          "source: XSDDate, target: XSDGMonth",
		srcTypes:      []string{XSDDateTerm.Name},
		srcValue:      "2006-01-2",
		targetTypes:   []string{XSDGMonthTerm.Name},
		expectedValue: "01",
	},
	{
		name:          "source: XSDDate, target: XSDGYear",
		srcTypes:      []string{XSDDateTerm.Name},
		srcValue:      "2006-01-2",
		targetTypes:   []string{XSDGYearTerm.Name},
		expectedValue: "2006",
	},
	{
		name:          "source: XSDDate, target: XSDGMonthDay",
		srcTypes:      []string{XSDDateTerm.Name},
		srcValue:      "2006-01-2",
		targetTypes:   []string{XSDGMonthDayTerm.Name},
		expectedValue: "01-02",
	},
	{
		name:          "source: XSDDate, target: XSDGYearMonth",
		srcTypes:      []string{XSDDateTerm.Name},
		srcValue:      "2006-01-2",
		targetTypes:   []string{XSDGYearMonthTerm.Name},
		expectedValue: "2006-01",
	},
	{
		name:          "source: JSONDate, target: XSDGDay",
		srcTypes:      []string{JSONDateTerm.Name},
		srcValue:      "2006-01-02",
		targetTypes:   []string{XSDGDayTerm.Name},
		expectedValue: "2",
	},
	{
		name:          "source: JSONDate, target: XSDGMonth",
		srcTypes:      []string{JSONDateTerm.Name},
		srcValue:      "2006-01-02",
		targetTypes:   []string{XSDGMonthTerm.Name},
		expectedValue: "01",
	},
	{
		name:          "source: JSONDate, target: XSDGYear",
		srcTypes:      []string{JSONDateTerm.Name},
		srcValue:      "2006-01-02",
		targetTypes:   []string{XSDGYearTerm.Name},
		expectedValue: "2006",
	},
	{
		name:          "source: JSONDate, target: XSDGMonthDay",
		srcTypes:      []string{JSONDateTerm.Name},
		srcValue:      "2006-01-02",
		targetTypes:   []string{XSDGMonthDayTerm.Name},
		expectedValue: "01-02",
	},
	{
		name:          "source: JSONDate, target: XSDGYearMonth",
		srcTypes:      []string{JSONDateTerm.Name},
		srcValue:      "2006-01-02",
		targetTypes:   []string{XSDGYearMonthTerm.Name},
		expectedValue: "2006-01",
	},
	{
		name:          "source: XSDDateTime, target: XSDGDay",
		srcTypes:      []string{XSDDateTimeTerm.Name},
		srcValue:      "2006-01-02T15:04:05Z",
		targetTypes:   []string{XSDGDayTerm.Name},
		expectedValue: "2",
	},
	{
		name:          "source: XSDDateTime, target: XSDGMonth",
		srcTypes:      []string{XSDDateTimeTerm.Name},
		srcValue:      "2006-01-02T15:04:05Z",
		targetTypes:   []string{XSDGMonthTerm.Name},
		expectedValue: "01",
	},
	{
		name:          "source: XSDDateTime, target: XSDGYear",
		srcTypes:      []string{XSDDateTimeTerm.Name},
		srcValue:      "2006-01-02T15:04:05Z",
		targetTypes:   []string{XSDGYearTerm.Name},
		expectedValue: "2006",
	},
	{
		name:          "source: XSDDateTime, target: XSDGMonthDay",
		srcTypes:      []string{XSDDateTimeTerm.Name},
		srcValue:      "2006-01-02T15:04:05Z",
		targetTypes:   []string{XSDGMonthDayTerm.Name},
		expectedValue: "01-02",
	},
	{
		name:          "source: XSDDateTime, target: XSDGYearMonth",
		srcTypes:      []string{XSDDateTimeTerm.Name},
		srcValue:      "2006-01-02T15:04:05Z",
		targetTypes:   []string{XSDGYearMonthTerm.Name},
		expectedValue: "2006-01",
	},
	{
		name:          "source: JSONDateTime, target: XSDGDay",
		srcTypes:      []string{JSONDateTimeTerm.Name},
		srcValue:      "2006-01-02T00:00:00Z",
		targetTypes:   []string{XSDGDayTerm.Name},
		expectedValue: "2",
	},
	{
		name:          "source: JSONDateTime, target: XSDGMonth",
		srcTypes:      []string{JSONDateTimeTerm.Name},
		srcValue:      "2006-01-02T00:00:00Z",
		targetTypes:   []string{XSDGMonthTerm.Name},
		expectedValue: "01",
	},
	{
		name:          "source: JSONDateTime, target: XSDGYear",
		srcTypes:      []string{JSONDateTimeTerm.Name},
		srcValue:      "2006-01-02T00:00:00Z",
		targetTypes:   []string{XSDGYearTerm.Name},
		expectedValue: "2006",
	},
	{
		name:          "source: JSONDateTime, target: XSDGMonthDay",
		srcTypes:      []string{JSONDateTimeTerm.Name},
		srcValue:      "2006-01-02T00:00:00Z",
		targetTypes:   []string{XSDGMonthDayTerm.Name},
		expectedValue: "01-02",
	},
	{
		name:          "source: JSONDateTime, target: XSDGYearMonth",
		srcTypes:      []string{JSONDateTimeTerm.Name},
		srcValue:      "2006-01-02T00:00:00Z",
		targetTypes:   []string{XSDGYearMonthTerm.Name},
		expectedValue: "2006-01",
	},
	{
		name:          "source: XSDDateTime, target: XSDTime",
		srcTypes:      []string{XSDDateTimeTerm.Name},
		srcValue:      "2006-01-02T15:04:05Z",
		targetTypes:   []string{XSDTimeTerm.Name},
		expectedValue: "09:00:00Z",
	},
	{
		name:          "source: JSONDateTime, target: XSDTime",
		srcTypes:      []string{XSDDateTimeTerm.Name},
		srcValue:      "2006-01-02T00:00:00Z",
		targetTypes:   []string{XSDTimeTerm.Name},
		expectedValue: "09:00:00Z",
	},
	{
		name:          "source: XSDDateTime, target: JSONTime",
		srcTypes:      []string{XSDDateTimeTerm.Name},
		srcValue:      "2006-01-02T15:04:05Z",
		targetTypes:   []string{JSONTimeTerm.Name},
		expectedValue: "09:00:00Z",
	},
	{
		name:          "source: JSONDateTime, target: JSONTime",
		srcTypes:      []string{JSONDateTimeTerm.Name},
		srcValue:      "2006-01-02T00:00:00Z",
		targetTypes:   []string{JSONTimeTerm.Name},
		expectedValue: "09:00:00Z",
	},
	{
		name:          "source: XSDDateTime, target: UnixTime",
		srcTypes:      []string{XSDDateTimeTerm.Name},
		srcValue:      "2006-01-02T15:04:05Z",
		targetTypes:   []string{UnixTimeTerm.Name},
		expectedValue: "1136214245",
	},
	{
		name:          "source: XSDDateTime, target: UnixTimeNano",
		srcTypes:      []string{XSDDateTimeTerm.Name},
		srcValue:      "2006-01-02T15:04:05Z",
		targetTypes:   []string{UnixTimeNanoTerm.Name},
		expectedValue: "1136214245000000000",
	},
	{
		name:          "source: XSDDate, target: UnixTime",
		srcTypes:      []string{XSDDateTerm.Name},
		srcValue:      "2006-01-02",
		targetTypes:   []string{UnixTimeTerm.Name},
		expectedValue: "1136160000",
	},
	{
		name:          "source: JSONDate, target: UnixTime",
		srcTypes:      []string{JSONDateTerm.Name},
		srcValue:      "2006-01-02",
		targetTypes:   []string{UnixTimeTerm.Name},
		expectedValue: "1136160000",
	},
	{
		name:          "source: XSDDate, target: UnixTimeNano",
		srcTypes:      []string{XSDDateTerm.Name},
		srcValue:      "2006-01-02",
		targetTypes:   []string{UnixTimeNanoTerm.Name},
		expectedValue: "1136160000000000000",
	},
	{
		name:          "source: JSONDate, target: UnixTimeNano",
		srcTypes:      []string{JSONDateTerm.Name},
		srcValue:      "2006-01-02",
		targetTypes:   []string{UnixTimeNanoTerm.Name},
		expectedValue: "1136160000000000000",
	},
	{
		name:          "source: JSONDateTime, target: UnixTime",
		srcTypes:      []string{JSONDateTimeTerm.Name},
		srcValue:      "2006-01-02T00:00:00Z",
		targetTypes:   []string{UnixTimeNanoTerm.Name},
		expectedValue: "1136160000000000000",
	},
	{
		name:          "source: JSONDateTime, target: UnixTimeNano",
		srcTypes:      []string{JSONDateTimeTerm.Name},
		srcValue:      "2006-01-02T00:00:00Z",
		targetTypes:   []string{UnixTimeNanoTerm.Name},
		expectedValue: "1136160000000000000",
	},
	{
		name:          "source: XSDTime, target: XSDDateTime",
		srcTypes:      []string{XSDTimeTerm.Name},
		srcValue:      "15:04:05Z",
		targetTypes:   []string{XSDDateTimeTerm.Name},
		expectedValue: "2006-01-02T15:04:05Z",
	},
	{
		name:          "source: XSDTime, target: JSONDateTime",
		srcTypes:      []string{XSDTimeTerm.Name},
		srcValue:      "15:04:05Z",
		targetTypes:   []string{JSONDateTimeTerm.Name},
		expectedValue: "2006-01-02T15:04:05Z",
	},
	{
		name:          "source: JSONTime, target: XSDDateTime",
		srcTypes:      []string{JSONTimeTerm.Name},
		srcValue:      "15:04:05Z",
		targetTypes:   []string{XSDDateTimeTerm.Name},
		expectedValue: "2006-01-02T15:04:05Z",
	},
	{
		name:          "source: JSONTime, target: JSONDateTime",
		srcTypes:      []string{JSONTimeTerm.Name},
		srcValue:      "15:04:05Z",
		targetTypes:   []string{JSONDateTimeTerm.Name},
		expectedValue: "2006-01-02T15:04:05Z",
	},
	{
		name:          "source: XSDTime, target: UnixTime",
		srcTypes:      []string{XSDTimeTerm.Name},
		srcValue:      "15:04:05Z",
		targetTypes:   []string{UnixTimeTerm.Name},
		expectedValue: "54245",
	},
	{
		name:          "source: JSONTime, target: UnixTime",
		srcTypes:      []string{JSONTimeTerm.Name},
		srcValue:      "15:04:05Z",
		targetTypes:   []string{UnixTimeTerm.Name},
		expectedValue: "54245",
	},
	{
		name:          "source: XSDTime, target: UnixTimeNano",
		srcTypes:      []string{XSDTimeTerm.Name},
		srcValue:      "15:04:05Z",
		targetTypes:   []string{UnixTimeNanoTerm.Name},
		expectedValue: "54245000000000",
	},
	{
		name:          "source: JSONTime, target: UnixTimeNano",
		srcTypes:      []string{JSONTimeTerm.Name},
		srcValue:      "15:04:05Z",
		targetTypes:   []string{UnixTimeNanoTerm.Name},
		expectedValue: "54245000000000",
	},
	{
		name:          "source: UnixTime, target: UnixTimeNano",
		srcTypes:      []string{UnixTimeTerm.Name},
		srcValue:      "1136160000",
		targetTypes:   []string{UnixTimeNanoTerm.Name},
		expectedValue: "1136160000000000000",
	},
	{
		name:             "source: XSDDateTime, target: PatternDateTime (goment)",
		srcTypes:         []string{XSDDateTimeTerm.Name},
		srcValue:         "2006-01-02T15:04:05Z",
		targetProperties: map[string]interface{}{MomentTimeFormatTerm.Name: ls.NewPropertyValue(MomentTimeFormatTerm.Name, "MM-DD-YYYY HH:mm:ss")},
		targetTypes:      []string{PatternDateTimeTerm.Name},
		expectedValue:    "01-02-2006 15:04:05",
	},
	{
		name:          "source: PatternDate (goment), target: xsdDate",
		srcTypes:      []string{PatternDateTerm.Name},
		srcValue:      "20150203",
		srcProperties: map[string]interface{}{MomentTimeFormatTerm.Name: ls.NewPropertyValue(MomentTimeFormatTerm.Name, "YYYYMMDD")},
		targetTypes:   []string{XSDDateTerm.Name},
		expectedValue: "2015-02-03",
	},
	{
		name:             "source: XSDTime, target: PatternDateTime (goment)",
		srcTypes:         []string{XSDTimeTerm.Name},
		srcValue:         "15:04:05Z",
		targetProperties: map[string]interface{}{MomentTimeFormatTerm.Name: ls.NewPropertyValue(MomentTimeFormatTerm.Name, "HH:mm:ss")},
		targetTypes:      []string{PatternDateTimeTerm.Name},
		expectedValue:    "15:04:05",
	},
	{
		name:          "source: PatternDateTime (no format), target: xsdDate",
		srcTypes:      []string{PatternDateTimeTerm.Name},
		srcValue:      "20200728000000",
		targetTypes:   []string{XSDDateTerm.Name},
		expectedValue: "2020-07-28Z",
	},
	// {
	// 	srcTypes:      []string{XSDDateTerm.Name},
	// 	srcValue:      "2006-01-2",
	// 	targetTypes:   []string{XSDGMonthDayTerm.Name},
	// 	expectedValue: "{1 2 2006 UTC}",
	// },
}

func TestDate(t *testing.T) {
	runGetSetTests(t, dateTests)
}
