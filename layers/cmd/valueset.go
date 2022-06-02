// Copyright 2021 Cloud Privacy Labs, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"unicode"

	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type Valuesets struct {
	Services     map[string]string   `json:"services" yaml:"services"`
	Spreadsheets []string            `json:"spreadsheets" yaml:"spreadsheets"`
	Sets         map[string]Valueset `json:"valuesets" yaml:"valuesets"`
}

type Valueset struct {
	ID      string          `json:"id" yaml:"id"`
	Values  []ValuesetValue `json:"values" yaml:"values"`
	Options Options         `json:"options" yaml:"options"`
}

type ValuesetValue struct {
	// Possible input values
	Values []string `json:"values" yaml:"values"`
	// Possible input value as key-value pairs
	KeyValues     map[string]string `json:"keyValues" yaml:"keyValues"`
	CaseSensitive bool              `json:"caseSensitive" yaml:"caseSensitive"`
	// Result output value
	Result string `json:"result" yaml:"result"`
	// Result output values as key-value pairs
	ResultValues map[string]string `json:"results" yaml:"results"`
}

type Options struct {
	// Order in which to search valueset
	LookupOrder []string `json:"lookupOrder" yaml:"lookupOrder"`
	// Which value to return
	Output []string `json:"output" yaml:"output"`
	// Types of string separation i.e. ";", "|", ",", " "
	Separator map[string]string `json:"separator" yaml:"separator"`
}

func (v ValuesetValue) buildResult() *ls.ValuesetLookupResponse {
	ret := ls.ValuesetLookupResponse{
		KeyValues: make(map[string]string),
	}
	if len(v.ResultValues) != 0 {
		for k, v := range v.ResultValues {
			ret.KeyValues[k] = v
		}
		return &ret
	}
	ret.KeyValues[""] = v.Result
	return &ret
}

func (v ValuesetValue) IsDefault() bool { return len(v.Values) == 0 && len(v.KeyValues) == 0 }

func wordCompare(s1, s2 string, caseSensitive bool) bool {
	toWords := func(in string) string {
		out := make([]rune, 0, len(in))
		lastWasSpace := true
		for _, x := range in {
			if unicode.IsSpace(x) {
				lastWasSpace = true
				continue
			}
			if lastWasSpace {
				if len(out) != 0 {
					out = append(out, ' ')
				}
				lastWasSpace = false
			}
			if !caseSensitive {
				x = unicode.ToLower(x)
			}
			out = append(out, x)
		}
		return string(out)
	}
	return toWords(s1) == toWords(s2)
}

func (v ValuesetValue) Match(req ls.ValuesetLookupRequest) (*ls.ValuesetLookupResponse, error) {
	if v.IsDefault() {
		return v.buildResult(), nil
	}
	if len(req.KeyValues) == 0 {
		return nil, nil
	}
	// If request has a single value:
	if len(req.KeyValues) == 1 {
		var key, value string
		for k, v := range req.KeyValues {
			key = k
			value = v
		}
		switch {
		case len(v.KeyValues) > 1:
			return nil, nil

		case len(v.KeyValues) == 0:
			// Check values array
			for _, val := range v.Values {
				if wordCompare(val, value, v.CaseSensitive) {
					return v.buildResult(), nil
				}
			}

		case len(v.KeyValues) == 1:
			// If input did not give a key, still applies
			if len(key) == 0 {
				for _, val := range v.KeyValues {
					if wordCompare(value, val, v.CaseSensitive) {
						return v.buildResult(), nil
					}
				}
				return nil, nil
			}
			// Input has key, must match
			val, ok := v.KeyValues[key]
			if !ok {
				return nil, nil
			}
			if wordCompare(val, value, v.CaseSensitive) {
				return v.buildResult(), nil
			}
		}

		return nil, nil
	}

	// Here, there are multiple key-values
	// they must all match
	if len(v.KeyValues) != len(req.KeyValues) {
		return nil, nil
	}

	for reqk, reqv := range req.KeyValues {
		vvalue, ok := v.KeyValues[reqk]
		if !ok {
			return nil, nil
		}
		if wordCompare(vvalue, reqv, v.CaseSensitive) {
			return v.buildResult(), nil
		}
	}
	return nil, nil
}

