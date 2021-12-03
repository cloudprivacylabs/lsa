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

package json

import (
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

func TestAnnotations(t *testing.T) {
	compiler := jsonschema.NewCompiler()
	compiler.AddResource("/schema", strings.NewReader(`{
 "type":"object",
 "properties": {
   "p1": {
     "type": "string",
     "x-ls": {
       "field": "value"
     }
   }
 }
}`))
	compiled, err := CompileWith(compiler, []Entity{{Name: "schema", Ref: "/schema", ID: "id"}})
	if err != nil {
		t.Error(err)
	}
	if compiled[0].Schema.Properties["p1"].Extensions[X_LS].(annotationExtSchema)["field"].AsString() != "value" {
		t.Errorf("No extension")
	}
}
