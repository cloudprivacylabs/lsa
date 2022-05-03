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
	"io/ioutil"
	"os"
	"text/template"

	dec "github.com/cloudprivacylabs/lsa/pkg/csv"
	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func init() {
	importCmd.AddCommand(importCSVCmd)
	importCSVCmd.Flags().String("spec", "", "Import specification JSON file")
	importCSVCmd.MarkFlagRequired("spec")
}

var importCSVCmd = &cobra.Command{
	Use:   "csv",
	Short: "Import a CSV file as a schema a slice it into layers",
	Long: `The CSV file is organized as follows:

  * Each row describes an attribute of the schema.
  * Each column describes a property of that attribute.

The import specification is as follows:

{
  // All string fields in the spec are Go templates.
  // Those that are evaluated once are  evaluated with {{.rows}} in the
  // template context. The 'rows' are the range of rows selected by startRow 
  // and nRows. That is, {{index (index .row 0) 0}} returns row[0][0] for the
  // selected range.
  // Those that are evaluated for each row have {{.row}} in the 
  // template context, which is the current row.

  "layerType": "Overlay or Schema. Evaluted once.",
  "layerId": "ID of the schema or the overlay. Evaluated once.",
  "valueType": "The value type of the layer. Evaluated once.",
  "rootId": "The ID of the schema root node. Evaluated once.",
  "entityIdRows": [ row indexes (0-based) containing the entity id],
  "entityId": "Evaluated for each row. If evaluates to "true", the attribute is part of entity id",
  "required": "Evaluated for each row. If evaluated to "true", attribute is required",
  "startRow": int (0), // First row in CSV file that has the schema spec
  "nRows":  int (all rows), // Number of rows in the schema spec

  "attributeId": "Evaluated for each row. Attribute id template.",
  "terms": [ // Defines the terms to be included in the schema
    {termSpec},
     ...
  ]
}

where termSpec is:

{
   "term": "string", // The attribute annotation key
   "template": "term value Go template, used to compute term value with {{.term}} and  {{.row}} variables.",
   "array": "boolean value denoting if the result of template evaluation is an array property. If so, its elements are separated with the separator char",
   "separator": "Array separator char"
}

For example:
{
  "layerType": "Schema",
  "layerId": "https://test.org/Person/schema
  "valueType": "Person",
  "rootId": "https://test.org/Person",
  "entityIdRows": [ 0 ],
  "startRow": 0,
  "nRows": 10,
  "attributeId": "https://test.org/Person/{{index .row 0}}",
  "entityId": "{{if eq (index .row 3) "yes"}}true{{end}}",
  "terms": [
    {
      "term": "https://lschema.org/valueType",
      "template": "{{index .row 1}}"
    },
    {
      "term": "https://lschema.org/description",
      "template": "{{index .row 2}}"
    }
  ]
}

With the input:
id,string,,yes
firstName,string,first name,
lastName,string,last name,

will result:

{
  "@id": "https://test.org/Person/schema",
  "@type": "Schema",
  "valueType": "Person",
  "entityIdFields": [ "https://test.org/Person/id"],
  "layer": {
    "@id": "https://test.org/Person",
    "@type": "Object",
    "attributeList": [
      {
        "@id": "https://test.org/Person/id",
        "@type": "Value",
        "valueType": "string"
      },
      {
        "@id": "https://test.org/Person/firstName",
        "@type": "Value",
        "valueType": "string",
        "description": "first name"
      },
      {
        "@id": "https://test.org/Person/lastName",
        "@type": "Value",
        "valueType": "string",
        "description": "last name"
      }
    ]
  }
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

		exec := func(tmpl string, data interface{}) string {
			t, err := template.New("").Parse(tmpl)
			if err != nil {
				failErr(err)
			}
			var buf bytes.Buffer
			if err := t.Execute(&buf, data); err != nil {
				failErr(err)
			}
			return buf.String()
		}

		type importSpec struct {
			AttributeID  string         `json:"attributeId"`
			LayerType    string         `json:"layerType"`
			LayerID      string         `json:"layerId"`
			RootID       string         `json:"rootId"`
			ValueType    string         `json:"valueType"`
			EntityIDRows []int          `json:"entityIdRows"`
			EntityID     string         `json:"entityId"`
			Required     string         `json:"required"`
			StartRow     int            `json:"startRow"`
			NRows        int            `json:"nrows"`
			Terms        []dec.TermSpec `json:"terms"`
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
		rows := records[spec.StartRow:]
		if spec.NRows > 0 && spec.NRows > len(rows) {
			rows = rows[:spec.NRows]
		}
		layerType := exec(spec.LayerType, map[string]interface{}{"rows": rows})
		if layerType == "Overlay" {
			layerType = ls.OverlayTerm
		} else if layerType == "Schema" {
			layerType = ls.SchemaTerm
		}
		layer, err := dec.Import(spec.AttributeID, spec.Terms, spec.StartRow, spec.NRows, spec.EntityIDRows, spec.EntityID, spec.Required, records)
		if err != nil {
			failErr(err)
		}
		if len(layerType) > 0 {
			layer.SetLayerType(layerType)
		}
		if layerID := exec(spec.LayerID, map[string]interface{}{"rows": rows}); len(layerID) > 0 {
			layer.SetID(layerID)
		}
		if rootID := exec(spec.RootID, map[string]interface{}{"rows": rows}); len(rootID) > 0 {
			ls.SetNodeID(layer.GetSchemaRootNode(), rootID)
		}
		if valueType := exec(spec.ValueType, map[string]interface{}{"rows": rows}); len(valueType) > 0 {
			layer.SetValueType(valueType)
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
