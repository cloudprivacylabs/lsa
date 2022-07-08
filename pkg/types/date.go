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

	"github.com/araddon/dateparse"
	"github.com/nleeper/goment"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
)

type ErrCannotParseTemporalValue string

type ErrIncompatibleTypes struct {
	cmp1, cmp2 interface{}
}

func (e ErrCannotParseTemporalValue) Error() string {
	return "Cannot parse temporal value: " + string(e)
}

func (e ErrIncompatibleTypes) Error() string {
	return fmt.Sprintf("Incompatible types: %v, %v", e.cmp1, e.cmp2)
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

func NewDate(t time.Time) Date {
	return Date{
		Month:    int(t.Month()),
		Day:      t.Day(),
		Year:     t.Year(),
		Location: t.Location(),
	}
}

func (d Date) ToTime() time.Time {
	if d.Location == nil {
		return time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
	}
	return time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, d.Location)
}

func (d Date) String() string {
	if d.Location == nil {
		return d.ToTime().Format("2006-01-02")
	}
	return d.ToTime().Format("2006-01-02 MST")
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

func (u UnixTime) String() string {
	return fmt.Sprint(u.ToTime().Unix())
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

func (u UnixTimeNano) String() string {
	return fmt.Sprint(u.ToTime().UnixNano())
}

// try to convert to go native time with function ToGoTime then pass result as parameter
func ToGomentTime(time time.Time) (*goment.Goment, error) {
	t, err := goment.New(time)
	if err != nil {
		return nil, err
	}
	return t, nil
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

func (dt DateTime) String() string {
	return dt.ToTime().Format(time.RFC3339)
}

func NewDateTime(t time.Time) DateTime {
	return DateTime{
		Month:        int(t.Month()),
		Day:          t.Day(),
		Year:         t.Year(),
		Nanoseconds:  int64(t.Nanosecond()),
		Milliseconds: int64(t.Second() * 1000),
		Seconds:      int64(t.Second()),
		Minute:       int64(t.Minute()),
		Hour:         int64(t.Hour()),
		Location:     t.Location(),
	}
}

type TimeOfDay struct {
	Nanoseconds  int64
	Milliseconds int64
	Seconds      int64
	Minute       int64
	Hour         int64
	Location     *time.Location
}

func NewTimeOfDay(t time.Time) TimeOfDay {
	return TimeOfDay{
		Nanoseconds:  int64(t.Nanosecond()),
		Milliseconds: int64(t.Second() * 1000),
		Seconds:      int64(t.Second()),
		Minute:       int64(t.Minute()),
		Hour:         int64(t.Hour()),
		Location:     t.Location(),
	}
}

func (t TimeOfDay) ToTime() time.Time {
	if t.Location == nil {
		return time.Date(0, 0, 0, int(t.Hour), int(t.Minute), int(t.Seconds), int(t.Nanoseconds), time.UTC)
	}
	return time.Date(0, 0, 0, int(t.Hour), int(t.Minute), int(t.Seconds), int(t.Nanoseconds), t.Location)
}

func (t TimeOfDay) String() string {
	return t.ToTime().Format("15:04:05.999999999Z07:00")
}

// GDay is XML Gregorian day part of date
type GDay int

func (g GDay) String() string { return strconv.Itoa(int(g)) }

// XSDGday can be used as a node-type to interpret the underlying value as a day (GDay)
var XSDGDayTerm = ls.NewTerm(XSD, "gDay", false, false, ls.OverrideComposition, struct {
	XSDGDayParser
}{
	XSDGDayParser: XSDGDayParser{},
}, "xsd:gDay", "xs:gDay")

// GMonth is XML Gregorian month part of date
type GMonth int

func (g GMonth) String() string { return strconv.Itoa(int(g)) }

// XSDGMonth can be used as node-type to interpret the underlying value as a month (int)
var XSDGMonthTerm = ls.NewTerm(XSD, "gMonth", false, false, ls.OverrideComposition, struct {
	XSDGMonthParser
}{
	XSDGMonthParser: XSDGMonthParser{},
}, "xsd:gMonth", "xs:gMonth")

// GMonth is XML Gregorian year part of date
type GYear int

func (g GYear) String() string { return strconv.Itoa(int(g)) }

// XSDGYear can be used as a node-type to interpret the underlying value as a year value (int)
var XSDGYearTerm = ls.NewTerm(XSD, "gYear", false, false, ls.OverrideComposition, struct {
	XSDGYearParser
}{
	XSDGYearParser: XSDGYearParser{},
}, "xsd:gYear", "xs:gYear")

// GMonthDay is XML Gregorian part of Month/Day
type GMonthDay struct {
	Day   int
	Month int
}

func (g GMonthDay) String() string { return fmt.Sprintf("%02d-%02d", g.Month, g.Day) }

// XSDMonthDay can be used as a node-type to interpret the underlying value as a MM-DD
var XSDGMonthDayTerm = ls.NewTerm(XSD, "gMonthDay", false, false, ls.OverrideComposition, struct {
	XSDGMonthDayParser
}{
	XSDGMonthDayParser: XSDGMonthDayParser{},
}, "xsd:gMonthDay", "xs:gMonthDay")

// GYearMonth is XML Gregorian part of Year/Month
type GYearMonth struct {
	Year  int
	Month int
}

func (g GYearMonth) String() string { return fmt.Sprintf("%04d-%02d", g.Year, g.Month) }

// XSDGYearMonth can be used as a node-type to interpret the underlying value as a YYYY-MM
var XSDGYearMonthTerm = ls.NewTerm(XSD, "gYearMonth", false, false, ls.OverrideComposition, struct {
	XSDGYearMonthParser
}{
	XSDGYearMonthParser: XSDGYearMonthParser{},
}, "xsd:gYearMonth", "xs:gYearMonth")

// XSDDate is a node-type that identifies the underlying value as an XML date. The format is:
//
//  [-]CCYY-MM-DD[Z|(+|-)hh:mm]
var XSDDateTerm = ls.NewTerm(XSD, "date", false, false, ls.OverrideComposition, struct {
	XSDDateParser
}{
	XSDDateParser: XSDDateParser{},
}, "xsd:date", "xs:date")

// XSDTime is a node-type that identifies the underlying value as an XML time.
var XSDTimeTerm = ls.NewTerm(XSD, "time", false, false, ls.OverrideComposition, struct {
	XSDTimeParser
}{
	XSDTimeParser: XSDTimeParser{},
}, "xsd:time", "xs:time")

// XSDDateTime is a node-type that identifies the underlying value as an XML date-time value
var XSDDateTimeTerm = ls.NewTerm(XSD, "dateTime", false, false, ls.OverrideComposition, struct {
	XSDDateTimeParser
}{
	XSDDateTimeParser: XSDDateTimeParser{},
}, "xsd:dateTime", "xs:dateTime")

// JSONDate is a node-type that identifies the underlying value as a JSON date value
//
//  YYYY-MM-DD
var JSONDateTerm = ls.NewTerm(JSON, "date", false, false, ls.OverrideComposition, struct {
	JSONDateParser
}{
	JSONDateParser: JSONDateParser{},
}, "json:date")

// JSONDateTime is a node-type that identifies the underlying value as
// a JSON datetime value, RFC3339 or RFC3339Nano
//
// YYYY-MM-DDTHH:mm:ssZ
// YYYY-MM-DDTHH:mm:ss.00000Z
var JSONDateTimeTerm = ls.NewTerm(JSON, "date-time", false, false, ls.OverrideComposition, struct {
	JSONDateTimeParser
}{
	JSONDateTimeParser: JSONDateTimeParser{},
}, "json:date-time")

// JSONTime is a node-type that identifies the underlying value as a
// JSON time value
//
//   HH:mm
//   HH:mm:ss
//   HH:mm:ssZ
var JSONTimeTerm = ls.NewTerm(JSON, "time", false, false, ls.OverrideComposition, struct {
	JSONTimeParser
}{
	JSONTimeParser: JSONTimeParser{},
}, "json:time")

var UnixTimeTerm = ls.NewTerm(Unix, "time", false, false, ls.OverrideComposition, struct {
	UnixTimeParser
}{
	UnixTimeParser: UnixTimeParser{},
}, "unix:time")

var UnixTimeNanoTerm = ls.NewTerm(Unix, "timeNano", false, false, ls.OverrideComposition, struct {
	UnixTimeNanoParser
}{
	UnixTimeNanoParser: UnixTimeNanoParser{},
}, "unix:timeNano")

var PatternDateTimeTerm = ls.NewTerm(ls.LS, "dateTime", false, false, ls.OverrideComposition, struct {
	PatternDateTimeParser
}{
	PatternDateTimeParser: PatternDateTimeParser{},
}, "ls:dateTime")

var PatternDateTerm = ls.NewTerm(ls.LS, "date", false, false, ls.OverrideComposition, struct {
	PatternDateParser
}{
	PatternDateParser: PatternDateParser{},
}, "ls:date")

var PatternTimeTerm = ls.NewTerm(ls.LS, "time", false, false, ls.OverrideComposition, struct {
	PatternTimeParser
}{
	PatternTimeParser: PatternTimeParser{},
}, "ls:time")

var GoTimeFormatTerm = ls.NewTerm(ls.LS, "goTimeFormat", false, false, ls.SetComposition, nil, "ls:goTimeFormat")
var MomentTimeFormatTerm = ls.NewTerm(ls.LS, "momentTimeFormat", false, false, ls.SetComposition, nil, "ls:momentTimeFormat")

type XSDDateParser struct{}

type goFormat string
type gomentFormat string

type dateFormatter interface {
	parseDate(string) (Date, error)
	formatDate(Date) (string, error)
}

type dateTimeFormatter interface {
	parseDateTime(string) (DateTime, error)
	formatDateTime(DateTime) (string, error)
}

type timeFormatter interface {
	parseTime(string) (TimeOfDay, error)
	formatTime(TimeOfDay) (string, error)
}

type goTimeFormatter interface {
	formatGoTime(time.Time) (string, error)
}

// type unixFormatter interface {
// 	parseUnix(string) (UnixTime, error)
// }

// type unixNanoFormatter interface {
// 	parseUnixNano(string) (UnixTimeNano, error)
// }

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
	return Date{Month: int(t.Month()), Day: t.Date(), Year: t.Year()}, nil
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
	return DateTime{Month: int(t.Month()), Day: t.Date(), Year: t.Year(),
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

func (f goFormat) formatDate(in Date) (string, error) {
	return in.ToTime().Format(string(f)), nil
}

func (f gomentFormat) formatDate(in Date) (string, error) {
	gmt, err := goment.New(in.ToTime())
	if err != nil {
		return "", err
	}
	return gmt.Format(string(f)), nil
}

func (f goFormat) formatDateTime(in DateTime) (string, error) {
	return in.ToTime().Format(string(f)), nil
}

func (f gomentFormat) formatDateTime(in DateTime) (string, error) {
	gmt, err := goment.New(in.ToTime())
	if err != nil {
		return "", err
	}
	return gmt.Format(string(f)), nil
}

func (f goFormat) formatTime(in TimeOfDay) (string, error) {
	return in.ToTime().Format(string(f)), nil

}

func (f goFormat) formatGoTime(in time.Time) (string, error) {
	return in.Format(string(f)), nil

}

func (f gomentFormat) formatGoTime(in time.Time) (string, error) {
	gmt, err := goment.New(in)
	if err != nil {
		return "", err
	}
	return gmt.Format(string(f)), nil
}

func (f gomentFormat) formatTime(in TimeOfDay) (string, error) {
	var gmt *goment.Goment
	var err error
	if in.Location == nil {
		gmt, err = goment.New(in.ToTime())
	} else {
		gmt, err = goment.New(in.ToTime().In(in.Location))
	}
	if err != nil {
		return "", err
	}
	return gmt.Format(string(f)), nil
}

// GetNativeValue parses an XSDDate value.
//
//  [-]CCYY-MM-DD[Z|(+|-)hh:mm]
func (XSDDateParser) GetNativeValue(value string, node graph.Node) (interface{}, error) {
	if len(value) == 0 {
		return nil, nil
	}
	return genericDateParse(value, goFormat("2006-1-2"), goFormat("2006-1-2Z"), goFormat("2006-1-2Z0700"), gomentFormat("YYYY-MM-DDZ"))
}

// FormatNativeValue gets a target node and it's go native value, and returns
// the value of the target node to an XSDDate
func (XSDDateParser) FormatNativeValue(newValue, oldValue interface{}, node graph.Node) (string, error) {
	if newValue == nil {
		ls.RemoveRawNodeValue(node)
		return "", nil
	}
	var oldDate Date
	var ok bool
	switch v := newValue.(type) {
	case time.Time:
		return v.Format("2006-01-02"), nil
	case Date:
		if v.Location == nil {
			return v.ToTime().Format("2006-01-02"), nil
		}
		return v.ToTime().In(v.Location).Format("2006-01-02Z0700"), nil
	case DateTime:
		if v.Location == nil {
			return v.ToTime().Format("2006-01-02"), nil
		}
		return v.ToTime().In(v.Location).Format("2006-01-02Z0700"), nil
	case GDay:
		if oldValue != nil {
			oldDate, ok = oldValue.(Date)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: XSDDateTerm, Value: oldValue, Msg: "Not a Date"}
			}
		}
		oldDate.Day = int(v)
		return oldDate.ToTime().Format("2006-01-02"), nil
	case GMonth:
		if oldValue != nil {
			oldDate, ok = oldValue.(Date)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: XSDDateTerm, Value: oldValue, Msg: "Not a Date"}
			}
		}
		oldDate.Month = int(v)
		return oldDate.ToTime().Format("2006-01-02"), nil
	case GYear:
		if oldValue != nil {
			oldDate, ok = oldValue.(Date)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: XSDDateTerm, Value: oldValue, Msg: "Not a Date"}
			}
		}
		oldDate.Year = int(v)
		return oldDate.ToTime().Format("2006-01-02"), nil
	case GMonthDay:
		if oldValue != nil {
			oldDate, ok = oldValue.(Date)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: XSDDateTerm, Value: oldValue, Msg: "Not a Date"}
			}
		}
		oldDate.Month = int(v.Month)
		oldDate.Day = int(v.Day)
		return oldDate.ToTime().Format("2006-01-02"), nil
	case GYearMonth:
		if oldValue != nil {
			oldDate, ok = oldValue.(Date)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: XSDDateTerm, Value: oldValue, Msg: "Not a Date"}
			}
		}
		oldDate.Month = int(v.Month)
		oldDate.Year = int(v.Year)
		return oldDate.ToTime().Format("2006-01-02"), nil
	case UnixTime:
		if v.Location == nil {
			return v.ToTime().Format("2006-01-02"), nil
		}
		return v.ToTime().Format("2006-01-02Z0700"), nil
	case UnixTimeNano:
		if v.Location == nil {
			return v.ToTime().Format("2006-01-02"), nil
		}
		return v.ToTime().In(v.Location).Format("2006-01-02Z0700"), nil
	default:
		return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: XSDDateTerm, Value: newValue}
	}
	return "", nil
}

