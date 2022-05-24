package cmd

import (
	"reflect"
	"sort"
	"testing"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
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

func TestValuesetSpreadSheet(t *testing.T) {
	const (
		sheetName = "sample"
	)
	vs := &Valuesets{
		Spreadsheets: []string{"testdata/valueset_sample.xlsx"},
	}
	err := LoadValuesetFiles(vs, nil)
	if err != nil {
		t.Error(err)
	}
	err = vs.LoadSpreadsheets(ls.DefaultContext())
	if err != nil {
		t.Error(err)
	}
	expected := []map[string]string{
		{"CODE": "A2:CODE"},
		{"CODE": "A3:CODE"},
		{"CODE": "A4"},
		{"CODE": "A5"},
		{"CODE": "A5"},
		{"CODE": "A5"},
		{"CODE": "A5"},
		{"CODE": "A6"},
		{"CODE": "A6"},
		{"CODE": "A6"},
		{"CODE": "A6"},
		{"CODE": "A7"},
		{"CODE": "A7"},
		{"CODE": "A7"},
		{"CODE": "A7"},
		{"DESCRIPTIVE_TEXT": "B2:DESCRIPTIVE_TEXT"},
		{"DESCRIPTIVE_TEXT": "B3:DESCRIPTIVE_TEXT"},
		{"DESCRIPTIVE_TEXT": "X"},
		{"DESCRIPTIVE_TEXT": "Y"},
		{"DESCRIPTIVE_TEXT": "Z"},
		{"DESCRIPTIVE_TEXT": "B5"},
		{"DESCRIPTIVE_TEXT": "B5"},
		{"DESCRIPTIVE_TEXT": "B5"},
		{"DESCRIPTIVE_TEXT": "B6"},
		{"DESCRIPTIVE_TEXT": "B6"},
		{"DESCRIPTIVE_TEXT": "B6"},
		{"DESCRIPTIVE_TEXT": "B7"},
		{"DESCRIPTIVE_TEXT": "B7"},
		{"DESCRIPTIVE_TEXT": "B7"},
	}
	if _, exists := vs.Sets[sheetName]; !exists {
		t.Errorf("Valueset with sheet name: %s does not exist", sheetName)
	}
	var got = make([]map[string]string, 0)
	// prune empty maps
	for _, vsv := range vs.Sets[sheetName].Values {
		if len(vsv.KeyValues) > 0 {
			for k, val := range vsv.KeyValues {
				got = append(got, map[string]string{k: val})
			}
		} else if len(vsv.ResultValues) > 0 {
			for k, val := range vsv.ResultValues {
				got = append(got, map[string]string{k: val})
			}
		}
	}
	sort.SliceStable(got, func(i, j int) bool {
		return got[i]["CODE"] < got[j]["CODE"]
	})
	sort.SliceStable(got, func(i, j int) bool {
		return got[i]["DESCRIPTIVE_TEXT"] < got[j]["DESCRIPTIVE_TEXT"]
	})
	sort.SliceStable(expected, func(i, j int) bool {
		return expected[i]["CODE"] < expected[j]["CODE"]
	})
	sort.SliceStable(expected, func(i, j int) bool {
		return expected[i]["DESCRIPTIVE_TEXT"] < expected[j]["DESCRIPTIVE_TEXT"]
	})
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Got %v expected %v", got, expected)
	}
}

func TestParseSpreadsheet(t *testing.T) {
	sheets, err := cmdutil.ReadSpreadsheetFile("testdata/valueset_sample.xlsx")
	if err != nil {
		t.Errorf("While reading file: %s, %v", "testdata/valueset_sample.xlsx", err)
	}
	testOptions := []struct {
		opt []string
	}{
		{[]string{"options.lookupOrder", "CODE", "DESCRIPTIVE_TEXT"}},
		{[]string{"options.separator", "DESCRIPTIVE_TEXT", ";"}},
	}
	testHeaders := []struct {
		hdr []string
	}{
		{[]string{"CODE", "DESCRIPTIVE_TEXT"}},
	}
	testData := []struct {
		data []string
	}{
		{[]string{"A2:CODE", "B2:DESCRIPTIVE_TEXT"}},
		{[]string{"A3:CODE", "B3:DESCRIPTIVE_TEXT"}},
		{[]string{"A4", "X; Y; Z"}},
	}
	for _, sheet := range sheets {
		optionRows, headers, data, err := parseSpreadsheet(sheet)
		if err != nil {
			t.Error(err)
		}
		for idx, tt := range testOptions {
			if !reflect.DeepEqual(optionRows[idx], tt.opt) {
				t.Errorf("Got %v, expected %v", optionRows[idx], tt.opt)
			}
		}
		for idx, tt := range testHeaders {
			if tt.hdr[idx] != headers[idx] {
				t.Errorf("Got %s, expected %s", tt.hdr[idx], headers[idx])
			}
		}
		for idx, tt := range testData {
			if !reflect.DeepEqual(tt.data, data[idx]) {
				t.Errorf("Got %v, expected %v", tt.data, data[idx])
			}
		}
	}
}

func TestSplitCell(t *testing.T) {
	optRows := [][]string{
		{"CODE", "ABC"},
	}
	opts := parseOptions(optRows)
	testdata := []struct {
		opt      Options
		header   string
		cell     string
		expected []string
	}{
		{Options{Separator: map[string]string{"DESCRIPTIVE_TEXT": ";"}}, "DESCRIPTIVE_TEXT", "a; b; c; d", []string{"a", "b", "c", "d"}},
		{Options{Separator: map[string]string{"DESCRIPTIVE_TEXT": ","}}, "DESCRIPTIVE_TEXT", "a, b, c, d", []string{"a", "b", "c", "d"}},
		{Options{Separator: map[string]string{"DESCRIPTIVE_TEXT": "|"}}, "DESCRIPTIVE_TEXT", "a | b | c | d", []string{"a", "b", "c", "d"}},
		{Options{Separator: map[string]string{"DESCRIPTIVE_TEXT": ""}}, "DESCRIPTIVE_TEXT", "a b c d", []string{"a", "b", "c", "d"}},
		{Options{Separator: map[string]string{"DESCRIPTIVE_TEXT": " "}}, "DESCRIPTIVE_TEXT", "a b c d", []string{"a", "b", "c", "d"}},
		{opts, "CODE", "ABC", []string{"ABC"}},
	}
	for _, tt := range testdata {
		got := tt.opt.splitCell(tt.header, tt.cell)
		if !reflect.DeepEqual(got, tt.expected) {
			t.Errorf("Got %v, expected %v", got, tt.expected)
		}
	}
}
