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
	compileCmd.PersistentFlags().String("repo", "", "Schema repository directory")
	compileCmd.Flags().String("schema", "", "If repo is given, the schema id. Otherwise schema file.")
}

var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile a schema",
	Run: func(cmd *cobra.Command, args []string) {
		var layer *ls.Layer
		repoDir, _ := cmd.Flags().GetString("repo")
		interner := ls.NewInterner()
		var repo *fs.Repository
		if len(repoDir) > 0 {
			var err error
			repo, err = getRepo(repoDir, interner)
			if err != nil {
				failErr(err)
			}
		}
		schemaName, _ := cmd.Flags().GetString("schema")
		if len(schemaName) == 0 {
			fail("schema required")
		}
		if repo != nil {
			var err error
			logf("Loading composed schema for %s", schemaName)
			layer, err = repo.GetComposedSchema(ls.DefaultContext(), schemaName)
			if err != nil {
				failErr(err)
			}
			compiler := ls.Compiler{
				Loader: func(x string) (*ls.Layer, error) {
					logf("Loading %s", x)
					if variant := repo.GetSchemaVariantByObjectType(x); variant != nil {
						x = variant.ID
					}
					return repo.LoadAndCompose(ls.DefaultContext(), x)
				},
			}
			logf("Compiling schema %s", schemaName)
			layer, err = compiler.Compile(ls.DefaultContext(), schemaName)
			if err != nil {
				failErr(err)
			}
			logf("Compilation complete")
		} else {
			var v interface{}
			err := cmdutil.ReadJSON(schemaName, &v)
			if err != nil {
				failErr(err)
			}
			layer, err = ls.UnmarshalLayer(v, interner)
			if err != nil {
				failErr(err)
			}
			compiler := ls.Compiler{
				Loader: func(x string) (*ls.Layer, error) {
					if x == schemaName || x == layer.GetID() {
						return layer, nil
					}
					return nil, fmt.Errorf("Not found")
				},
			}
			layer, err = compiler.Compile(ls.DefaultContext(), schemaName)
			if err != nil {
				failErr(err)
			}
		}
		marshaler := ls.JSONMarshaler{}
		x, _ := marshaler.Marshal(layer.Graph)
		fmt.Println(string(x))
	},
}
