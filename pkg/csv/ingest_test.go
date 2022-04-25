package csv

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
	"github.com/stretchr/testify/require"
)

func TestIngest(t *testing.T) {
	schStr := `{
		"@context": ["../../schemas/ls.json"],
		"@type": "Schema",
		"@id": "http://example.com/id",
		"layer": {
			"@type": "Object",
      "@id": "root",
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

	parser := Parser{
		SchemaNode:  schema.GetSchemaRootNode(),
		ColumnNames: []string{"field1", "field2", "field3", "field4", "field5", "field6"},
	}
	builder := ls.NewGraphBuilder(nil, ls.GraphBuilderOptions{
		EmbedSchemaNodes:     false,
		OnlySchemaAttributes: false,
	})

	// Test with OnlySchemaAttributes flag set to false (ingest all nodes)
	for idx, tt := range inputStrColData {
		parsed, err := parser.ParseDoc(ls.DefaultContext(), "https://www.example.com/id", tt)
		if err != nil {
			t.Error(err)
		}
		_, err = ls.Ingest(builder, parsed)
		if err != nil {
			t.Error(err)
		}
		nodesRow := make([][]string, 0, len(inputStrColData))
		const nodeID = "https://www.example.com/id.field"
		for i := 0; i < len(tt); i++ {
			nodes := make([]graph.Node, 0)
			for nx := builder.GetGraph().GetNodes(); nx.Next(); {
				node := nx.Node()
				t.Logf("NodeID: %s", ls.GetNodeID(node))
				if ls.GetNodeID(node) == (nodeID + strconv.Itoa(idx+1)) {
					nodes = append(nodes, node)
				}
			}
			if len(nodes) == 0 {
				t.Errorf("node not found: %s", nodeID+strconv.Itoa(idx+1))
			}
			nodesRow = append(nodesRow, expectedNodes_OSA_FlagFalse[ls.GetNodeIndex(nodes[idx])])
		}
		require.Equalf(t, expectedNodes_OSA_FlagFalse[idx], nodesRow[idx], "inequal data, expected: %s, received: %s", expectedNodes_OSA_FlagFalse[idx], nodesRow[idx])
	}

	// Test with OnlySchemaAttributes flag set
	builder = ls.NewGraphBuilder(nil, ls.GraphBuilderOptions{
		EmbedSchemaNodes:     false,
		OnlySchemaAttributes: true,
	})
	for idx, tt := range inputStrColData {
		parsed, err := parser.ParseDoc(ls.DefaultContext(), "https://www.example.com/id", tt)
		if err != nil {
			t.Error(err)
		}
		_, err = ls.Ingest(builder, parsed)
		if err != nil {
			t.Error(err)
		}
		nodesRow := make([][]string, 0, len(inputStrColData))
		require.NoError(t, err)
		const nodeID = "https://www.example.com/id.field"
		for i := 0; i < len(tt); i++ {
			nodes := make([]graph.Node, 0)
			for nx := builder.GetGraph().GetNodes(); nx.Next(); {
				node := nx.Node()
				if ls.GetNodeID(node) == (nodeID + strconv.Itoa(idx+1)) {
					nodes = append(nodes, node)
				}
			}
			if len(nodes) == 0 {
				t.Errorf("node not found: %s", nodeID+strconv.Itoa(idx+1))
			}
			nodesRow = append(nodesRow, expectedNodes_OSA_FlagTrue[ls.GetNodeIndex(nodes[idx])])
		}
		require.Equalf(t, expectedNodes_OSA_FlagTrue[idx], nodesRow[idx], "inequal data, expected: %s, received: %s", expectedNodes_OSA_FlagTrue[idx], nodesRow[idx])
	}
}