func (vs Valueset) Lookup(req ls.ValuesetLookupRequest) (ls.ValuesetLookupResponse, error) {
	var nondef *ls.ValuesetLookupResponse
	var def *ls.ValuesetLookupResponse
	for _, x := range vs.Values {
		if x.IsDefault() {
			if def != nil {
				return ls.ValuesetLookupResponse{}, fmt.Errorf("Multiple defaults in %s", vs.ID)
			}
			def = x.buildResult()
			continue
		}
		res, err := x.Match(req)
		if err != nil {
			return ls.ValuesetLookupResponse{}, err
		}
		if res != nil {
			if nondef != nil {
				return ls.ValuesetLookupResponse{}, fmt.Errorf("Multiple matches for %v in %s", req, vs.ID)
			}
			nondef = res
		}
	}
	if nondef != nil {
		return *nondef, nil
	}
	if def != nil {
		return *def, nil
	}
	return ls.ValuesetLookupResponse{}, nil
}

// Lookup can be used as the external lookup func of LookupProcessor
func (vsets Valuesets) Lookup(ctx *ls.Context, req ls.ValuesetLookupRequest) (ls.ValuesetLookupResponse, error) {
	found := ls.ValuesetLookupResponse{}
	lookup := func(v Valueset) error {
		rsp, err := v.Lookup(req)
		if err != nil {
			return err
		}
		if len(rsp.KeyValues) > 0 {
			if len(found.KeyValues) > 0 {
				return fmt.Errorf("Ambiguous lookup for %s", req)
			}
			found = rsp
		}
		return nil
	}
	ctx.GetLogger().Debug(map[string]interface{}{"valueset.lookup": req})
	if len(req.TableIDs) == 0 {
		for _, v := range vsets.Sets {
			if err := lookup(v); err != nil {
				ctx.GetLogger().Debug(map[string]interface{}{"valueset.err": err})
				return ls.ValuesetLookupResponse{}, err
			}
		}
		ctx.GetLogger().Debug(map[string]interface{}{"valueset.found": found})
		return found, nil
	}
	var n int = len(req.TableIDs)
	var wg sync.WaitGroup
	results := make([]ls.ValuesetLookupResponse, n)
	errs := make([]error, n)
	for idx, id := range req.TableIDs {
		if v, ok := vsets.Services[id]; ok {
			id := id
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				base, err := url.Parse(v)
				if err != nil {
					errs[i] = err
					return
				}
				qparams := base.Query()
				qparams["tableId"] = append(qparams["tableId"], id)
				for k, v := range req.KeyValues {
					qparams.Set(k, v)
				}
				base.RawQuery = qparams.Encode()
				str := base.String()
				ctx.GetLogger().Debug(map[string]interface{}{"valueset.lookup": id, "req": str})
				resp, err := http.Get(str)
				if err != nil {
					errs[i] = err
					return
				}
				var m map[string]string
				err = json.NewDecoder(resp.Body).Decode(&m)
				defer resp.Body.Close()
				if err != nil {
					errs[i] = err
					return
				}
				ctx.GetLogger().Debug(map[string]interface{}{"valueset.lookup": id, "rsp": m})
				results[i] = ls.ValuesetLookupResponse{KeyValues: m}
			}(idx)
		} else if v, ok := vsets.Sets[id]; ok {
			if err := lookup(v); err != nil {
				errs[idx] = err
			}
		} else {
			errs[idx] = fmt.Errorf("Valueset not found: %s", id)
		}
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return ls.ValuesetLookupResponse{}, err
		}
	}
	var counter int
	var resultIdx int
	for idx, res := range results {
		if len(res.KeyValues) > 0 {
			counter++
			if counter >= 2 {
				return ls.ValuesetLookupResponse{}, fmt.Errorf("Ambiguous lookup for %s", req)
			}
			resultIdx = idx
		}
	}
	if counter == 1 {
		return results[resultIdx], nil
	}
	return found, nil
}

type valuesetMarshal struct {
	Valueset
	Services map[string]string `json:"services" yaml:"services"`
	Sets     []Valueset        `json:"valuesets" yaml:"valuesets"`
}

