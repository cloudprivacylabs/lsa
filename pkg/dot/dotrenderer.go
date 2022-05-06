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

package dot

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
)

type HorizontalAlignment string

const HALIGN_CENTER = "CENTER"
const HALIGN_LEFT = "LEFT"
const HALIGN_RIGHT = "RIGHT"
const HALIGN_TEXT = "TEXT"

type VerticalAlignment string

const VALIGN_TOP = "TOP"
const VALIGN_BOTTOM = "BOTTOM"
const VALIGN_MIDDLE = "MIDDLE"

type TableOptions struct {
	Align       HorizontalAlignment `dot:"ALIGN"`
	BGColor     string              `dot:"BGCOLOR"`
	Border      string              `dot:"BORDER"`
	CellBorder  string              `dot:"CELLBORDER"`
	CellPadding string              `dot:"CELLPADDING"`
	CellSpacing string              `dot:"CELLSPACING"`
	Color       string              `dot:"COLOR"`
	Columns     int                 `dot:"COLUMNS"`
	HRef        string              `dot:"HREF"`
	ID          string              `dot:"ID"`
	Port        string              `dot:"PORT"`
	Rows        int                 `dot:"ROWS"`
	Sides       int                 `dot:"SIDES"`
	Style       string              `dot:"STYLE"`
	Target      string              `dot:"TARGET"`
	Title       string              `dot:"TITLE"`
	Valign      VerticalAlignment   `dot:"VALIGN"`
}

type TableCellOptions struct {
	Align       HorizontalAlignment `dot:"ALIGN"`
	Balign      HorizontalAlignment `dot:"BALIGN"`
	BGColor     string              `dot:"BGCOLOR"`
	Border      string              `dot:"BORDER"`
	CellPadding string              `dot:"CELLPADDING"`
	CellSpacing string              `dot:"CELLSPACING"`
	Color       string              `dot:"COLOR"`
	ColSpan     int                 `dot:"COLSPAN"`
	HRef        string              `dot:"HREF"`
	Port        string              `dot:"PORT"`
	RowSpan     int                 `dot:"ROWSPAN"`
	Style       string              `dot:"STYLE"`
	Target      string              `dot:"TARGET"`
	Valign      VerticalAlignment   `dot:"VALIGN"`
	Font        FontConfig
}

type FontConfig struct {
	Face  string
	Size  int
	Color string
}

func (f FontConfig) String() string {
	ret := ""
	if len(f.Face) > 0 {
		ret += fmt.Sprintf(" fontname=\"%s\" ", f.Face)
	}
	if f.Size > 0 {
		ret += fmt.Sprintf(" fontsize=\"%d\" ", f.Size)
	}
	if len(f.Color) > 0 {
		ret += fmt.Sprintf(" fontcolor=\"%s\" ", f.Color)
	}
	return ret
}

func (f FontConfig) Write(s string) string {
	if len(f.Face) == 0 && f.Size == 0 && len(f.Color) == 0 {
		return s
	}
	out := "<font "
	if len(f.Face) > 0 {
		out += fmt.Sprintf("face=\"%s\" ", f.Face)
	}
	if f.Size > 0 {
		out += fmt.Sprintf("point-size=\"%d\" ", f.Size)
	}
	if len(f.Color) > 0 {
		out += fmt.Sprintf("color=\"%s\" ", f.Color)
	}
	return out + ">" + s + "</font>"
}

func buildOptions(data interface{}) []string {
	ret := make([]string, 0)
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	ty := val.Type()
	for i := 0; i < ty.NumField(); i++ {
		tag := ty.Field(i).Tag.Get("dot")
		if len(tag) > 0 {
			value := val.Field(i)
			if !value.IsZero() && value.Interface() != nil {
				ret = append(ret, fmt.Sprintf(`%s="%v"`, tag, value.Interface()))
			}
		}
	}
	return ret
}

func (t TableOptions) String() string {
	return "<TABLE " + strings.Join(buildOptions(t), " ") + ">"
}

func (t TableCellOptions) String() string {
	return "<TD " + strings.Join(buildOptions(t), " ") + ">"
}

type EdgeOptions struct {
	Font  FontConfig
	Color string
}

func (e EdgeOptions) String() string {
	ret := e.Font.String()
	if len(e.Color) > 0 {
		ret += fmt.Sprintf(" color=\"%s\" ", e.Color)
	}
	return ret
}

type Options struct {
	Font    FontConfig
	Color   string
	Rankdir string

	Table      TableOptions
	Labels     TableCellOptions
	Properties TableCellOptions
	Edges      EdgeOptions
}

var DefaultFontConfig = FontConfig{
	Face:  "Courier",
	Size:  10,
	Color: "gray20",
}

func DefaultOptions() Options {
	return Options{
		Font:  DefaultFontConfig,
		Color: "gray20",
		Table: TableOptions{
			CellSpacing: "0",
			Border:      "0",
			Color:       "gray20",
		},
		Labels: TableCellOptions{
			Border: "1",
			Align:  "CENTER",
			Color:  "gray20",
			Font:   DefaultFontConfig,
		},
		Properties: TableCellOptions{
			Border:      "1",
			Balign:      "LEFT",
			CellPadding: "5",
			Font:        DefaultFontConfig,
		},
		Edges: EdgeOptions{
			Font:  DefaultFontConfig,
			Color: "gray20",
		},
	}
}

