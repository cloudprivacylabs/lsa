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

package template

// import (
// 	"bytes"
// 	"encoding/json"
// 	"testing"
// 	"text/template"

// 	"github.com/cloudprivacylabs/lsa/pkg/ls"
// )

// func TestTemplateProcessing(t *testing.T) {
// 	var v interface{}
// 	json.Unmarshal([]byte(`[ {
//             "@id": "http://testschema",
//             "@type": [
//                 "https://lschema.org/Schema"
//             ],
//             "https://lschema.org/layer": [{
//                 "@type": ["https://lschema.org/Attribute",
//                           "https://lschema.org/Object",
//                           "https://lschema.org/targetType"],
//                 "https://lschema.org/Object#attributes": [
//                     {
//                         "@id": "http://attr1",
//                         "@type": ["https://lschema.org/Attribute","https://lschema.org/Value"]
//                     },
//                     {
//                         "@id": "http://attr2",
//                         "@type": ["https://lschema.org/Attribute","https://lschema.org/Object"],
//                         "https://lschema.org/Object#attributes": [
//                             {
//                                 "@id": "http://attr3",
//                                 "@type":  ["https://lschema.org/Attribute","https://lschema.org/Value"]
//                             },
//                             {
//                                 "@id": "http://attr4",
//                                 "@type":  ["https://lschema.org/Attribute","https://lschema.org/Array"],
//                                 "https://lschema.org/Array#items": [
//                                     {
//                                         "@id": "http://attr5",
//                                         "@type": ["https://lschema.org/Attribute","https://lschema.org/Value"]
//                                     }
//                                 ]
//                             }
//                         ]
//                     }
//                 ]
//             }]
// }]`), &v)
// 	layer, err := ls.UnmarshalLayer(v)
// 	if err != nil {
// 		panic(err)
// 	}
// 	tmp := template.New("")
// 	tmp.Funcs(Functions)
// 	tmp, err = tmp.Parse(`{{$x := (gnode .l "http://attr2")}} {{$x.GetID}}`)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	out := bytes.Buffer{}
// 	err = tmp.Execute(&out, map[string]interface{}{"l": layer.Graph})
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	t.Log(out.String())
// }
