package cmd

import (
	"testing"

	"github.com/cloudprivacylabs/lsa/layers/cmd/pipeline"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func TestCSVJoinIngest(t *testing.T) {
	cji := CSVJoinIngester{
		BaseIngestParams: BaseIngestParams{
			Bundle: []string{"testdata/ingest-csvjoin.bundle.json"},
		},
		HeaderRow: 0,
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
	}
	pctx := pipeline.NewContext(ls.DefaultContext(), p, nil, pipeline.InputsFromFiles([]string{"testdata/ingest_csvjoin_pipeline.json"}))

	// pctx, err := runPipeline(p, "", []string{"testdata/csvjoin.csv"})
	err := cji.Run(pctx)
	if err != nil {
		t.Error(err)
	}
	t.Fatal()
}
