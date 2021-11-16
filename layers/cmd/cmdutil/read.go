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

package cmdutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/text/encoding"
)

// ReadURL reads a file or a URL
func ReadURL(input string, enc ...encoding.Encoding) ([]byte, error) {
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

// ReadJSON reads JSON from a file or a URL
func ReadJSON(input string, output interface{}, enc ...encoding.Encoding) error {
	data, err := ReadURL(input, enc...)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, output)
}

// ReadJSONFileOrStdin reads a JSON file(s), or if there are none, reads from stdin
func ReadJSONFileOrStdin(input []string, output interface{}, enc ...encoding.Encoding) error {
	if len(input) == 0 {
		var rd io.Reader
		rd = os.Stdin
		if len(enc) == 1 {
			rd = enc[0].NewDecoder().Reader(os.Stdin)
		}
		dec := json.NewDecoder(rd)
		return dec.Decode(output)
	}
	return ReadJSON(input[0], output, enc...)
}

// StreamFileOrStdin reads  file(s), or if there are none, reads frm stdin
func StreamFileOrStdin(input []string, enc ...encoding.Encoding) (io.Reader, error) {
	if len(input) == 0 {
		var rd io.Reader
		rd = os.Stdin
		if len(enc) == 1 {
			rd = enc[0].NewDecoder().Reader(os.Stdin)
		}
		return rd, nil
	}
	data, err := ReadURL(input[0], enc...)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}

// ReadFileOrStdin reads a file or reads stdin
func ReadFileOrStdin(input []string) ([]byte, error) {
	if len(input) == 0 {
		return ioutil.ReadAll(os.Stdin)
	}
	return ioutil.ReadFile(input[0])
}

// ReadJSONMultiple reads multiple JSON files
func ReadJSONMultiple(input []string) ([]interface{}, error) {
	out := make([]interface{}, 0, len(input))
	for _, x := range input {
		var o interface{}
		if err := ReadJSON(x, &o); err != nil {
			return nil, fmt.Errorf("While reading %s: %w", x, err)
		}
		out = append(out, o)
	}
	return out, nil
}
