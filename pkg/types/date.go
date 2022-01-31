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
	"strconv"
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
const Unix = "https://unixtime.org/"

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

type UnixTime struct {
	Seconds  int64
	Location *time.Location
}

func (u UnixTime) ToTime() time.Time {
	if u.Location == nil {
		return time.Unix(u.Seconds, 0)
	}
	return time.Unix(u.Seconds, 0).In(u.Location)
}

type UnixTimeNano struct {
	Nanoseconds int64
	Location    *time.Location
}

func (u UnixTimeNano) ToTime() time.Time {
	if u.Location == nil {
		return time.Unix(0, u.Nanoseconds)
	}
	return time.Unix(0, u.Nanoseconds).In(u.Location)
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
		return time.Date(0, 0, 0, int(t.Hour), int(t.Minute), int(t.Seconds), int(t.Nanoseconds), time.UTC)
	}
	return time.Date(0, 0, 0, int(t.Hour), int(t.Minute), int(t.Seconds), int(t.Nanoseconds), t.Location)
}

// GDay is XML Gregorian day part of date
type GDay int

// XSDGday can be used as a node-type to interpret the underlying value as a day (GDay)
var XSDGDayTerm = ls.NewTerm(XSD+"gDay", false, false, ls.OverrideComposition, struct {
	XSDGDayParser
}{
	XSDGDayParser: XSDGDayParser{},
})

// GMonth is XML Gregorian month part of date
type GMonth int

// XSDGMonth can be used as node-type to interpret the underlying value as a month (int)
var XSDGMonthTerm = ls.NewTerm(XSD+"gMonth", false, false, ls.OverrideComposition, struct {
	XSDGMonthParser
}{
	XSDGMonthParser: XSDGMonthParser{},
})

// GMonth is XML Gregorian year part of date
type GYear int

// XSDGYear can be used as a node-type to interpret the underlying value as a year value (int)
var XSDGYearTerm = ls.NewTerm(XSD+"gYear", false, false, ls.OverrideComposition, struct {
	XSDGYearParser
}{
	XSDGYearParser: XSDGYearParser{},
})

// GMonthDay is XML Gregorian part of Month/Day
type GMonthDay struct {
	Day   int
	Month int
}

// XSDMonthDay can be used as a node-type to interpret the underlying value as a MM-DD
var XSDGMonthDayTerm = ls.NewTerm(XSD+"gMonthDay", false, false, ls.OverrideComposition, struct {
	XSDGMonthDayParser
}{
	XSDGMonthDayParser: XSDGMonthDayParser{},
})

// GYearMonth is XML Gregorian part of Year/Month
type GYearMonth struct {
	Year  int
	Month int
}

// XSDGYearMonth can be used as a node-type to interpret the underlying value as a YYYY-MM
var XSDGYearMonthTerm = ls.NewTerm(XSD+"gYearMonth", false, false, ls.OverrideComposition, struct {
	XSDGYearMonthParser
}{
	XSDGYearMonthParser: XSDGYearMonthParser{},
})

// XSDDate is a node-type that identifies the underlying value as an XML date. The format is:
//
//  [-]CCYY-MM-DD[Z|(+|-)hh:mm]
var XSDDateTerm = ls.NewTerm(XSD+"date", false, false, ls.OverrideComposition, struct {
	XSDDateParser
}{
	XSDDateParser: XSDDateParser{},
})

// XSDTime is a node-type that identifies the underlying value as an XML time.
var XSDTimeTerm = ls.NewTerm(XSD+"time", false, false, ls.OverrideComposition, struct {
	XSDTimeParser
}{
	XSDTimeParser: XSDTimeParser{},
})

// XSDDateTime is a node-type that identifies the underlying value as an XML date-time value
var XSDDateTimeTerm = ls.NewTerm(XSD+"dateTime", false, false, ls.OverrideComposition, struct {
	XSDDateTimeParser
}{
	XSDDateTimeParser: XSDDateTimeParser{},
})

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

var UnixTimeTerm = ls.NewTerm(Unix+"time", false, false, ls.OverrideComposition, struct {
	UnixTimeParser
}{
	UnixTimeParser: UnixTimeParser{},
})

