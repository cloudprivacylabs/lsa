package csv

import (
	"encoding/json"
	"strconv"
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
		{"data1", "data2", "data3", "data4", "data5"},
	}
	expectedNodes_OSA_FlagTrue := [][]string{
		{"data1", "data2", "data3", "data4", "data5", "data6"},
		{"data1", "data2", "data3", "data4", "data5", "data6"},
		{"data1", "data2", "data3", "data4", "data5", "data6"},
		{"data1", "data2", "data3", "data4", "data5"},
		{"data1", "data2", "data3", "data4", "data5"},
	}
	expectedNodes_OSA_FlagFalse := [][]string{
		{"data1", "data2", "data3", "data4", "data5", "data6"},
		{"data1", "data2", "data3", "data4", "data5", "data6", "data7"},
		{"data1", "data2", "data3", "data4", "data5", "data6", "data7", "data8", "data9"},
		{"data1", "data2", "data3", "data4", "data5"},
		{"data1", "data2", "data3", "data4", "data5"},
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

	// Test with OnlySchemaAttributes flag set to false (ingest all nodes)
	ingester.PreserveNodePaths = true
	ingester.OnlySchemaAttributes = false
	target := digraph.New()
	for idx, tt := range inputStrColData {
		node, err := ingester.Ingest(tt, "https://www.example.com/id")
		nodesRow := make([][]string, 0, len(inputStrColData))
		require.NoError(t, err)
		target.AddNode(node)
		ix := target.GetIndex()
		const nodeID = "https://www.example.com/id"
		for i := 0; i < len(tt); i++ {
			nodes := ix.NodesByLabelSlice(nodeID + "." + strconv.Itoa(idx))
			if len(nodes) == 0 {
				t.Errorf("node not found: %s", nodeID)
			}
			nodesRow = append(nodesRow, expectedNodes_OSA_FlagFalse[nodes[idx].(ls.Node).GetIndex()])
		}
		require.Equalf(t, expectedNodes_OSA_FlagFalse[idx], nodesRow[idx], "inequal data, expected: %s, received: %s", expectedNodes_OSA_FlagFalse[idx], nodesRow[idx])
	}

	// Test with OnlySchemaAttributes flag set
	ingester.OnlySchemaAttributes = true
	target = digraph.New()
	for idx, tt := range inputStrColData {
		node, err := ingester.Ingest(tt, "https://www.example.com/id")
		nodesRow := make([][]string, 0, len(inputStrColData))
		require.NoError(t, err)
		target.AddNode(node)
		ix := target.GetIndex()
		const nodeID = "https://www.example.com/id"
		for i := 0; i < len(tt); i++ {
			nodes := ix.NodesByLabelSlice(nodeID + "." + strconv.Itoa(idx))
			if len(nodes) == 0 {
				t.Errorf("node not found: %s", nodeID)
			}
			nodesRow = append(nodesRow, expectedNodes_OSA_FlagTrue[nodes[idx].(ls.Node).GetIndex()])
		}
		require.Equalf(t, expectedNodes_OSA_FlagTrue[idx], nodesRow[idx], "inequal data, expected: %s, received: %s", expectedNodes_OSA_FlagTrue[idx], nodesRow[idx])
	}
}
