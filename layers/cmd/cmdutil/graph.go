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

package cmdutil

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/pkg/dot"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

func ReadGraph(gfile []string, interner ls.Interner, inputFormat string) (graph.Graph, error) {
	if inputFormat == "json" {
		return ReadJSONGraph(gfile, interner)
	}
	if inputFormat == "jsonld" {
		return ReadJSONLDGraph(gfile, interner)
	}
	return nil, fmt.Errorf("Unrecognized input format: %s", inputFormat)
}

func ReadJSONLDGraph(gfile []string, interner ls.Interner) (graph.Graph, error) {
	data, err := ReadFileOrStdin(gfile)
	if err != nil {
		return nil, err
	}
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	g := graph.NewOCGraph()
	err = ls.UnmarshalJSONLDGraph(v, g, interner)
	return g, err
}

func ReadJSONGraph(gfile []string, interner ls.Interner) (graph.Graph, error) {
	data, err := ReadFileOrStdin(gfile)
	if err != nil {
		return nil, err
	}
	target := graph.NewOCGraph()
	m := ls.JSONMarshaler{}
	err = m.Unmarshal(data, target)
	return target, err
}

func WriteGraph(cmd *cobra.Command, graph graph.Graph, format string, out io.Writer) error {
	switch format {
	case "json":
		m := ls.JSONMarshaler{}
		return m.Encode(graph, out)
	case "jsonld":
		marshaler := ls.LDMarshaler{}
		intf := marshaler.Marshal(graph)
		enc := json.NewEncoder(out)
		return enc.Encode(intf)
	case "dot":
		renderer := dot.Renderer{Options: dot.DefaultOptions()}
		renderer.Options.Rankdir, _ = cmd.Flags().GetString("rankdir")
		renderer.Render(graph, "g", out)
		return nil
	}

	return fmt.Errorf("Unrecognized output format: %s", format)
}
