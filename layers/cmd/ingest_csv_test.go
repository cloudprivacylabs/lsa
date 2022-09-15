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
	}
	err = cji.ingestCSVJoin(entities, f)
	if err != nil {
		t.Error(err)
	}
	t.Fatal()
}
