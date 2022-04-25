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
	"io"

	"github.com/bserdar/jsonom"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
)

func IngestBytes(ctx *ls.Context, baseID string, input []byte, parser Parser, builder ls.GraphBuilder) (graph.Node, error) {
	return IngestStream(ctx, baseID, bytes.NewReader(input), parser, builder)
}

func IngestStream(ctx *ls.Context, baseID string, input io.Reader, parser Parser, builder ls.GraphBuilder) (graph.Node, error) {
	node, err := jsonom.UnmarshalReader(input, ctx.GetInterner())
	if err != nil {
		return nil, err
	}
	pd, err := parser.ParseDoc(ctx, baseID, node)
	if err != nil {
		return nil, err
	}
	return ls.Ingest(builder, pd)
}
