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
	"encoding/json"
	"io/ioutil"

	"github.com/spf13/cobra"

	jsonsch "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/jsonld"
	"github.com/cloudprivacylabs/lsa/pkg/layers"
)

func init() {
	importCmd.AddCommand(importJSONCmd)
}

var importJSONCmd = &cobra.Command{
	Use:   "jsonschema",
	Short: "Import a JSON schema and slice into its layers",
	Long: `Input a JSON file of the format:

{
  "entities": [
     {
       "ref": "<reference to schema>",
       "name": "entity name"
     },
    ...
   ],
  "schemaManifest": "schema output file",
  "schemaId": "schema id",
  "objectType": "object type",
  "layers": [
     {
         "@id": "output object id",
         "type": "Schema" or "Overlay",
         "terms": [ terms to include in the layer ],
         "file": "output file"
     },
     ...
   ]
}

Each element in the input will be compiled, and sliced into a layers, and saved 
as a schema+layers. The file names, @ids, and terms
are Go templates, you can reference entity names and references using {{.name}} and {{.ref}}.
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		inputData, err := ioutil.ReadFile(args[0])
		if err != nil {
			failErr(err)
		}

		type request struct {
			Entities   []jsonsch.Entity   `json:"entities"`
			SchemaID   string             `json:"schemaId"`
			Schema     string             `json:"schemaManifest"`
			ObjectType string             `json:"objectType"`
			Layers     []SliceByTermsSpec `json:"layers"`
		}

		var req request
		if err := json.Unmarshal(inputData, &req); err != nil {
			failErr(err)
		}
		if len(req.Entities) == 0 {
			return
		}

		for i := range req.Entities {
			req.Entities[i].SchemaName = execTemplate(req.SchemaID, map[string]interface{}{"name": req.Entities[i].Name, "ref": req.Entities[i].Ref})
		}

		results, err := jsonsch.CompileAndImport(req.Entities)
		if err != nil {
			failErr(err)
		}

		for _, item := range results {
			layerIDs := make([]string, 0)
			baseID := ""
			tdata := map[string]interface{}{"name": item.Entity.Name, "ref": item.Entity.Ref}
			for _, ovl := range req.Layers {
				layer, err := ovl.Slice(item.Layer, execTemplate(req.ObjectType, tdata), tdata)
				if err != nil {
					failErr(err)
				}
				if layer.GetLayerType() == layers.SchemaTerm {
					if baseID != "" {
						fail("Multiple schemas")
					}
					baseID = layer.GetID()
				} else {
					layerIDs = append(layerIDs, layer.GetID())
				}
				data, err := json.MarshalIndent(jsonld.MarshalLayer(layer), "", "  ")
				if err != nil {
					failErr(err)
				}
				ioutil.WriteFile(execTemplate(ovl.File, tdata), data, 0664)
			}

			if len(req.Schema) > 0 {
				sch := layers.SchemaManifest{
					ID:         execTemplate(req.SchemaID, tdata),
					TargetType: execTemplate(req.ObjectType, tdata),
					Schema:     baseID,
					Overlays:   layerIDs,
				}
				data, _ := json.MarshalIndent(jsonld.MarshalSchemaManifest(&sch), "", "  ")
				ioutil.WriteFile(execTemplate(req.Schema, tdata), data, 0664)
			}
		}
	},
}
