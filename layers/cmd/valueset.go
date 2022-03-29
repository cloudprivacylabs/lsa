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
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type Valuesets struct {
	Sets map[string]Valueset `json:"valuesets" yaml:"valuesets"`
}

type Valueset struct {
	ID     string          `json:"id" yaml:"id"`
	Values []ValuesetValue `json:"values" yaml:"values"`
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

func (v ValuesetValue) Match(req ls.ValuesetLookupRequest) (*ls.ValuesetLookupResponse, error) {
	if v.IsDefault() {
		return v.buildResult(), nil
	}
	filter := func(s string) string { return strings.ToLower(s) }
	if v.CaseSensitive {
		filter = func(s string) string { return s }
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
				if filter(val) == filter(value) {
					return v.buildResult(), nil
				}
			}

		case len(v.KeyValues) == 1:
			// If input did not give a key, still applies
			if len(key) == 0 {
				for _, val := range v.KeyValues {
					if filter(value) == filter(val) {
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
			if filter(val) == filter(value) {
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
		if filter(vvalue) == filter(reqv) {
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
func (vsets Valuesets) Lookup(req ls.ValuesetLookupRequest) (ls.ValuesetLookupResponse, error) {
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
	if len(req.TableIDs) == 0 {
		for _, v := range vsets.Sets {
			if err := lookup(v); err != nil {
				return ls.ValuesetLookupResponse{}, err
			}
		}
		return found, nil
	}
	for _, id := range req.TableIDs {
		if v, ok := vsets.Sets[id]; ok {
			if err := lookup(v); err != nil {
				return ls.ValuesetLookupResponse{}, err
			}
		} else {
			return found, fmt.Errorf("Valueset not found: %s", id)
		}
	}
	return found, nil
}

type valuesetMarshal struct {
	Valueset
	Sets []Valueset `json:"valuesets" yaml:"valuesets"`
}

func LoadValuesetFiles(vs *Valuesets, files []string) error {
	if vs.Sets == nil {
		vs.Sets = make(map[string]Valueset)
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
