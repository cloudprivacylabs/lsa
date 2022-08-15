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
package bundle

import (
	"fmt"

	"github.com/cloudprivacylabs/lsa/pkg/csv"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type SpreadsheetReference struct {
	Name    string             `json:"name" yaml:"name"`
	Spec    *csv.CSVImportSpec `json:"spec" yaml:"spec"`
	Context []interface{}      `json:"context" yaml:"context"`
}

func (s SpreadsheetReference) Import(ctx *ls.Context, sheetReader func(*ls.Context, string) ([][][]string, error)) (map[string]*ls.Layer, error) {
	ctx.GetLogger().Debug(map[string]interface{}{"spreadSheet": s.Name})
	records, err := sheetReader(ctx, s.Name)
	if err != nil {
		return nil, err
	}

	var toMap func(interface{}) interface{}
	toMap = func(in interface{}) interface{} {
		if arr, ok := in.([]interface{}); ok {
			out := make([]interface{}, 0, len(arr))
			for _, x := range arr {
				out = append(out, toMap(x))
			}
			return out
		}
		if m, ok := in.(map[interface{}]interface{}); ok {
			out := map[string]interface{}{}
			for k, v := range m {
				out[fmt.Sprint(k)] = toMap(v)
			}
			return out
		}
		return in
	}

	if s.Spec != nil {
		if len(records) != 1 {
			return nil, fmt.Errorf("Use a spreadsheet with a single sheet to import with spec")
		}
		layer, err := s.Spec.Import(records[0])
		if err != nil {
			return nil, err
		}
		return map[string]*ls.Layer{layer.GetID(): layer}, nil
	}
	var context map[string]interface{}
	if len(s.Context) > 0 {
		context = map[string]interface{}{"@context": toMap(s.Context)}
	}

	ret := make(map[string]*ls.Layer)
	for _, sheet := range records {
		layers, err := csv.ImportSchema(ctx, sheet, context)
		if err != nil {
			return nil, err
		}
		for _, l := range layers {
			ret[l.GetID()] = l
			ctx.GetLogger().Debug(map[string]interface{}{"spreadSheet": s.Name, "layer": l.GetID()})
		}
	}
	return ret, nil
}