type XSDDateTimeParser struct{}

func (XSDDateTimeParser) GetNativeValue(value string, node graph.Node) (interface{}, error) {
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

func (XSDDateTimeParser) FormatNativeValue(newValue, oldValue interface{}, node graph.Node) (string, error) {
	if newValue == nil {
		return "", nil
	}
	var old DateTime
	var ok bool
	switch v := newValue.(type) {
	case time.Time:
		return v.Format("2006-01-02T15:04:05Z"), nil
	case Date:
		if v.Location == nil {
			return v.ToTime().Format("2006-01-02T15:04:05Z"), nil
		}
		return v.ToTime().In(v.Location).Format("2006-01-02T15:04:05Z"), nil
	case DateTime:
		if v.Location == nil {
			return v.ToTime().Format("2006-01-02T15:04:05Z"), nil
		}
		return v.ToTime().In(v.Location).Format("2006-01-02T15:04:05Z"), nil //2002-05-30T09:30:10-06:00
	case TimeOfDay:
		if v.Location == nil {
			return "2006-01-02T" + v.ToTime().Format("15:04:05Z"), nil
		}
		return "2006-01-02T" + v.ToTime().In(v.Location).Format("15:04:05Z"), nil
	case GDay:
		if oldValue != nil {
			old, ok = oldValue.(DateTime)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: XSDDateTerm, Value: oldValue, Msg: "Not a DateTime"}
			}
		}
		old.Day = int(v)
		return old.ToTime().Format("2006-01-02T15:04:05Z"), nil
	case GMonth:
		if oldValue != nil {
			old, ok = oldValue.(DateTime)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: XSDDateTerm, Value: oldValue, Msg: "Not a DateTime"}
			}
		}
		old.Month = int(v)
		return old.ToTime().Format("2006-01-02T15:04:05Z"), nil
	case GYear:
		if oldValue != nil {
			old, ok = oldValue.(DateTime)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: XSDDateTerm, Value: oldValue, Msg: "Not a DateTime"}
			}
		}
		old.Year = int(v)
		return old.ToTime().Format("2006-01-02T15:04:05Z"), nil
	case GMonthDay:
		if oldValue != nil {
			old, ok = oldValue.(DateTime)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: XSDDateTerm, Value: oldValue, Msg: "Not a DateTime"}
			}
		}
		old.Month = int(v.Month)
		old.Day = int(v.Day)
		return old.ToTime().Format("2006-01-02T15:04:05Z"), nil
	case GYearMonth:
		if oldValue != nil {
			old, ok = oldValue.(DateTime)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: XSDDateTerm, Value: oldValue, Msg: "Not a DateTime"}
			}
		}
		old.Year = int(v.Year)
		old.Month = int(v.Month)
		return old.ToTime().Format("2006-01-02T15:04:05Z"), nil
	case UnixTime:
		if v.Location == nil {
			return fmt.Sprintf("%02d", v.Seconds), nil
		}
		return v.ToTime().Format("2002-05-30T09:30:10.5-06:00"), nil
	case UnixTimeNano:
		if v.Location == nil {
			return v.ToTime().Format("2002-05-30T09:30:10.5"), nil
		}
		return v.ToTime().In(v.Location).Format("2002-05-30T09:30:10.5-06:00"), nil
	default:
		return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: XSDDateTimeTerm, Value: newValue}
	}
	return "", nil
}

