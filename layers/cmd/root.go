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
	"errors"
	"fmt"
	"log"
	"os"
	"runtime/pprof"

	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/repo/fs"
)

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
		},
		PersistentPostRun: func(cmd *cobra.Command, _ []string) {
			if f, _ := cmd.Flags().GetString("cpuprofile"); len(f) > 0 {
				pprof.StopCPUProfile()
			}
		},
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().String("cpuprofile", "", "Write cpu profile to file")
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

func getRepo(repodir string, interner ls.Interner) (*fs.Repository, error) {
	repo := fs.NewWithInterner(repodir, interner)
	if err := repo.Load(); err != nil {
		if errors.Is(err, fs.ErrNoIndex) || errors.Is(err, fs.ErrBadIndex) {
			warnings, err := repo.UpdateIndex()
			if len(warnings) > 0 {
				for _, x := range warnings {
					fmt.Println(x)
				}
			}
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	if repo.IsIndexStale() {
		warnings, err := repo.UpdateIndex()
		if len(warnings) > 0 {
			for _, x := range warnings {
				fmt.Println(x)
			}
		}
		if err != nil {
			return nil, err
		}
	}
	return repo, nil
}
