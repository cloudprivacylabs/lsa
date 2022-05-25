package cmd

import (
	"reflect"
	"testing"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// Match looks at req.
func TestMatch(t *testing.T) {
	sheets, err := cmdutil.ReadSpreadsheetFile("testdata/valueset_sample.xlsx")
	if err != nil {
		t.Errorf("While reading file: %s, %v", "testdata/valueset_sample.xlsx", err)
	}
	var vs Valueset
	for sheetName, sheet := range sheets {
		optionRows, headers, data, err := parseSpreadsheet(sheet)
		if err != nil {
			t.Error(err)
		}
		options := parseOptions(optionRows)
		vs, err = parseData(sheetName, headers, data, options)
		if err != nil {
			t.Error(err)
		}
	}
	tests := []ls.ValuesetLookupRequest{
		{TableIDs: []string{"sample"}, KeyValues: map[string]string{"CODE": "A2", "DESCRIPTIVE_TEXT": "B2"}},
		{TableIDs: []string{"sample"}, KeyValues: map[string]string{"CODE": "A4", "DESCRIPTIVE_TEXT": "X"}},
		{TableIDs: []string{"sample"}, KeyValues: map[string]string{"CODE": "A4", "DESCRIPTIVE_TEXT": "Y"}},
		{TableIDs: []string{"sample"}, KeyValues: map[string]string{"CODE": "A4", "DESCRIPTIVE_TEXT": "Z"}},
		{TableIDs: []string{"sample"}, KeyValues: map[string]string{"CODE": "A5", "DESCRIPTIVE_TEXT": "B5"}},
		{TableIDs: []string{"sample"}, KeyValues: map[string]string{"CODE": "A5", "DESCRIPTIVE_TEXT": "B6"}},
		{TableIDs: []string{"sample"}, KeyValues: map[string]string{"CODE": "A5", "DESCRIPTIVE_TEXT": "B7"}},
		{TableIDs: []string{"sample"}, KeyValues: map[string]string{"CODE": "A6", "DESCRIPTIVE_TEXT": "B5"}},
		{TableIDs: []string{"sample"}, KeyValues: map[string]string{"CODE": "A6", "DESCRIPTIVE_TEXT": "B6"}},
		{TableIDs: []string{"sample"}, KeyValues: map[string]string{"CODE": "A6", "DESCRIPTIVE_TEXT": "B7"}},
		{TableIDs: []string{"sample"}, KeyValues: map[string]string{"CODE": "A7", "DESCRIPTIVE_TEXT": "B5"}},
		{TableIDs: []string{"sample"}, KeyValues: map[string]string{"CODE": "A7", "DESCRIPTIVE_TEXT": "B6"}},
		{TableIDs: []string{"sample"}, KeyValues: map[string]string{"CODE": "A7", "DESCRIPTIVE_TEXT": "B7"}},
		//{KeyValues: map[string]string{"x": "some_value", "y": "another_value"}},
	}

	for ix, tt := range tests {
		// 	vsv := ValuesetValue{
		// 		Result:       "some_value",
		// 		ResultValues: tt.KeyValues,
		// 	}
		var vslResp *ls.ValuesetLookupResponse
		var err error
		if ix < len(vs.Options.LookupOrder) {
			vslResp, err = vs.Values[ix].Match(tt, vs.Options.LookupOrder[ix])
		} else {
			vslResp, err = vs.Values[ix].Match(tt, "")
		}
		if err != nil || vslResp == nil {
			t.Errorf("Match failed %v", err)
		}
		// for k, v := range vslResp.KeyValues {
		// 	t.Log(k, v)
		// }
		// t.Log(vslResp.KeyValues)
	}
	// t.Fail()
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
		{"CODE": "A2", "DESCRIPTIVE_TEXT": "B2"},
		{"CODE": "A4", "DESCRIPTIVE_TEXT": "X"},
		{"CODE": "A4", "DESCRIPTIVE_TEXT": "Y"},
		{"CODE": "A4", "DESCRIPTIVE_TEXT": "Z"},
		{"CODE": "A5", "DESCRIPTIVE_TEXT": "B5"},
		{"CODE": "A5", "DESCRIPTIVE_TEXT": "B6"},
		{"CODE": "A5", "DESCRIPTIVE_TEXT": "B7"},
		{"CODE": "A6", "DESCRIPTIVE_TEXT": "B5"},
		{"CODE": "A6", "DESCRIPTIVE_TEXT": "B6"},
		{"CODE": "A6", "DESCRIPTIVE_TEXT": "B7"},
		{"CODE": "A7", "DESCRIPTIVE_TEXT": "B5"},
		{"CODE": "A7", "DESCRIPTIVE_TEXT": "B6"},
		{"CODE": "A7", "DESCRIPTIVE_TEXT": "B7"},
	}

	if _, exists := vs.Sets[sheetName]; !exists {
		t.Errorf("Valueset with sheet name: %s does not exist", sheetName)
	}
	var got = make([]map[string]string, 0)

	for _, vsv := range vs.Sets[sheetName].Values {
		if len(vsv.KeyValues) > 0 {
			got = append(got, vsv.KeyValues)
		} else if len(vsv.ResultValues) > 0 {
			got = append(got, vsv.ResultValues)
		}
	}
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
		{"options.separator", "CODE", ";"},
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
		{opts, "CODE", "A;B;C", []string{"A", "B", "C"}},
	}
	for _, tt := range testdata {
		got := tt.opt.splitCell(tt.header, tt.cell)
		if !reflect.DeepEqual(got, tt.expected) {
			t.Errorf("Got %v, expected %v", got, tt.expected)
		}
	}
}