type XSDTimeParser struct{}

func (XSDTimeParser) GetNativeValue(value string, node graph.Node) (interface{}, error) {
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

func (XSDTimeParser) FormatNativeValue(newValue, oldValue interface{}, node graph.Node) (string, error) {
	if newValue == nil {
		return "", nil
	}
	switch v := newValue.(type) {
	case time.Time:
		return v.Format("15:04:05"), nil
	case Date:
		return "", ErrIncompatibleTypes{node.GetLabels().String(), v}
	case DateTime:
		if v.Location == nil {
			return v.ToTime().Format("09:00:00Z"), nil
		}
		return v.ToTime().In(v.Location).Format("09:00:00Z"), nil
	case UnixTime:
		if v.Location == nil {
			return strconv.FormatInt(v.ToTime().Unix(), 10), nil
		}
		return strconv.FormatInt((v.ToTime().In(v.Location).Unix()), 10), nil
	case UnixTimeNano:
		if v.Location == nil {
			return strconv.FormatInt(v.ToTime().UnixNano(), 10), nil
		}
		return strconv.FormatInt((v.ToTime().In(v.Location).UnixNano()), 10), nil
	default:
		return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: XSDTimeTerm, Value: newValue}
	}
	return "", nil
}

type JSONDateParser struct{}

