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
