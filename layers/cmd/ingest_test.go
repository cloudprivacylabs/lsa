package cmd

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	jsoningest "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type ingestTest struct {
	Labels []string `json:"labels"`
	Path   []string `json:"path"`
}

func TestIngest(t *testing.T) {
	var defaultsCase = ingestTest{
		Path: nil,
	}
	defaultsCase.testDefaultValues(t)
}

func (tc ingestTest) testDefaultValues(t *testing.T) {
	var schMap interface{}
	schStr, err := ioutil.ReadFile("testdata/defaultcases.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(schStr), &schMap); err != nil {
		t.Fatal(err)
	}

	schema, err := ls.UnmarshalLayer(schMap, nil)
	if err != nil {
		t.Error(err)
	}

	builder := ls.NewGraphBuilder(nil, ls.GraphBuilderOptions{
		OnlySchemaAttributes: false,
		EmbedSchemaNodes:     true,
	})

	parser := jsoningest.Parser{
		Layer:                schema,
		OnlySchemaAttributes: false,
		IngestNullValues:     true,
	}

	// Ingest
	_, err = jsoningest.IngestBytes(ls.DefaultContext(), "http://example.org/id", schStr, parser, builder)

	for nodeItr := builder.GetGraph().GetNodes(); nodeItr.Next(); {
		node := nodeItr.Node()
		// Check if default values have been added
		if x, exists := node.GetProperty(ls.NodeValueTerm); exists {
			val := ls.AsPropertyValue(x, true).AsString()
			if val == "valDefault" || val == "objDefault" || val == "xyz" {
				t.Logf("Default value has been placed: %v", val)
			} else {
				continue
			}
		}
	}
}