var UnixTimeNanoTerm = ls.NewTerm(Unix+"timeNano", false, false, ls.OverrideComposition, struct {
	UnixTimeNanoParser
}{
	UnixTimeNanoParser: UnixTimeNanoParser{},
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
	parseDate(string) (Date, error)
}

type dateTimeFormatter interface {
	parseDateTime(string) (DateTime, error)
}

type timeFormatter interface {
	parseTime(string) (TimeOfDay, error)
}

type unixFormatter interface {
	parseUnix(string) (UnixTime, error)
}

type unixNanoFormatter interface {
	parseUnixNano(string) (UnixTimeNano, error)
}

// Time?
// Unix?

func (f goFormat) parseDate(s string) (Date, error) {
	t, err := time.Parse(string(f), s)
	if err != nil {
		return Date{}, err
	}
	return Date{Year: t.Year(), Month: int(t.Month()), Day: t.Day()}, nil
}

func (f gomentFormat) parseDate(s string) (Date, error) {
	t, err := goment.New(s, string(f))
	if err != nil {
		return Date{}, err
	}
	return Date{Month: int(t.Month()), Day: t.Day(), Year: t.Year()}, nil
}

func (f goFormat) parseDateTime(s string) (DateTime, error) {
	t, err := time.Parse(string(f), s)
	if err != nil {
		return DateTime{}, err
	}
	return DateTime{Month: int(t.Month()), Day: t.Day(), Year: t.Year(),
		Nanoseconds: int64(t.Nanosecond()), Milliseconds: int64(t.Second() * 1000), Seconds: int64(t.Second()),
		Minute: int64(t.Minute()), Hour: int64(t.Hour())}, nil
}

func (f gomentFormat) parseDateTime(s string) (DateTime, error) {
	t, err := goment.New(s, string(f))
	if err != nil {
		return DateTime{}, err
	}
	return DateTime{Month: int(t.Month()), Day: t.Day(), Year: t.Year(),
		Nanoseconds: int64(t.Nanosecond()), Milliseconds: int64(t.Second() * 1000), Seconds: int64(t.Second()),
		Minute: int64(t.Minute()), Hour: int64(t.Hour())}, nil
}

func (f goFormat) parseTime(s string) (TimeOfDay, error) {
	t, err := time.Parse(string(f), s)
	if err != nil {
		return TimeOfDay{}, err
	}
	return TimeOfDay{Nanoseconds: int64(t.Nanosecond()), Milliseconds: int64(t.Second()) * 1000,
		Seconds: int64(t.Second()), Minute: int64(t.Minute()), Hour: int64(t.Hour())}, nil
}

func (f gomentFormat) parseTime(s string) (TimeOfDay, error) {
	t, err := goment.New(s, string(f))
	if err != nil {
		return TimeOfDay{}, err
	}
	return TimeOfDay{Nanoseconds: int64(t.Nanosecond()), Milliseconds: int64(t.Second()) * 1000,
		Seconds: int64(t.Second()), Minute: int64(t.Minute()), Hour: int64(t.Hour())}, nil
}

func (f goFormat) parseUnix(s string) (UnixTime, error) {
	t, err := time.Parse(string(f), s)
	if err != nil {
		return UnixTime{}, err
	}
	return UnixTime{Seconds: int64(t.Second())}, nil
}

func (f gomentFormat) parseUnix(s string) (UnixTime, error) {
	t, err := time.Parse(string(f), s)
	if err != nil {
		return UnixTime{}, err
	}
	return UnixTime{Seconds: int64(t.Second())}, nil
}

func (f goFormat) parseUnixNano(s string) (UnixTimeNano, error) {
	t, err := time.Parse(string(f), s)
	if err != nil {
		return UnixTimeNano{}, err
	}
	return UnixTimeNano{Nanoseconds: int64(t.Nanosecond())}, nil
}

func (f gomentFormat) parseUnixNano(s string) (UnixTimeNano, error) {
	t, err := time.Parse(string(f), s)
	if err != nil {
		return UnixTimeNano{}, err
	}
	return UnixTimeNano{Nanoseconds: int64(t.Nanosecond())}, nil
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
		node.SetValue(v.Format("2006-01-02"))
	case Date:
		if v.Location == nil {
			node.SetValue(v.ToTime().Format("2006-01-02"))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format("2006-01-02Z0700"))
		}
	case DateTime:
		if v.Location == nil {
			node.SetValue(v.ToTime().Format("2006-01-02"))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format("2006-01-02Z0700"))
		}
	case GDay:
		date, err := XSDDateParser.GetNodeValue(XSDDateParser{}, node)
		if err != nil {
			return err
		}
		x := date.(Date)
		x.Day = int(v)
		node.SetValue(x.ToTime().Format("2006-01-02"))
	case GMonth:
		date, err := XSDDateParser.GetNodeValue(XSDDateParser{}, node)
		if err != nil {
			return err
		}
		x := date.(Date)
		x.Month = int(v)
		node.SetValue(x.ToTime().Format("2006-01-02"))
	case GYear:
		date, err := XSDDateParser.GetNodeValue(XSDDateParser{}, node)
		if err != nil {
			return err
		}
		x := date.(Date)
		x.Year = int(v)
		node.SetValue(x.ToTime().Format("2006-01-02"))
	case GMonthDay:
		date, err := XSDDateParser.GetNodeValue(XSDDateParser{}, node)
		if err != nil {
			return err
		}
		x := date.(Date)
		x.Month = int(v.Month)
		x.Day = int(v.Day)
		node.SetValue(x.ToTime().Format("2006-01-02"))
	case GYearMonth:
		date, err := XSDDateParser.GetNodeValue(XSDDateParser{}, node)
		if err != nil {
			return err
		}
		x := date.(Date)
		x.Month = int(v.Month)
		x.Year = int(v.Year)
	case UnixTime:
		if v.Location == nil {
			node.SetValue(v.ToTime().Format("2006-01-02"))
		} else {
			node.SetValue(v.ToTime().Format("2006-01-02Z0700"))
		}
	case UnixTimeNano:
		if v.Location == nil {
			node.SetValue(v.ToTime().Format("2006-01-02"))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format("2006-01-02Z0700"))
		}

	default:
		return ls.ErrInvalidValue{ID: node.GetID(), Type: XSDDateTerm, Value: value}
	}
	return nil
}

