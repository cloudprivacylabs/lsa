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
	"bytes"
	"encoding/json"
	"testing"
)

func TestDecode(t *testing.T) {
	rd := bytes.NewReader([]byte(`{
    "firstName": "John",
    "lastName": "Doe",
    "extraField": true,
    "contact": [
        {
            "type": "cell",
            "value": "123-123123"
        },
        {
            "type": "work",
            "value": "234-234234"
        }
    ]
}`))
	dec := json.NewDecoder(rd)
	dec.UseNumber()
	out, err := Decode(dec)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%+v", out)
}
