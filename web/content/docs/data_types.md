---
title: Data Types
---
# Data Types

LSA stores data as uninterpreted string of bytes. Values can be stored
as node or edge properties. The default node property to store an
ingested value is `https://lschema.org/nodeValue`. An optional
`https://lschema.org/valueType` term specifies the data type of the
value. If a value has no value type or an unrecognized value type, it
is stored and processed as a raw string. LSA recognizes the following
value types and can perform conversions between compatible values:

<table class="table table-sm">
 <thead>
   <tr>
     <th>valueType</th>
     <th>Examples</th>
     <th>Description</th>
   </tr>
 </thead>
 <tbody>
 
  <tr>
    <td><code>https:/json-schema.org/boolean</code><br><code>json:boolean</code></td>
    <td>true, false</td>
    <td>JSON boolean value</td>
  </tr>

  <tr>
    <td><code>http://www.w3.org/2001/XMLSchema/boolean</code><br><code>xsd:boolean</code><br><code>xs:boolean</code></td>
    <td>true, 1<br>
    false, 0</td>
    <td>XML boolean value, can be "true", "false", "0", or "1".</td>
  </tr>

  <tr>
    <td><code>http://www.w3.org/2001/XMLSchema/gDay</code><br><code>xsd:gDay</code><br><code>xs:gDay</code></td>
    <td>1, 02</td>
    <td><a href="https://www.w3.org/TR/xmlschema-2/#gDay">XML Gregorian date day part</a></td>
  </tr>

  <tr>
    <td><code>http://www.w3.org/2001/XMLSchema/gMonth</code><br><code>xsd:gMonth</code><br><code>xs:gMonth</code></td>
    <td>06, 6, 11</td>
    <td><a href="https://www.w3.org/TR/xmlschema-2/#gMonth">XML Gregorian date month part</a></td>
  </tr>


  <tr>
    <td><code>http://www.w3.org/2001/XMLSchema/gYear</code><br><code>xsd:gYear</code><br><code>xs:gYear</code></td>
    <td>2001</td>
    <td><a href="https://www.w3.org/TR/xmlschema-2/#gYear">XML Gregorian date year part.</a></td>
  </tr>

  <tr>
    <td><code>http://www.w3.org/2001/XMLSchema/gMonthDay</code><br><code>xsd:gMonthDay</code><br><code>xs:gMonthDay</code></td>
    <td>01-05, 1-5</td>
    <td><a href="https://www.w3.org/TR/xmlschema-2/#gMonthDay">XML Gregorian date month and day.</a></td>
  </tr>

  <tr>
    <td><code>http://www.w3.org/2001/XMLSchema/gYearMonth</code><br><code>xsd:gYearMonth</code><br><code>xs:gYearMonth</code></td>
    <td>2002-01, 2002-1</td>
    <td><a href="https://www.w3.org/TR/xmlschema-2/#gYearMonth">XML Gregorian date year and month.</a></td>
  </tr>

  <tr>
    <td><code>http://www.w3.org/2001/XMLSchema/date</code><br><code>xsd:date</code><br><code>xs:date</code></td>
    <td>2002-02-15<br><nobr>2002-02-15+0300</nobr><br>2002-02-15Z</td>
    <td><a href="https://www.w3.org/TR/xmlschema-2/#date">XML date.</a></td>
  </tr>

  <tr>
    <td><code>http://www.w3.org/2001/XMLSchema/time</code><br><code>xsd:time</code><br><code>xs:time</code></td>
    <td><nobr>13:20:00-05:00</nobr><br>
    13:20:00
    </td>
    <td><a href="https://www.w3.org/TR/xmlschema-2/#time">XML time.</a></td>
  </tr>

  <tr>
    <td><code>http://www.w3.org/2001/XMLSchema/dateTime</code><br><code>xsd:dateTime</code><br><code>xs:dateTime</code></td>
    <td><nobr>2002-10-10T12:00:00-05:00</nobr><br>
    <nobr>2002-10-10T17:00:00Z</nobr><br></td>
    <td><a href="https://www.w3.org/TR/xmlschema-2/#dateTime">XML date-time.</a></td>
  </tr>

  <tr>
    <td><code>https:/json-schema.org/date</code><br><code>json:date</code></td>
    <td>2006-01-02</td>
    <td><a href="https://datatracker.ietf.org/doc/html/rfc3339">RFC3339 JSON date.</a></td>
  </tr>
  
  <tr>
    <td><code>https:/json-schema.org/date-time</code><br><code>json:date-time</code></td>
    <td><nobr>2006-01-02T15:04:05Z07:00</nobr></td>
    <td><a href="https://datatracker.ietf.org/doc/html/rfc3339">RFC3339 JSON date-time.</a></td>
  </tr>

  <tr>
    <td><code>https:/json-schema.org/time</code><br><code>json:time</code></td>
    <td>15:04:05Z07:00</td>
    <td><a href="https://datatracker.ietf.org/doc/html/rfc3339">RFC3339 JSON time.</a></td>
  </tr>
  
  <tr>
    <td><code>https://unixtime.org/time</code><br><code>unix:time</code></td>
    <td>1649979010</td>
    <td>UNIX Epoch time in seconds.</td>
  </tr>

  <tr>
    <td><code>https://unixtime.org/timeNano</code><br><code>unix:timeNano</code></td>
    <td><nobr>1649979038230000000</nobr></td>
    <td>UNIX Epoch time in nanoseconds</td>
  </tr>
  
  <tr>
    <td><code>https://lschema.org/dateTime</code><br><code>ls:dateTime</code></td>
    <td><em><a href="#pattern-based-date-time">See below</a></em></td>
    <td>Date-time based on a pattern. The pattern is given in 
    <code>https://lschema.org/goTimeFormat</code>, <code>ls:goTimeFormat</code>, 
    <code>https://lschema.org/momentTimeFormat</code>, or <code>ls:momentTimeFormat</code>.
    </td>
  </tr>

  <tr>
    <td><code>https://lschema.org/date</code><br><code>ls:date</code></td>
    <td><em><a href="#pattern-based-date-time">See below</a></em></td>
    <td>Date based on a pattern. The pattern is given in 
    <code>https://lschema.org/goTimeFormat</code>, <code>ls:goTimeFormat</code>, 
    <code>https://lschema.org/momentTimeFormat</code>, or <code>ls:momentTimeFormat</code>.
    </td>
  </tr>

  <tr>
    <td><code>https://lschema.org/time</code><br><code>ls:time</code></td>
    <td><em><a href="#pattern-based-date-time">See below</a></em></td>
    <td>Time based on a pattern.  The pattern is given in 
    <code>https://lschema.org/goTimeFormat</code>, <code>ls:goTimeFormat</code>, 
    <code>https://lschema.org/momentTimeFormat</code>, or <code>ls:momentTimeFormat</code>.
    </td>
  </tr>

 </tbody>
</table>

## Pattern date-time

If none of the standard date-time patterns do not match the data at
hand, it is possible to provide a pattern to process date, time, or
date-time values. 

To process a date value that looks like `January 1 2020`, use:

{{< highlight json >}}
{
  "@id": "attr1",
  "@type": "Value",
  "valueType": "ls:date",
  "momentTimeFormat": "MMMM D YYYY"
}
{{</highlight>}}

The annotation `ls:momentTimeFormat` allows specifying format using
the [Go implementation of JS Moment library](https://github.com/nleeper/goment)
conventions.

The `ls:goTimeFormat` allows specifying date/time format using the Go
standard library [time
format](https://pkg.go.dev/time@go1.18.1#Time.Format).


