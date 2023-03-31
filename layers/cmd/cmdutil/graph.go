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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/bserdar/jsonstream"
	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/lsa/pkg/dot"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func ReadGraph(gfile []string, interner ls.Interner, inputFormat string) (*lpg.Graph, error) {
	if inputFormat == "json" {
		return ReadJSONGraph(gfile, interner)
	}
	if inputFormat == "jsonld" {
		return ReadJSONLDGraph(gfile, interner)
	}
	return nil, fmt.Errorf("Unrecognized input format: %s", inputFormat)
}

func StreamGraph(ctx context.Context, file []string, interner ls.Interner, inputFormat string) (<-chan GraphStream, error) {
	if inputFormat == "json" {
		return StreamJSONGraph(ctx, file, interner)
	}
	if inputFormat == "jsonld" {
		return StreamJSONLDGraph(ctx, file, interner)
	}
	return nil, fmt.Errorf("Unrecognized input format: %s", inputFormat)
}

func ReadGraphFromReader(ctx context.Context, reader io.Reader, interner ls.Interner, inputFormat string) (<-chan GraphStream, error) {
	rd := jsonstream.NewConcatReader(reader)
	ret := make(chan GraphStream)
	go func() {
		defer close(ret)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				data, err := rd.ReadRaw()
				if err == io.EOF {
					return
				}
				target := lpg.NewGraph()
				switch inputFormat {
				case "jsonld":
					var v interface{}
					if err := json.Unmarshal(data, &v); err != nil {
						ret <- GraphStream{
							Err: err,
						}
						break
					}
					g := lpg.NewGraph()
					err := ls.UnmarshalJSONLDGraph(v, g, interner)
					ret <- GraphStream{
						G:   g,
						Err: err,
					}

				case "json":
					m := ls.JSONMarshaler{}
					err := m.Unmarshal(data, target)
					ret <- GraphStream{
						G:   target,
						Err: err,
					}
				default:
					ret <- GraphStream{
						Err: fmt.Errorf("Unrecognized input format: %s", inputFormat),
					}
				}
			}
		}
	}()
	return ret, nil
}

func ReadJSONLDGraph(gfile []string, interner ls.Interner) (*lpg.Graph, error) {
	data, err := ReadFileOrStdin(gfile)
	if err != nil {
		return nil, err
	}
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	g := lpg.NewGraph()
	err = ls.UnmarshalJSONLDGraph(v, g, interner)
	return g, err
}

func ReadJSONGraph(gfile []string, interner ls.Interner) (*lpg.Graph, error) {
	data, err := ReadFileOrStdin(gfile)
	if err != nil {
		return nil, err
	}
	target := lpg.NewGraph()
	m := ls.JSONMarshaler{}
	err = m.Unmarshal(data, target)
	return target, err
}

type GraphStream struct {
	G   *lpg.Graph
	Err error
}

func StreamJSONGraph(ctx context.Context, file []string, interner ls.Interner) (<-chan GraphStream, error) {
	reader, err := StreamFileOrStdin(file)
	if err != nil {
		return nil, err
	}
	rd := jsonstream.NewConcatReader(reader)
	ret := make(chan GraphStream)
	go func() {
		defer close(ret)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				data, err := rd.ReadRaw()
				if err == io.EOF {
					return
				}
				target := lpg.NewGraph()
				m := ls.JSONMarshaler{}
				err = m.Unmarshal(data, target)
				ret <- GraphStream{
					G:   target,
					Err: err,
				}
			}
		}
	}()
	return ret, nil
}

func StreamJSONLDGraph(ctx context.Context, file []string, interner ls.Interner) (<-chan GraphStream, error) {
	reader, err := StreamFileOrStdin(file)
	if err != nil {
		return nil, err
	}
	rd := jsonstream.NewConcatReader(reader)
	ret := make(chan GraphStream)
	go func() {
		defer close(ret)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				data, err := rd.ReadRaw()
				if err == io.EOF {
					return
				}
				var v interface{}
				if err := json.Unmarshal(data, &v); err != nil {
					ret <- GraphStream{
						Err: err,
					}
					return
				}
				g := lpg.NewGraph()
				err = ls.UnmarshalJSONLDGraph(v, g, interner)
				if err != nil {
					ret <- GraphStream{
						Err: err,
					}
					return
				}
				ret <- GraphStream{
					G: g,
				}
			}
		}
	}()
	return ret, nil
}

func WriteGraph(cmd *cobra.Command, graph *lpg.Graph, format string, out io.Writer) error {
	switch format {
	case "json":
		m := ls.JSONMarshaler{}
		if err := m.Encode(graph, out); err != nil {
			return err
		}
		_, err := fmt.Fprintln(out, "")
		return err
	case "jsonld":
		marshaler := ls.LDMarshaler{}
		intf := marshaler.Marshal(graph)
		enc := json.NewEncoder(out)
		if err := enc.Encode(intf); err != nil {
			return err
		}
		_, err := fmt.Fprintln(out, "")
		return err
	case "dot":
		renderer := dot.Renderer{Options: dot.DefaultOptions()}
		renderer.Options.Rankdir, _ = cmd.Flags().GetString("rankdir")
		renderer.Render(graph, "g", out)
		return nil
	case "web":
		dotOut := bytes.Buffer{}
		renderer := dot.Renderer{Options: dot.DefaultOptions()}
		renderer.Options.Rankdir, _ = cmd.Flags().GetString("rankdir")
		renderer.Render(graph, "g", &dotOut)
		renderDot(dotOut.Bytes())
		return nil
	}

	return fmt.Errorf("Unrecognized output format: %s", format)
}
