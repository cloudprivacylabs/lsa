package ls

import (
	"testing"

	"github.com/cloudprivacylabs/lpg/v2"
)

func TestProcessLabeledAs(t *testing.T) {
	schStr := `
{
  "nodes": [
    {
      "n": 0,
      "labels": [
        "https://lschema.org/Schema"
      ],
      "properties": {
        "https://lschema.org/nodeId": "http://testschema"
      },
      "edges": [
        {
          "to": 1,
          "label": "https://lschema.org/layer"
        }
      ]
    },
    {
      "n": 1,
      "labels": [
        "https://lschema.org/valueType",
        "https://lschema.org/Attribute",
        "https://lschema.org/Object"
      ],
      "properties": {
        "https://lschema.org/nodeId": "root"
      },
      "edges": [
        {
          "to": 2,
          "label": "https://lschema.org/Object/attributes"
        },
        {
          "to": 3,
          "label": "https://lschema.org/Object/attributes"
        },
        {
          "to": 4,
          "label": "https://lschema.org/Object/attributes"
        }
      ]
    },
    {
      "n": 2,
      "labels": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "properties": {
        "https://lschema.org/attributeIndex": 0,
        "https://lschema.org/labeledAs": [
          "SOMELABEL"
        ],
        "https://lschema.org/nodeId": "attr1"
      }
    },
    {
      "n": 3,
      "labels": [
        "https://lschema.org/Attribute",
        "https://lschema.org/Value"
      ],
      "properties": {
        "https://lschema.org/attributeIndex": 1,
        "https://lschema.org/labeledAs": [
          "ANOTHERLABEL"
        ],
        "https://lschema.org/nodeId": "attr2",
        "https://lschema.org/privacy": "flg1"
      }
    },
    {
      "n": 4,
      "labels": [
        "https://lschema.org/Value",
        "https://lschema.org/Attribute"
      ],
      "properties": {
        "https://lschema.org/attributeIndex": 2,
        "https://lschema.org/labeledAs": [
          "thirdLabel",
          "fourthLabel"
        ],
        "https://lschema.org/nodeId": "attr3",
        "https://lschema.org/privacy": [
          "flg2",
          "flg3"
        ]
      }
    }
  ]
}
	`

	schema, err := UnmarshalLayerFromSlice([]byte(schStr))
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
