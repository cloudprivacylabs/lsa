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

package csv

import (
	"errors"

	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

const CSV = ls.LS + "csv/"

func ParseIngest(context *ls.Context, ingester *ls.Ingester, parser Parser, builder ls.GraphBuilder, baseID string, data []string) (*lpg.Node, error) {
	parsed, err := parser.ParseDoc(context, baseID, data)
	if err != nil {
		return nil, err
	}
	if parsed == nil {
		return nil, errors.New("Parsed CSV document is nil")
	}

	r, err := ingester.Ingest(builder, parsed)
	if err != nil {
		return nil, err
	}
	if err := builder.LinkNodes(context, ingester.Schema); err != nil {
		return nil, err
	}
	return r, nil
}
