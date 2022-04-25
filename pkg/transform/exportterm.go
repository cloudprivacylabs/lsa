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

package transform

import (
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
)

var ExportTerm = ls.NewTerm(TRANSFORM, "export", false, false, ls.OverrideComposition, ExportTermSemantics)

type exportTermSemantics struct{}

var ExportTermSemantics = exportTermSemantics{}

func (exportTermSemantics) CompileTerm(target ls.CompilablePropertyContainer, term string, value *ls.PropertyValue) error {
	if value == nil {
		return nil
	}
	target.SetProperty("$compiled_"+ExportTerm, value.MustStringSlice())
	return nil
}

// GetExportVars returns the contents of the compiled export term
func (exportTermSemantics) GetExportVars(node graph.Node) []string {
	v, _ := node.GetProperty("$compiled_" + ExportTerm)
	if v == nil {
		return nil
	}
	return v.([]string)
}
