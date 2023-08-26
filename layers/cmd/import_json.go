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

	"github.com/cloudprivacylabs/lpg/v2"
	jsonsch "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/jsonld"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func init() {
	importCmd.AddCommand(importJSONCmd)
	importJSONCmd.Flags().String("output", "graph", "Output format: graph (single graph containing the schema(s), or sliced for sliced output)")
}

type ImportJSONSchemaRequest struct {
	Entities []jsonsch.Entity   `json:"entities"`
	SchemaID string             `json:"schemaId"`
	Schema   string             `json:"schemaVariant"`
	Layers   []SliceByTermsSpec `json:"layers"`
}

func (req *ImportJSONSchemaRequest) CompileAndImport() (*lpg.Graph, []jsonsch.EntityLayer, error) {
	c, err := jsonsch.CompileEntities(req.Entities...)
	if err != nil {
		return nil, nil, err
	}
	g := lpg.NewGraph()
	layers, err := jsonsch.BuildEntityGraph(g, ls.SchemaTerm.Name, jsonsch.LinkRefsByLayerID, c...)
	return g, layers, err
}

func makeEntityTemplateData(e jsonsch.Entity) map[string]interface{} {
	return map[string]interface{}{
		"ref":        e.Ref,
		"layerId":    e.LayerID,
		"rootNodeId": e.RootNodeID,
		"valueType":  e.ValueType,
	}
}

func (req *ImportJSONSchemaRequest) Slice(index int, item jsonsch.EntityLayer) ([]*ls.Layer, error) {
	layerIDs := make([]string, 0)
	baseID := ""
	returnLayers := make([]*ls.Layer, 0)
	tdata := makeEntityTemplateData(item.Entity.Entity)
	for _, ovl := range req.Layers {
		layer, err := ovl.Slice(item.Layer, item.Entity.ValueType, tdata)
		if err != nil {
			return nil, err
		}
		if layer.GetLayerType() == ls.SchemaTerm.Name {
			if baseID != "" {
				return nil, fmt.Errorf("Multiple schemas")
			}
			baseID = layer.GetID()
		} else {
			layerIDs = append(layerIDs, layer.GetID())
		}
		returnLayers = append(returnLayers, layer)
	}

	return returnLayers, nil
}

var importJSONCmd = &cobra.Command{
	Use:   "jsonschema",
	Short: "Import a JSON schema and slice into its layers",
	Long: `Input a JSON file of the format:

{
  "entities": [
     {
       "ref": "reference to schema, e.g. myFile.json#/definitions/item",
       "layerId": "Id of the schema or the overlay",
       "rootNodeId": "Id of the root node. If none, same as layerId",
       "valueType" "Type of the value defined in the schema" 
     },
    ...
   ],
  "schemaVariant": "schema output file",
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
as a schema+layers. All fields are Go templates evaluated using the current entity: 
{{.entity.ref}}, {{.entity.layerId}}...
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

		g, results, err := req.CompileAndImport()
		if err != nil {
			failErr(err)
		}

		output, _ := cmd.Flags().GetString("output")
		switch output {
		case "graph":
			m := ls.JSONMarshaler{}
			out, _ := m.Marshal(g)
			fmt.Println(string(out))
		case "sliced":
			for index, item := range results {
				layers, err := req.Slice(index, item)
				if err != nil {
					failErr(err)
				}
				tdata := makeEntityTemplateData(item.Entity.Entity)
				for i := range layers {
					marshaled, err := jsonld.MarshalLayer(layers[i])
					if err != nil {
						failErr(err)
					}
					data, err := json.MarshalIndent(marshaled, "", "  ")
					if err != nil {
						failErr(err)
					}
					ioutil.WriteFile(execTemplate(req.Layers[i].File, tdata), data, 0664)
				}
			}
		}
	},
}
