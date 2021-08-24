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

	"github.com/bserdar/digraph"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/project"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.Flags().String("schema", "", "If repo is given, the schema id. Otherwise schema file.")
	projectCmd.Flags().String("repo", "", "Schema repository directory")
	projectCmd.PersistentFlags().String("format", "json", "Output format, json(ld), rdf, or dot")
	projectCmd.Flags().String("compiledschema", "", "Use the given compiled schema")
}

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Project a graph using a target schema",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		g, err := ReadGraph(args[0])
		if err != nil {
			failErr(err)
		}
		compiledSchema, _ := cmd.Flags().GetString("compiledschema")
		repoDir, _ := cmd.Flags().GetString("repo")
		schemaName, _ := cmd.Flags().GetString("schema")
		layer, err := LoadSchemaFromFileOrRepo(compiledSchema, repoDir, schemaName)
		if err != nil {
			failErr(err)
		}

		projector := project.Projector{TargetSchema: layer}
		nix := g.GetNodeIndex()
		sources := nix.Sources()
		if len(sources) != 1 {
			fail("Cannot determine the root of the graph")
		}
		rootNode := sources[0]
		result, err := projector.Project(rootNode.(ls.Node))
		if err != nil {
			failErr(err)
		}
		outFormat, _ := cmd.Flags().GetString("format")
		gr := digraph.New()
		gr.AddNode(result)
		err = OutputIngestedGraph(outFormat, gr, os.Stdout, false)
		if err != nil {
			failErr(err)
		}
	},
}