type XSDDateTimeParser struct{}

func (XSDDateTimeParser) GetNodeValue(node ls.Node) (interface{}, error) {
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
	return genericDateTimeParse(value,
		goFormat("2006-1-2T15:04:05"),
		goFormat("2006-1-2T15:04:05Z"),
		// goFormat("2002-11-11T09:00:00:10-06:00"),
		// goFormat("2002-11-11T09:00:00:10+06:00"),
		// goFormat("2002-11-11T09:00:00:10.5"),
		gomentFormat("YYYY-MM-DDThh:mm:ss"),
		// gomentFormat("YYYY-MM-DDThh:mm:ssZ"),
	)
}

func (XSDDateTimeParser) SetNodeValue(node ls.Node, value interface{}) error {
	if value == nil {
		node.SetValue(nil)
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		node.SetValue(v.Format("2006-01-02T15:04:05Z"))
	case Date:
		if v.Location == nil {
			node.SetValue(v.ToTime().Format("2006-01-02T15:04:05Z"))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format("2006-01-02T15:04:05Z"))
		}
	case DateTime:
		if v.Location == nil {
			node.SetValue(v.ToTime().Format("2006-01-02T15:04:05Z"))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format("2006-01-02T15:04:05Z")) //2002-05-30T09:30:10-06:00
		}
	case TimeOfDay:
		if v.Location == nil {
			node.SetValue("2006-01-02T" + v.ToTime().Format("15:04:05Z"))
		} else {
			node.SetValue("2006-01-02T" + v.ToTime().In(v.Location).Format("15:04:05Z"))
		}
	case GDay:
		dateTime, err := XSDDateTimeParser.GetNodeValue(XSDDateTimeParser{}, node)
		if err != nil {
			return err
		}
		x := dateTime.(DateTime)
		x.Day = int(v)
		node.SetValue(x.ToTime().Format("2006-01-02T15:04:05Z"))
	case GMonth:
		dateTime, err := XSDDateTimeParser.GetNodeValue(XSDDateTimeParser{}, node)
		if err != nil {
			return err
		}
		x := dateTime.(DateTime)
		x.Month = int(v)
		node.SetValue(x.ToTime().Format("2006-01-02T15:04:05Z"))
	case GYear:
		dateTime, err := XSDDateTimeParser.GetNodeValue(XSDDateTimeParser{}, node)
		if err != nil {
			return err
		}
		x := dateTime.(DateTime)
		x.Year = int(v)
		node.SetValue(x.ToTime().Format("2006-01-02T15:04:05Z"))
	case GMonthDay:
		dateTime, err := XSDDateTimeParser.GetNodeValue(XSDDateTimeParser{}, node)
		if err != nil {
			return err
		}
		x := dateTime.(DateTime)
		x.Month = int(v.Month)
		x.Day = int(v.Day)
		node.SetValue(x.ToTime().Format("2006-01-02T15:04:05Z"))
	case GYearMonth:
		dateTime, err := XSDDateTimeParser.GetNodeValue(XSDDateTimeParser{}, node)
		if err != nil {
			return err
		}
		x := dateTime.(DateTime)
		x.Year = int(v.Year)
		x.Month = int(v.Month)
		node.SetValue(x.ToTime().Format("2006-01-02T15:04:05Z"))
	case UnixTime:
		if v.Location == nil {
			node.SetValue(fmt.Sprintf("%02d", v.Seconds))
		} else {
			node.SetValue(v.ToTime().Format("2002-05-30T09:30:10.5-06:00"))
		}
	case UnixTimeNano:
		if v.Location == nil {
			node.SetValue(v.ToTime().Format("2002-05-30T09:30:10.5"))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format("2002-05-30T09:30:10.5-06:00"))
		}
	default:
		return ls.ErrInvalidValue{ID: node.GetID(), Type: XSDDateTimeTerm, Value: value}
	}
	return nil
}

