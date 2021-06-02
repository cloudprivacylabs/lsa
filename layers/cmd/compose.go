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
	//	"encoding/json"
	//	"fmt"
	//	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(composeCmd)
	composeCmd.Flags().String("format", "jsonld", "Output format (jsonld, jsonschema)")
	composeCmd.Flags().StringP("output", "o", "", "Output file")
	composeCmd.Flags().String("repo", "", "Schema repository directory. If a repository is given, all layers are resolved using that repository. Otherwise, all layers are read as files.")
}

var composeCmd = &cobra.Command{
	Use:   "compose",
	Short: "Compose a schema from components",
	Long:  `Compose a schema from components and output the resulting schema layer.`,

	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// var repo *fs.Repository
		// repoDir, _ := cmd.Flags().GetString("repo")
		// if len(repoDir) > 0 {
		// 	repo = fs.New(repoDir, ls.Terms, func(fname string, err error) {
		// 		fmt.Printf("%s: %s\n", fname, err)
		// 	})
		// 	if err := repo.Load(true); err != nil {
		// 		failErr(err)
		// 	}
		// }

		// var output *layers.Layer
		// var err error
		// for _, arg := range args {
		// 	var obj interface{}
		// 	if repo == nil {
		// 		obj, err = fs.ReadRepositoryObject(arg)
		// 		if err != nil {
		// 			failErr(err)
		// 		}
		// 	} else {
		// 		manifest := repo.GetSchemaManifest(arg)
		// 		if manifest == nil {
		// 			layer := repo.GetLayer(arg)
		// 			if layer != nil {
		// 				obj = layer
		// 			}
		// 		} else {
		// 			obj = manifest
		// 		}
		// 		if obj == nil {
		// 			fail("Not found: " + arg)
		// 		}
		// 	}

		// 			switch t := obj.(type) {
		// 			case *ls.Layer:
		// 				if output == nil {
		// 					output = t
		// 				} else {
		// 					if err := output.Compose(ls.ComposeOptions{}, ls.Terms, t); err != nil {
		// 						failErr(err)
		// 					}
		// 				}
		// 			case *ls.SchemaManifest:
		// 				var err error
		// 				output, err = repo.GetComposedSchema(t.ID)
		// 				if err != nil {
		// 					failErr(err)
		// 				}
		// 			}
		// 		}

		// 		var data []byte
		// 		format, _ := cmd.Flags().GetString("format")
		// 		switch format {
		// 		case "jsonld":
		// 			data, _ = json.MarshalIndent(output.MarshalExpanded(), "", "  ")
		// 		case "jsonschema":
		// 			// ctx := terms.DefaultJSONOutputContext()
		// 			// out, err := terms.OutputJSONSchema(ctx, output)
		// 			// if err != nil {
		// 			// 	failErr(err)
		// 			// }
		// 			// for _, warn := range ctx.Warnings {
		// 			// 	log.Print(warn)
		// 			// }
		// 			// data, _ = json.MarshalIndent(out, "", "  ")
		// 		}

		// 		outputFlag, _ := cmd.Flags().GetString("output")
		// 		if len(outputFlag) == 0 {
		// 			fmt.Println(string(data))
		// 		} else {
		// 			outFile, err := os.Create(outputFlag)
		// 			if err != nil {
		// 				failErr(err)
		// 			}
		// 			defer outFile.Close()
		// 			fmt.Fprintln(outFile, string(data))
		//		}
	},
}
