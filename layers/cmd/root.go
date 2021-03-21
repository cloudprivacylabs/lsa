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
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "layers",
		Short: "Layered schema processing CLI",
		Long:  `Use this CLI to compose, slice, process layered schemas.`,
	}
)

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
}

func readURL(input string) ([]byte, error) {
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
	return data, nil
}

func readJSON(input string, output interface{}) error {
	data, err := readURL(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, output)
}

// reads input[0] if it exists, otherwise reads from stdin
func readJSONFileOrStdin(input []string, output interface{}) error {
	if len(input) == 0 {
		dec := json.NewDecoder(os.Stdin)
		return dec.Decode(output)
	}
	return readJSON(input[0], output)
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
