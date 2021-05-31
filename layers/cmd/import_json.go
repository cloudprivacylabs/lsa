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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"text/template"

	"github.com/bserdar/digraph"
	"github.com/spf13/cobra"

	jsonsch "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/jsonld"
	"github.com/cloudprivacylabs/lsa/pkg/layers"
)

func init() {
	importCmd.AddCommand(importJSONCmd)
}

var importJSONCmd = &cobra.Command{
	Use:   "json",
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

		type overlay struct {
			Type         string   `json:"type"`
			Terms        []string `json:"terms"`
			termst       []*template.Template
			File         string `json:"file"`
			filet        *template.Template
			ID           string `json:"@id"`
			idt          *template.Template
			IncludeEmpty bool `json:"includeEmpty"`
		}

		type request struct {
			Entities    []jsonsch.Entity `json:"entities"`
			SchemaID    string           `json:"schemaId"`
			schemaIDt   *template.Template
			Schema      string `json:"schemaManifest"`
			schemat     *template.Template
			ObjectType  string `json:"objectType"`
			objectTypet *template.Template
			Layers      []overlay `json:"layers"`
		}

		var req request
		if err := json.Unmarshal(inputData, &req); err != nil {
			failErr(err)
		}
		if len(req.Entities) == 0 {
			return
		}

		req.schemaIDt = template.Must(template.New("schemaID").Parse(req.SchemaID))
		req.schemat = template.Must(template.New("schema").Parse(req.Schema))
		req.objectTypet = template.Must(template.New("objectType").Parse(req.ObjectType))
		for i := range req.Layers {
			for _, x := range req.Layers[i].Terms {
				req.Layers[i].termst = append(req.Layers[i].termst, template.Must(template.New(fmt.Sprintf("terms-%d", i)).Parse(x)))
			}
			req.Layers[i].filet = template.Must(template.New(fmt.Sprintf("file-%d", i)).Parse(req.Layers[i].File))
			req.Layers[i].idt = template.Must(template.New(fmt.Sprintf("id-%d", i)).Parse(req.Layers[i].ID))
		}

		exec := func(t *template.Template, entity jsonsch.Entity) string {
			tdata := map[string]interface{}{"name": entity.Name, "ref": entity.Ref}
			out := bytes.Buffer{}
			if err := t.Execute(&out, tdata); err != nil {
				panic(fmt.Errorf("During %s: %w", entity.Name, err))
			}
			return out.String()
		}
		for i := range req.Entities {
			req.Entities[i].SchemaName = exec(req.schemaIDt, req.Entities[i])
		}

		compiled, err := jsonsch.Compile(req.Entities)
		if err != nil {
			failErr(err)
		}
		results, err := jsonsch.Import(compiled)
		if err != nil {
			failErr(err)
		}

		for i, item := range results {
			layerIDs := make([]string, 0)
			baseID := ""
			for _, ovl := range req.Layers {
				inclterms := make(map[string]struct{})
				for _, x := range ovl.termst {
					inclterms[exec(x, req.Entities[i])] = struct{}{}
				}
				var layerType string
				if ovl.Type == "Overlay" || ovl.Type == layers.OverlayTerm {
					layerType = layers.OverlayTerm
				}
				if ovl.Type == "Schema" || ovl.Type == layers.SchemaTerm {
					layerType = layers.SchemaTerm
				}
				if len(layerType) == 0 {
					fail("Layer type unspecified")
				}

				layer := item.Layer.Slice(layerType, func(layer *layers.Layer, node *digraph.Node) *digraph.Node {
					properties := make(map[string]interface{})
					payload := node.Payload.(*layers.SchemaNode)
					for k, v := range payload.Properties {
						if _, ok := inclterms[k]; ok {
							properties[k] = v
						}
					}
					if ovl.IncludeEmpty || len(properties) > 0 {
						newNode := layer.NewNode(node.Label(), payload.GetTypes()...)
						newNode.Payload.(*layers.SchemaNode).Properties = properties
						return newNode
					}
					return nil
				})
				layer.SetID(exec(ovl.idt, req.Entities[i]))
				if ovl.Type == "Overlay" || ovl.Type == layers.OverlayTerm {
					layer.RootNode.Payload.(*layers.SchemaNode).AddTypes(layers.OverlayTerm)
					layerIDs = append(layerIDs, layer.GetID())
				}
				if ovl.Type == "Schema" || ovl.Type == layers.SchemaTerm {
					layer.RootNode.Payload.(*layers.SchemaNode).AddTypes(layers.SchemaTerm)
					if baseID != "" {
						fail("Multiple schemas")
					}
					baseID = layer.GetID()
				}
				if len(layer.GetLayerType()) == 0 {
					fail("Layer type not specified")
				}
				layer.RootNode.Payload.(*layers.SchemaNode).Properties[layers.TargetType] = exec(req.objectTypet, req.Entities[i])
				data, err := json.MarshalIndent(jsonld.MarshalLayer(layer), "", "  ")
				if err != nil {
					failErr(err)
				}
				ioutil.WriteFile(exec(ovl.filet, req.Entities[i]), data, 0664)
			}

			if len(req.Schema) > 0 {
				sch := layers.SchemaManifest{
					ID:         exec(req.schemaIDt, req.Entities[i]),
					TargetType: exec(req.objectTypet, req.Entities[i]),
					Schema:     baseID,
					Overlays:   layerIDs,
				}
				data, _ := json.MarshalIndent(jsonld.MarshalSchemaManifest(&sch), "", "  ")
				ioutil.WriteFile(exec(req.schemat, req.Entities[i]), data, 0664)
			}
		}
	},
}
