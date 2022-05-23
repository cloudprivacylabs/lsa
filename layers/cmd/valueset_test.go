package cmd

import (
	"reflect"
	"sort"
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// Match looks at req.
func TestMatch(t *testing.T) {
	tests := []ls.ValuesetLookupRequest{
		{KeyValues: map[string]string{"": "some_value"}},
		{KeyValues: map[string]string{"k": "some_value"}},
		{KeyValues: map[string]string{"x": "some_value", "y": "another_value"}},
	}

	for _, tt := range tests {
		vsv := ValuesetValue{
			Result:       "some_value",
			ResultValues: tt.KeyValues,
		}
		vslResp, err := vsv.Match(tt)
		if err != nil || vslResp == nil {
			t.Errorf("Match failed %v", err)
		}
		t.Log(vslResp)
	}
}

func TestValuesetSpreadSheet_IO(t *testing.T) {
	const (
		sheetName = "sample"
	)
	vs := &Valuesets{
		Spreadsheets: []string{"testdata/sample.xlsx"},
	}
	err := LoadValuesetFiles(vs, nil)
	if err != nil {
		t.Error(err)
	}
	err = vs.LoadSpreadsheets(ls.DefaultContext())
	if err != nil {
		t.Error(err)
	}
	io_tt := []map[string]string{
		{"Input (i)": "F"},
		{"Input (i)": "Female"},
		{"Input (i)": "Male"},
		{"Concept_id (o)": "300"},
		{"Concept_id (o)": "512"},
		{"Concept_id (o)": "512"},
	}
	if _, exists := vs.Sets[sheetName]; !exists {
		t.Errorf("Valueset with sheet name: %s does not exist", sheetName)
	}
	var expected = make([]map[string]string, 0)
	// prune empty maps
	for _, vsv := range vs.Sets[sheetName].Values {
		if len(vsv.KeyValues) > 0 {
			for k, val := range vsv.KeyValues {
				expected = append(expected, map[string]string{k: val})
			}
		} else if len(vsv.ResultValues) > 0 {
			for k, val := range vsv.ResultValues {
				expected = append(expected, map[string]string{k: val})
			}
		}
	}
	// for idx, vsv := range input_tt {
	// 	for key, val := range vsv.KeyValues {
	// 		if val != expected[idx].KeyValues[key] {
	// 			t.Errorf("Got %v, expected: %v", val, expected[idx].KeyValues[key])
	// 		}
	// 	}
	// 	for key, val := range vsv.ResultValues {
	// 		if val != expected[idx].ResultValues[key] {
	// 			t.Errorf("Got %v, expected: %v", val, expected[idx].ResultValues[key])
	// 		}
	// 	}
	// }
	sort.Slice(expected, func(i, j int) bool {
		return expected[i]["Input (i)"] < expected[j]["Input (i)"]
	})
	sort.Slice(expected, func(i, j int) bool {
		return expected[i]["Concept_id (o)"] < expected[j]["Concept_id (o)"]
	})
	if !reflect.DeepEqual(io_tt, expected) {
		t.Errorf("Got %v expected %v", io_tt, expected)
	}
}

func TestValuesetSpreadSheet_KVPair(t *testing.T) {
	const (
		sheetName = "sample"
	)
	vs := &Valuesets{
		Spreadsheets: []string{"testdata/kv-pair.xlsx"},
	}
	err := LoadValuesetFiles(vs, nil)
	if err != nil {
		t.Error(err)
	}
	err = vs.LoadSpreadsheets(ls.DefaultContext())
	if err != nil {
		t.Error(err)
	}
	io_tt := []map[string]string{
		{"CODE": "A2:CODE"},
		{"CODE": "A3:CODE"},
		{"CODE": "A4:CODE"},
		{"DESCRIPTIVE_TEXT": "B2:DESCRIPTIVE_TEXT"},
		{"DESCRIPTIVE_TEXT": "B3:DESCRIPTIVE_TEXT"},
		{"DESCRIPTIVE_TEXT": "B4:DESCRIPTIVE_TEXT"},
	}
	if _, exists := vs.Sets[sheetName]; !exists {
		t.Errorf("Valueset with sheet name: %s does not exist", sheetName)
	}
	var expected = make([]map[string]string, 0)
	// prune empty maps
	for _, vsv := range vs.Sets[sheetName].Values {
		if len(vsv.KeyValues) > 0 {
			for k, val := range vsv.KeyValues {
				expected = append(expected, map[string]string{k: val})
			}
		} else if len(vsv.ResultValues) > 0 {
			for k, val := range vsv.ResultValues {
				expected = append(expected, map[string]string{k: val})
			}
		}
	}
	// for idx, vsv := range input_tt {
	// 	for key, val := range vsv.KeyValues {
	// 		if val != expected[idx].KeyValues[key] {
	// 			t.Errorf("Got %v, expected: %v", val, expected[idx].KeyValues[key])
	// 		}
	// 	}
	// 	for key, val := range vsv.ResultValues {
	// 		if val != expected[idx].ResultValues[key] {
	// 			t.Errorf("Got %v, expected: %v", val, expected[idx].ResultValues[key])
	// 		}
	// 	}
	// }
	sort.Slice(expected, func(i, j int) bool {
		return expected[i]["CODE"] < expected[j]["CODE"]
	})
	sort.Slice(expected, func(i, j int) bool {
		return expected[i]["DESCRIPTIVE_TEXT"] < expected[j]["DESCRIPTIVE_TEXT"]
	})
	if !reflect.DeepEqual(io_tt, expected) {
		t.Errorf("Got %v expected %v", io_tt, expected)
	}
}

func TestValuesetSpreadSheet(t *testing.T) {
	const (
		sheetName = "sample"
	)
	vs := &Valuesets{
		Spreadsheets: []string{"testdata/kv-pair.xlsx"},
	}
	err := LoadValuesetFiles(vs, nil)
	if err != nil {
		t.Error(err)
	}
	err = vs.LoadSpreadsheets(ls.DefaultContext())
	if err != nil {
		t.Error(err)
	}
	io_tt := []map[string]string{
		{"CODE": "A2:CODE"},
		{"CODE": "A3:CODE"},
		{"CODE": "A4:CODE"},
		{"DESCRIPTIVE_TEXT": "B2:DESCRIPTIVE_TEXT"},
		{"DESCRIPTIVE_TEXT": "B3:DESCRIPTIVE_TEXT"},
		{"DESCRIPTIVE_TEXT": "B4:DESCRIPTIVE_TEXT"},
	}
	if _, exists := vs.Sets[sheetName]; !exists {
		t.Errorf("Valueset with sheet name: %s does not exist", sheetName)
	}
	var expected = make([]map[string]string, 0)
	// prune empty maps
	for _, vsv := range vs.Sets[sheetName].Values {
		if len(vsv.KeyValues) > 0 {
			for k, val := range vsv.KeyValues {
				expected = append(expected, map[string]string{k: val})
			}
		} else if len(vsv.ResultValues) > 0 {
			for k, val := range vsv.ResultValues {
				expected = append(expected, map[string]string{k: val})
			}
		}
	}
	// for idx, vsv := range input_tt {
	// 	for key, val := range vsv.KeyValues {
	// 		if val != expected[idx].KeyValues[key] {
	// 			t.Errorf("Got %v, expected: %v", val, expected[idx].KeyValues[key])
	// 		}
	// 	}
	// 	for key, val := range vsv.ResultValues {
	// 		if val != expected[idx].ResultValues[key] {
	// 			t.Errorf("Got %v, expected: %v", val, expected[idx].ResultValues[key])
	// 		}
	// 	}
	// }
	sort.Slice(expected, func(i, j int) bool {
		return expected[i]["CODE"] < expected[j]["CODE"]
	})
	sort.Slice(expected, func(i, j int) bool {
		return expected[i]["DESCRIPTIVE_TEXT"] < expected[j]["DESCRIPTIVE_TEXT"]
	})
	if !reflect.DeepEqual(io_tt, expected) {
		t.Errorf("Got %v expected %v", io_tt, expected)
	}
}