type XSDTimeParser struct{}

func (XSDTimeParser) GetNodeValue(node ls.Node) (interface{}, error) {
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
	return genericTimeParse(value,
		goFormat("09:00:00"),
		goFormat("09:00:00.5"),
		goFormat("09:00:00Z"),
		goFormat("09:00:00-06:00"),
		goFormat("09:00:00+06:00"),
		gomentFormat("hh:mm:ss"))
}

func (XSDTimeParser) SetNodeValue(node ls.Node, value interface{}) error {
	if value == nil {
		node.SetValue(nil)
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		node.SetValue(v.Format("15:04:05"))
	case DateTime:
		if v.Location == nil {
			node.SetValue(v.ToTime().Format("09:00:00Z"))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format("09:00:00Z"))
		}
	case UnixTime:
		if v.Location == nil {
			node.SetValue(strconv.FormatInt(v.ToTime().Unix(), 10))
		} else {
			node.SetValue(strconv.FormatInt((v.ToTime().In(v.Location).Unix()), 10))
		}
	case UnixTimeNano:
	default:
		return ls.ErrInvalidValue{ID: node.GetID(), Type: XSDTimeTerm, Value: value}
	}
	return nil
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
			node.SetValue(v.ToTime().In(v.Location).Format("2006-01-02Z0700"))
		}
	case DateTime:
		if v.Location == nil {
			node.SetValue(v.ToTime().Format("2006-01-02"))
			// target.SetValue(fmt.Sprintf("%04d-%02d-%02d"+"T"+"%02d:%02d:%02d", value.Year, value.Month, value.Day, value.Hour, value.Minute, (value.Nanoseconds / 1000000000)))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format("2006-01-02Z0700"))
		}
	case TimeOfDay:
		if v.Location == nil {
			node.SetValue(v.ToTime().Format("2006-01-02"))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format("2006-01-02Z0700"))
		}
	case GDay:
		date, err := JSONDateParser.GetNodeValue(JSONDateParser{}, node)
		if err != nil {
			return err
		}
		x := date.(Date)
		x.Day = int(v)
		node.SetValue(x.ToTime().Format("2006-01-02"))
	case GMonth:
		date, err := JSONDateParser.GetNodeValue(JSONDateParser{}, node)
		if err != nil {
			return err
		}
		x := date.(Date)
		x.Month = int(v)
		node.SetValue(x.ToTime().Format("2006-01-02"))
	case GYear:
		date, err := JSONDateParser.GetNodeValue(JSONDateParser{}, node)
		if err != nil {
			return err
		}
		x := date.(Date)
		x.Year = int(v)
		node.SetValue(x.ToTime().Format("2006-01-02"))
	case GMonthDay:
		date, err := JSONDateParser.GetNodeValue(JSONDateParser{}, node)
		if err != nil {
			return err
		}
		x := date.(Date)
		x.Month = int(v.Month)
		x.Day = int(v.Day)
		node.SetValue(x.ToTime().Format("2006-01-02"))
	case GYearMonth:
		date, err := JSONDateParser.GetNodeValue(JSONDateParser{}, node)
		if err != nil {
			return err
		}
		x := date.(Date)
		x.Year = int(v.Year)
		x.Month = int(v.Month)
		node.SetValue(x.ToTime().Format("2006-01-02"))
	case UnixTime:
		if v.Location == nil {
			node.SetValue(v.ToTime().Format("2006-01-02"))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format("2006-01-02Z0700"))
		}
	case UnixTimeNano:
		if v.Location == nil {
			node.SetValue(v.ToTime().Format("2006-01-02"))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format("2006-01-02Z0700"))
		}
	default:
		return ls.ErrInvalidValue{ID: node.GetID(), Type: JSONDateTerm, Value: value}
	}
	return nil
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
	return genericDateTimeParse(value, goFormat(time.RFC3339), goFormat(time.RFC3339Nano))
}

