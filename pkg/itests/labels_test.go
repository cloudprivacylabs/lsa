package itests

import (
	"encoding/json"
	"testing"

	jsoningest "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func TestProcessLabeledAs(t *testing.T) {
	schStr := `
	{
		"@context": {
			"ls":"https://lschema.org/"
		},
		"@type":"ls:Schema",
		"@id": "http://testschema",
		"https://lschema.org/layer": {
			"@type": ["https://lschema.org/Object",
					  "https://lschema.org/valueType"],
			"@id": "root",
			"ls:Object/attributes": [
				{
					"@id":  "attr1",
					"@type": "ls:Value",
					"labeledAs": "SOMELABEL"
				},
				{
					"@id":  "attr2" ,
					"@type": "ls:Value",
					"ls:privacy": [
						{
							"@value": "flg1",
							"labeledAs": "ANOTHERLABEL"
						}
					]
				},
				{
					"@id":"attr3",
					"@type": "ls:Value",
					"ls:privacy": [
						{"@value": "flg2"},
						{"@value": "flg3"}
					]
				}
			]
		}
	}
	`

	// inputStr := `{
	// 	"field1": "value1",
	// 	"field2": {
	// 	   "t": "type1"
	// 	}
	// }`

	var schMap interface{}
	if err := json.Unmarshal([]byte(schStr), &schMap); err != nil {
		t.Fatal(err)
	}

	schema, err := ls.UnmarshalLayer(schMap, nil)
	if err != nil {
		t.Error(err)
	}

	builder := ls.NewGraphBuilder(nil, ls.GraphBuilderOptions{
		EmbedSchemaNodes: true,
	})

	parser := jsoningest.Parser{Layer: schema}

	_, err = jsoningest.IngestBytes(ls.DefaultContext(), "", []byte(schStr), parser, builder)
	if err != nil {
		t.Fatal(err)
	}

	ls.ProcessLabeledAs(builder.GetGraph())

	var seenL1, seenL2 = false, false
	for nodeItr := builder.GetGraph().GetNodes(); nodeItr.Next(); {
		node := nodeItr.Node()
		if node.HasLabel("SOMELABEL") {
			seenL1 = true
		}
		if node.HasLabel("ANOTHERLABEL") {
			seenL2 = true
		}
	}
	if !seenL1 && !seenL2 {
		t.Fatal("Labels cannot be found")
	}
}
