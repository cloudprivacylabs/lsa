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

package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/bserdar/digraph"
	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/pkg/dot"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/repo/fs"
)

func init() {
	rootCmd.AddCommand(ingestCmd)
	ingestCmd.PersistentFlags().String("repo", "", "Schema repository directory")
	ingestCmd.PersistentFlags().String("format", "json", "Output format, json(ld), rdf, or dot")
	ingestCmd.PersistentFlags().Bool("includeSchema", false, "Include schema in the output")
}

var ingestCmd = &cobra.Command{
	Use:   "ingest",
	Short: "Ingest and enrich data with a schema",
}

func LoadSchemaFromFileOrRepo(compiledSchema, repoDir, schemaName string, interner ls.Interner) (*ls.Layer, error) {
	var layer *ls.Layer
	if len(compiledSchema) > 0 {
		sch, err := ioutil.ReadFile(compiledSchema)
		if err != nil {
			return nil, err
		}
		var v interface{}
		err = json.Unmarshal(sch, &v)
		if err != nil {
			return nil, err
		}
		layer, err = ls.UnmarshalLayer(v, interner)
		if err != nil {
			return nil, err
		}
		compiler := ls.Compiler{}
		layer, err = compiler.CompileSchema(layer)
		if err != nil {
			return nil, err
		}
	}
	if layer == nil {
		var repo *fs.Repository
		if len(repoDir) > 0 {
			var err error
			repo, err = getRepo(repoDir, interner)
			if err != nil {
				return nil, err
			}
		}
		if len(schemaName) > 0 {
			if repo != nil {
				var err error
				layer, err = repo.GetComposedSchema(schemaName)
				if err != nil {
					return nil, err
				}
				compiler := ls.Compiler{
					Loader: func(x string) (*ls.Layer, error) {
						if manifest := repo.GetSchemaManifestByObjectType(x); manifest != nil {
							x = manifest.ID
						}
						return repo.LoadAndCompose(x)
					},
				}
				layer, err = compiler.Compile(schemaName)
				if err != nil {
					return nil, err
				}
			} else {
				var v interface{}
				err := readJSON(schemaName, &v)
				if err != nil {
					return nil, err
				}
				layer, err = ls.UnmarshalLayer(v, interner)
				if err != nil {
					return nil, err
				}
				compiler := ls.Compiler{
					Loader: func(x string) (*ls.Layer, error) {
						if x == schemaName || x == layer.GetID() {
							return layer, nil
						}
						return nil, fmt.Errorf("Not found")
					},
				}
				layer, err = compiler.Compile(schemaName)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return layer, nil
}

func OutputIngestedGraph(outFormat string, target *digraph.Graph, wr io.Writer, includeSchema bool) error {
	if !includeSchema {
		schemaNodes := make(map[ls.Node]struct{})
		for nodes := target.GetAllNodes(); nodes.HasNext(); {
			node := nodes.Next().(ls.Node)
			if ls.IsDocumentNode(node) {
				for _, edge := range node.OutWith(ls.InstanceOfTerm).All() {
					schemaNodes[edge.GetTo().(ls.Node)] = struct{}{}
				}
			}
		}
		newTarget := digraph.New()
		ls.Copy(target, newTarget, func(n ls.Node) bool {
			if !ls.IsAttributeNode(n) {
				return true
			}
			if _, ok := schemaNodes[n]; ok {
				return true
			}
			return false
		},
			func(edge ls.Edge) bool {
				return !ls.IsAttributeTreeEdge(edge)
			})
		target = newTarget
	}
	switch outFormat {
	case "dot":
		renderer := dot.Renderer{Options: dot.DefaultOptions()}
		renderer.Render(target, "g", wr)
	case "json", "jsonld":
		marshaler := ls.LDMarshaler{}
		out := marshaler.Marshal(target)
		x, _ := json.MarshalIndent(out, "", "  ")
		wr.Write(x)
	case "rdf":
	default:
		return fmt.Errorf("unknown output format")
	}
	return nil
}
