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
// 	"fmt"

// 	"github.com/piprate/json-gold/ld"
// 	"github.com/spf13/cobra"
// )

// func init() {
// 	rootCmd.AddCommand(rdfCmd)
// 	rdfCmd.AddCommand(nquadsCmd)
// 	rdfCmd.AddCommand(parseNquadsCmd)
// }

// var rdfCmd = &cobra.Command{
// 	Use:   "rdf",
// 	Short: "Export/Import RDF",
// }

// var nquadsCmd = &cobra.Command{
// 	Use:   "nquads",
// 	Short: "Return nquads from a JSON-LD document",
// 	Args:  cobra.MaximumNArgs(1),
// 	Run: func(cmd *cobra.Command, args []string) {
// 		var input interface{}
// 		if err := readJSONFileOrStdin(args, &input); err != nil {
// 			failErr(err)
// 		}
// 		options := ld.NewJsonLdOptions("")
// 		options.Format = "application/n-quads"
// 		proc := ld.NewJsonLdProcessor()
// 		ret, err := proc.ToRDF(input, options)
// 		if err != nil {
// 			failErr(err)
// 		}
// 		fmt.Println(ret)
// 	},
// }

// var parseNquadsCmd = &cobra.Command{
// 	Use:   "parsenquads",
// 	Short: "Parse nquads and output JSON-LD document",
// 	Args:  cobra.MaximumNArgs(1),
// 	Run: func(cmd *cobra.Command, args []string) {
// 		data, err := readFileOrStdin(args)
// 		if err != nil {
// 			failErr(err)
// 		}
// 		proc := ld.NewJsonLdProcessor()
// 		jsonld, err := proc.FromRDF(string(data), nil)
// 		if err != nil {
// 			failErr(err)
// 		}
// 		data, _ = json.MarshalIndent(jsonld, "", "  ")
// 		fmt.Println(string(data))
// 	},
// }