type Renderer struct {
	Options          Options
	NodeSelectorFunc func(graph.Node) bool
	EdgeSelectorFunc func(graph.Edge) bool
}

// escapeForDot escapes double quotes and backslashes, and replaces Graphviz's
// "center" character (\n) with a left-justified character.
// See https://graphviz.org/docs/attr-types/escString/ for more info.
func escapeForDot(str string) string {
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(str, `\`, `\\`), `"`, `\"`), "\n", `\l`)
}

func (r Renderer) NodeBoxRenderer(ID string, node graph.Node, wr io.Writer) (bool, error) {
	label := strings.Builder{}
	for l := range node.GetLabels() {
		label.WriteRune(':')
		label.WriteString(l)
		label.WriteString(`\n`)
	}
	label.WriteString(`\n`)
	if id := ls.GetNodeID(node); len(id) > 0 {
		label.WriteString(fmt.Sprintf("id=%s\\l", escapeForDot(id)))
	}
	node.ForEachProperty(func(k string, v interface{}) bool {
		if pv, ok := v.(*ls.PropertyValue); ok {
			label.WriteString(fmt.Sprintf("%s=%v\\l", escapeForDot(k), escapeForDot(fmt.Sprint(pv.GetNativeValue()))))
		}
		return true
	})

	io.WriteString(wr, fmt.Sprintf("%s [shape=box %s label=\"%s\"];\n", ID, r.Options.Font.String(), label.String()))

	return true, nil
}

func (r Renderer) NodeTableRenderer(ID string, node graph.Node, wr io.Writer) (bool, error) {
	to := r.Options.Table
	to.ID = ID
	io.WriteString(wr, fmt.Sprintf("%s [shape=plaintext label=<", ID))
	io.WriteString(wr, to.String())

	io.WriteString(wr, "<TR>")
	io.WriteString(wr, r.Options.Labels.String())
	lbl := bytes.Buffer{}
	for l := range node.GetLabels() {
		io.WriteString(&lbl, ":")
		io.WriteString(&lbl, l)
	}
	io.WriteString(wr, r.Options.Labels.Font.Write(lbl.String()))
	io.WriteString(wr, "</TD></TR>")

	io.WriteString(wr, "<TR>")
	io.WriteString(wr, r.Options.Properties.String())

	if id := ls.GetNodeID(node); len(id) > 0 {
		io.WriteString(wr, r.Options.Properties.Font.Write("id="+id)+"<br/>")
	}
	node.ForEachProperty(func(k string, v interface{}) bool {
		if pv, ok := v.(*ls.PropertyValue); ok {
			io.WriteString(wr, r.Options.Properties.Font.Write(fmt.Sprintf("%s=%v", k, pv))+"<br/>")
		}
		return true
	})
	io.WriteString(wr, "</TD></TR></TABLE>>];\n")
	return true, nil
}

func (r Renderer) renderEdge(fromNode, toNode string, edge graph.Edge, w io.Writer) error {
	lbl := edge.GetLabel()
	if len(lbl) != 0 {
		if _, err := fmt.Fprintf(w, "  %s -> %s [label=\"%s\" %s];\n", fromNode, toNode, escapeForDot(lbl), r.Options.Edges); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(w, "  %s -> %s;\n", fromNode, toNode); err != nil {
			return err
		}
	}
	return nil
}

func (r Renderer) EdgeRenderer(fromID, toID string, edge graph.Edge, w io.Writer) (bool, error) {
	if r.EdgeSelectorFunc == nil || r.EdgeSelectorFunc(edge) {
		return true, r.renderEdge(fromID, toID, edge, w)
	}
	return false, nil
}

func (r Renderer) Render(g graph.Graph, graphName string, out io.Writer) error {
	dr := graph.DOTRenderer{NodeRenderer: r.NodeBoxRenderer, EdgeRenderer: r.EdgeRenderer}
	if _, err := fmt.Fprintf(out, "digraph %s {\n", graphName); err != nil {
		return err
	}
	if len(r.Options.Rankdir) > 0 {
		if _, err := fmt.Fprintf(out, "rankdir=\"%s\";\n", r.Options.Rankdir); err != nil {
			return err
		}
	}

	if len(r.Options.Font.Face) > 0 {
		if _, err := fmt.Fprintf(out, "fontname=\"%s\";\n", r.Options.Font.Face); err != nil {
			return err
		}
	}
	if len(r.Options.Font.Color) > 0 {
		if _, err := fmt.Fprintf(out, "fontcolor=\"%s\";\n", r.Options.Font.Color); err != nil {
			return err
		}
	}
	if r.Options.Font.Size > 0 {
		if _, err := fmt.Fprintf(out, "fontsize=\"%d\";\n", r.Options.Font.Size); err != nil {
			return err
		}
	}
	if len(r.Options.Color) > 0 {
		if _, err := fmt.Fprintf(out, "color=\"%s\";\n", r.Options.Color); err != nil {
			return err
		}
	}

	if err := dr.RenderNodesEdges(g, out); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(out, "}\n"); err != nil {
		return err
	}
	return nil
}