// "2006-01-02T11:11:11Z07:00" -> Note: uses a 24Hour based clock   "2006-01-02T11:11:11.999999999Z07:00"
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
			node.SetValue(v.ToTime().Format("2006-01-02T15:04:05Z"))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format("2006-01-02T15:04:05Z"))
		}
	case DateTime:
		if v.Location == nil && v.Nanoseconds == 0 {
			node.SetValue(v.ToTime().Format("2006-01-02T15:04:05Z"))
		} else if v.Location != nil && v.Nanoseconds != 0 {
			node.SetValue(v.ToTime().In(v.Location).Format(time.RFC3339Nano))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format(time.RFC3339))
		}
	case TimeOfDay:
		if v.Location == nil {
			node.SetValue("2006-01-02T" + v.ToTime().Format("15:04:05Z"))
		} else {
			node.SetValue("2006-01-02T" + v.ToTime().In(v.Location).Format("15:04:05Z"))
		}
	case UnixTime:
		if v.Location == nil {
			node.SetValue(v.ToTime().Format("2006-01-02T15:04:05Z"))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format(time.RFC3339))
		}
	case UnixTimeNano:
		if v.Location == nil {
			node.SetValue(v.ToTime().Format("2006-01-02T15:04:05Z"))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format(time.RFC3339Nano))
		}

	default:
		return ls.ErrInvalidValue{ID: node.GetID(), Type: JSONDateTimeTerm, Value: value}
	}
	return nil
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
	return genericTimeParse(value, gomentFormat("HH:mm:ssZ"),
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
		node.SetValue(v.Format("HH:mm:ssZ")) // 11:11:11Z
	case Date:
		if v.Location == nil {
			node.SetValue(v.ToTime().Format("09:00:00"))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format("09:00:00Z"))
		}
	case DateTime:
		if v.Location == nil {
			//node.SetValue(v.ToTime().Format(fmt.Sprintf("%02d:%02d:%02d", v.Hour, v.Minute, v.Seconds)))
			node.SetValue(v.ToTime().Format("09:00:00Z"))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format("09:00:00Z"))
		}
	case TimeOfDay:
		if v.Location == nil && v.Nanoseconds == 0 {
			node.SetValue(v.ToTime().Format("HH:mm"))
		} else if v.Location == nil && v.Nanoseconds != 0 {
			node.SetValue(v.ToTime().Format("HH:mm:ss"))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format("HH:mm:ssZ"))
		}
	case UnixTime:
		if v.Location == nil {
			node.SetValue(v.ToTime().Format("ss"))
		} else {
			node.SetValue(v.ToTime().Format("HH:mm:ssZ"))
		}
	case UnixTimeNano:
		if v.Location == nil {
			node.SetValue(v.ToTime().Format("HH:mm:ss"))
		} else {
			node.SetValue(v.ToTime().Format("HH:mm:ssZ"))
		}
	default:
		return ls.ErrInvalidValue{ID: node.GetID(), Type: JSONTimeTerm, Value: value}
	}
	return nil
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
		if _, ok := node.GetProperties()[GoTimeFormatTerm]; ok {
			node.SetValue(v.Format(node.GetProperties()[GoTimeFormatTerm].AsString()))
		} else {
			node.SetValue(v.Format(node.GetProperties()[MomentTimeFormatTerm].AsString()))
		}
	case Date:
		if _, ok := node.GetProperties()[GoTimeFormatTerm]; ok {
			if v.Location == nil {
				node.SetValue(v.ToTime().Format(node.GetProperties()[GoTimeFormatTerm].AsString()))
			} else {
				node.SetValue(v.ToTime().In(v.Location).Format(node.GetProperties()[GoTimeFormatTerm].AsString()))
			}
		} else {
			if v.Location == nil {
				node.SetValue(v.ToTime().Format(node.GetProperties()[MomentTimeFormatTerm].AsString()))
			} else {
				node.SetValue(v.ToTime().In(v.Location).Format(node.GetProperties()[MomentTimeFormatTerm].AsString()))
			}
		}
	case DateTime:
		if _, ok := node.GetProperties()[GoTimeFormatTerm]; ok {
			if v.Location == nil {
				node.SetValue(v.ToTime().Format(node.GetProperties()[GoTimeFormatTerm].AsString()))
			} else {
				node.SetValue(v.ToTime().In(v.Location).Format(node.GetProperties()[GoTimeFormatTerm].AsString()))
			}
		} else {
			if v.Location == nil {
				node.SetValue(v.ToTime().Format(node.GetProperties()[MomentTimeFormatTerm].AsString()))
			} else {
				node.SetValue(v.ToTime().In(v.Location).Format(node.GetProperties()[MomentTimeFormatTerm].AsString()))
			}
		}
	case TimeOfDay:
		if _, ok := node.GetProperties()[GoTimeFormatTerm]; ok {
			if v.Location == nil {
				node.SetValue(v.ToTime().Format(node.GetProperties()[GoTimeFormatTerm].AsString()))
			} else {
				node.SetValue(v.ToTime().In(v.Location).Format(node.GetProperties()[GoTimeFormatTerm].AsString()))
			}
		} else {
			if v.Location == nil {
				node.SetValue(v.ToTime().Format(node.GetProperties()[MomentTimeFormatTerm].AsString()))
			} else {
				node.SetValue(v.ToTime().In(v.Location).Format(node.GetProperties()[MomentTimeFormatTerm].AsString()))
			}
		}
	default:
		return ls.ErrInvalidValue{ID: node.GetID(), Type: PatternDateTimeTerm, Value: value}
	}
	return nil
}

