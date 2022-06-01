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

package cmdutil

import (
	"encoding/csv"
	"os"
	"strings"

	"github.com/xuri/excelize/v2"
)

// ReadSpreadsheetFile reads a CSV or Excel file. For CSV, it will
// look at the environment variable CSV_SEPARATOR. For Excel, it will
// load all the spreadsheets
func ReadSheets(fileName string) ([][][]string, error) {
	if strings.HasSuffix(strings.ToLower(fileName), ".csv") {
		file, err := os.Open(fileName)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		reader := csv.NewReader(file)
		reader.LazyQuotes = true
		reader.FieldsPerRecord = -1
		if s := os.Getenv("CSV_SEPARATOR"); len(s) > 0 {
			reader.Comma = rune(s[0])
		}
		data, err := reader.ReadAll()
		if err != nil {
			return nil, err
		}
		return [][][]string{data}, nil
	}
	f, err := excelize.OpenFile(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sheets := f.GetSheetList()
	ret := make([][][]string, 0, len(sheets))
	for _, sheet := range sheets {
		rows, err := f.GetRows(sheet)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rows)
	}
	return ret, nil
}

func ReadSpreadsheetFile(fileName string) (map[string][][]string, error) {
	if strings.HasSuffix(strings.ToLower(fileName), ".csv") {
		file, err := os.Open(fileName)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		reader := csv.NewReader(file)
		reader.LazyQuotes = true
		reader.FieldsPerRecord = -1
		if s := os.Getenv("CSV_SEPARATOR"); len(s) > 0 {
			reader.Comma = rune(s[0])
		}
		data, err := reader.ReadAll()
		if err != nil {
			return nil, err
		}
		return map[string][][]string{fileName: data}, nil
	}
	f, err := excelize.OpenFile(fileName)
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
