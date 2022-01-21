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

func ToGoTime(t interface{}) time.Time {
	return t.(time.Time)
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
	Month       int
	Day         int
	Year        int
	Nanoseconds int64
	Minute      int64
	Hour        int64
	Location    *time.Location
}

type TimeOfDay struct {
	Nanoseconds int64
	Minute      int64
	Hour        int64
	Location    *time.Location
}

func (t TimeOfDay) GetHour() int64 {
	return t.Hour
}

func (t TimeOfDay) GetMinute() int64 {
	return t.Minute
}

func (t TimeOfDay) GetNano() int64 {
	return t.Nanoseconds
}

func (t *TimeOfDay) SetMinute(min int64) {
	t.Minute = min
}

func (t *TimeOfDay) SetHour(hour int64) {
	t.Hour = hour
}

func (t *TimeOfDay) SetNano(nano int64) {
	t.Nanoseconds = nano
}

func timeToString(time interface{}) string {
	return time.(string)
}

// GDay is XML Gregorian day part of date
type GDay int

// XSDGday can be used as a node-type to interpret the underlying value as a day (GDay)
var XSDGday = ls.NewTerm(XSD+"gDay", false, false, ls.OverrideComposition, nil)

// GMonth is XML Gregorian month part of date
type GMonth int

// XSDGMonth can be used as node-type to interpret the underlying value as a month (int)
var XSDGMonth = ls.NewTerm(XSD+"gMonth", false, false, ls.OverrideComposition, nil)

// GMonth is XML Gregorian year part of date
type GYear int

// XSDGYear can be used as a node-type to interpret the underlying value as a year value (int)
var XSDGYear = ls.NewTerm(XSD+"gYear", false, false, ls.OverrideComposition, nil)

// GMonthDay is XML Gregorian part of Month/Day
type GMonthDay struct {
	Day   int
	Month int
}

// XSDMonthDay can be used as a node-type to interpret the underlying value as a MM-DD
var XSDGMonthDay = ls.NewTerm(XSD+"gMonthDay", false, false, ls.OverrideComposition, nil)

// GYearMonth is XML Gregorian part of Year/Month
type GYearMonth struct {
	Year  int
	Month int
}

// XSDGYearMonth can be used as a node-type to interpret the underlying value as a YYYY-MM
var XSDGYearMonth = ls.NewTerm(XSD+"gYearMonth", false, false, ls.OverrideComposition, nil)

// XSDDate is a node-type that identifies the underlying value as an XML date. The format is:
//
//  [-]CCYY-MM-DD[Z|(+|-)hh:mm]
var XSDDate = ls.NewTerm(XSD+"date", false, false, ls.OverrideComposition, struct {
	XSDDateParser
}{
	XSDDateParser: XSDDateParser{},
})

// XSDTime is a node-type that identifies the underlying value as an XML time.
var XSDTime = ls.NewTerm(XSD+"time", false, false, ls.OverrideComposition, nil)

// XSDDateTime is a node-type that identifies the underlying value as an XML date-time value
var XSDDateTime = ls.NewTerm(XSD+"dateTime", false, false, ls.OverrideComposition, nil)

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
type XSDDateFormatter struct{}

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

// SetValue gets a target node and it's go native value, and sets
// the value of the target node to an XSDDate
func (XSDDateFormatter) SetValue(target ls.Node, val interface{}) error {
	switch value := val.(type) {
	case time.Time:
		target.SetValue(value)
	case Date:
		if value.Location == nil {
			target.SetValue(fmt.Sprintf("%04d-%02d-%02d", value.Year, value.Month, value.Day))
		} else {
			target.SetValue("")
		}
	case DateTime:
		if value.Location == nil {
			target.SetValue(fmt.Sprintf("%04d-%02d-%02d"+"T"+"%02d:%02d:%02d", value.Year, value.Month, value.Day, value.Hour, value.Minute, (value.Nanoseconds / 1000000000)))
		}
	case GDay:
		target.SetValue(timeToString(value))
	case GMonth:
		target.SetValue(timeToString(value))
	case GYear:
		target.SetValue(timeToString(value))
	case GMonthDay:
		target.SetValue(fmt.Sprintf("%02d-%02d", value.Month, value.Day))
	case GYearMonth:
		target.SetValue(fmt.Sprintf("%04d-%02d", value.Year, value.Month))
	}
	return nil
}

type JSONDateParser struct{}
type JSONDateFormatter struct{}

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

// SetValue gets a target node and it's go native value, and sets
// the value of the target node to an JSONDate
func (JSONDateFormatter) SetValue(target ls.Node, val interface{}) error {
	switch value := val.(type) {
	case time.Time:
		target.SetValue(value)
	case Date:
		if value.Location == nil {
			target.SetValue(fmt.Sprintf("%04d-%02d-%02d", value.Year, value.Month, value.Day))
		} else {
			target.SetValue("")
		}
	}
	return nil
}

type JSONDateTimeParser struct{}
type JSONDateTimeFormatter struct{}

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

// "2006-01-02T15:04:05Z07:00" -> Note: uses a 24Hour based clock   "2006-01-02T15:04:05.999999999Z07:00"
func (JSONDateTimeFormatter) SetValue(target ls.Node, val interface{}) error {
	switch value := val.(type) {
	case time.Time:
		target.SetValue(value)
	case DateTime:
		if value.Nanoseconds != 0 {
			target.SetValue(fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d.%09dZ", value.Year, value.Month, value.Day,
				value.Hour, value.Minute, (value.Nanoseconds / 1000000000), value.Nanoseconds))
		}
		if value.Location == nil {
			target.SetValue(fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02dZ", value.Year, value.Month, value.Day,
				value.Hour, value.Minute, (value.Nanoseconds / 1000000000)))
		}
	}
	return nil
}

type JSONTimeParser struct{}
type JSONTimeFormatter struct{}

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

func (JSONTimeFormatter) SetValue(target ls.Node, val interface{}) error {
	switch value := val.(type) {
	case time.Time:
		target.SetValue(value)
	case TimeOfDay:
		if value.Location != nil {
			target.SetValue(fmt.Sprintf("%02d:%02d:%02dZ", value.Hour, value.Minute, (value.Nanoseconds / 1000000000)))
		} else if value.Nanoseconds != 0 {
			target.SetValue(fmt.Sprintf("%02d:%02d:%02d", value.Hour, value.Minute, (value.Nanoseconds / 1000000000)))
		} else {
			target.SetValue(fmt.Sprintf("%02d:%02d", value.Hour, value.Minute))
		}
	}
	return nil
}

type PatternDateTimeParser struct{}
type PatternDateTimeFormatter struct{}

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

func (PatternDateTimeFormatter) SetValue(target ls.Node, p PatternDateTimeParser) error {
	value, err := p.ParseValue(target)
	if err != nil {
		return err
	}
	target.SetValue(value.(time.Time).String())
	return nil
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
