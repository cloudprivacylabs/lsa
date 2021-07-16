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
	"github.com/santhosh-tekuri/jsonschema/v3"
	"github.com/spf13/cobra"
)

func init() {
	validateCmd.AddCommand(validateJsonCmd)
}

var validateJsonCmd = &cobra.Command{
	Use:   "json",
	Short: "Validate a JSON document using a schema",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		schemaFile, err := cmd.Flags().GetString("schema")
		if err != nil {
			failErr(err)
		}
		compiler := jsonschema.NewCompiler()
		sch, err := compiler.Compile(schemaFile)
		if err != nil {
			failErr(err)
		}
		var data interface{}
		err = readJSON(args[0], &data)
		if err != nil {
			failErr(err)
		}
		err = sch.ValidateInterface(data)
		if err != nil {
			failErr(err)
		}
	},
}
