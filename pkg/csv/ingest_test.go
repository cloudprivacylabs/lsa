package csv

import (
	"encoding/json"
	"testing"

	"github.com/bserdar/digraph"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/stretchr/testify/require"
)

func TestIngest(t *testing.T) {
	schStr := `{
		"@context": ["../../schemas/ls.json"],
		"@type": "Schema",
		"@id": "http://example.com/id",
		"layer": {
			"@type": "Object",
			"attributeList": [
				{
					"@id": "https://www.example.com/id1",
					"@type": "Value",
					"attributeName": "field1"
				},
				{
					"@id": "https://www.example.com/id2",
					"@type": "Value",
					"attributeName": "field2"
				},
				{
					"@id": "https://www.example.com/id3",
					"@type": "Value",
					"attributeName": "field3"
				},
				{
					"@id": "https://www.example.com/id4",
					"@type": "Value",
					"attributeName": "field4"
				},
				{
					"@id": "https://www.example.com/id5",
					"@type": "Value",
					"attributeName": "field5"
				},
				{
					"@id": "https://www.example.com/id6",
					"@type": "Value",
					"attributeName": "field6"
				}
			]
		}
	}
	`

	inputStrColData := [][]string{
		{"data1", "data2", "data3", "data4", "data5", "data6"},
		{"data1", "data2", "data3", "data4", "data5", "data6", "data7"},
		{"data1", "data2", "data3", "data4", "data5", "data6", "data7", "data8", "data9"},
		{"data1", "data2", "data3", "data4", "data5"},
		{"data1", "data2", "data3"},
	}

	var schMap interface{}
	if err := json.Unmarshal([]byte(schStr), &schMap); err != nil {
		t.Fatal(err)
	}
	schema, err := ls.UnmarshalLayer(schMap, nil)
	if err != nil {
		t.Error(err)
	}

	ingester := Ingester{
		Ingester: ls.Ingester{
			Schema: schema,
		},
	}

	ingester.PreserveNodePaths = true
	target := digraph.New()
	for _, tt := range inputStrColData {
		node, err := ingester.Ingest(tt, "http://base")
		require.NoError(t, err)
		target.AddNode(node)
		ix := target.GetIndex()
		checkNodeValue := func(nodeId string, expected interface{}) {
			nodes := ix.NodesByLabelSlice(nodeId)

			if len(nodes) == 0 {
				t.Errorf("node not found: %s", nodeId)
			}
			if nodes[0].(ls.Node).GetValue() != expected {
				t.Errorf("Wrong value for %s: %v", nodeId, nodes[0].(ls.Node).GetValue())
			}
		}
		checkNodeValue("http://base.0", "data1")
		checkNodeValue("http://base.1", "data2")
		checkNodeValue("http://base.2", "data3")
		checkNodeValue("http://base.3", "data4")
		checkNodeValue("http://base.4", "data5")
	}
}
