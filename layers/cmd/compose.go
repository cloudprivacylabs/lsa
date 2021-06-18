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

	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/pkg/jsonld"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func init() {
	rootCmd.AddCommand(composeCmd)
	composeCmd.Flags().StringP("output", "o", "", "Output file")
	composeCmd.Flags().String("repo", "", "Schema repository directory. If a repository is given, all layers are resolved using that repository. Otherwise, all layers are read as files.")
}

var composeCmd = &cobra.Command{
	Use:   "compose",
	Short: "Compose a schema from components",
	Long:  `Compose a schema from components and output the resulting schema layer.`,

	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoDir, _ := cmd.Flags().GetString("repo")
		var output *ls.Layer
		if len(repoDir) == 0 {
			inputs, err := readJSONMultiple(args)
			if err != nil {
				failErr(err)
			}
			for i, input := range inputs {
				layer, err := jsonld.UnmarshalLayer(input)
				if err != nil {
					fail(fmt.Sprintf("Cannot unmarshal %s: %v", args[i], err))
				}
				if output == nil {
					output = layer
				} else {
					if err := output.Compose(layer); err != nil {
						fail(fmt.Sprintf("Cannot compose %s: %s", args[i], err))
					}
				}
			}
		} else {
			repo, err := getRepo(repoDir)
			if err != nil {
				failErr(err)
			}
			output, err = repo.GetComposedSchema(args[0])
			if err != nil {
				failErr(err)
			}
		}
		if output != nil {
			out := jsonld.MarshalLayer(output)
			d, _ := json.MarshalIndent(out, "", "  ")
			fmt.Println(string(d))
		}
	},
}
