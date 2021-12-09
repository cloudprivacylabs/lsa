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
	"time"

	"github.com/nleeper/goment"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type ErrCannotParseTemporalValue string

func (e ErrCannotParseTemporalValue) Error() string {
	return "Cannot parse temporal value: " + string(e)
}

const XSD = "http://www.w3.org/2001/XMLSchema"
const JSON = "https:/json-schema.org/"

// XSDGday can be used as a node-type to interpret the underlying value as a day (int)
var XSDGday = ls.NewTerm(XSD+":gDay", false, false, ls.OverrideComposition, nil)

// XSDGMonth can be used as node-type to interpret the underlying value as a month (int)
var XSDGMonth = ls.NewTerm(XSD+":gMonth", false, false, ls.OverrideComposition, nil)

// XSDGYear can be used as a node-type to interpret the underlying value as a year value (int)
var XSDGYear = ls.NewTerm(XSD+":gYear", false, false, ls.OverrideComposition, nil)

// XSDMonthDay can be used as a node-type to interpret the underlying value as a MM-DD
var XSDGMonthDay = ls.NewTerm(XSD+":gMonthDay", false, false, ls.OverrideComposition, nil)

// XSDGYearMonth can be used as a node-type to interpret the underlying value as a YYYY-MM
var XSDGYearMonth = ls.NewTerm(XSD+":gYearMonth", false, false, ls.OverrideComposition, nil)

// XSDDate is a node-type that identifies the underlying value as an XML date. The format is:
//
//  [-]CCYY-MM-DD[Z|(+|-)hh:mm]
var XSDDate = ls.NewTerm(XSD+":date", false, false, ls.OverrideComposition, struct {
	XSDDateParser
}{
	XSDDateParser: XSDDateParser{},
})

// XSDTime is a node-type that identifies the underlying value as an XML time.
var XSDTime = ls.NewTerm(XSD+":time", false, false, ls.OverrideComposition, nil)

// XSDDateTime is a node-type that identifies the underlying value as an XML date-time value
var XSDDateTime = ls.NewTerm(XSD+":dateTime", false, false, ls.OverrideComposition, nil)

// JSONDate is a node-type that identifies the underlying value as a JSON date value
//
//  YYYY-MM-DD
var JSONDate = ls.NewTerm(JSON+"date", false, false, ls.OverrideComposition, struct {
	JSONDateParser
}{
	JSONDateParser: JSONDateParser{},
})

// JSONDateTime is a node-type that identifies the underlying value as
// a JSON datetime value, RFC3339 or RFC3339Nano
//
// YYYY-MM-DDTHH:mm:ssZ
// YYYY-MM-DDTHH:mm:ss.00000Z
var JSONDateTime = ls.NewTerm(JSON+"date-time", false, false, ls.OverrideComposition, struct {
	JSONDateTimeParser
}{
	JSONDateTimeParser: JSONDateTimeParser{},
})

// JSONTime is a node-type that identifies the underlying value as a
// JSON time value
//
//   HH:mm
//   HH:mm:ss
//   HH:mm:ssZ
var JSONTime = ls.NewTerm(JSON+"time", false, false, ls.OverrideComposition, struct {
	JSONTimeParser
}{
	JSONTimeParser: JSONTimeParser{},
})

var PatternDateTime = ls.NewTerm(ls.LS+"dateTime", false, false, ls.OverrideComposition, struct {
	PatternDateTimeParser
}{
	PatternDateTimeParser: PatternDateTimeParser{},
})

var GoTimeFormatTerm = ls.NewTerm(ls.LS+"goTimeFormat", false, false, ls.SetComposition, nil)
var MomentTimeFormatTerm = ls.NewTerm(ls.LS+"momentTimeFormat", false, false, ls.SetComposition, nil)

type XSDDateParser struct{}

type goFormat string
type gomentFormat string

type dateFormatter interface {
	parse(string) (time.Time, error)
}

func (f goFormat) parse(s string) (time.Time, error) {
	return time.Parse(string(f), s)
}

func (f gomentFormat) parse(s string) (time.Time, error) {
	t, err := goment.New(s, string(f))
	if err != nil {
		return time.Time{}, err
	}
	return t.ToTime(), nil
}

// ParseValue parses an XSDDate value.
//
//  [-]CCYY-MM-DD[Z|(+|-)hh:mm]
func (XSDDateParser) ParseValue(node ls.Node) (interface{}, error) {
	value, exists, err := getStringNodeValue(node)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	if len(value) == 0 {
		return nil, nil
	}
	return genericDateParse(value, goFormat("2006-1-2"), goFormat("2006-1-2Z"), goFormat("2006-1-2Z0700"), gomentFormat("YYYY-MM-DDZ"))
}

type JSONDateParser struct{}

// ParseValue parses a JSON date
//
//   YYYY-MM-DD
func (JSONDateParser) ParseValue(node ls.Node) (interface{}, error) {
	value, exists, err := getStringNodeValue(node)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	if len(value) == 0 {
		return nil, nil
	}
	return genericDateParse(value, goFormat("2006-01-02"))
}

type JSONDateTimeParser struct{}

// ParseValue parses a JSON date-time
//
//   YYYY-MM-DD
func (JSONDateTimeParser) ParseValue(node ls.Node) (interface{}, error) {
	value, exists, err := getStringNodeValue(node)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	if len(value) == 0 {
		return nil, nil
	}
	return genericDateParse(value, goFormat(time.RFC3339), goFormat(time.RFC3339Nano))
}

type JSONTimeParser struct{}

// ParseValue parses a JSON time
//
//   HH:mm
//   HH:mm:ss
//   HH:mm:ssZ
func (JSONTimeParser) ParseValue(node ls.Node) (interface{}, error) {
	value, exists, err := getStringNodeValue(node)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	if len(value) == 0 {
		return nil, nil
	}
	return genericDateParse(value, gomentFormat("HH:mm:ssZ"),
		gomentFormat("HH:mm:ss"),
		gomentFormat("HH:mm"))
}

type PatternDateTimeParser struct{}

// ParseValue looks at the goTimeFormat, momentTimeFormat properties
// in the node, and parses the datetime using that. The format
// property can be an array, giving all possible formats
func (PatternDateTimeParser) ParseValue(node ls.Node) (interface{}, error) {
	value, exists, err := getStringNodeValue(node)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	if len(value) == 0 {
		return nil, nil
	}
	gf := node.GetProperties()[GoTimeFormatTerm].Slice()
	mf := node.GetProperties()[MomentTimeFormatTerm].Slice()
	garr := make([]dateFormatter, 0, len(gf)+len(mf))
	for _, x := range gf {
		garr = append(garr, goFormat(x))
	}
	for _, x := range mf {
		garr = append(garr, gomentFormat(x))
	}
	return genericDateParse(value, garr...)
}

// genericDateParse parses a node value using the given format(s)
func genericDateParse(value string, format ...dateFormatter) (time.Time, error) {
	for _, f := range format {
		t, err := f.parse(value)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, ErrCannotParseTemporalValue(value)
}
