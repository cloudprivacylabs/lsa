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
package rdf

import (
	"fmt"
	"io"
)

type DOTNode struct {
	Node
	ID string
}

type DOTEdge struct {
	From, To string
	Label    string
}

func ToDOT(graphName string, nodes []DOTNode, edges []DOTEdge, out io.Writer) error {
	if _, err := fmt.Fprintf(out, "digraph %s {\n", graphName); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "rankdir=\"LR\";\n"); err != nil {
		return err
	}
	for _, node := range nodes {
		if _, err := fmt.Fprintf(out, "  %s [label=\"%s\"];\n", node.ID, node.GetValue()); err != nil {
			return err
		}
	}
	for _, e := range edges {
		if len(e.Label) > 0 {
			if _, err := fmt.Fprintf(out, "  %s -> %s [label=\"%s\"];\n", e.From, e.To, e.Label); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(out, "  %s -> %s;\n", e.From, e.To); err != nil {
				return err
			}
		}
	}
	if _, err := fmt.Fprintf(out, "}\n"); err != nil {
		return err
	}
	return nil
}