func LoadValuesetFiles(vs *Valuesets, files []string) error {
	if vs.Sets == nil {
		vs.Sets = make(map[string]Valueset)
		vs.Services = make(map[string]string)
	}
	for _, file := range files {
		var vm valuesetMarshal
		err := cmdutil.ReadJSON(file, &vm)
		if err != nil {
			return err
		}
		for _, v := range vm.Sets {
			if _, exists := vs.Sets[v.ID]; exists {
				return fmt.Errorf("Value set %s already defined", v.ID)
			}
			vs.Sets[v.ID] = v
		}
		if len(vm.ID) > 0 {
			if _, exists := vs.Sets[vm.ID]; exists {
				return fmt.Errorf("Value set %s already defined", vm.ID)
			}
			vs.Sets[vm.ID] = vm.Valueset
		}
		for k, v := range vm.Services {
			if _, exists := vs.Services[k]; exists {
				return fmt.Errorf("Service %s already defined", k)
			}
			vs.Services[k] = v
		}
	}
	return nil
}

func (opt Options) splitCell(header, cell string) []string {
	if sep, exists := opt.Separator[header]; !exists {
		return []string{cell}
	} else {
		var ret []string
		spl := strings.Split(cell, sep)
		for _, str := range spl {
			str = strings.TrimSpace(str)
			if str != "" {
				ret = append(ret, str)
			}
		}
		return ret
	}
}

var options_re = regexp.MustCompile(`options.`)

func parseSpreadsheet(sheet [][]string) (optionsRows [][]string, headers []string, data [][]string, err error) {
	var headerIdx int
ROWS:
	for rowIdx, row := range sheet {
		if len(row) == 0 {
			continue
		}
		if options_re.MatchString(row[0]) {
			optionsRows = append(optionsRows, row)
			continue ROWS
		} else {
			headers = row
			headerIdx = rowIdx
			break ROWS
		}
	}
	data = sheet[headerIdx+1:][:]
	if len(data) == 0 {
		err = fmt.Errorf("Empty data")
	}
	return optionsRows, headers, data, err
}

func parseOptions(optionsRows [][]string) Options {
	options := Options{
		LookupOrder: make([]string, 0),
		Output:      make([]string, 0),
		Separator:   map[string]string{},
	}
	for _, opt := range optionsRows {
		if len(opt) == 0 {
			continue
		}
		optionType := opt[0]
		switch optionType {
		case "options.lookupOrder":
			options.LookupOrder = opt[1:]
		case "options.output":
			options.Output = opt[1:]
		case "options.separator":
			// options.separator | DESCRIPTIVE_TEXT | ; | CODE | ;
			for i := 1; i < len(opt)-1; i += 2 {
				sep := strings.TrimSpace(opt[i+1])
				if sep == "" {
					sep = " "
				}
				options.Separator[strings.TrimSpace(opt[i])] = sep
			}
		}
	}
	return options
}

func cartesianProduct(arr [][]string) [][]string {
	n := 1
	for _, a := range arr {
		n *= len(a)
	}
	ans := make([][]string, 0, n)
	if len(arr) == 0 {
		return ans
	}
	if len(arr) == 1 {
		for _, val := range arr[0] {
			ans = append(ans, []string{val})
		}
		return ans
	}
	cross := cartesianProduct(arr[1:])
	for _, val := range arr[0] {
		for _, perm := range cross {
			ans = append(ans, append([]string{val}, perm...))
		}
	}
	return ans
}

func parseData(sheetName string, headers []string, data [][]string, options Options) (Valueset, error) {
	vs := Valueset{Values: make([]ValuesetValue, 0), Options: options}

	for rowIdx := range data {
		splits := make([][]string, len(headers))
		for hdrIdx, header := range headers {
			cellData := data[rowIdx][hdrIdx]
			if cellData == "" {
				continue
			}
			sep_split := options.splitCell(header, cellData)
			splits[hdrIdx] = sep_split
		}
		for _, permute := range cartesianProduct(splits) {
			vsv := ValuesetValue{KeyValues: map[string]string{}}
			for headerIdx, hdr := range headers {
				vsv.KeyValues[hdr] = permute[headerIdx]
			}
			vs.Values = append(vs.Values, vsv)
		}
	}
	return vs, nil
}

