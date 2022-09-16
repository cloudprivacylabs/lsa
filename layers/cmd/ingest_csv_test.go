package cmd

import (
	"testing"

	"github.com/cloudprivacylabs/lsa/layers/cmd/pipeline"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func TestCSVJoinIngest(t *testing.T) {
	cji := CSVJoinIngester{
		BaseIngestParams: BaseIngestParams{
			Schema: "Patient",
			Bundle: []string{"testdata/ingest-csvjoin.bundle.json"},
		},
		StartRow: 1,
		entities: []CSVJoinConfig{
			{
				VariantID: "Patient",
				StartCol:  0,
				EndCol:    2,
				IDCols:    []int{0, 1, 2},
			},
			{
				VariantID: "Nest",
				StartCol:  3,
				EndCol:    5,
				IDCols:    []int{0, 1, 2, 3, 4, 5},
			},
			{
				VariantID: "West",
				StartCol:  6,
				EndCol:    8,
				IDCols:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8},
			},
		},
	}
	cji2 := CSVJoinIngester{
		BaseIngestParams: BaseIngestParams{
			Schema: "West",
			Bundle: []string{"testdata/ingest-csvjoin.bundle.json"},
		},
		StartRow: 1,
		entities: []CSVJoinConfig{
			{
				VariantID: "Patient",
				StartCol:  0,
				EndCol:    2,
				IDCols:    []int{0, 1, 2},
			},
			{
				VariantID: "Nest",
				StartCol:  3,
				EndCol:    5,
				IDCols:    []int{0, 1, 2, 3, 4, 5},
			},
			{
				VariantID: "West",
				StartCol:  6,
				EndCol:    8,
				IDCols:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8},
			},
		},
	}
	cji3 := CSVJoinIngester{
		BaseIngestParams: BaseIngestParams{
			Schema: "Nest",
			Bundle: []string{"testdata/ingest-csvjoin.bundle.json"},
		},
		StartRow: 1,
		entities: []CSVJoinConfig{
			{
				VariantID: "Patient",
				StartCol:  0,
				EndCol:    2,
				IDCols:    []int{0, 1, 2},
			},
			{
				VariantID: "Nest",
				StartCol:  3,
				EndCol:    5,
				IDCols:    []int{0, 1, 2, 3, 4, 5},
			},
			{
				VariantID: "West",
				StartCol:  6,
				EndCol:    8,
				IDCols:    []int{0, 1, 2, 3, 4, 5, 6, 7, 8},
			},
		},
	}

	p := []pipeline.Step{
		&cji,
		&cji2,
		&cji3,
	}
	pctx := pipeline.NewContext(ls.DefaultContext(), p, nil, pipeline.InputsFromFiles([]string{"testdata/csvjoin.csv"}))
	err := cji.Run(pctx)
	if err != nil {
		t.Error(err)
	}
}
