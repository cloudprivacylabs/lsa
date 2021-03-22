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
	"encoding/json"
	"fmt"
	"strings"

	ljson "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// ToJSON converts the CSV row into a JSON document using the given profile
func ToJSON(row []string, profile IngestionProfile) (map[string]interface{}, error) {
	output := make(map[string]interface{}, len(profile.Columns))
	for _, col := range profile.Columns {
		if col.Index < len(row) {
			data := row[col.Index]
			var value interface{}
			switch col.Type {
			case "string", "":
				value = data
			case "number":
				value = json.Number([]byte(data))
			case "boolean":
				data = strings.ToLower(data)
				if data == "true" {
					value = true
				} else if data == "false" || data == "" {
					value = false
				} else {
					return nil, fmt.Errorf("Invalid boolean value: %s", data)
				}
			}
			output[col.Name] = value
		}
	}
	return output, nil
}

// Ingest converts a CSV row into JSON, and then uses JSON ingest to
// convert data based on the schema
func Ingest(baseID string, row []string, profile IngestionProfile, schema *ls.SchemaLayer) (ls.DocumentNode, error) {
	data, err := ToJSON(row, profile)
	if err != nil {
		return nil, err
	}
	return ljson.Ingest(baseID, data, schema)
}