func (vsets Valuesets) LoadSpreadsheets(ctx *ls.Context) error {
	for _, spreadsheet := range vsets.Spreadsheets {
		sheets, err := cmdutil.ReadSpreadsheetFile(spreadsheet)
		if err != nil {
			return fmt.Errorf("While reading file: %s, %w", spreadsheet, err)
		}
		for sheetName, sheet := range sheets {
			if _, exists := vsets.Sets[sheetName]; exists {
				return fmt.Errorf("Value set %s already defined -- in file: %s", sheetName, sheet)
			}
			optionsRows, headers, data, err := parseSpreadsheet(sheet)
			if err != nil {
				return fmt.Errorf("Error parsing sheet %s, %w", sheetName, err)
			}
			options := parseOptions(optionsRows)
			vsets.Sets[sheetName], err = parseData(sheetName, headers, data, options)
			if err != nil {
				return fmt.Errorf("Error parsing sheet data %s, %w", sheetName, err)
			}
		}
	}
	return nil
}

func loadValuesetsCmd(cmd *cobra.Command, valuesets *Valuesets) {
	vsf, _ := cmd.Flags().GetStringSlice("valueset")
	if len(vsf) > 0 {
		err := LoadValuesetFiles(valuesets, vsf)
		if err != nil {
			failErr(err)
		}
	}
}

type ValuesetStep struct {
	BaseIngestParams
	ValuesetFiles []string `json:"valuesetFiles" yaml:"valuesetFiles"`
	initialized   bool
	valuesets     Valuesets
	layer         *ls.Layer
}

func (ValuesetStep) Help() {
	fmt.Println(`Valueset Lookup
Perform valueset lookup on an ingested graph

operation: valueset
params:
  valuesetFiles:
  - f1
  - f2

  # Specify the schema the input graph was ingested with`)
	fmt.Println(baseIngestParamsHelp)
}

func (vs *ValuesetStep) Run(pipeline *PipelineContext) error {
	if !vs.initialized {
		err := LoadValuesetFiles(&vs.valuesets, vs.ValuesetFiles)
		if err != nil {
			return err
		}
		if vs.IsEmptySchema() {
			vs.layer, _ = pipeline.Properties["layer"].(*ls.Layer)
		} else {
			vs.layer, err = LoadSchemaFromFileOrRepo(pipeline.Context, vs.CompiledSchema, vs.Repo, vs.Schema, vs.Type, vs.Bundle)
			if err != nil {
				return err
			}
		}
		if vs.layer == nil {
			return fmt.Errorf("No schema")
		}
		vs.initialized = true
	}
	builder := ls.NewGraphBuilder(pipeline.GetGraphRW(), ls.GraphBuilderOptions{
		EmbedSchemaNodes: true,
	})

	pipeline.Context.GetLogger().Debug(map[string]interface{}{"pipeline": "valueset"})
	prc := ls.NewValuesetProcessor(vs.layer, vs.valuesets.Lookup)
	err := prc.ProcessGraph(pipeline.Context, builder)
	if err != nil {
		return err
	}
	return pipeline.Next()
}

func init() {
	rootCmd.AddCommand(valuesetCmd)
	valuesetCmd.Flags().String("input", "json", "Input graph format (json, jsonld)")
	valuesetCmd.Flags().String("output", "json", "Output format, json, jsonld, or dot")
	valuesetCmd.Flags().StringSlice("valueset", nil, "Valueset file(s)")
	addSchemaFlags(valuesetCmd.Flags())

	operations["valueset"] = func() Step { return &ValuesetStep{} }
}

var valuesetCmd = &cobra.Command{
	Use:   "valueset",
	Short: "Apply valueset to a graph",
	Long: `Apply valueset processing to a graph.

The valuesets are defined in JSON or YAML files with the following structure:
{
  "valuesets": [
    valueSet
  ]
}

Individual valueset objects can be given as separate files as well:

{
  "id": "valueset id",
  "values": [
     // Multiple input values mapping to a result value
   {
     "values": [ "possible input values" ],
     "caseSensitive": "true",
     "result": "resultValue"
   },
    // Input key-value pairs mapping to a key-value object
   {
     "keyValues": {
        "i1": "v1",
        "i2": "v2"
     },
     "results": {
        "r1": "result1",
        "r2": "result2"
     }
   }
  ]
}

`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		step := &ValuesetStep{}
		step.fromCmd(cmd)
		step.ValuesetFiles, _ = cmd.Flags().GetStringSlice("valueset")
		p := []Step{
			NewReadGraphStep(cmd),
			step,
			NewWriteGraphStep(cmd),
		}
		_, err := runPipeline(p, "", args)
		return err
	},
}
