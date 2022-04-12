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

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

var ErrMultipleNodesMatched = errors.New("Multiple nodes match query")

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
	RowRootQuery string
	// Optional openCypher queries for each column. The map key is the
	// column name, and the map value is an opencypher query that is
	// evaluated with `root` node set to the current root node.
	ColumnQueries map[string]string

	// The column names in the output. If the column name does not have
	// a column query, then the column query is assumed to be
	//
	//  match (root)-[]->(n {attributeName: <attributeName>}) return n
	Columns []string
}

// WriteHeader writes the header to the given writer
func (wr Writer) WriteHeader(writer csv.Writer) error {
	return writer.Write(wr.Columns)
}

func (wr Writer) WriteRows(writer csv.Writer, g graph.Graph) error {
	var roots []graph.Node
	if len(wr.RowRootQuery) == 0 {
		roots = graph.Sources(g)
	} else {
		evalctx := opencypher.NewEvalContext(g)
		v, err := opencypher.ParseAndEvaluate(wr.RowRootQuery, evalctx)
		if err != nil {
			return err
		}
		rs, ok := v.Value.(opencypher.ResultSet)
		if !ok {
			return opencypher.ErrExpectingResultSet
		}
		for _, row := range rs.Rows {
			if len(row) == 1 {
				for _, v := range row {
					if node, ok := v.Value.(graph.Node); ok {
						roots = append(roots, node)
					} else {
						return opencypher.ErrExpectingNode
					}
				}
			}
		}
	}

	// Are there column queries? Parse them
	parsedQueries := make(map[string]opencypher.Evaluatable)
	for k, v := range wr.ColumnQueries {
		ev, err := opencypher.Parse(v)
		if err != nil {
			return err
		}
		parsedQueries[k] = ev
	}
	// Missing queries? fill them
	for _, colName := range wr.Columns {
		if _, exists := parsedQueries[colName]; exists {
			continue
		}
		var err error
		parsedQueries[colName], err = opencypher.Parse(fmt.Sprintf(`match (root)-[]->(n {`+"`"+ls.AttributeNameTerm+"`"+`:"%s"}) return n`, opencypher.EscapeStringLiteral(colName)))
		if err != nil {
			panic(err)
		}
	}

	ctx := opencypher.NewEvalContext(g)
	for _, root := range roots {
		ctx.SetVar("root", opencypher.Value{Value: root})
		row := make([]string, 0, len(wr.Columns))
		for _, col := range wr.Columns {
			result, err := parsedQueries[col].Evaluate(ctx)
			if err != nil {
				return err
			}
			// Expexting a single result
			rs, ok := result.Value.(opencypher.ResultSet)
			if !ok {
				return opencypher.ErrExpectingResultSet
			}
			if len(rs.Rows) == 0 {
				row = append(row, "")
				continue
			}
			if len(rs.Rows) > 1 {
				return ErrMultipleNodesMatched
			}
			for _, v := range rs.Rows[0] {
				node, ok := v.Value.(graph.Node)
				if !ok {
					return opencypher.ErrExpectingNode
				}
				val, _ := ls.GetRawNodeValue(node)
				row = append(row, val)
			}
		}
		writer.Write(row)
	}

	return nil
}
