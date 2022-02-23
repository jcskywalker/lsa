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
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
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

type Options struct {
	Table TableOptions
	TD    TableCellOptions
}

func DefaultOptions() Options {
	return Options{
		Table: TableOptions{
			CellSpacing: "0",
			Border:      "0",
		},
		TD: TableCellOptions{
			Border: "1",
		},
	}
}

// SchemaNodeRenderer renders the node as an HTML table
func SchemaNodeRenderer(ID string, node graph.Node, options *Options) string {
	to := options.Table
	to.ID = ID
	wr := &bytes.Buffer{}
	io.WriteString(wr, fmt.Sprintf("%s [shape=plaintext label=<", ID))
	io.WriteString(wr, to.String())

	io.WriteString(wr, "<TR>")
	io.WriteString(wr, options.TD.String())
	io.WriteString(wr, ls.GetNodeID(node))
	io.WriteString(wr, "</TD></TR>")

	io.WriteString(wr, "<TR>")
	io.WriteString(wr, options.TD.String())
	node.ForEachProperty(func(k string, v interface{}) bool {
		if pv, ok := v.(*ls.PropertyValue); ok {
			io.WriteString(wr, fmt.Sprintf("%s=%v<br/>", k, pv))
		}
		return true
	})
	io.WriteString(wr, "</TD></TR></TABLE>>];\n")
	return wr.String()
}

func DocNodeRenderer(ID string, node graph.Node, options *Options) string {
	to := options.Table
	to.ID = ID
	wr := &bytes.Buffer{}
	io.WriteString(wr, fmt.Sprintf("%s [shape=plaintext label=<", ID))
	io.WriteString(wr, to.String())

	io.WriteString(wr, "<TR>")
	io.WriteString(wr, options.TD.String())
	io.WriteString(wr, ls.GetNodeID(node))

	io.WriteString(wr, "</TD></TR>")

	io.WriteString(wr, "<TR>")
	io.WriteString(wr, options.TD.String())
	if v, ok := ls.GetRawNodeValue(node); ok {
		io.WriteString(wr, fmt.Sprintf("@value=%v<br/>", v))
	}
	node.ForEachProperty(func(k string, v interface{}) bool {
		if pv, ok := v.(*ls.PropertyValue); ok {
			io.WriteString(wr, fmt.Sprintf("%s=%v<br/>", k, pv))
		}
		return true
	})
	io.WriteString(wr, "</TD></TR>")
	io.WriteString(wr, "</TABLE>>];\n")
	return wr.String()
}

type Renderer struct {
	Options          Options
	NodeSelectorFunc func(graph.Node) bool
	EdgeSelectorFunc func(graph.Edge) bool
}

func (r Renderer) NodeRenderer(ID string, n graph.Node, wr io.Writer) (bool, error) {
	node := n.(graph.Node)
	if r.NodeSelectorFunc != nil && !r.NodeSelectorFunc(node) {
		return false, nil
	}
	if node.GetLabels().Has(ls.AttributeNodeTerm) {
		_, err := io.WriteString(wr, SchemaNodeRenderer(ID, node, &r.Options))
		return true, err
	}
	if node.GetLabels().Has(ls.DocumentNodeTerm) {
		_, err := io.WriteString(wr, DocNodeRenderer(ID, node, &r.Options))
		return true, err
	}
	return true, graph.DefaultDOTNodeRender(ID, node, wr)
}

func (r Renderer) EdgeRenderer(fromID, toID string, edge graph.Edge, w io.Writer) (bool, error) {
	if r.EdgeSelectorFunc == nil || r.EdgeSelectorFunc(edge) {
		return true, graph.DefaultDOTEdgeRender(fromID, toID, edge, w)
	}
	return false, nil
}

func (r Renderer) Render(g graph.Graph, graphName string, out io.Writer) error {
	dr := graph.DOTRenderer{NodeRenderer: r.NodeRenderer, EdgeRenderer: r.EdgeRenderer}
	return dr.Render(g, graphName, out)
}