type XSDGDayParser struct{}

func (XSDGDayParser) GetNodeValue(node ls.Node) (interface{}, error) {
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
	x, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}
	return GDay(x), nil
	//return genericDateParse(value, goFormat("02"), goFormat("2"), gomentFormat("DD"))
}

func (XSDGDayParser) SetNodeValue(node ls.Node, value interface{}) error {
	if value == nil {
		node.SetValue(nil)
		return nil
	}
	switch v := value.(type) {
	// case time.Time:
	// 	node.SetValue(v.Format("02")) // DD
	case Date:
		node.SetValue(strconv.Itoa((v.Day)))
	case DateTime:
		node.SetValue(strconv.Itoa((v.Day)))
	case GDay:
		node.SetValue(strconv.Itoa(int(v)))
	case GMonthDay:
		node.SetValue(strconv.Itoa(int(v.Day)))
	default:
		return ls.ErrInvalidValue{ID: node.GetID(), Type: XSDGDayTerm, Value: value}
	}
	return nil
}

type XSDGMonthParser struct{}

func (XSDGMonthParser) GetNodeValue(node ls.Node) (interface{}, error) {
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
	return genericDateParse(value, goFormat("01"), gomentFormat("MM"))
}

func (XSDGMonthParser) SetNodeValue(node ls.Node, value interface{}) error {
	if value == nil {
		node.SetValue(nil)
		return nil
	}
	switch v := value.(type) {
	// case time.Time:
	// 	node.SetValue(v.Format("01")) // MM
	case Date:
		node.SetValue(v.ToTime().Format("01"))
	case DateTime:
		node.SetValue(v.ToTime().Format("01"))
	case GMonth:
		node.SetValue(strconv.Itoa(int(v)))
	case GMonthDay:
		node.SetValue(strconv.Itoa(int(v.Month)))
	case GYearMonth:
		node.SetValue(strconv.Itoa(int(v.Month)))
	default:
		return ls.ErrInvalidValue{ID: node.GetID(), Type: XSDGMonthTerm, Value: value}
	}
	return nil
}

