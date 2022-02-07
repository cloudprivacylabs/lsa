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

// import (
// 	"encoding/json"
// 	"errors"
// 	"io"
// 	"os"

// 	"github.com/spf13/cobra"
// )

// func init() {
// 	//getschemaCSVCmd.Flags().String("file", "", "Use the given file")
// 	getschemaCmd.AddCommand(getschemaJSONCmd)
// 	getschemaJSONCmd.Flags().Int("startRow", 1, "Start row 0-based (default 1)")
// 	getschemaJSONCmd.Flags().Int("endRow", -1, "End row 0-based")
// 	getschemaJSONCmd.Flags().Int("headerRow", -1, "Header row 0-based (default: no header)")
// }

// var getschemaJSONCmd = &cobra.Command{
// 	Use:   "json",
// 	Short: "Write layered schema from JSON file",
// 	Args:  cobra.MaximumNArgs(1),
// 	Run: func(cmd *cobra.Command, args []string) {
// 		f, err := os.Open(args[0])
// 		if err != nil {
// 			failErr(err)
// 		}
// 		decoder := json.NewDecoder(f)
// 		err = decoder.Decode(os.Stdout)
// 		if err != nil {
// 			var syntaxError *json.SyntaxError
// 			var unmarshalTypeError *json.UnmarshalTypeError
// 			var invalidUnmarshalError *json.InvalidUnmarshalError

// 			switch {
// 			case errors.As(err, &syntaxError):
// 			case errors.Is(err, io.EOF):
// 			default:
// 				failErr(err)
// 			}
// 		}

// 	},
// }