// ParseValue parses a JSON date
//
//   YYYY-MM-DD
func (JSONDateParser) GetNativeValue(value string, node graph.Node) (interface{}, error) {
	if len(value) == 0 {
		return nil, nil
	}
	return genericDateParse(value, goFormat("2006-01-02"))
}

// FormatNativeValue gets a target node and it's go native value, and returns
// the value of the target node to an JSONDate
func (JSONDateParser) FormatNativeValue(newValue, oldValue interface{}, node graph.Node) (string, error) {
	if newValue == nil {
		return "", nil
	}
	var old Date
	var ok bool
	switch v := newValue.(type) {
	case time.Time:
		return v.Format("2006-01-02"), nil
	case Date:
		if v.Location == nil {
			return v.ToTime().Format("2006-01-02"), nil
		}
		return v.ToTime().In(v.Location).Format("2006-01-02Z0700"), nil
	case DateTime:
		if v.Location == nil {
			return v.ToTime().Format("2006-01-02"), nil
		}
		return v.ToTime().In(v.Location).Format("2006-01-02Z0700"), nil
	case TimeOfDay:
		if v.Location == nil {
			return v.ToTime().Format("2006-01-02"), nil
		}
		return v.ToTime().In(v.Location).Format("2006-01-02Z0700"), nil
	case GDay:
		if oldValue != nil {
			old, ok = oldValue.(Date)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: JSONDateTerm, Value: oldValue, Msg: "Not a Date"}
			}
		}
		old.Day = int(v)
		return old.ToTime().Format("2006-01-02"), nil
	case GMonth:
		if oldValue != nil {
			old, ok = oldValue.(Date)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: JSONDateTerm, Value: oldValue, Msg: "Not a Date"}
			}
		}
		old.Month = int(v)
		return old.ToTime().Format("2006-01-02"), nil
	case GYear:
		if oldValue != nil {
			old, ok = oldValue.(Date)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: JSONDateTerm, Value: oldValue, Msg: "Not a Date"}
			}
		}
		old.Year = int(v)
		return old.ToTime().Format("2006-01-02"), nil
	case GMonthDay:
		if oldValue != nil {
			old, ok = oldValue.(Date)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: JSONDateTerm, Value: oldValue, Msg: "Not a Date"}
			}
		}
		old.Month = int(v.Month)
		old.Day = int(v.Day)
		return old.ToTime().Format("2006-01-02"), nil
	case GYearMonth:
		if oldValue != nil {
			old, ok = oldValue.(Date)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: JSONDateTerm, Value: oldValue, Msg: "Not a Date"}
			}
		}
		old.Year = int(v.Year)
		old.Month = int(v.Month)
		return old.ToTime().Format("2006-01-02"), nil
	case UnixTime:
		if v.Location == nil {
			return v.ToTime().Format("2006-01-02"), nil
		}
		return v.ToTime().In(v.Location).Format("2006-01-02Z0700"), nil
	case UnixTimeNano:
		if v.Location == nil {
			return v.ToTime().Format("2006-01-02"), nil
		}
		return v.ToTime().In(v.Location).Format("2006-01-02Z0700"), nil
	default:
		return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: JSONDateTerm, Value: newValue}
	}
	return "", nil
}

type JSONDateTimeParser struct{}

