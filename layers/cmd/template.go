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

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"text/template"

// 	"github.com/spf13/cobra"

// 	"github.com/cloudprivacylabs/lsa/pkg/ls"
// 	lstemplate "github.com/cloudprivacylabs/lsa/pkg/template"
// )

// func init() {
// 	rootCmd.AddCommand(templateCmd)
// 	templateCmd.Flags().String("graph", "", "Input graph")
// 	templateCmd.Flags().String("template", "", "Template file")
// 	templateCmd.MarkFlagRequired("template")
// 	templateCmd.MarkFlagRequired("graph")
// }

// var templateCmd = &cobra.Command{
// 	Use:   "template",
// 	Short: "Generate output using a Go template from a graph",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		tfile, _ := cmd.Flags().GetString("template")
// 		tmp := template.New("")
// 		tmp.Funcs(lstemplate.Functions)
// 		data, err := ioutil.ReadFile(tfile)
// 		if err != nil {
// 			failErr(err)
// 		}
// 		_, err = tmp.Parse(string(data))
// 		if err != nil {
// 			failErr(err)
// 		}

// 		gfile, _ := cmd.Flags().GetString("graph")
//    graph, err:=ReadGraph(gFile)
// 		if err != nil {
// 			failErr(err)
// 		}
// 		var out bytes.Buffer
// 		err = tmp.Execute(&out, map[string]interface{}{"g": graph})
// 		if err != nil {
// 			failErr(err)
// 		}
// 		fmt.Print(out.String())
// 	},
// }
