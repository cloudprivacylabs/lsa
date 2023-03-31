package ls

import (
	"encoding/json"
	"testing"

	"github.com/cloudprivacylabs/lpg/v2"
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

	schema, err := UnmarshalLayer(schMap, nil)
	if err != nil {
		t.Error(err)
	}

	c := Compiler{}
	layer, err := c.CompileSchema(DefaultContext(), schema)
	if err != nil {
		t.Error(err)
	}

	attr1 := layer.GetAttributeByID("attr1")
	t.Logf("%v", attr1)

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
