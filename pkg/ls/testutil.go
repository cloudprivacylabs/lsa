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
	"io/ioutil"
	"os"
	"testing"
)

type TestCase interface {
	GetName() string
	Run(*testing.T)
}

func RunTestsFromFile(t *testing.T, file string, unmarshal func(json.RawMessage) (TestCase, error)) {
	d, err := ioutil.ReadFile(file)
	if err != nil {
		t.Error(err)
		return
	}
	var cases []json.RawMessage
	err = json.Unmarshal(d, &cases)
	if err != nil {
		t.Error(err)
		return
	}

	for _, c := range cases {
		tc, err := unmarshal(c)
		if err != nil {
			t.Error(err)
		} else {
			if run := os.Getenv("CASE"); run == "" || run == tc.GetName() {
				tc.Run(t)
			}
		}
	}
}

func ReadLayerFromFile(f string) (*Layer, error) {
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return nil, err
	}
	return UnmarshalLayerFromSlice(data)
}

func UnmarshalLayerFromSlice(in []byte) (*Layer, error) {
	var v interface{}
	if err := json.Unmarshal(in, &v); err != nil {
		return nil, err
	}
	return UnmarshalLayer(v, nil)
}
