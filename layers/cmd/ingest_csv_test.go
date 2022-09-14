package cmd

import (
	"os"
	"testing"
)

func TestCSVJoinIngest(t *testing.T) {
	cji := CSVJoinIngester{HeaderRow: 0}
	f, err := os.Open("testdata/csvjoin.csv")
	if err != nil {
		t.Error(err)
	}
	defer f.Close()
	entities := []CSVJoinConfig{
		{
			VariantID: "Patient",
			StartCol:  0,
			EndCol:    2,
			IDCols:    make([]int, 0),
		},
		{
			VariantID: "Nest",
			StartCol:  3,
			EndCol:    5,
			IDCols:    make([]int, 0),
		},
		{
			VariantID: "West",
			StartCol:  6,
			EndCol:    8,
			IDCols:    make([]int, 0),
		},
	}
	err = cji.ingestCSVJoin(entities, f)
	if err != nil {
		t.Error(err)
	}
	t.Fatal()
}