// ParseValue parses a JSON date-time
//
//   YYYY-MM-DD
func (JSONDateTimeParser) GetNativeValue(value string, node graph.Node) (interface{}, error) {
	if len(value) == 0 {
		return nil, nil
	}
	return genericDateTimeParse(value, goFormat(time.RFC3339), goFormat(time.RFC3339Nano))
}

// "2006-01-02T11:11:11Z07:00" -> Note: uses a 24Hour based clock   "2006-01-02T11:11:11.999999999Z07:00"
func (JSONDateTimeParser) FormatNativeValue(newValue, oldValue interface{}, node graph.Node) (string, error) {
	if newValue == nil {
		return "", nil
	}
	switch v := newValue.(type) {
	case time.Time:
		return v.Format(time.RFC3339), nil
	case Date:
		if v.Location == nil {
			return v.ToTime().Format("2006-01-02T15:04:05Z"), nil
		}
		return v.ToTime().In(v.Location).Format("2006-01-02T15:04:05Z"), nil
	case DateTime:
		if v.Location == nil && v.Nanoseconds == 0 {
			return v.ToTime().Format("2006-01-02T15:04:05Z"), nil
		}
		if v.Location != nil && v.Nanoseconds != 0 {
			return v.ToTime().In(v.Location).Format(time.RFC3339Nano), nil
		}
		return v.ToTime().In(v.Location).Format(time.RFC3339), nil
	case TimeOfDay:
		if v.Location == nil {
			return "2006-01-02T" + v.ToTime().Format("15:04:05Z"), nil
		}
		return "2006-01-02T" + v.ToTime().In(v.Location).Format("15:04:05Z"), nil
	case UnixTime:
		if v.Location == nil {
			return v.ToTime().Format("2006-01-02T15:04:05Z"), nil
		}
		return v.ToTime().In(v.Location).Format(time.RFC3339), nil
	case UnixTimeNano:
		if v.Location == nil {
			return v.ToTime().Format("2006-01-02T15:04:05Z"), nil
		}
		return v.ToTime().In(v.Location).Format(time.RFC3339Nano), nil
	default:
		return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: JSONDateTimeTerm, Value: newValue}
	}
	return "", nil
}

type JSONTimeParser struct{}

// ParseValue parses a JSON time
//
//   HH:mm
//   HH:mm:ss
//   HH:mm:ssZ
func (JSONTimeParser) GetNativeValue(value string, node graph.Node) (interface{}, error) {
	if len(value) == 0 {
		return nil, nil
	}
	return genericTimeParse(value, gomentFormat("HH:mm:ssZ"),
		gomentFormat("HH:mm:ss"),
		gomentFormat("HH:mm"))
}

func (JSONTimeParser) FormatNativeValue(newValue, oldValue interface{}, node graph.Node) (string, error) {
	if newValue == nil {
		return "", nil
	}
	switch v := newValue.(type) {
	case time.Time:
		return v.Format("HH:mm:ssZ"), nil // 11:11:11Z
	case Date:
		if v.Location == nil {
			return v.ToTime().Format("09:00:00"), nil
		}
		return v.ToTime().In(v.Location).Format("09:00:00Z"), nil
	case DateTime:
		if v.Location == nil {
			return v.ToTime().Format("09:00:00Z"), nil
		}
		return v.ToTime().In(v.Location).Format("09:00:00Z"), nil
	case TimeOfDay:
		if v.Location == nil {
			return v.ToTime().Format("HH:mm:ss"), nil
		}
		return v.ToTime().In(v.Location).Format("HH:mm:ssZ"), nil
	case UnixTime:
		if v.Location == nil {
			return v.ToTime().Format("ss"), nil
		}
		return v.ToTime().Format("HH:mm:ssZ"), nil
	case UnixTimeNano:
		if v.Location == nil {
			return v.ToTime().Format("HH:mm:ss"), nil
		}
		return v.ToTime().Format("HH:mm:ssZ"), nil
	default:
		return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: JSONTimeTerm, Value: newValue}
	}
	return "", nil
}

type PatternDateTimeParser struct{}

// GetNativeValue looks at the goTimeFormat, momentTimeFormat properties
// in the node, and parses the datetime using that. The format
// property can be an array, giving all possible formats. If none existsm guesses format
func (PatternDateTimeParser) GetNativeValue(value string, node graph.Node) (interface{}, error) {
	if len(value) == 0 {
		return nil, nil
	}
	gf := ls.AsPropertyValue(node.GetProperty(GoTimeFormatTerm)).Slice()
	mf := ls.AsPropertyValue(node.GetProperty(MomentTimeFormatTerm)).Slice()
	garr := make([]dateTimeFormatter, 0, len(gf)+len(mf))
	for _, x := range gf {
		garr = append(garr, goFormat(x))
	}
	for _, x := range mf {
		garr = append(garr, gomentFormat(x))
	}
	if len(garr) > 0 {
		return genericDateTimeParse(value, garr...)
	}
	t, err := dateparse.ParseStrict(value)
	if err != nil {
		return nil, err
	}
	return NewDateTime(t), nil
}

