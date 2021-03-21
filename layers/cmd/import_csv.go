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
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"os"

	dec "github.com/cloudprivacylabs/lsa/pkg/csv"
	"github.com/spf13/cobra"
)

func init() {
	importCmd.AddCommand(importCSVCmd)
	importCSVCmd.Flags().Bool("noheader", false, "The first row is not a header row")
	importCSVCmd.Flags().String("spec", "", "Import specification JSON file")
	importCSVCmd.MarkFlagRequired("spec")
}

var importCSVCmd = &cobra.Command{
	Use:   "csv",
	Short: "Import a CSV file as a schema a slice it into layers",
	Long: `The input CSV file is of the following format:
Term1, Term2, Term3, ...
value1, value2, value3,...
value1, value2, value3,...

The first line lists the terms in the schema base or the overlay. The
values are the values for the terms. A slicing specification
file can be supplied to create schema base and overlays.

{
  "attributeIdColumn": <columnIndex>,
  "objectType": "<target object type>",
  "layers": [
    "output": "outputFile",
    "type": "SchemaBase or Overlay",
    "columns": [
       {
          "index": 1,
          "name": "http://schemas.cloudprivacylabs.com/attributes/attributeName",
          "type": "@id, @value, @idlist, @valuelist"
       }
    ]
  ]
}

`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		f, err := os.Open(args[0])
		if err != nil {
			failErr(err)
		}
		records, err := csv.NewReader(f).ReadAll()
		if err != nil {
			failErr(err)
		}
		f.Close()

		var spec dec.ImportSpec
		s, _ := cmd.Flags().GetString("spec")
		data, err := ioutil.ReadFile(s)
		if err != nil {
			failErr(err)
		}
		if err := json.Unmarshal(data, &spec); err != nil {
			failErr(err)
		}
		noHeader, _ := cmd.Flags().GetBool("noheader")

		if !noHeader {
			records = records[1:]
		}
		overlays, err := dec.Import(spec, records)
		if err != nil {
			failErr(err)
		}
		for i := range overlays {
			data, err := json.MarshalIndent(overlays[i].MarshalExpanded(), "", "  ")
			if err != nil {
				failErr(err)
			}
			ioutil.WriteFile(spec.Layers[i].Output, data, 0664)
		}
	},
}
