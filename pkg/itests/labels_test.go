package itests

import (
	"encoding/json"
	"testing"

	"github.com/cloudprivacylabs/lpg"
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
					"ls:labeledAs": [
						"thirdLabel",
						"fourthLabel"
					],
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

	var seenL1, seenL2, seenL3 = false, false, false
	layer.ForEachAttribute(func(n1 *lpg.Node, n2 []*lpg.Node) bool {
		if n1.HasLabel("SOMELABEL") {
			seenL1 = true
		}
		if n1.HasLabel("ANOTHERLABEL") {
			seenL2 = true
		}
		if n1.GetLabels().HasAll("thirdLabel", "fourthLabel") {
			seenL3 = true
		}
		return true
	})
	if !seenL1 {
		t.Errorf("SOMELABEL cannot be found")
	}
	if !seenL2 {
		t.Errorf("ANOTHERLABEL cannot be found")
	}
	if !seenL3 {
		t.Errorf("thirdLabel, fourthLabel cannot be found")
	}
}
