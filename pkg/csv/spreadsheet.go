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

package csv

import (
	"encoding/csv"
	"io"

	"github.com/xuri/excelize/v2"
)

// ReadCSV reads a CSV file
func ReadCSV(input io.Reader, csvSeparator string) ([][]string, error) {
	reader := csv.NewReader(input)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1
	if len(csvSeparator) > 0 {
		reader.Comma = rune(csvSeparator[0])
	}
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	return data, nil
}

type Row struct {
	Index   int
	Headers []string
	Row     []string
	Err     error
}

func StreamCSVRows(input io.Reader, csvSeparator string, headerRow int) (<-chan Row, error) {
	reader := csv.NewReader(input)
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1
	if len(csvSeparator) > 0 {
		reader.Comma = rune(csvSeparator[0])
	}
	ch := make(chan Row)
	go func() {
		defer close(ch)
		n := 0
		var header []string
		for {
			row, err := reader.Read()
			if err == io.EOF {
				return
			}
			if err != nil {
				ch <- Row{Index: n, Err: err}
				return
			}
			if n < headerRow {
				n++
				continue
			}
			if n == headerRow {
				header = make([]string, len(row))
				copy(header, row)
				n++
				continue
			}
			rowData := make([]string, len(row))
			copy(rowData, row)
			ch <- Row{Index: n, Headers: header, Row: rowData}
			n++
		}
	}()
	return ch, nil
}

// ReadExcel reads an excel file
func ReadExcel(input io.Reader) (map[string][][]string, error) {
	f, err := excelize.OpenReader(input)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sheets := f.GetSheetList()
	ret := make(map[string][][]string)
	for _, sheet := range sheets {
		rows, err := f.GetRows(sheet)
		if err != nil {
			return nil, err
		}
		ret[sheet] = rows
	}
	return ret, nil
}

// StreamExcelSheetRows reads an excel file and streams its rows. Close the reader to stop streaming
func StreamExcelSheetRows(input io.Reader, sheet string, headerRow int) (<-chan Row, error) {
	f, err := excelize.OpenReader(input)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	rows, err := f.Rows(sheet)
	if err != nil {
		return nil, err
	}
	ch := make(chan Row)
	go func() {
		defer rows.Close()
		defer close(ch)
		n := 0
		var header []string
		for rows.Next() {
			row, err := rows.Columns()
			if err != nil {
				ch <- Row{Index: n, Err: err}
				return
			}
			if n < headerRow {
				n++
				continue
			}
			if n == headerRow {
				header = make([]string, len(row))
				copy(header, row)
				n++
				continue
			}
			rowData := make([]string, len(row))
			copy(rowData, row)
			ch <- Row{Index: n, Headers: header, Row: rowData}
			n++
		}
	}()
	return ch, nil
}