func (PatternDateTimeParser) FormatNativeValue(newValue, oldValue interface{}, node graph.Node) (string, error) {
	if newValue == nil {
		return "", nil
	}
	var formatter interface{}
	if s, ok := node.GetProperty(GoTimeFormatTerm); ok {
		formatter = goFormat(ls.AsPropertyValue(s, true).AsString())
	}
	if s, ok := node.GetProperty(MomentTimeFormatTerm); ok {
		formatter = gomentFormat(ls.AsPropertyValue(s, true).AsString())
	}
	if formatter == nil {
		formatter = goFormat(time.RFC3339)
	}
	var old Date
	var ok bool
	switch v := newValue.(type) {
	case time.Time:
		return formatter.(goTimeFormatter).formatGoTime(v)
	case Date:
		return formatter.(dateFormatter).formatDate(v)
	case DateTime:
		return formatter.(dateTimeFormatter).formatDateTime(v)
	case TimeOfDay:
		return formatter.(timeFormatter).formatTime(v)
	case GDay:
		if oldValue != nil {
			old, ok = oldValue.(Date)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: PatternDateTimeTerm, Value: oldValue, Msg: "Not a Date"}
			}
		}
		old.Day = int(v)
		return formatter.(dateFormatter).formatDate(old)
	case GMonth:
		if oldValue != nil {
			old, ok = oldValue.(Date)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: PatternDateTimeTerm, Value: oldValue, Msg: "Not a Date"}
			}
		}
		old.Month = int(v)
		return formatter.(dateFormatter).formatDate(old)
	case GYear:
		if oldValue != nil {
			old, ok = oldValue.(Date)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: PatternDateTimeTerm, Value: oldValue, Msg: "Not a Date"}
			}
		}
		old.Year = int(v)
		return formatter.(dateFormatter).formatDate(old)
	case GMonthDay:
		if oldValue != nil {
			old, ok = oldValue.(Date)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: PatternDateTimeTerm, Value: oldValue, Msg: "Not a Date"}
			}
		}
		old.Month = int(v.Month)
		old.Day = int(v.Day)
		return formatter.(dateFormatter).formatDate(old)
	case GYearMonth:
		if oldValue != nil {
			old, ok = oldValue.(Date)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: PatternDateTimeTerm, Value: oldValue, Msg: "Not a Date"}
			}
		}
		old.Year = int(v.Year)
		old.Month = int(v.Month)
		return formatter.(dateFormatter).formatDate(old)
	default:
		return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: PatternDateTimeTerm, Value: newValue}
	}
	return "", nil
}

type PatternDateParser struct{}

// ParseValue looks at the goTimeFormat, momentTimeFormat properties
// in the node, and parses the datetime using that. The format
// property can be an array, giving all possible formats
func (PatternDateParser) GetNativeValue(value string, node graph.Node) (interface{}, error) {
	if len(value) == 0 {
		return nil, nil
	}
	gf := ls.AsPropertyValue(node.GetProperty(GoTimeFormatTerm)).Slice()
	mf := ls.AsPropertyValue(node.GetProperty(MomentTimeFormatTerm)).Slice()
	garr := make([]dateFormatter, 0, len(gf)+len(mf))
	for _, x := range gf {
		garr = append(garr, goFormat(x))
	}
	for _, x := range mf {
		garr = append(garr, gomentFormat(x))
	}
	if len(garr) > 0 {
		return genericDateParse(value, garr...)
	}
	t, err := dateparse.ParseStrict(value)
	if err != nil {
		return nil, err
	}
	return NewDate(t), nil
}

func (PatternDateParser) FormatNativeValue(newValue, oldValue interface{}, node graph.Node) (string, error) {
	if newValue == nil {
		return "", nil
	}
	var formatter interface{}
	if s, ok := node.GetProperty(GoTimeFormatTerm); ok {
		formatter = goFormat(ls.AsPropertyValue(s, true).AsString())
	}
	if s, ok := node.GetProperty(MomentTimeFormatTerm); ok {
		formatter = gomentFormat(ls.AsPropertyValue(s, true).AsString())
	}
	if formatter == nil {
		formatter = goFormat("2006-01-02")
	}
	var old Date
	var ok bool
	switch v := newValue.(type) {
	case time.Time:
		return "", ErrIncompatibleTypes{node.GetLabels().String(), v}
	case Date:
		return formatter.(dateFormatter).formatDate(v)
	case DateTime:
		return formatter.(dateTimeFormatter).formatDateTime(v)
	case TimeOfDay:
		return "", ErrIncompatibleTypes{node.GetLabels().String(), v}
	case GDay:
		if oldValue != nil {
			old, ok = oldValue.(Date)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: PatternDateTimeTerm, Value: oldValue, Msg: "Not a Date"}
			}
		}
		old.Day = int(v)
		return formatter.(dateFormatter).formatDate(old)
	case GMonth:
		if oldValue != nil {
			old, ok = oldValue.(Date)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: PatternDateTimeTerm, Value: oldValue, Msg: "Not a Date"}
			}
		}
		old.Month = int(v)
		return formatter.(dateFormatter).formatDate(old)
	case GYear:
		if oldValue != nil {
			old, ok = oldValue.(Date)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: PatternDateTimeTerm, Value: oldValue, Msg: "Not a Date"}
			}
		}
		old.Year = int(v)
		return formatter.(dateFormatter).formatDate(old)
	case GMonthDay:
		if oldValue != nil {
			old, ok = oldValue.(Date)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: PatternDateTimeTerm, Value: oldValue, Msg: "Not a Date"}
			}
		}
		old.Month = int(v.Month)
		old.Day = int(v.Day)
		return formatter.(dateFormatter).formatDate(old)
	case GYearMonth:
		if oldValue != nil {
			old, ok = oldValue.(Date)
			if !ok {
				return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: PatternDateTimeTerm, Value: oldValue, Msg: "Not a Date"}
			}
		}
		old.Year = int(v.Year)
		old.Month = int(v.Month)
		return formatter.(dateFormatter).formatDate(old)
	default:
		return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: PatternDateTerm, Value: newValue}
	}
	return "", nil
}

type PatternTimeParser struct{}

// ParseValue looks at the goTimeFormat, momentTimeFormat properties
// in the node, and parses the datetime using that. The format
// property can be an array, giving all possible formats
func (PatternTimeParser) GetNativeValue(value string, node graph.Node) (interface{}, error) {
	if len(value) == 0 {
		return nil, nil
	}
	gf := ls.AsPropertyValue(node.GetProperty(GoTimeFormatTerm)).Slice()
	mf := ls.AsPropertyValue(node.GetProperty(MomentTimeFormatTerm)).Slice()
	garr := make([]timeFormatter, 0, len(gf)+len(mf))
	for _, x := range gf {
		garr = append(garr, goFormat(x))
	}
	for _, x := range mf {
		garr = append(garr, gomentFormat(x))
	}
	if len(garr) > 0 {
		return genericTimeParse(value, garr...)
	}
	t, err := dateparse.ParseStrict(value)
	if err != nil {
		return nil, err
	}
	return NewTimeOfDay(t), nil
}

