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

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/transform"
)

func init() {
	rootCmd.AddCommand(applyLayerCmd)
	applyLayerCmd.Flags().String("repo", "", "Schema repository directory")
	applyLayerCmd.Flags().String("schema", "", "If repo is given, the schema id. Otherwise schema file.")
	applyLayerCmd.Flags().String("type", "", "Use if a bundle is given for data types. The type name to ingest.")
	applyLayerCmd.Flags().String("bundle", "", "Schema bundle.")
}

var applyLayerCmd = &cobra.Command{
	Use:   "applylayer",
	Short: "Apply a layer (schema/overlay) onto an existing graph",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := getContext()
		layer := loadSchemaCmd(ctx, cmd)
		g, err := cmdutil.ReadJSONGraph(args, nil)
		if err != nil {
			failErr(err)
		}
		err = transform.ApplyLayer(ctx, g, layer, false)
		if err != nil {
			failErr(err)
		}
		cmdutil.WriteGraph(cmd, g, "json", os.Stdout)
	},
}
