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
	"bytes"
	"encoding/csv"
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func TestBasicWrite(t *testing.T) {
	input := [][]string{
		{"1", "a", "b", "c", "d"},
		{"2", "e", "f", "g", "h"},
		{"3", "i", "j", "k", "l"},
		{"4", "m", "n", "o", "p"},
	}

	parser := Parser{
		ColumnNames: []string{"v", "w", "x", "y", "z"},
	}
	builder := ls.NewGraphBuilder(nil, ls.GraphBuilderOptions{
		EmbedSchemaNodes: true,
	})
	for _, row := range input {
		doc, err := parser.ParseDoc(ls.DefaultContext(), "row", row)
		if err != nil {
			t.Error(err)
			return
		}
		_, err = ls.Ingest(builder, doc)
		if err != nil {
			t.Error(err)
			return
		}
	}

	wr := Writer{
		Columns: []WriterColumn{
			{Name: "v"},
			{Name: "w"},
			{Name: "x"},
			{Name: "y"},
			{Name: "z"},
		},
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := wr.WriteRows(writer, builder.GetGraph()); err != nil {
		t.Error(err)
		return
	}
	writer.Flush()
	t.Log(buf.String())
	if buf.String() != `1,a,b,c,d
2,e,f,g,h
3,i,j,k,l
4,m,n,o,p
` {
		t.Errorf(buf.String())
	}
}