func (PatternTimeParser) FormatNativeValue(newValue, oldValue interface{}, node graph.Node) (string, error) {
	if newValue == nil {
		return "", nil
	}
	var formatter interface{}
	if s, ok := node.GetProperty(GoTimeFormatTerm); ok {
		formatter = goFormat(ls.AsPropertyValue(s, true).AsString())
	}
	if s, ok := node.GetProperty(MomentTimeFormatTerm); ok {
		formatter = gomentFormat(ls.AsPropertyValue(s, true).AsString())
	}
	if formatter == nil {
		formatter = goFormat("15:04:05.999999999Z07:00")
	}
	switch v := newValue.(type) {
	case time.Time:
		return formatter.(goTimeFormatter).formatGoTime(v)
	case Date:
		return "", ErrIncompatibleTypes{node.GetLabels().String(), v}
	case DateTime:
		return formatter.(dateTimeFormatter).formatDateTime(v)
	case TimeOfDay:
		return formatter.(timeFormatter).formatTime(v)
	case GDay:
		return "", ErrIncompatibleTypes{node.GetLabels().String(), v}
	case GMonth:
		return "", ErrIncompatibleTypes{node.GetLabels().String(), v}
	case GYear:
		return "", ErrIncompatibleTypes{node.GetLabels().String(), v}
	case GMonthDay:
		return "", ErrIncompatibleTypes{node.GetLabels().String(), v}
	case GYearMonth:
		return "", ErrIncompatibleTypes{node.GetLabels().String(), v}
	default:
		return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: PatternTimeTerm, Value: newValue}
	}
	return "", nil
}

type XSDGDayParser struct{}

func (XSDGDayParser) GetNativeValue(value string, node graph.Node) (interface{}, error) {
	if len(value) == 0 {
		return nil, nil
	}
	x, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}
	return GDay(x), nil
}

func (XSDGDayParser) FormatNativeValue(newValue, oldValue interface{}, node graph.Node) (string, error) {
	if newValue == nil {
		return "", nil
	}
	switch v := newValue.(type) {
	case time.Time:
		return "", ErrIncompatibleTypes{node.GetLabels().String(), v}
	case Date:
		return strconv.Itoa((v.Day)), nil
	case DateTime:
		return strconv.Itoa((v.Day)), nil
	case GDay:
		return strconv.Itoa(int(v)), nil
	case GMonthDay:
		return strconv.Itoa(int(v.Day)), nil
	case GYearMonth:
		return "", ErrIncompatibleTypes{node.GetLabels().String(), v}
	case UnixTime:
		return "", ErrIncompatibleTypes{node.GetLabels().String(), v}
	case UnixTimeNano:
		return "", ErrIncompatibleTypes{node.GetLabels().String(), v}
	default:
		return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: XSDGDayTerm, Value: newValue}
	}
	return "", nil
}

type XSDGMonthParser struct{}

func (XSDGMonthParser) GetNativeValue(value string, node graph.Node) (interface{}, error) {
	if len(value) == 0 {
		return nil, nil
	}
	return genericDateParse(value, goFormat("01"), gomentFormat("MM"))
}

func (XSDGMonthParser) FormatNativeValue(newValue, oldValue interface{}, node graph.Node) (string, error) {
	if newValue == nil {
		return "", nil
	}
	switch v := newValue.(type) {
	case time.Time:
		return "", ErrIncompatibleTypes{node.GetLabels().String(), v}
	case Date:
		return v.ToTime().Format("01"), nil
	case DateTime:
		return v.ToTime().Format("01"), nil
	case GMonth:
		return strconv.Itoa(int(v)), nil
	case GMonthDay:
		return strconv.Itoa(int(v.Month)), nil
	case GYearMonth:
		return strconv.Itoa(int(v.Month)), nil
	default:
		return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: XSDGMonthTerm, Value: newValue}
	}
	return "", nil
}

type XSDGMonthDayParser struct{}

func (XSDGMonthDayParser) GetNativeValue(value string, node graph.Node) (interface{}, error) {
	if len(value) == 0 {
		return nil, nil
	}
	return genericDateParse(value, goFormat("1-2"), goFormat("01-2"), goFormat("01-02"), gomentFormat("MM-DD"))
}

func (XSDGMonthDayParser) FormatNativeValue(newValue, oldValue interface{}, node graph.Node) (string, error) {
	if newValue == nil {
		return "", nil
	}
	switch v := newValue.(type) {
	case time.Time:
		return "", ErrIncompatibleTypes{node.GetLabels().String(), v}
	case Date:
		return v.ToTime().Format("01-02"), nil
	case DateTime:
		return v.ToTime().Format("01-02"), nil
	case GDay:
		return strconv.Itoa(int(v)), nil
	case GMonth:
		return strconv.Itoa(int(v)), nil
	case GMonthDay:
		return fmt.Sprintf("%02d-%02d", v.Month, v.Day), nil
	case GYearMonth:
		return strconv.Itoa(int(v.Month)), nil
	case UnixTime:
		return "", ErrIncompatibleTypes{node.GetLabels().String(), v}
	case UnixTimeNano:
		return "", ErrIncompatibleTypes{node.GetLabels().String(), v}
	default:
		return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: XSDGMonthDayTerm, Value: newValue}
	}
	return "", nil
}

type XSDGYearParser struct{}

