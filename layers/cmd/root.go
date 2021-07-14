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
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime/pprof"

	"golang.org/x/text/encoding"

	"github.com/spf13/cobra"

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

func readURL(input string, enc ...encoding.Encoding) ([]byte, error) {
	var data []byte
	urlInput, err := url.Parse(input)
	if err != nil {
		return nil, err
	}
	if len(urlInput.Scheme) == 0 {
		data, err = ioutil.ReadFile(urlInput.String())
		if err != nil {
			return nil, err
		}
	} else {
		rsp, err := http.Get(urlInput.String())
		if err != nil {
			return nil, err
		}
		if (rsp.StatusCode / 100) != 2 {
			return nil, fmt.Errorf(rsp.Status)
		}
		data, err = ioutil.ReadAll(rsp.Body)
		rsp.Body.Close()
		if err != nil {
			return nil, err
		}
	}
	if len(enc) == 1 {
		return enc[0].NewDecoder().Bytes(data)
	}
	return data, nil
}

func readJSON(input string, output interface{}, enc ...encoding.Encoding) error {
	data, err := readURL(input, enc...)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, output)
}

// reads input[0] if it exists, otherwise reads from stdin
func readJSONFileOrStdin(input []string, output interface{}, enc ...encoding.Encoding) error {
	if len(input) == 0 {
		var rd io.Reader
		rd = os.Stdin
		if len(enc) == 1 {
			rd = enc[0].NewDecoder().Reader(os.Stdin)
		}
		dec := json.NewDecoder(rd)
		return dec.Decode(output)
	}
	return readJSON(input[0], output, enc...)
}

func readFileOrStdin(input []string) ([]byte, error) {
	if len(input) == 0 {
		return ioutil.ReadAll(os.Stdin)
	}
	return ioutil.ReadFile(input[0])
}

func readJSONMultiple(input []string) ([]interface{}, error) {
	out := make([]interface{}, 0, len(input))
	for _, x := range input {
		var o interface{}
		if err := readJSON(x, &o); err != nil {
			return nil, fmt.Errorf("While reading %s: %w", x, err)
		}
		out = append(out, o)
	}
	return out, nil
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

func getRepo(repodir string) (*fs.Repository, error) {
	repo := fs.New(repodir)
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
