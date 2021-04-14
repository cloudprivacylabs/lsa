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
package fs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/piprate/json-gold/ld"
)

type ErrUnrecognizedObject string

func (e ErrUnrecognizedObject) Error() string { return "Unrecognized object: " + string(e) }

// ReadRepositoryObject reads a schema manifest or schema layer
func ReadRepositoryObject(file string) (interface{}, error) {
	proc := ld.NewJsonLdProcessor()
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}
	expanded, err := proc.Expand(obj, nil)
	if err != nil {
		return nil, err
	}
	ret, err := ParseRepositoryObject(expanded)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func ParseRepositoryObject(expanded interface{}) (interface{}, error) {
	manifest, err1 := ls.SchemaManifestFromLD(expanded)
	if err1 == nil {
		return manifest, nil
	}
	layer, err2 := ls.LayerFromLD(expanded)
	if err2 == nil {
		return layer, nil
	}
	return nil, fmt.Errorf("Cannot parse object: Not a manifest: %v Not a layer: %v", err1, err2)
}
