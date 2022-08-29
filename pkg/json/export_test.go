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
	"io/ioutil"
	"testing"

	"github.com/cloudprivacylabs/lpg"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func TestExport(t *testing.T) {
	var v interface{}

	data, err := ioutil.ReadFile("testdata/vcgraph.json")
	if err != nil {
		t.Error(err)
		return
	}
	if err := json.Unmarshal(data, &v); err != nil {
		t.Error(err)
		return
	}

	target := lpg.NewGraph()
	err = ls.UnmarshalJSONLDGraph(v, target, nil)
	if err != nil {
		t.Error(err)
		return
	}
	source := lpg.Sources(target)[0]
	node, err := Export(source, ExportOptions{})
	if err != nil {
		t.Error(err)
		return
	}

	out := bytes.Buffer{}
	node.Encode(&out)

	t.Log(out.String())

}
