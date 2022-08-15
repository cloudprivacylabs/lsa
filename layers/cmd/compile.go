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
	addSchemaFlags(compileCmd.Flags())
}

var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile schema(s)",
	Long:  `Compile schemas. If a bundle is given, all schemas in the bundle are compiled`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := getContext()
		repoDir, _ := cmd.Flags().GetString("repo")
		bundleNames, _ := cmd.Flags().GetStringSlice("bundle")
		schemaName, _ := cmd.Flags().GetString("schema")
		typeName, _ := cmd.Flags().GetString("type")
		if len(repoDir) > 0 && len(bundleNames) > 0 {
			fail("One of repo or bundle is required")
		}
		var layer *ls.Layer
		if len(bundleNames) == 0 {
			if len(schemaName) == 0 {
				fail("Schema is required")
			}
			if len(repoDir) > 0 {
				var repo *fs.Repository
				var err error
				repo, err = getRepo(repoDir, ctx.GetInterner())
				if err != nil {
					failErr(err)
				}
				layer, err = repo.GetComposedSchema(ctx, schemaName)
				if err != nil {
					failErr(err)
				}
				compiler := ls.Compiler{
					Loader: ls.SchemaLoaderFunc(func(x string) (*ls.Layer, error) {
						return repo.LoadAndCompose(ctx, x)
					}),
				}
				layer, err = compiler.Compile(ctx, schemaName)
				if err != nil {
					failErr(err)
				}
			} else {
				data, err := cmdutil.ReadURL(schemaName)
				if err != nil {
					failErr(err)
				}
				layers, err := ReadLayers(data, ctx.GetInterner())
				if err != nil {
					failErr(err)
				}
				if len(layers) > 1 {
					fail("There are more than one layers in input")
				}
				layer = layers[0]
				compiler := ls.Compiler{
					Loader: ls.SchemaLoaderFunc(func(x string) (*ls.Layer, error) {
						if x == schemaName || x == layer.GetID() {
							return layer, nil
						}
						return nil, fmt.Errorf("Not found")
					}),
				}
				layer, err = compiler.Compile(ctx, schemaName)
				if err != nil {
					failErr(err)
				}
			}
		} else {
			loader, err := LoadBundle(ctx, bundleNames)
			if err != nil {
				failErr(err)
			}
			compiler := ls.Compiler{
				Loader: loader,
			}
			name := typeName
			if len(name) == 0 {
				name = schemaName
			}
			layer, err = compiler.Compile(ctx, name)
			if err != nil {
				failErr(err)
			}
		}
		marshaler := ls.JSONMarshaler{}
		x, _ := marshaler.Marshal(layer.Graph)
		fmt.Println(string(x))
		return
	},
}
