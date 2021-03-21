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
	// For schema import, type is @id, @value, @list, or @idlist. If
	// omitted, it is assumed to be @value, unless the term is
	// known. @list or @set is comma separated list
	//
	// For data ingestion, type is the JSON data type, string, number, or boolean
	Type string `json:"type" yaml:"column"`
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
