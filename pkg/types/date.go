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
	"fmt"
	"time"

	"github.com/nleeper/goment"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type ErrCannotParseTemporalValue string

func (e ErrCannotParseTemporalValue) Error() string {
	return "Cannot parse temporal value: " + string(e)
}

const XSD = "http://www.w3.org/2001/XMLSchema/"
const JSON = "https:/json-schema.org/"

type Date struct {
	Month    int
	Day      int
	Year     int
	Location *time.Location
}

func (d Date) ToTime() time.Time {
	if d.Location == nil {
		return time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
	}
	return time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, d.Location)
}

// try to convert to go native time with function ToGoTime then pass result as parameter
func ToGomentTime(time time.Time) interface{} {
	t, err := goment.New(time)
	if err != nil {
		return err
	}
	return t
}

type DateTime struct {
	Month        int
	Day          int
	Year         int
	Nanoseconds  int64
	Milliseconds int64
	Seconds      int64
	Minute       int64
	Hour         int64
	Location     *time.Location
}

func (dt DateTime) ToTime() time.Time {
	if dt.Location == nil {
		return time.Date(dt.Year, time.Month(dt.Month), dt.Day, int(dt.Hour), int(dt.Minute), int(dt.Seconds), int(dt.Nanoseconds), time.UTC)
	}
	return time.Date(dt.Year, time.Month(dt.Month), dt.Day, int(dt.Hour), int(dt.Minute), int(dt.Seconds), int(dt.Nanoseconds), dt.Location)
}

type TimeOfDay struct {
	Nanoseconds  int64
	Milliseconds int64
	Seconds      int64
	Minute       int64
	Hour         int64
	Location     *time.Location
}

func (t TimeOfDay) ToTime() time.Time {
	if t.Location == nil {
		return time.Date(0, 0, 0, int(t.Hour), int(t.Minute), int(t.Nanoseconds)/1000000000, int(t.Nanoseconds), time.UTC)
	}
	return time.Date(0, 0, 0, int(t.Hour), int(t.Minute), int(t.Nanoseconds)/1000000000, int(t.Nanoseconds), t.Location)
}

func timeToString(time interface{}) string {
	return time.(string)
}

// GDay is XML Gregorian day part of date
type GDay int

// XSDGday can be used as a node-type to interpret the underlying value as a day (GDay)
var XSDGdayTerm = ls.NewTerm(XSD+"gDay", false, false, ls.OverrideComposition, nil)

// GMonth is XML Gregorian month part of date
type GMonth int

// XSDGMonth can be used as node-type to interpret the underlying value as a month (int)
var XSDGMonthTerm = ls.NewTerm(XSD+"gMonth", false, false, ls.OverrideComposition, nil)

// GMonth is XML Gregorian year part of date
type GYear int

// XSDGYear can be used as a node-type to interpret the underlying value as a year value (int)
var XSDGYearTerm = ls.NewTerm(XSD+"gYear", false, false, ls.OverrideComposition, nil)

// GMonthDay is XML Gregorian part of Month/Day
type GMonthDay struct {
	Day   int
	Month int
}

// XSDMonthDay can be used as a node-type to interpret the underlying value as a MM-DD
var XSDGMonthDayTerm = ls.NewTerm(XSD+"gMonthDay", false, false, ls.OverrideComposition, nil)

// GYearMonth is XML Gregorian part of Year/Month
type GYearMonth struct {
	Year  int
	Month int
}

// XSDGYearMonth can be used as a node-type to interpret the underlying value as a YYYY-MM
var XSDGYearMonthTerm = ls.NewTerm(XSD+"gYearMonth", false, false, ls.OverrideComposition, nil)

// XSDDate is a node-type that identifies the underlying value as an XML date. The format is:
//
//  [-]CCYY-MM-DD[Z|(+|-)hh:mm]
var XSDDateTerm = ls.NewTerm(XSD+"date", false, false, ls.OverrideComposition, struct {
	XSDDateParser
}{
	XSDDateParser: XSDDateParser{},
})

// XSDTime is a node-type that identifies the underlying value as an XML time.
var XSDTimeTerm = ls.NewTerm(XSD+"time", false, false, ls.OverrideComposition, nil)

