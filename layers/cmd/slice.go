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
	"bytes"
	"fmt"
	"text/template"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type SliceByTermsSpec struct {
	Type  string   `json:"type"`
	Terms []string `json:"terms"`
	ID    string   `json:"@id"`
	File  string   `json:"file"`
}

func execTemplate(tmpl string, data interface{}) string {
	compiled := template.Must(template.New("").Parse(tmpl))
	out := bytes.Buffer{}
	if err := compiled.Execute(&out, data); err != nil {
		panic(err)
	}
	return out.String()
}

func (spec SliceByTermsSpec) Slice(sourceLayer *ls.Layer, targetType string, templateData interface{}) (*ls.Layer, error) {
	var layer *ls.Layer
	id := execTemplate(spec.ID, templateData)
	switch spec.Type {
	case "Overlay", ls.OverlayTerm:
		f := ls.GetSliceByTermsFunc(spec.Terms, false)
		if spec.Terms == nil {
			f = ls.IncludeAllNodesInSliceFunc
		}
		layer = sourceLayer.Slice(ls.OverlayTerm, f)
	case "Schema", ls.SchemaTerm:
		f := ls.GetSliceByTermsFunc(spec.Terms, true)
		if spec.Terms == nil {
			f = ls.IncludeAllNodesInSliceFunc
		}
		layer = sourceLayer.Slice(ls.SchemaTerm, f)
	default:
		return nil, fmt.Errorf("Layer type unspecified")
	}
	layer.SetID(id)
	layer.SetValueType(targetType)
	return layer, nil
}
