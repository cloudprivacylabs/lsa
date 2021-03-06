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
	"os"

	"github.com/spf13/cobra"
)

func init() {
	graphCmd.AddCommand(graphDotCmd)
}

var graphDotCmd = &cobra.Command{
	Use:   "dot",
	Short: "Write graph as a DOT file",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		g, err := ReadGraph(args[0])
		if err != nil {
			failErr(err)
		}
		err = OutputIngestedGraph("dot", g, os.Stdout, true)
		if err != nil {
			failErr(err)
		}
	},
}
