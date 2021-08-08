// Copyright 2021 Cloud Privacy Labs, LLC

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package csv

import (
	"strings"
)

// IngestionProfile defines how CSV columns are mapped to data by
// providing JSON field names and JSON data types for each column
type IngestionProfile struct {
	Columns []ColumnProfile `json:"columns" yaml:"columns"`
}

// ColumnProfile describes a CSV column
type ColumnProfile struct {
	// The 0-based column index
	Index int `json:"column" yaml:"column"`
	// Name specifies the JSON field name during data ingestion/output
	// and schema field name during schema import
	Name string `json:"name" yaml:"column"`
}

// DefaultProfile creates a column profile using the colum names as
// field names
func DefaultProfile(row []string) ([]ColumnProfile, error) {
	ret := make([]ColumnProfile, 0, len(row))
	for index, c := range row {
		if len(c) > 0 {
			ret = append(ret, ColumnProfile{Index: index, Name: strings.TrimSpace(c)})
		}
	}
	return ret, nil
}

// ToJSON converts the CSV row into a JSON document using the given profile
func (profile IngestionProfile) ToJSON(row []string) (map[string]interface{}, error) {
	output := make(map[string]interface{}, len(profile.Columns))
	for _, col := range profile.Columns {
		if col.Index < len(row) {
			data := row[col.Index]
			output[col.Name] = data
		}
	}
	return output, nil
}
