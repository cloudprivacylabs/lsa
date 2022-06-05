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
	"net/http"
	"net/url"

	"github.com/cloudprivacylabs/lsa/pkg/types"
)

type ucumUnitService struct {
	serviceURL string
}

func (u ucumUnitService) Parse(in string) (types.Measure, error) {
	query := url.Values{}
	query.Set("value", in)
	rsp, err := http.Get(u.serviceURL + "/unit?" + query.Encode())
	if err != nil {
		return types.Measure{}, err
	}
	defer rsp.Body.Close()
	data, _ := ioutil.ReadAll(rsp.Body)
	if rsp.StatusCode != 200 {
		return types.Measure{}, fmt.Errorf("%d: %s", rsp.StatusCode, string(data))
	}
	var ret types.Measure
	if err := json.Unmarshal(data, &ret); err != nil {
		return types.Measure{}, err
	}
	return ret, nil
}

type ucumResponse struct {
	Status string      `json:"status"`
	ToVal  json.Number `json:"toVal"`
	Msg    []string    `json:"msg"`
	ToUnit struct {
		Name   string `json:"name_"`
		CSCode string `json:"csCode_"`
	} `json:"toUnit"`
}

func (u ucumUnitService) Convert(measure types.Measure, targetUnit string, domain string) (types.Measure, error) {
	query := url.Values{}
	query.Set("value", measure.Value)
	query.Set("unit", measure.Unit)
	query.Set("output", targetUnit)
	rsp, err := http.Get(u.serviceURL + "/convert?" + query.Encode())
	if err != nil {
		return types.Measure{}, err
	}
	defer rsp.Body.Close()
	data, _ := ioutil.ReadAll(rsp.Body)
	if rsp.StatusCode != 200 {
		return types.Measure{}, fmt.Errorf("%d: %s", rsp.StatusCode, string(data))
	}
	var result ucumResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return types.Measure{}, err
	}
	if result.Status != "succeeded" {
		return types.Measure{}, fmt.Errorf("%+v", result)
	}
	ret := types.Measure{
		Value: string(result.ToVal),
		Unit:  result.ToUnit.CSCode,
	}
	return ret, nil

}
