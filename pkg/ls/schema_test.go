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
"schemaBase": "http://hl7.org/fhir/Account/schemaBase",
"objectType":  "http://hl7.org/fhir/Account",
"overlays": [
          "http://hl7.org/fhir/Account/overlay/format",
          "http://hl7.org/fhir/Account/overlay/type",
          "http://hl7.org/fhir/Account/overlay/enum",
          "http://hl7.org/fhir/Account/overlay/info"
        ]
}`)
	t.Logf("Expanded: %v", sch)
	s, err := NewSchema(sch)
	if err != nil {
		t.Error(err)
	}
	if s.ID != "http://hl7.org/fhir/Account/schema" ||
		s.SchemaBase != "http://hl7.org/fhir/Account/schemaBase" ||
		s.ObjectType != "http://hl7.org/fhir/Account" {
		t.Errorf("Wrong schema data")
	}
	if s.Overlays[0] != "http://hl7.org/fhir/Account/overlay/format" ||
		s.Overlays[1] != "http://hl7.org/fhir/Account/overlay/type" ||
		s.Overlays[2] != "http://hl7.org/fhir/Account/overlay/enum" ||
		s.Overlays[3] != "http://hl7.org/fhir/Account/overlay/info" {
		t.Errorf("Wrong overlays")
	}
}
