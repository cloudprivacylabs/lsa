package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/layers/cmd/valueset"
	vs "github.com/cloudprivacylabs/lsa/layers/cmd/valueset"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/mitchellh/mapstructure"
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
		vslResp, err := vs.Values[ix].Match(tt)
		if err != nil || vslResp == nil {
			t.Errorf("Match failed %v", err)
		}
	}
}

func TestValuesetSpreadSheet(t *testing.T) {
	const (
		sheetName = "sample"
	)
	vs := &Valuesets{
		Spreadsheets: []string{"testdata/valueset_sample.xlsx"},
		cache:        vs.NoCache{},
	}
	env := make(map[string]string)
	err := LoadValuesetFiles(ls.DefaultContext(), env, vs, vs.cache, nil)
	if err != nil {
		t.Error(err)
	}
	err = vs.LoadSpreadsheets(ls.DefaultContext(), "")
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
		{[]string{"options.separator", "CODE", ";", "DESCRIPTIVE_TEXT", ";"}},
	}
	testHeaders := []struct {
		hdr []string
	}{
		{[]string{"CODE", "DESCRIPTIVE_TEXT"}},
	}
	testData := []struct {
		data []string
	}{
		{[]string{"A2", "B2"}},
		{[]string{"A4", "X; Y; Z"}},
		{[]string{"A5; A6; A7", "B5; B6; B7"}},
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
			if !reflect.DeepEqual(data[idx], tt.data) {
				fmt.Println(tt.data, data[idx])
				t.Errorf("Got %v, expected %v", data[idx], tt.data)
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

func init() {
	vs.RegisterDB("pgx", NewMockPostgresqlDataStore)
}

type mockPgxDataStore struct {
	MockDatabase `mapstructure:",squash"`
}

type MockDatabase struct {
	DB        *sql.DB
	Params    Params         `json:"params" yaml:"params"`
	Valuesets []MockValueset `json:"valuesets" yaml:"valuesets"`
	tableIds  map[string]struct{}
	once      sync.Once
}

type Params struct {
	DatabaseName string `json:"db" yaml:"db"`
	User         string `json:"user" yaml:"user"`
	Pwd          string `json:"pwd" yaml:"pwd"`
	URI          string `json:"uri" yaml:"uri"`
}

type MockValueset struct {
	TableId string      `yaml:"tableId"`
	Queries []mockQuery `yaml:"queries"`
}

type mockQuery struct {
	Query string `yaml:"query"`
}

func (pgx *mockPgxDataStore) ValueSetLookup(ctx context.Context, tableId string, queryParams map[string]string) (map[string]string, error) {
	return nil, nil
}

func (pgx *mockPgxDataStore) GetTableIds() map[string]struct{} {
	return nil
}

func (pgx *mockPgxDataStore) Close() error {
	return nil
}

func NewMockPostgresqlDataStore(value interface{}, env map[string]string) (valueset.ValuesetDB, error) {
	psqlDs := &mockPgxDataStore{}
	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: true,
		Result:           &psqlDs,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		panic(err)
	}
	if err := decoder.Decode(value); err != nil {
		return psqlDs, err
	}
	// for _, v := range value.(map[string]interface{}) {
	// 	if err := mapstructure.Decode(v, &psqlDs); err != nil {
	// 		return psqlDs, err
	// 	}
	// }
	psqlDs.tableIds = make(map[string]struct{})
	for _, vs := range psqlDs.Valuesets {
		psqlDs.tableIds[vs.TableId] = struct{}{}
	}
	psqlDs.Params.URI = env["uri"]
	psqlDs.Params.User = env["users"]
	psqlDs.Params.Pwd = env["pwd"]
	return psqlDs, nil
}

func TestLoadValuesetConfig(t *testing.T) {
	env := map[string]string{
		"uri":      "SOME_URI",
		"pgx_user": "someUser",
		"password": "somePwd",
	}
	f := "testdata/config/valueset-databases.yaml"
	var uc struct {
		Databases []map[string]interface{} `json:"databases" yaml:"databases"`
	}
	if err := cmdutil.ReadJSONOrYAML(f, &uc); err != nil {
		t.Error(err)
	}
	cfg, err := vs.LoadConfig(f, env)
	if err != nil {
		t.Error(err)
	}
	for ix, dt := range uc.Databases {
		db, err := vs.UnmarshalSingleDatabaseConfig(dt, env)
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(db, cfg.ValuesetDBs[ix]) {
			t.Errorf("mismatched databases ValusetDBs, got %v, expected :%v", db, cfg.ValuesetDBs[ix])
		}
	}
}