func (XSDGYearParser) GetNativeValue(value string, node graph.Node) (interface{}, error) {
	if len(value) == 0 {
		return nil, nil
	}
	return genericDateParse(value, goFormat("2006"), gomentFormat("YYYY"))
}

func (XSDGYearParser) FormatNativeValue(newValue, oldValue interface{}, node graph.Node) (string, error) {
	if newValue == nil {
		return "", nil
	}
	switch v := newValue.(type) {
	case Date:
		return v.ToTime().Format("2006"), nil
	case DateTime:
		return v.ToTime().Format("2006"), nil
	case GYear:
		return strconv.Itoa(int(v)), nil
	case GYearMonth:
		return strconv.Itoa(int(v.Year)), nil
	default:
		return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: XSDGYearTerm, Value: newValue}
	}
	return "", nil
}

type XSDGYearMonthParser struct{}

func (XSDGYearMonthParser) GetNativeValue(value string, node graph.Node) (interface{}, error) {
	if len(value) == 0 {
		return nil, nil
	}
	return genericDateParse(value, goFormat("2006-01"), gomentFormat("YY-MM"))
}

func (XSDGYearMonthParser) FormatNativeValue(newValue, oldValue interface{}, node graph.Node) (string, error) {
	if newValue == nil {
		return "", nil
	}
	switch v := newValue.(type) {
	case Date:
		return v.ToTime().Format("2006-01"), nil
	case DateTime:
		return v.ToTime().Format("2006-01"), nil
	case GMonth:
		return strconv.Itoa(int(v)), nil
	case GYear:
		return strconv.Itoa(int(v)), nil
	case GYearMonth:
		return fmt.Sprintf("%02d-%02d", v.Year, v.Month), nil
	default:
		return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: XSDGYearMonthTerm, Value: newValue}
	}
	return "", nil
}

type UnixTimeParser struct{}

func (UnixTimeParser) GetNativeValue(value string, node graph.Node) (interface{}, error) {
	if len(value) == 0 {
		return nil, nil
	}
	x, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}
	return UnixTime{int64(x), nil}, nil
}

func (UnixTimeParser) FormatNativeValue(newValue, oldValue interface{}, node graph.Node) (string, error) {
	if newValue == nil {
		return "", nil
	}
	switch v := newValue.(type) {
	case time.Time:
		return strconv.FormatInt(v.Unix(), 10), nil
	case Date:
		if v.Location == nil {
			return strconv.FormatInt(v.ToTime().Unix(), 10), nil
		}
		return strconv.FormatInt(v.ToTime().In(v.Location).Unix(), 10), nil
	case DateTime:
		if v.Location == nil {
			return strconv.FormatInt(v.ToTime().Unix(), 10), nil
		}
		return strconv.FormatInt(v.ToTime().In(v.Location).Unix(), 10), nil
	case TimeOfDay:
		if v.Location == nil {
			t := time.Date(1970, time.January, 1, int(v.Hour), int(v.Minute), int(v.Seconds), int(v.Nanoseconds), time.UTC)
			return strconv.FormatInt(t.Unix(), 10), nil
		}
		t := time.Date(1970, time.January, 1, int(v.Hour), int(v.Minute), int(v.Seconds), int(v.Nanoseconds), v.Location)
		return strconv.FormatInt(t.Unix(), 10), nil
	case UnixTimeNano:
		if v.Location == nil {
			return fmt.Sprintf("%d", v.ToTime().UnixNano()), nil
		}
		return strconv.FormatInt(v.ToTime().In(v.Location).UnixNano(), 10), nil
	default:
		return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: UnixTimeTerm, Value: newValue}
	}
	return "", nil
}

type UnixTimeNanoParser struct{}

func (UnixTimeNanoParser) GetNativeValue(value string, node graph.Node) (interface{}, error) {
	if len(value) == 0 {
		return nil, nil
	}
	x, err := strconv.Atoi(value)
	if err != nil {
		return nil, err
	}
	return UnixTimeNano{int64(x), nil}, nil
}

func (UnixTimeNanoParser) FormatNativeValue(newValue, oldValue interface{}, node graph.Node) (string, error) {
	if newValue == nil {
		return "", nil
	}
	switch v := newValue.(type) {
	case time.Time:
		return strconv.FormatInt(v.UnixNano(), 10), nil
	case Date:
		if v.Location == nil {
			return strconv.FormatInt(v.ToTime().UnixNano(), 10), nil
		}
		return strconv.FormatInt(v.ToTime().In(v.Location).UnixNano(), 10), nil
	case DateTime:
		if v.Location == nil {
			return strconv.FormatInt(v.ToTime().UnixNano(), 10), nil
		}
		return strconv.FormatInt(v.ToTime().In(v.Location).UnixNano(), 10), nil
	case TimeOfDay:
		if v.Location == nil {
			t := time.Date(1970, time.January, 1, int(v.Hour), int(v.Minute), int(v.Seconds), int(v.Nanoseconds), time.UTC)
			return strconv.FormatInt(t.UnixNano(), 10), nil
		}
		t := time.Date(1970, time.January, 1, int(v.Hour), int(v.Minute), int(v.Seconds), int(v.Nanoseconds), v.Location)
		return strconv.FormatInt(t.UnixNano(), 10), nil
	case GDay:
		return "", ErrIncompatibleTypes{node.GetLabels().String(), v}
	case GMonth:
		return "", ErrIncompatibleTypes{node.GetLabels().String(), v}
	case GYear:
		return "", ErrIncompatibleTypes{node.GetLabels().String(), v}
	case GMonthDay:
		return "", ErrIncompatibleTypes{node.GetLabels().String(), v}
	case GYearMonth:
		return "", ErrIncompatibleTypes{node.GetLabels().String(), v}
	case UnixTime:
		return strconv.FormatInt(v.Seconds*1e9, 10), nil
	default:
		return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: UnixTimeTerm, Value: newValue}
	}
	return "", nil
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
