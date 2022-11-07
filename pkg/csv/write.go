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
	"encoding/csv"
	"errors"
	"fmt"

	"github.com/cloudprivacylabs/lpg"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher"
)

var ErrMultipleNodesMatched = errors.New("Multiple nodes match query")

type WriterColumn struct {
	Name string `json:"name" yaml:"name"`
	// Optional openCypher queries for each column. The map key is the
	// column name, and the map value is an opencypher query that is
	// evaluated with `root` node set to the current root node.
	Query string `json:"query" yaml:"query"`

	parsedQuery opencypher.Evaluatable
}

// Writer writes CSV output.
//
// The writer specifies how to interpret the input graph. The output
// object specifies an opencypher query that determines each row of
// data.
type Writer struct {
	// openCypher query giving the root nodes for each row of data. This
	// should be of the form:
	//
	//  match (n ...) return n
	//
	// If empty, all root nodes of the graph are included in the output
	RowRootQuery string `json:"rowQuery" yaml:"rowQuery"`

	// The column names in the output. If the column name does not have
	// a column query, then the column query is assumed to be
	//
	//  match (root)-[]->(n:DocumentNode {attributeName: <attributeName>}) return n
	Columns []WriterColumn `json:"columns" yaml:"columns"`
}

// WriteHeader writes the header to the given writer
func (wr *Writer) WriteHeader(writer *csv.Writer) error {
	c := make([]string, 0, len(wr.Columns))
	for _, x := range wr.Columns {
		c = append(c, x.Name)
	}
	return writer.Write(c)
}

func (wr *Writer) BuildRow(root *lpg.Node) ([]string, error) {
	if err := wr.parseColumnQueries(); err != nil {
		return nil, err
	}

	row := make([]string, len(wr.Columns))

	for edges := root.GetEdges(lpg.OutgoingEdge); edges.Next(); {
		node := edges.Edge().GetTo()
		if !node.HasLabel(ls.DocumentNodeTerm) {
			continue
		}
		attrName := ls.AsPropertyValue(node.GetProperty(ls.AttributeNameTerm)).AsString()
		if len(attrName) == 0 {
			continue
		}
		for i := range wr.Columns {
			if wr.Columns[i].Name == attrName {
				row[i], _ = ls.GetRawNodeValue(node)
				break
			}
		}
	}

	for i, col := range wr.Columns {
		if col.parsedQuery == nil {
			continue
		}
		ctx := ls.NewEvalContext(root.GetGraph())
		ctx.SetVar("root", opencypher.RValue{Value: root})
		result, err := col.parsedQuery.Evaluate(ctx)
		if err != nil {
			return nil, err
		}
		// Expexting a single result
		rs, ok := result.Get().(opencypher.ResultSet)
		if !ok {
			return nil, opencypher.ErrExpectingResultSet
		}
		if len(rs.Rows) == 0 {
			row[i] = ""
			continue
		}
		if len(rs.Rows) > 1 {
			return nil, ErrMultipleNodesMatched
		}
		for _, v := range rs.Rows[0] {
			node, ok := v.Get().(*lpg.Node)
			if !ok {
				return nil, fmt.Errorf("Expecting a node in resultset")
			}
			val, _ := ls.GetRawNodeValue(node)
			row[i] = val
		}
	}
	return row, nil
}

func (wr *Writer) WriteRow(writer *csv.Writer, root *lpg.Node) error {
	row, err := wr.BuildRow(root)
	if err != nil {
		return nil
	}
	writer.Write(row)
	return nil
}

func (wr *Writer) parseColumnQueries() error {
	// Are there column queries? Parse them
	for k, col := range wr.Columns {
		if len(col.Query) > 0 {
			ev, err := opencypher.Parse(col.Query)
			if err != nil {
				return err
			}
			wr.Columns[k].parsedQuery = ev
		}
	}
	return nil
}

func (wr *Writer) WriteRows(writer *csv.Writer, g *lpg.Graph) error {
	var roots []*lpg.Node
	if len(wr.RowRootQuery) == 0 {
		roots = lpg.Sources(g)
	} else {
		evalctx := ls.NewEvalContext(g)
		v, err := opencypher.ParseAndEvaluate(wr.RowRootQuery, evalctx)
		if err != nil {
			return err
		}
		rs, ok := v.Get().(opencypher.ResultSet)
		if !ok {
			return opencypher.ErrExpectingResultSet
		}
		for _, row := range rs.Rows {
			if len(row) == 1 {
				for _, v := range row {
					if node, ok := v.Get().(*lpg.Node); ok {
						roots = append(roots, node)
					} else {
						return fmt.Errorf("Expecting a node in resultset")
					}
				}
			}
		}
	}

	for _, root := range roots {
		if !root.HasLabel(ls.DocumentNodeTerm) {
			continue
		}
		if err := wr.WriteRow(writer, root); err != nil {
			return err
		}
	}

	return nil
}
