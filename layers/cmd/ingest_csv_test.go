package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/cloudprivacylabs/lpg"
	"github.com/cloudprivacylabs/lsa/layers/cmd/pipeline"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func TestCSVJoinIngest(t *testing.T) {
	cji := CSVJoinIngester{
		BaseIngestParams: BaseIngestParams{
			Bundle:           []string{"testdata/ingest-csvjoin.bundle.json"},
			EmbedSchemaNodes: true,
		},
		StartRow: 1,
		EndRow:   -1,
		entities: []CSVJoinConfig{
			{
				VariantID: "Patient",
				StartCol:  0,
				EndCol:    2,
				IDCols:    []int{0, 1, 2},
			},
			{
				VariantID: "Foo",
				StartCol:  3,
				EndCol:    5,
				IDCols:    []int{0, 1, 2, 3, 4, 5},
			},
			{
				VariantID: "Bar",
				StartCol:  6,
				EndCol:    8,
				IDCols:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8},
			},
		},
	}
	cji.ingester = make(map[string]*ls.Ingester)
	p := []pipeline.Step{
		&cji,
	}
	pctx := pipeline.NewContext(ls.DefaultContext(), p, nil, pipeline.InputsFromFiles([]string{"testdata/csvjoin.csv"}))
	err := cji.Run(pctx)

	m := ls.JSONMarshaler{}
	f, err := os.Open("testdata/csvingest-test.json")
	buf := bytes.Buffer{}
	buf.ReadFrom(f)
	// x.Encode(pctx.Graph, f)
	expectedGraph := lpg.NewGraph()
	err = m.Unmarshal(buf.Bytes(), expectedGraph)
	if err != nil {
		t.Error(err)
	}
	eq := lpg.CheckIsomorphism(pctx.Graph, expectedGraph, func(n1, n2 *lpg.Node) bool {
		// t.Logf("Cmp: %+v %+v\n", n1, n2)
		if !n1.GetLabels().IsEqual(n2.GetLabels()) {
			return false
		}
		s1, _ := ls.GetRawNodeValue(n1)
		s2, _ := ls.GetRawNodeValue(n2)
		if s1 != s2 {
			return false
		}
		// Expected properties must be a subset
		propertiesOK := true
		n2.ForEachProperty(func(k string, v interface{}) bool {
			pv, ok := v.(*ls.PropertyValue)
			if !ok {
				return true
			}
			v2, ok := n1.GetProperty(k)
			if !ok {
				propertiesOK = false
				return false
			}
			pv2, ok := v2.(*ls.PropertyValue)
			if !ok {
				propertiesOK = false
				return false
			}
			if !pv2.IsEqual(pv) {
				propertiesOK = false
				return false
			}
			return true
		})
		if !propertiesOK {
			return false
		}
		t.Logf("True\n")
		return true
	}, func(e1, e2 *lpg.Edge) bool {
		return e1.GetLabel() == e2.GetLabel() && ls.IsPropertiesEqual(ls.PropertiesAsMap(e1), ls.PropertiesAsMap(e2))
	})

	if !eq {
		result, _ := m.Marshal(pctx.Graph)
		expected, _ := m.Marshal(expectedGraph)
		t.Errorf("Result is different from the expected: Result:\n%s\nExpected:\n%s", string(result), string(expected))
	}

	if err != nil {
		t.Error(err)
	}
}

func TestCSVJoinLinkedIngest(t *testing.T) {
	cji := CSVJoinIngester{
		BaseIngestParams: BaseIngestParams{
			Bundle:           []string{"testdata/ingest-csvjoin-linked.bundle.json"},
			EmbedSchemaNodes: true,
		},
		StartRow: 1,
		EndRow:   -1,
		entities: []CSVJoinConfig{
			{
				VariantID: "Patient",
				StartCol:  0,
				EndCol:    2,
				IDCols:    []int{0, 1, 2},
			},
			{
				VariantID: "Foo",
				StartCol:  3,
				EndCol:    5,
				IDCols:    []int{0, 1, 2, 3, 4, 5},
			},
			{
				VariantID: "Bar",
				StartCol:  6,
				EndCol:    8,
				IDCols:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8},
			},
		},
	}
	cji.ingester = make(map[string]*ls.Ingester)
	p := []pipeline.Step{
		&cji,
	}
	pctx := pipeline.NewContext(ls.DefaultContext(), p, nil, pipeline.InputsFromFiles([]string{"testdata/csvjoin.csv"}))
	err := cji.Run(pctx)

	m := ls.JSONMarshaler{}
	f, err := os.Open("testdata/csvjoiningest-linked-test.json")
	buf := bytes.Buffer{}
	buf.ReadFrom(f)
	// x.Encode(pctx.Graph, f)
	expectedGraph := lpg.NewGraph()
	err = m.Unmarshal(buf.Bytes(), expectedGraph)
	if err != nil {
		t.Error(err)
	}
	eq := lpg.CheckIsomorphism(pctx.Graph, expectedGraph, func(n1, n2 *lpg.Node) bool {
		// t.Logf("Cmp: %+v %+v\n", n1, n2)
		if !n1.GetLabels().IsEqual(n2.GetLabels()) {
			return false
		}
		s1, _ := ls.GetRawNodeValue(n1)
		s2, _ := ls.GetRawNodeValue(n2)
		if s1 != s2 {
			return false
		}
		// Expected properties must be a subset
		propertiesOK := true
		n2.ForEachProperty(func(k string, v interface{}) bool {
			pv, ok := v.(*ls.PropertyValue)
			if !ok {
				return true
			}
			v2, ok := n1.GetProperty(k)
			if !ok {
				propertiesOK = false
				return false
			}
			pv2, ok := v2.(*ls.PropertyValue)
			if !ok {
				propertiesOK = false
				return false
			}
			if !pv2.IsEqual(pv) {
				propertiesOK = false
				return false
			}
			return true
		})
		if !propertiesOK {
			return false
		}
		t.Logf("True\n")
		return true
	}, func(e1, e2 *lpg.Edge) bool {
		return e1.GetLabel() == e2.GetLabel() && ls.IsPropertiesEqual(ls.PropertiesAsMap(e1), ls.PropertiesAsMap(e2))
	})

	if !eq {
		result, _ := m.Marshal(pctx.Graph)
		expected, _ := m.Marshal(expectedGraph)
		t.Errorf("Result is different from the expected: Result:\n%s\nExpected:\n%s", string(result), string(expected))
	}

	if err != nil {
		t.Error(err)
	}
}
