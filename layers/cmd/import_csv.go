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
	"fmt"
	"io/ioutil"
	"os"

	dec "github.com/cloudprivacylabs/lsa/pkg/csv"
	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func init() {
	importCmd.AddCommand(importCSVCmd)
	importCSVCmd.Flags().String("spec", "", "Import specification JSON file")
	importCSVCmd.Flags().String("layerId", "", "Layer ID")
	importCSVCmd.MarkFlagRequired("spec")
}

var importCSVCmd = &cobra.Command{
	Use:   "csv",
	Short: "Import a CSV file as a schema a slice it into layers",
	Long: `The import specification is as follows:

{
  "attributeId": { attrSpec 
     "term": "string",
     "column": 0-based column index containing term data,
     "template": "term Go template, used to compute term value with {{.term}}, {{.data}}, and {{.row}} variables",
     "arrayTemplate": "Go template that determines array element type."
  },
  "layerType": "Overlay or Schema",
  "layerId": "id",
  "startRow": int (0),
  "nRows":  int (all rows),
  "terms": [
    {termSpec},
     ...
  ]
}

where termSpec is:

{
   "term": "string",
   "column": 0-based column index containing term data,
   "template": "term Go template, used to compute term value with {{.term}} and {{.data}} variables,
   "array": "boolean value denoting if the term is an array",
   "separator": "Array separator char"
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

		type importSpec struct {
			AttributeID dec.AttributeSpec `json:"attributeId"`
			LayerType   string            `json:"layerType"`
			LayerID     string            `json:"layerId"`
			TargetType  string            `json:"targetType"`
			StartRow    int               `json:"startRow"`
			NRows       int               `json:"nrows"`
			Terms       []dec.TermSpec    `json:"terms"`
		}

		var spec importSpec
		s, _ := cmd.Flags().GetString("spec")
		data, err := ioutil.ReadFile(s)
		if err != nil {
			failErr(err)
		}
		if err := json.Unmarshal(data, &spec); err != nil {
			failErr(err)
		}
		s, _ = cmd.Flags().GetString("layerId")
		if len(s) > 0 {
			spec.LayerID = s
		}
		if spec.LayerType == "Overlay" {
			spec.LayerType = ls.OverlayTerm
		} else if spec.LayerType == "Schema" {
			spec.LayerType = ls.SchemaTerm
		}
		layer, err := dec.Import(spec.AttributeID, spec.Terms, spec.StartRow, spec.NRows, records)
		if err != nil {
			failErr(err)
		}
		if len(spec.LayerType) > 0 {
			layer.SetLayerType(spec.LayerType)
		}
		if len(spec.LayerID) > 0 {
			layer.SetID(spec.LayerID)
		}
		if len(spec.TargetType) > 0 {
			layer.SetTargetType(spec.TargetType)
		}
		marshaled, err := ls.MarshalLayer(layer)
		if err != nil {
			failErr(err)
		}
		data, err = json.MarshalIndent(marshaled, "", "  ")
		if err != nil {
			failErr(err)
		}
		fmt.Println(string(data))
	},
}
