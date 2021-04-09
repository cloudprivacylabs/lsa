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
	"encoding/json"
	"testing"
)

func TestMarshal(t *testing.T) {
	base := expand(t, `{
"@context": "../../schemas/ls.jsonld",
"@type":"Schema",
"attributes": [
 {
  "@id": "attr1",
  "@type": "Value"
 },
 {
  "@id":  "attr2",
  "@type": "Value",
  "privacyClassification": [
    {
      "@value": "flg1"
    }
  ]
 }
]
}`)
	s := Layer{}
	if err := s.UnmarshalExpanded(base); err != nil {
		t.Error(err)
	}
	x, _ := json.MarshalIndent(s.MarshalExpanded(), "", "")
	expected := `[
{
"@type": [
"` + uriBase + `/Schema"
],
"` + uriBase + `/Object/attributes": [
{
"@id": "attr1",
"@type": [
"` + uriBase + `/Value"
]
},
{
"@id": "attr2",
"@type": [
"` + uriBase + `/Value"
],
"` + uriBase + `/attr/privacyClassification": [
{
"@value": "flg1"
}
]
}
]
}
]`
	if string(x) != expected {
		t.Errorf("Unexpected: %s\n Expected: %s", string(x), expected)
	}

	attr := s.Attributes.FindByID("attr2")
	if _, ok := attr.GetParent().(*ObjectType); !ok {
		t.Errorf("Wrong parent")
	}
	if GetNodeValue(attr.Values[AttributeAnnotations.Privacy.ID].([]interface{})[0]) != "flg1" {
		t.Errorf("Wrong flag: %v", attr.Values)
	}
}
