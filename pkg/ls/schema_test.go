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
package ls

import (
	"testing"
)

func TestSchemaMarshal(t *testing.T) {
	sch := expand(t, `{
"@context": "../../schemas/schema.jsonld",
"@id": "http://hl7.org/fhir/Account/schema",
"@type": "Schema",
"objectType":  "http://hl7.org/fhir/Account",
"layers": [
          "http://hl7.org/fhir/Account/base",
          "http://hl7.org/fhir/Account/format",
          "http://hl7.org/fhir/Account/type",
          "http://hl7.org/fhir/Account/enum",
          "http://hl7.org/fhir/Account/info"
        ]
}`)
	t.Logf("Expanded: %v", sch)
	s, err := NewSchema(sch)
	if err != nil {
		t.Error(err)
	}
	if s.ID != "http://hl7.org/fhir/Account/schema" ||
		s.ObjectType != "http://hl7.org/fhir/Account" {
		t.Errorf("Wrong schema data")
	}
	if s.Layers[0] != "http://hl7.org/fhir/Account/base" ||
		s.Layers[1] != "http://hl7.org/fhir/Account/format" ||
		s.Layers[2] != "http://hl7.org/fhir/Account/type" ||
		s.Layers[3] != "http://hl7.org/fhir/Account/enum" ||
		s.Layers[4] != "http://hl7.org/fhir/Account/info" {
		t.Errorf("Wrong layers")
	}
}
