package csv

import (
	"encoding/json"
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/jsonld"
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
	schema, err := jsonld.UnmarshalLayer(schMap, nil)
	if err != nil {
		t.Error(err)
	}

	parser := Parser{
		SchemaNode:  schema.GetSchemaRootNode(),
		ColumnNames: []string{"field1", "field2", "field3", "field4", "field5", "field6"},
	}
	builder := ls.NewGraphBuilder(nil, ls.GraphBuilderOptions{
		EmbedSchemaNodes:     true,
		OnlySchemaAttributes: false,
	})

	ing := ls.Ingester{Schema: schema}

	// Test with OnlySchemaAttributes flag set to false (ingest all nodes)
	for idx, tt := range inputStrColData {
		parsed, err := parser.ParseDoc(ls.DefaultContext(), "https://www.example.com/id", tt)
		if err != nil {
			t.Error(err)
		}
		_, err = ing.Ingest(builder, parsed)
		if err != nil {
			t.Error(err)
		}
		for _, expected := range expectedNodes_OSA_FlagFalse[idx] {
			found := false
			for nx := builder.GetGraph().GetNodes(); nx.Next(); {
				node := nx.Node()
				v, _ := ls.GetRawNodeValue(node)
				if v == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("node not found: %s - %d", expected, idx)
			}
		}
	}

	// Test with OnlySchemaAttributes flag set
	builder = ls.NewGraphBuilder(nil, ls.GraphBuilderOptions{
		EmbedSchemaNodes:     true,
		OnlySchemaAttributes: true,
	})
	for idx, tt := range inputStrColData {
		parsed, err := parser.ParseDoc(ls.DefaultContext(), "https://www.example.com/id", tt)
		if err != nil {
			t.Error(err)
		}
		ing := ls.Ingester{Schema: schema}
		_, err = ing.Ingest(builder, parsed)
		if err != nil {
			t.Error(err)
		}
		require.NoError(t, err)
		for _, expected := range expectedNodes_OSA_FlagTrue[idx] {
			found := false
			for nx := builder.GetGraph().GetNodes(); nx.Next(); {
				node := nx.Node()
				v, _ := ls.GetRawNodeValue(node)
				if v == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("node not found: %s - %d", expected, idx)
			}
		}
	}
}
