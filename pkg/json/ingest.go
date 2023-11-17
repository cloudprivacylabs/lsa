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

	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func IngestBytes(ctx *ls.Context, baseID string, input []byte, parser Parser, builder ls.GraphBuilder, ingester *ls.Ingester) (*lpg.Node, error) {
	return IngestStream(ctx, baseID, bytes.NewReader(input), parser, builder, ingester)
}

func IngestStream(ctx *ls.Context, baseID string, input io.Reader, parser Parser, builder ls.GraphBuilder, ingester *ls.Ingester) (*lpg.Node, error) {
	node, err := jsonom.UnmarshalReader(input, ctx.GetInterner())
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, nil
	}
	return IngestNode(ctx, baseID, node, parser, builder, ingester)
}

func IngestNode(ctx *ls.Context, baseID string, node jsonom.Node, parser Parser, builder ls.GraphBuilder, ingester *ls.Ingester) (*lpg.Node, error) {
	pd, err := parser.ParseDoc(ctx, baseID, node)
	if err != nil {
		return nil, err
	}
	if pd == nil {
		return nil, nil
	}
	root, err := ingester.Ingest(builder, pd)
	if err != nil {
		return nil, err
	}
	if err := builder.LinkNodes(ctx, parser.Layer); err != nil {
		return nil, err
	}
	return root, nil
}
