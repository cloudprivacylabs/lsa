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
	"encoding/json"
	"fmt"
	"io"

	"github.com/bserdar/digraph"
	"github.com/cloudprivacylabs/lsa/pkg/dot"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func readGraph(gfile []string, interner ls.Interner, inputFormat string) (*digraph.Graph, error) {
	if inputFormat == "json" {
		return readJSONGraph(gfile, interner)
	}
	if inputFormat == "jsonld" {
		return readJSONLDGraph(gfile, interner)
	}
	return nil, fmt.Errorf("Unrecognized input format: %s", inputFormat)
}

func readJSONLDGraph(gfile []string, interner ls.Interner) (*digraph.Graph, error) {
	data, err := readFileOrStdin(gfile)
	if err != nil {
		return nil, err
	}
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return ls.UnmarshalJSONLDGraph(v, interner)
}

func readJSONGraph(gfile []string, interner ls.Interner) (*digraph.Graph, error) {
	data, err := readFileOrStdin(gfile)
	if err != nil {
		return nil, err
	}
	target := digraph.New()
	err = ls.UnmarshalGraphJSON(data, target, interner)
	return target, err
}

func writeGraph(graph *digraph.Graph, format string, out io.Writer) error {
	switch format {
	case "json":
		return ls.EncodeGraphJSON(graph, out)
	case "jsonld":
		marshaler := ls.LDMarshaler{}
		intf := marshaler.Marshal(graph)
		enc := json.NewEncoder(out)
		return enc.Encode(intf)
	case "dot":
		renderer := dot.Renderer{Options: dot.DefaultOptions()}
		renderer.Render(graph, "g", out)
		return nil
	}

	return fmt.Errorf("Unrecognized output format: %s", format)
}