type XSDGMonthDayParser struct{}

func (XSDGMonthDayParser) GetNodeValue(node ls.Node) (interface{}, error) {
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
	return genericDateParse(value, goFormat("1-2"), goFormat("01-2"), goFormat("01-02"), gomentFormat("MM-DD"))
}

func (XSDGMonthDayParser) SetNodeValue(node ls.Node, value interface{}) error {
	if value == nil {
		node.SetValue(nil)
		return nil
	}
	switch v := value.(type) {
	// case time.Time:
	// 	node.SetValue(v.Format("01")) // MM
	case Date:
		node.SetValue(v.ToTime().Format("01-02"))
	case DateTime:
		node.SetValue(v.ToTime().Format("01-02"))
	case GDay:
		node.SetValue(strconv.Itoa(int(v)))
	case GMonth:
		node.SetValue(strconv.Itoa(int(v)))
	case GMonthDay:
		node.SetValue(fmt.Sprintf("%02d-%02d", v.Month, v.Day))
	case GYearMonth:
		node.SetValue(strconv.Itoa(int(v.Month)))
	default:
		return ls.ErrInvalidValue{ID: node.GetID(), Type: XSDGMonthDayTerm, Value: value}
	}
	return nil
}

type XSDGYearParser struct{}

func (XSDGYearParser) GetNodeValue(node ls.Node) (interface{}, error) {
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
	return genericDateParse(value, goFormat("2006"), gomentFormat("YYYY"))
}

func (XSDGYearParser) SetNodeValue(node ls.Node, value interface{}) error {
	if value == nil {
		node.SetValue(nil)
		return nil
	}
	switch v := value.(type) {
	// case time.Time:
	// 	node.SetValue(v.Format("2006"))
	case Date:
		node.SetValue(v.ToTime().Format("2006"))
	case DateTime:
		node.SetValue(v.ToTime().Format("2006"))
	case GYear:
		node.SetValue(strconv.Itoa(int(v)))
	case GYearMonth:
		node.SetValue(strconv.Itoa(int(v.Year)))
	default:
		return ls.ErrInvalidValue{ID: node.GetID(), Type: XSDGYearTerm, Value: value}
	}
	return nil
}

type XSDGYearMonthParser struct{}

func (XSDGYearMonthParser) GetNodeValue(node ls.Node) (interface{}, error) {
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
	return genericDateParse(value, goFormat("2006-01"), gomentFormat("YY-MM"))
}

func (XSDGYearMonthParser) SetNodeValue(node ls.Node, value interface{}) error {
	if value == nil {
		node.SetValue(nil)
		return nil
	}
	switch v := value.(type) {
	// case time.Time:
	// 	node.SetValue(v.Format("01")) // MM
	case Date:
		node.SetValue(v.ToTime().Format("2006-01"))
	case DateTime:
		node.SetValue(v.ToTime().Format("2006-01"))
	case GMonth:
		node.SetValue(strconv.Itoa(int(v)))
	case GYear:
		node.SetValue(strconv.Itoa(int(v)))
	case GYearMonth:
		node.SetValue(fmt.Sprintf("%02d-%02d", v.Year, v.Month))
	default:
		return ls.ErrInvalidValue{ID: node.GetID(), Type: XSDGYearMonthTerm, Value: value}
	}
	return nil
}

type UnixTimeParser struct{}

func (UnixTimeParser) GetNodeValue(node ls.Node) (interface{}, error) {
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
	// x, err := strconv.Atoi(value)
	// if err != nil {
	// 	return nil, err
	// }
	//return UnixTime{int64(x), nil}, nil
	return genericUnixParse(value,
		goFormat("1000000000"),
		goFormat("1000000000000000000"),
		goFormat("1970-01-01 00:00:00"),
		goFormat("1970-01-01 00:00:00 +0000 UTC"),
		goFormat(time.RFC3339),
		goFormat(time.RFC3339Nano),
	)
}

