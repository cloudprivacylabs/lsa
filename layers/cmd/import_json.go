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
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"

	jsonsch "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func init() {
	importCmd.AddCommand(importJSONCmd)
}

type importEntity struct {
	jsonsch.Entity
	Name       string `json:"name"`
	SchemaName string `json:"fileName"`
}

type ImportJSONSchemaRequest struct {
	Entities   []importEntity     `json:"entities"`
	SchemaID   string             `json:"schemaId"`
	Schema     string             `json:"schemaManifest"`
	ObjectType string             `json:"objectType"`
	Layers     []SliceByTermsSpec `json:"layers"`
}

func (req *ImportJSONSchemaRequest) SetSchemaNames() {
	for i := range req.Entities {
		req.Entities[i].SchemaName = execTemplate(req.SchemaID, map[string]interface{}{"name": req.Entities[i].Name, "ref": req.Entities[i].Ref})
	}
}

func (req *ImportJSONSchemaRequest) CompileAndImport() ([]jsonsch.EntityLayer, error) {
	req.SetSchemaNames()
	e := make([]jsonsch.Entity, 0, len(req.Entities))
	for _, x := range req.Entities {
		e = append(e, x.Entity)
	}
	c, err := jsonsch.CompileEntities(e...)
	if err != nil {
		return nil, err
	}
	return jsonsch.BuildEntityGraph(ls.SchemaTerm, c...)
}

func (req *ImportJSONSchemaRequest) Slice(index int, item jsonsch.EntityLayer) ([]*ls.Layer, *ls.SchemaManifest, error) {
	layerIDs := make([]string, 0)
	baseID := ""
	returnLayers := make([]*ls.Layer, 0)
	tdata := map[string]interface{}{"name": req.Entities[index].Name, "ref": item.Entity.Ref}
	for _, ovl := range req.Layers {
		layer, err := ovl.Slice(item.Layer, execTemplate(req.ObjectType, tdata), tdata)
		if err != nil {
			return nil, nil, err
		}
		if layer.GetLayerType() == ls.SchemaTerm {
			if baseID != "" {
				return nil, nil, fmt.Errorf("Multiple schemas")
			}
			baseID = layer.GetID()
		} else {
			layerIDs = append(layerIDs, layer.GetID())
		}
		returnLayers = append(returnLayers, layer)
	}
	sch := ls.SchemaManifest{
		ID:         execTemplate(req.SchemaID, tdata),
		TargetType: execTemplate(req.ObjectType, tdata),
		Schema:     baseID,
		Overlays:   layerIDs,
	}
	return returnLayers, &sch, nil
}

var importJSONCmd = &cobra.Command{
	Use:   "jsonschema",
	Short: "Import a JSON schema and slice into its layers",
	Long: `Input a JSON file of the format:

{
  "entities": [
     {
       "ref": "<reference to schema>",
       "name": "entity name",
       "id": "ID of the output type"
     },
    ...
   ],
  "schemaManifest": "schema output file",
  "schemaId": "schema id",
  "objectType": "object type",
  "layers": [
     {
         "@id": "output layer id",
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

		var req ImportJSONSchemaRequest
		if err := json.Unmarshal(inputData, &req); err != nil {
			failErr(err)
		}
		if len(req.Entities) == 0 {
			return
		}

		results, err := req.CompileAndImport()
		if err != nil {
			failErr(err)
		}
		for index, item := range results {
			layers, sch, err := req.Slice(index, item)
			if err != nil {
				failErr(err)
			}
			tdata := map[string]interface{}{"name": req.Entities[index].Name, "ref": item.Entity.Ref}
			for i := range layers {
				data, err := json.MarshalIndent(ls.MarshalLayer(layers[i]), "", "  ")
				if err != nil {
					failErr(err)
				}
				ioutil.WriteFile(execTemplate(req.Layers[i].File, tdata), data, 0664)
			}
			if len(req.Schema) > 0 {
				data, _ := json.MarshalIndent(ls.MarshalSchemaManifest(sch), "", "  ")
				ioutil.WriteFile(execTemplate(req.Schema, tdata), data, 0664)
			}
		}
	},
}