// XSDDateTime is a node-type that identifies the underlying value as an XML date-time value
var XSDDateTimeTerm = ls.NewTerm(XSD+"dateTime", false, false, ls.OverrideComposition, nil)

// JSONDate is a node-type that identifies the underlying value as a JSON date value
//
//  YYYY-MM-DD
var JSONDateTerm = ls.NewTerm(JSON+"date", false, false, ls.OverrideComposition, struct {
	JSONDateParser
}{
	JSONDateParser: JSONDateParser{},
})

// JSONDateTime is a node-type that identifies the underlying value as
// a JSON datetime value, RFC3339 or RFC3339Nano
//
// YYYY-MM-DDTHH:mm:ssZ
// YYYY-MM-DDTHH:mm:ss.00000Z
var JSONDateTimeTerm = ls.NewTerm(JSON+"date-time", false, false, ls.OverrideComposition, struct {
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
var JSONTimeTerm = ls.NewTerm(JSON+"time", false, false, ls.OverrideComposition, struct {
	JSONTimeParser
}{
	JSONTimeParser: JSONTimeParser{},
})

var PatternDateTimeTerm = ls.NewTerm(ls.LS+"dateTime", false, false, ls.OverrideComposition, struct {
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
	parse(string) (Date, error)
}

func (f goFormat) parse(s string) (Date, error) {
	t, err := time.Parse(string(f), s)
	if err != nil {
		return Date{Month: int(t.Month()), Day: t.Day(), Year: t.Year()}, err
	}
	return Date{}, nil
}

func (f gomentFormat) parse(s string) (Date, error) {
	t, err := goment.New(s, string(f))
	if err != nil {
		return Date{Month: int(t.Month()), Day: t.Day(), Year: t.Year()}, err
	}
	return Date{}, nil
}

// ParseValue parses an XSDDate value.
//
//  [-]CCYY-MM-DD[Z|(+|-)hh:mm]
func (XSDDateParser) GetNodeValue(node ls.Node) (interface{}, error) {
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

// SetValue gets a target node and it's go native value, and sets
// the value of the target node to an XSDDate
func (XSDDateParser) SetNodeValue(node ls.Node, value interface{}) error {
	if value == nil {
		node.SetValue(nil)
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		node.SetValue(v.Format("2006-01-2"))
	case Date:
		if v.Location == nil {
			node.SetValue(v.ToTime().Format("2006-01-2"))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format("2006-01-2Z0700"))
		}
	case DateTime:
		if v.Location == nil {
			node.SetValue(v.ToTime().Format("2006-01-2"))
			// target.SetValue(fmt.Sprintf("%04d-%02d-%02d"+"T"+"%02d:%02d:%02d", value.Year, value.Month, value.Day, value.Hour, value.Minute, (value.Nanoseconds / 1000000000)))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format("2006-01-2Z0700"))
		}
	case GDay:
		node.SetValue(timeToString(v))
	case GMonth:
		node.SetValue(timeToString(v))
	case GYear:
		node.SetValue(timeToString(v))
	case GMonthDay:
		node.SetValue(fmt.Sprintf("%02d-%02d", v.Month, v.Day))
	case GYearMonth:
		node.SetValue(fmt.Sprintf("%04d-%02d", v.Year, v.Month))
	}
	return ls.ErrInvalidValue{ID: node.GetID(), Type: XSDDateTerm, Value: value}
}

type JSONDateParser struct{}

// ParseValue parses a JSON date
//
//   YYYY-MM-DD
func (JSONDateParser) GetNodeValue(node ls.Node) (interface{}, error) {
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

// SetValue gets a target node and it's go native value, and sets
// the value of the target node to an JSONDate
func (JSONDateParser) SetNodeValue(node ls.Node, value interface{}) error {
	if value == nil {
		node.SetValue(nil)
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		node.SetValue(v.Format("2006-01-02"))
	case Date:
		if v.Location == nil {
			node.SetValue(v.ToTime().Format("2006-01-02"))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format("2006-01-02"))
		}
	case DateTime:
		if v.Location == nil {
			node.SetValue(v.ToTime().Format("2006-01-02"))
			// target.SetValue(fmt.Sprintf("%04d-%02d-%02d"+"T"+"%02d:%02d:%02d", value.Year, value.Month, value.Day, value.Hour, value.Minute, (value.Nanoseconds / 1000000000)))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format("2006-01-02Z0700"))
		}
	case GDay:
		node.SetValue(timeToString(v))
	case GMonth:
		node.SetValue(timeToString(v))
	case GYear:
		node.SetValue(timeToString(v))
	case GMonthDay:
		node.SetValue(fmt.Sprintf("%02d-%02d", v.Month, v.Day))
	case GYearMonth:
		node.SetValue(fmt.Sprintf("%04d-%02d", v.Year, v.Month))
	}
	return ls.ErrInvalidValue{ID: node.GetID(), Type: JSONDateTerm, Value: value}
}

type JSONDateTimeParser struct{}

// ParseValue parses a JSON date-time
//
//   YYYY-MM-DD
func (JSONDateTimeParser) GetNodeValue(node ls.Node) (interface{}, error) {
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

// "2006-01-02T15:04:05Z07:00" -> Note: uses a 24Hour based clock   "2006-01-02T15:04:05.999999999Z07:00"
func (JSONDateTimeParser) SetNodeValue(node ls.Node, value interface{}) error {
	if value == nil {
		node.SetValue(nil)
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		node.SetValue(v.Format(time.RFC3339))
	case Date:
		if v.Location == nil {
			node.SetValue(v.ToTime().Format("2006-01-02"))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format("2006-01-02"))
		}
	case DateTime:
		if v.Location == nil && v.Nanoseconds == 0 {
			node.SetValue(v.ToTime().Format("2006-01-02T15:04:05"))
		} else if v.Location != nil && v.Nanoseconds != 0 {
			node.SetValue(v.ToTime().In(v.Location).Format(time.RFC3339Nano))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format(time.RFC3339))
		}
	}
	return ls.ErrInvalidValue{ID: node.GetID(), Type: JSONDateTimeTerm, Value: value}
}

type JSONTimeParser struct{}

// ParseValue parses a JSON time
//
//   HH:mm
//   HH:mm:ss
//   HH:mm:ssZ
func (JSONTimeParser) GetNodeValue(node ls.Node) (interface{}, error) {
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

func (JSONTimeParser) SetNodeValue(node ls.Node, value interface{}) error {
	if value == nil {
		node.SetValue(nil)
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		node.SetValue(v.Format("HH:mm:ssZ"))
	case TimeOfDay:
		if v.Location == nil && v.Nanoseconds == 0 {
			node.SetValue(v.ToTime().Format("HH:mm"))
		} else if v.Location == nil && v.Nanoseconds != 0 {
			node.SetValue(v.ToTime().Format("HH:mm:ss"))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format("HH:mm:ssZ"))
		}
	}
	return ls.ErrInvalidValue{ID: node.GetID(), Type: JSONTimeTerm, Value: value}
}

type PatternDateTimeParser struct{}

// ParseValue looks at the goTimeFormat, momentTimeFormat properties
// in the node, and parses the datetime using that. The format
// property can be an array, giving all possible formats
func (PatternDateTimeParser) GetNodeValue(node ls.Node) (interface{}, error) {
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

func (PatternDateTimeParser) SetNodeValue(node ls.Node, value interface{}) error {
	if value == nil {
		node.SetValue(nil)
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		node.SetValue(v.Format(node.GetProperties()[v.String()].AsString()))
	case Date:
	case DateTime:
	case TimeOfDay:
		if v.Location == nil && v.Nanoseconds == 0 {
			node.SetValue(v.ToTime().Format(node.GetProperties()[v.ToTime().String()].AsString()))
		} else if v.Location == nil && v.Nanoseconds != 0 {
			node.SetValue(v.ToTime().Format("HH:mm:ss"))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format("HH:mm:ssZ"))
		}
	}
	return ls.ErrInvalidValue{ID: node.GetID(), Type: PatternDateTimeTerm, Value: value}
}

// genericDateParse parses a node value using the given format(s)
func genericDateParse(value string, format ...dateFormatter) (Date, error) {
	for _, f := range format {
		t, err := f.parse(value)
		if err == nil {
			return t, nil
		}
	}
	return Date{}, ErrCannotParseTemporalValue(value)
}
