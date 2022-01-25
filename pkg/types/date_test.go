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
	"time"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func testTime(t *testing.T, expected, got interface{}) {
	if expected == nil && got == nil {
		return
	}
	if expected == nil || got == nil {
		t.Errorf("Expected %v got %v", expected, got)
		return
	}
	if !expected.(time.Time).Equal(got.(time.Time)) {
		t.Errorf("Expected %v got %v", expected, got)
	}
}

// func TestXSDDate(t *testing.T) {
// 	noerr := func(v interface{}, err error) interface{} {
// 		if err != nil {
// 			t.Error(err)
// 		}
// 		return v
// 	}
// 	mkNode := func(v interface{}) ls.Node {
// 		node := ls.NewNode("id")
// 		node.SetValue(v)
// 		return node
// 	}
// 	testTime(t, nil, noerr(XSDDateParser{}.GetNodeValue(mkNode(nil))))
// 	testTime(t, nil, noerr(XSDDateParser{}.GetNodeValue(mkNode(""))))
// 	testTime(t, Date{2001, 9, 26, time.UTC}, noerr(XSDDateParser{}.GetNodeValue(mkNode("2001-9-26"))).(Date))
// 	testTime(t, time.Date(2001, 9, 26, 0, 0, 0, 0, time.UTC), noerr(XSDDateParser{}.GetNodeValue(mkNode("2001-9-26"))))
// 	testTime(t, time.Date(2001, 10, 26, 0, 0, 0, 0, time.FixedZone("+2", 2*60*60)), noerr(XSDDateParser{}.GetNodeValue(mkNode("2001-10-26+02:00"))))
// 	testTime(t, time.Date(2001, 11, 26, 0, 0, 0, 0, time.UTC), noerr(XSDDateParser{}.GetNodeValue(mkNode("2001-11-26Z"))))
// 	testTime(t, time.Date(2001, 12, 26, 0, 0, 0, 0, time.UTC), noerr(XSDDateParser{}.GetNodeValue(mkNode("2001-12-26+00:00"))))
// }

func TestJSONDate(t *testing.T) {
	noerr := func(v interface{}, err error) interface{} {
		if err != nil {
			t.Error(err)
		}
		return v
	}
	mkNode := func(v interface{}) ls.Node {
		node := ls.NewNode("id")
		node.SetValue(v)
		return node
	}
	testTime(t, nil, noerr(JSONDateParser{}.GetNodeValue(mkNode(nil))))
	testTime(t, nil, noerr(JSONDateParser{}.GetNodeValue(mkNode(""))))
	testTime(t, time.Date(2001, 9, 26, 0, 0, 0, 0, time.UTC), noerr(JSONDateParser{}.GetNodeValue(mkNode("2001-09-26"))))
}

func TestJSONDateTime(t *testing.T) {
	noerr := func(v interface{}, err error) interface{} {
		if err != nil {
			t.Error(err)
		}
		return v
	}
	mkNode := func(v interface{}) ls.Node {
		node := ls.NewNode("id")
		node.SetValue(v)
		return node
	}
	testTime(t, nil, noerr(JSONDateParser{}.GetNodeValue(mkNode(nil))))
	testTime(t, nil, noerr(JSONDateParser{}.GetNodeValue(mkNode(""))))
	testTime(t, time.Date(2001, 9, 26, 10, 11, 12, 0, time.UTC), noerr(JSONDateTimeParser{}.GetNodeValue(mkNode("2001-09-26T10:11:12Z"))))
	testTime(t, time.Date(2001, 9, 26, 10, 11, 12, 0, time.FixedZone("+2", 2*60*60)), noerr(JSONDateTimeParser{}.GetNodeValue(mkNode("2001-09-26T10:11:12+02:00"))))
}

var dateTests = []getSetTestCase{
	{
		srcTypes:      []string{JSONDateTerm},
		srcValue:      "2006-01-02",
		targetTypes:   []string{XSDDateTerm},
		expectedValue: "2006-01-2",
	},
	// {
	// 	srcTypes:      []string{JSONBooleanTerm},
	// 	srcValue:      "false",
	// 	targetTypes:   []string{JSONBooleanTerm},
	// 	expectedValue: "false",
	// },
	// {
	// 	srcTypes:       []string{JSONBooleanTerm},
	// 	srcValue:       "False",
	// 	expectGetError: true,
	// },
	// {
	// 	srcTypes:      []string{XMLBooleanTerm},
	// 	srcValue:      "1",
	// 	targetTypes:   []string{JSONBooleanTerm},
	// 	expectedValue: "true",
	// },
	// {
	// 	srcTypes:      []string{XMLBooleanTerm},
	// 	srcValue:      "0",
	// 	targetTypes:   []string{JSONBooleanTerm},
	// 	expectedValue: "false",
	// },
}

func TestDate(t *testing.T) {
	runGetSetTests(t, dateTests)
}
