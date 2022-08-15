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
	"os"
	"strings"

	lscsv "github.com/cloudprivacylabs/lsa/pkg/csv"
)

// ReadSpreadsheetFile reads a CSV or Excel file. For CSV, it will
// look at the environment variable CSV_SEPARATOR. For Excel, it will
// load all the spreadsheets
func ReadSheets(fileName string) ([][][]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if strings.HasSuffix(strings.ToLower(fileName), ".csv") {
		data, err := lscsv.ReadCSV(file, os.Getenv("CSV_SEPARATOR"))
		if err != nil {
			return nil, err
		}
		return [][][]string{data}, nil
	}
	data, err := lscsv.ReadExcel(file)
	if err != nil {
		return nil, err
	}
	ret := make([][][]string, 0, len(data))
	for _, v := range data {
		ret = append(ret, v)
	}
	return ret, nil
}

func ReadSpreadsheetFile(fileName string) (map[string][][]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if strings.HasSuffix(strings.ToLower(fileName), ".csv") {
		data, err := lscsv.ReadCSV(file, os.Getenv("CSV_SEPARATOR"))
		if err != nil {
			return nil, err
		}
		return map[string][][]string{fileName: data}, nil
	}
	return lscsv.ReadExcel(file)
}
