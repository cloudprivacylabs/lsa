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

package cmd

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	getschemaCmd.AddCommand(getschemaCSVCmd)
	getschemaCSVCmd.Flags().Int("headerRow", -1, "Header row 0-based (default: no header)")
}

// go run main.go getschema csv ../examples/ghp/GHP_data_capture_global-en-ca.csv
var getschemaCSVCmd = &cobra.Command{
	Use:   "csv",
	Short: "Write layered schema from CSV file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		f, err := os.Open(args[0])
		if err != nil {
			failErr(err)
		}

		reader := csv.NewReader(f)
		if err != nil {
			failErr(err)
		}
		headerRow, err := cmd.Flags().GetInt("headerRow")
		if err != nil {
			failErr(err)
		}
		for row := 0; row < headerRow; row++ {
			rowData, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				failErr(err)
			} else {
				templateData := map[string]interface{}{
					"columns": rowData,
				}
				js, err := json.MarshalIndent(templateData, "", "\t")
				if err != nil {
					failErr(err)
				}
				var buf bytes.Buffer
				buf.Write(js)
				str := buf.String()
				fmt.Println(str)
				//fmt.Println(string(js))
			}
		}
	},
}
