package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"text/template"

	jsonsch "github.com/cloudprivacylabs/lsa/pkg/json"
)

var inputFile = flag.String("input", "", "Input file")
var refBase = flag.String("base", "", "URL base for refs. If empty, all refs will be of the form #/...")
var layerID = flag.String("layerId", "https://hl7.org/fhir/{{.ValueType}}", "Layer ID template ({{.ValueType}}, {{.Ref}})")
var rootNodeID = flag.String("rootNodeId", "", "Root node ID template ({{.ValueType}}, {{.Ref}})")

func main() {
	flag.Parse()
	data, err := ioutil.ReadFile(*inputFile)
	if err != nil {
		panic(err)
	}
	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		panic(err)
	}
	mapping := v["discriminator"].(map[string]interface{})["mapping"].(map[string]interface{})
	layerIdT, err := template.New("layerId").Parse(*layerID)
	if err != nil {
		panic(err)
	}
	rootNodeIDT, err := template.New("rootNodeId").Parse(*rootNodeID)
	if err != nil {
		panic(err)
	}
	entities := make([]jsonsch.Entity, 0)
	for k, v := range mapping {
		e := jsonsch.Entity{
			Ref:       *refBase + v.(string),
			ValueType: k,
		}
		var b bytes.Buffer
		if err := layerIdT.Execute(&b, e); err != nil {
			panic(err)
		}
		e.LayerID = b.String()
		b = bytes.Buffer{}
		if err := rootNodeIDT.Execute(&b, e); err != nil {
			panic(err)
		}
		e.RootNodeID = b.String()
		entities = append(entities, e)
	}
	b, _ := json.MarshalIndent(entities, "", "  ")
	fmt.Println(string(b))
}
