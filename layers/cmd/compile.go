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
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/repo/fs"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
)

func init() {
	rootCmd.AddCommand(compileCmd)
	compileCmd.Flags().String("repo", "", "Schema repository directory")
	compileCmd.Flags().String("schema", "", "If repo is given, the schema id. Otherwise schema file.")
	compileCmd.Flags().String("bundle", "", "Schema bundle.")
}

var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile schema(s)",
	Long:  `Compile schemas. If a bundle is given, all schemas in the bundle are compiled`,
	Run: func(cmd *cobra.Command, args []string) {
		interner := ls.NewInterner()
		repoDir, _ := cmd.Flags().GetString("repo")
		bundleName, _ := cmd.Flags().GetString("bundle")
		schemaName, _ := cmd.Flags().GetString("schema")
		if len(schemaName) == 0 || len(bundleName) == 0 {
			fail("One of schema or bundle is required")
		}
		if len(schemaName) > 0 && len(bundleName) > 0 {
			fail("One of schema or bundle is required")
		}
		if len(schemaName) > 0 {
			var layer *ls.Layer
			if len(repoDir) > 0 {
				var repo *fs.Repository
				var err error
				repo, err = getRepo(repoDir, interner)
				if err != nil {
					failErr(err)
				}
				logf("Loading composed schema for %s", schemaName)
				layer, err = repo.GetComposedSchema(ls.DefaultContext(), schemaName)
				if err != nil {
					failErr(err)
				}
				compiler := ls.Compiler{
					Loader: ls.SchemaLoaderFunc(func(x string) (*ls.Layer, error) {
						logf("Loading %s", x)
						if variant := repo.GetSchemaVariantByObjectType(x); variant != nil {
							x = variant.ID
						}
						return repo.LoadAndCompose(ls.DefaultContext(), x)
					}),
				}
				logf("Compiling schema %s", schemaName)
				layer, err = compiler.Compile(ls.DefaultContext(), schemaName)
				if err != nil {
					failErr(err)
				}
				logf("Compilation complete")
			} else {
				data, err := cmdutil.ReadURL(schemaName)
				if err != nil {
					failErr(err)
				}
				layers, err := ReadLayers(data, interner)
				if err != nil {
					failErr(err)
				}
				if len(layers) > 1 {
					fail("There are more than one layers in input")
				}
				layer = layers[1]
				compiler := ls.Compiler{
					Loader: ls.SchemaLoaderFunc(func(x string) (*ls.Layer, error) {
						if x == schemaName || x == layer.GetID() {
							return layer, nil
						}
						return nil, fmt.Errorf("Not found")
					}),
				}
				layer, err = compiler.Compile(ls.DefaultContext(), schemaName)
				if err != nil {
					failErr(err)
				}
			}
			marshaler := ls.JSONMarshaler{}
			x, _ := marshaler.Marshal(layer.Graph)
			fmt.Println(string(x))
			return
		}

	},
}
