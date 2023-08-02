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
	"log"
	"os"
	"runtime/pprof"

	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/types"
)

var logger = ls.NewDefaultLogger()

var (
	rootCmd = &cobra.Command{
		Use:   "layers",
		Short: "Layered schema processing CLI",
		Long:  `Use this CLI to compose, slice, process layered schemas.`,
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			if f, _ := cmd.Flags().GetString("cpuprofile"); len(f) > 0 {
				file, err := os.Create(f)
				if err != nil {
					panic(err)
				}
				pprof.StartCPUProfile(file)
			}
			if b, _ := cmd.Flags().GetBool("log"); b {
				logger.Level = ls.LogLevelDebug
			}
		},
		PersistentPostRun: func(cmd *cobra.Command, _ []string) {
			if f, _ := cmd.Flags().GetString("cpuprofile"); len(f) > 0 {
				pprof.StopCPUProfile()
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := cmdutil.LoadConfig(cmd)
			if err != nil {
				return err
			}
			return nil
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().String("cpuprofile", "", "Write cpu profile to file")
	rootCmd.PersistentFlags().Bool("log", false, "Enable logging")
	rootCmd.PersistentFlags().Bool("log.debug", false, "Enable logging at debug level")
	rootCmd.PersistentFlags().Bool("log.info", false, "Enable logging at info level")

	rootCmd.PersistentFlags().String("cfg", "", "configuration spec for node properties and labels (default: layers.config.yaml)")
	rootCmd.PersistentFlags().String("rankdir", "LR", "DOT: rankdir option")
	rootCmd.PersistentFlags().String("units", "", "Units service URL")
}

func getContext() *ls.Context {
	l1, _ := rootCmd.PersistentFlags().GetBool("log")
	l2, _ := rootCmd.PersistentFlags().GetBool("log.debug")
	if l1 || l2 {
		logger.Level = ls.LogLevelDebug
	}
	l1, _ = rootCmd.PersistentFlags().GetBool("log.info")
	if l1 {
		logger.Level = ls.LogLevelInfo
	}

	ctx := ls.DefaultContext().SetLogger(logger)

	units, _ := rootCmd.PersistentFlags().GetString("units")
	if len(units) > 0 {
		svc := ucumUnitService{serviceURL: units}
		types.SetMeasureService(ctx, svc)
	}
	return ctx
}

func failErr(err error) {
	log.Fatalf(err.Error())
}

func fail(msg string) {
	log.Fatalf(msg)
}

func unroll(in interface{}, depth int) interface{} {
	if depth == 0 {
		return nil
	}
	depth--
	switch t := in.(type) {
	case map[string]interface{}:
		ret := map[string]interface{}{}
		for k, v := range t {
			ret[k] = unroll(v, depth)
		}
		return ret
	case []interface{}:
		ret := []interface{}{}
		for i := range t {
			ret = append(ret, unroll(t[i], depth))
		}
		return ret
	}
	return in
}
