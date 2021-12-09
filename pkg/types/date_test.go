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

func TestXSDDate(t *testing.T) {
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
	testTime(t, nil, noerr(XSDDateParser{}.ParseValue(mkNode(nil))))
	testTime(t, nil, noerr(XSDDateParser{}.ParseValue(mkNode(""))))
	testTime(t, time.Date(2001, 9, 26, 0, 0, 0, 0, time.UTC), noerr(XSDDateParser{}.ParseValue(mkNode("2001-9-26"))))
	testTime(t, time.Date(2001, 10, 26, 0, 0, 0, 0, time.FixedZone("+2", 2*60*60)), noerr(XSDDateParser{}.ParseValue(mkNode("2001-10-26+02:00"))))
	testTime(t, time.Date(2001, 11, 26, 0, 0, 0, 0, time.UTC), noerr(XSDDateParser{}.ParseValue(mkNode("2001-11-26Z"))))
	testTime(t, time.Date(2001, 12, 26, 0, 0, 0, 0, time.UTC), noerr(XSDDateParser{}.ParseValue(mkNode("2001-12-26+00:00"))))
}

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
	testTime(t, nil, noerr(JSONDateParser{}.ParseValue(mkNode(nil))))
	testTime(t, nil, noerr(JSONDateParser{}.ParseValue(mkNode(""))))
	testTime(t, time.Date(2001, 9, 26, 0, 0, 0, 0, time.UTC), noerr(JSONDateParser{}.ParseValue(mkNode("2001-09-26"))))
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
	testTime(t, nil, noerr(JSONDateParser{}.ParseValue(mkNode(nil))))
	testTime(t, nil, noerr(JSONDateParser{}.ParseValue(mkNode(""))))
	testTime(t, time.Date(2001, 9, 26, 10, 11, 12, 0, time.UTC), noerr(JSONDateTimeParser{}.ParseValue(mkNode("2001-09-26T10:11:12Z"))))
	testTime(t, time.Date(2001, 9, 26, 10, 11, 12, 0, time.FixedZone("+2", 2*60*60)), noerr(JSONDateTimeParser{}.ParseValue(mkNode("2001-09-26T10:11:12+02:00"))))
}
