package itests

import (
	"encoding/json"
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func TestProcessLabeledAs(t *testing.T) {
	schStr := `
	{
		"@context": "../../schemas/ls.json",
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
					"ls:labeledAs": "SOMELABEL"
				},
				{
					"@id":  "attr2" ,
					"@type": "ls:Value",
					"ls:labeledAs": "ANOTHERLABEL",
					"ls:privacy": [
						{
							"@value": "flg1"
						}
					]
				},
				{
					"@id":"attr3",
					"@type": "ls:Value",
					"ls:labeledAs": "thirdlabel",
					"ls:privacy": [
						{"@value": "flg2"},
						{"@value": "flg3"}
					]
				}
			]
		}
	}
	`

	var schMap interface{}
	if err := json.Unmarshal([]byte(schStr), &schMap); err != nil {
		t.Fatal(err)
	}

	schema, err := ls.UnmarshalLayer(schMap, nil)
	if err != nil {
		t.Error(err)
	}

	c := ls.Compiler{}
	layer, err := c.CompileSchema(ls.DefaultContext(), schema)
	if err != nil {
		t.Error(err)
	}

	ls.ProcessLabeledAs(layer.Graph)

	var seenL1, seenL2, seenL3 = false, false, false
	for nodeItr := layer.Graph.GetNodes(); nodeItr.Next(); {
		node := nodeItr.Node()
		if node.HasLabel("SOMELABEL") {
			seenL1 = true
		}
		if node.HasLabel("ANOTHERLABEL") {
			seenL2 = true
		}
		if node.HasLabel("thirdlabel") {
			seenL3 = true
		}
		if _, ok := node.GetProperty(ls.LabeledAsTerm); ok {
			t.Fatalf("Did not remove LabeledAsTerm")
		}
	}
	if !seenL1 && !seenL2 && !seenL3 {
		t.Fatal("Labels cannot be found")
	}
}