func (UnixTimeParser) SetNodeValue(node ls.Node, value interface{}) error {
	if value == nil {
		node.SetValue(nil)
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		node.SetValue(strconv.FormatInt(v.Unix(), 10))
	case DateTime:
		if v.Location == nil {
			node.SetValue(strconv.FormatInt(v.ToTime().Unix(), 10))
		} else {
			node.SetValue(strconv.FormatInt(v.ToTime().In(v.Location).Unix(), 10))
		}
	case TimeOfDay:
		if v.Location == nil {
			t := time.Date(1970, time.January, 1, int(v.Hour), int(v.Minute), int(v.Seconds), int(v.Nanoseconds), time.UTC)
			node.SetValue(strconv.FormatInt(t.Unix(), 10))
		} else {
			t := time.Date(1970, time.January, 1, int(v.Hour), int(v.Minute), int(v.Seconds), int(v.Nanoseconds), v.Location)
			node.SetValue(strconv.FormatInt(t.Unix(), 10))
		}
	case UnixTimeNano:
		if v.Location == nil {
			node.SetValue(fmt.Sprintf("%d", v.ToTime().UnixNano()))
		} else {
			node.SetValue(strconv.FormatInt(v.ToTime().In(v.Location).UnixNano(), 10))
		}
	default:
		return ls.ErrInvalidValue{ID: node.GetID(), Type: UnixTimeTerm, Value: value}
	}
	return nil
}

type UnixTimeNanoParser struct{}

func (UnixTimeNanoParser) GetNodeValue(node ls.Node) (interface{}, error) {
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
	return genericUnixNanoParse(value, goFormat(time.RFC3339), goFormat(time.RFC3339Nano))
}

func (UnixTimeNanoParser) SetNodeValue(node ls.Node, value interface{}) error {
	if value == nil {
		node.SetValue(nil)
		return nil
	}
	switch v := value.(type) {
	case time.Time:
		node.SetValue(strconv.FormatInt(v.UnixNano(), 10))
	case DateTime:
		if v.Location == nil {
			node.SetValue(strconv.FormatInt(v.ToTime().UnixNano(), 10))
		} else {
			node.SetValue(strconv.FormatInt(v.ToTime().In(v.Location).UnixNano(), 10))
		}
	case TimeOfDay:
		if v.Location == nil {
			t := time.Date(1970, time.January, 1, int(v.Hour), int(v.Minute), int(v.Seconds), int(v.Nanoseconds), time.UTC)
			node.SetValue(strconv.FormatInt(t.UnixNano(), 10))
		} else {
			t := time.Date(1970, time.January, 1, int(v.Hour), int(v.Minute), int(v.Seconds), int(v.Nanoseconds), v.Location)
			node.SetValue(strconv.FormatInt(t.UnixNano(), 10))
		}
	case UnixTimeNano:
		if v.Location == nil {
			node.SetValue(v.ToTime().Format(time.RFC3339Nano))
		} else {
			node.SetValue(v.ToTime().In(v.Location).Format(time.RFC3339Nano))
		}
	default:
		return ls.ErrInvalidValue{ID: node.GetID(), Type: UnixTimeTerm, Value: value}
	}
	return nil
}

// genericDateParse parses a node value using the given format(s)
func genericDateParse(value string, format ...dateFormatter) (Date, error) {
	for _, f := range format {
		t, err := f.parseDate(value)
		if err == nil {
			return t, nil
		}
	}
	return Date{}, ErrCannotParseTemporalValue(value)
}

func genericDateTimeParse(value string, format ...dateTimeFormatter) (DateTime, error) {
	for _, f := range format {
		t, err := f.parseDateTime(value)
		if err == nil {
			return t, nil
		}
	}
	return DateTime{}, ErrCannotParseTemporalValue(value)
}

func genericTimeParse(value string, format ...timeFormatter) (TimeOfDay, error) {
	for _, f := range format {
		t, err := f.parseTime(value)
		if err == nil {
			return t, nil
		}
	}
	return TimeOfDay{}, ErrCannotParseTemporalValue(value)
}

func genericUnixParse(value string, format ...unixFormatter) (UnixTime, error) {
	for _, f := range format {
		t, err := f.parseUnix(value)
		if err == nil {
			return t, nil
		}
	}
	return UnixTime{}, ErrCannotParseTemporalValue(value)
}

func genericUnixNanoParse(value string, format ...unixNanoFormatter) (UnixTimeNano, error) {
	for _, f := range format {
		t, err := f.parseUnixNano(value)
		if err == nil {
			return t, nil
		}
	}
	return UnixTimeNano{}, ErrCannotParseTemporalValue(value)
}
