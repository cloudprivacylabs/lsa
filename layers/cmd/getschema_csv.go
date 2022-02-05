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
	getschemaCSVCmd.Flags().Int("headerRow", 0, "Header row 0-based (default: 1st row)")
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
		for row := 0; row <= headerRow; row++ {
			rowData, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				failErr(err)
			} else if row < headerRow {
				continue
			} else {
				// var jsObj map[string]interface{}
				// for i := range rowData {
				// 	jsObj = map[string]interface{}{
				// 		"@id":           rowData[i],
				// 		"@type":         "Value",
				// 		"attributeName": rowData[i],
				// 	}
				// }
				// attributes := []interface{}{}
				// for _, hdr := range jsObj {
				// 	attributes = append(attributes, hdr)
				// }
				// attributeList := make(map[string]interface{})
				// attributeList["attributesList"] = attributes
				// layer := make(map[string]interface{})
				// layer["layer"] = attributeList
				// var LS struct {
				// 	Layer struct {
				// 		AttributeList []struct {
				// 			ID            string `json:"@id"`
				// 			AttributeName string `json:"attributeName"`
				// 			Type          string `json:"@type"`
				// 		} `json:"attributeList"`
				// 	} `json:"layer"`
				// }

				type AttributeList struct {
					Id            string `json:"@id"`
					AttributeName string `json:"attributeName"`
					Types         string `json:"@type"`
				}

				type Layer struct {
					AttributeList []AttributeList `json:"attributeList"`
				}

				type LS struct {
					Layer Layer `json:"layer"`
				}

				attHeaders := []AttributeList{}
				for i := range rowData {
					attHeaders = append(attHeaders, AttributeList{
						Id:            rowData[i],
						AttributeName: rowData[i],
						Types:         "Value",
					})
				}

				test := LS{
					Layer: Layer{
						AttributeList: attHeaders,
					},
				}

				// for i := range rowData {
				// }
				// marshal csv to json, unmarshal json to LS struct
				js, err := json.Marshal(rowData)
				if err != nil {
					failErr(err)
				}

				json.Unmarshal(js, &test)
				//buffer := &bytes.Buffer{}
				//fmt.Println(test)

				// gob.NewEncoder(buffer).Encode(rowData[0])
				// byteSlice := buffer.Bytes()
				// fmt.Println(string(byteSlice))
				//json.Unmarshal(byteSlice, &LS)

				// type Inner struct {
				// 	Object map[string]interface{} `json:"attributes"`
				// 	// AttributeName map[string]interface{} `json:"@attributeName"`
				// 	// Type          map[string]interface{} `json:"@type"`
				// }
				// type Outer struct {
				// 	AttributeList []Inner `json:"attributeList"`
				// }
				// type Outmost struct {
				// 	Layer Outer `json:"layer"`
				// }
				// var cont Outmost
				// templateData := map[string]interface{}{
				// 	"layer": map[string]interface{}{
				// 		"attributeList": map[string]interface{}{
				// 			"@id":           rowData,
				// 			"attributeName": rowData,
				// 			"Value":         rowData,
				// 		},
				// 	},
				// }

				//json.Unmarshal(byteSlice, &LS)

				// cont.Layer.AttributeList = append(cont.Layer.AttributeList, Inner{Object: map[string]interface{}{
				// 	"@id":            rowData[0],
				// 	"@attributeName": rowData[0],
				// 	"@type":          rowData[0],
				// }})

				// json.Unmarshal(byteSlice, &cont)

				js, err = json.MarshalIndent(test, "", "\t")
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
