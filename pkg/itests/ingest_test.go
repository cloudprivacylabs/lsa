package itests

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"text/template"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/bundle"
	csvingest "github.com/cloudprivacylabs/lsa/pkg/csv"
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

func loadJSONLDSchema(fname string) (*ls.Layer, error) {
	data, err := os.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	var v interface{}
	if err := json.Unmarshal([]byte(data), &v); err != nil {
		return nil, err
	}

	return ls.UnmarshalLayer(v, nil)
}

func (tc ingestTest) testDefaultValues(t *testing.T) {
	schema, err := loadJSONLDSchema("testdata/defaultcases.json")
	if err != nil {
		t.Fatal(err)
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

	data, err := os.ReadFile("testdata/defaults.json")
	if err != nil {
		t.Fatal(err)
	}
	// Ingest
	_, err = jsoningest.IngestBytes(ls.DefaultContext(), "http://example.org/id", data, parser, builder)

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

func TestParseEmptyCSV(t *testing.T) {
	schema, err := loadJSONLDSchema("testdata/csv/dmhmreport.schema.json")
	if err != nil {
		t.Fatal(err)
		return
	}
	parser := csvingest.Parser{
		OnlySchemaAttributes: true,
		SchemaNode:           schema.GetSchemaRootNode(),
		IngestNullValues:     false,
	}
	idTemplate := "row_{{.rowIndex}}"
	idTmp, err := template.New("id").Parse(idTemplate)
	if err != nil {
		t.Fatal(err)
		return
	}
	input, err := os.Open("testdata/csv/dmhm.csv")
	if err != nil {
		t.Fatal(err)
		return
	}
	reader := csv.NewReader(input)
	done := false
	ctx := ls.DefaultContext()
	for row := 0; !done; row++ {
		rowData, err := reader.Read()
		if err == io.EOF {
			done = true
		} else if err != nil {
			t.Error(err)
			return
		}
		if row == 0 {
			parser.ColumnNames = rowData
			continue
		}

		templateData := map[string]interface{}{
			"rowIndex":  row,
			"dataIndex": row - 1,
			"columns":   rowData,
		}
		buf := bytes.Buffer{}
		if err := idTmp.Execute(&buf, templateData); err != nil {
			t.Error(err)
			return
		}
		parsed, err := parser.ParseDoc(ctx, strings.TrimSpace(buf.String()), rowData)
		if err != nil {
			t.Error(err)
		}
		if row == 1 && parsed == nil {
			t.Errorf("Nil parsed")
		}
		if row > 1 && parsed != nil {
			t.Errorf("Nil expected")
		}

	}
}

// layers ingest json --bundle testdata/fhir/schemas/fhir.discriminator.bundle.yaml --type https://hl7.org/fhir/Patient testdata/fhir/schemas/Aaron697_Brekke496_2fa15bc7-8866-461a-9000-f739e425860a.json --output web > abc.json
// layers ingest json --bundle testdata/fhir/schemas/fhir.bundle.yaml --type https://hl7.org/fhir/Patient testdata/fhir/schemas/Aaron697_Brekke496_2fa15bc7-8866-461a-9000-f739e425860a.json --output web > abc.json
func TestIngestPolyHint(t *testing.T) {
	b, err := bundle.LoadBundle("testdata/fhir/schemas/fhir.bundle.yaml", func(parentBundle string, loadBundle string) (bundle.Bundle, error) {
		var bnd bundle.Bundle
		if err := cmdutil.ReadJSONOrYAML(loadBundle, &bnd); err != nil {
			return bnd, err
		}
		return bnd, nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := b.Build(ls.DefaultContext(), func(ctx *ls.Context, fname string) ([][][]string, error) {
		return cmdutil.ReadSheets(fname)
	}, func(ctx *ls.Context, fname string) (io.ReadCloser, error) {
		return os.Open(fname)
	}, func(ctx *ls.Context, fname string) (*ls.Layer, error) {
		data, err := os.ReadFile(fname)
		if err != nil {
			return nil, err
		}
		var v interface{}
		err = json.Unmarshal(data, &v)
		if err != nil {
			return nil, err
		}
		return ls.UnmarshalLayer(v, ctx.GetInterner())
	}); err != nil {
		t.Fatal(err)
	}

	layer, err := b.GetLayer(ls.DefaultContext(), "https://hl7.org/fhir/Patient")
	if err != nil {
		t.Fatal(err)
	}

	// in, err := ls.MarshalLayer(layer)
	// fx, err := os.Create("test.json")
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// defer fx.Close()
	// by, _ := json.Marshal(in)
	// fx.Write(by)

	parser := jsoningest.Parser{Layer: layer}

	builder := ls.NewGraphBuilder(nil, ls.GraphBuilderOptions{
		EmbedSchemaNodes: true,
	})

	f, err := os.Open("testdata/fhir/schemas/Aaron697_Brekke496_2fa15bc7-8866-461a-9000-f739e425860a.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	_, err = jsoningest.IngestStream(ls.DefaultContext(), b.Base, f, parser, builder)
	if err != nil {
		t.Fatal(err)
	}

	findPoly := func() bool {
		for nx := layer.Graph.GetNodes(); nx.Next(); {
			node := nx.Node()
			if node.HasLabel(ls.TypeDiscriminatorTerm) {
				return true
			}
		}
		return false
	}

	if !findPoly() {
		t.Fatalf("Expecting type hint")
	}
}
