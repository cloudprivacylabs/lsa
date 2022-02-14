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

package types

import (
	"fmt"
	"strings"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

var JSONBooleanTerm = ls.NewTerm(JSON, "boolean", false, false, ls.OverrideComposition, struct {
	JSONBooleanParser
}{
	JSONBooleanParser{},
})

var XMLBooleanTerm = ls.NewTerm(XSD, "boolean", false, false, ls.OverrideComposition, struct {
	XMLBooleanParser
}{
	XMLBooleanParser{},
})

type JSONBooleanParser struct{}

func (JSONBooleanParser) GetNodeValue(node graph.Node) (interface{}, error) {
	value, exists, err := getStringNodeValue(node)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	if value == "false" {
		return false, nil
	}
	if value == "true" {
		return true, nil
	}
	return nil, ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: JSONBooleanTerm, Value: value}
}

func (JSONBooleanParser) SetNodeValue(node graph.Node, value interface{}) error {
	if value == nil {
		ls.SetRawNodeValue(node, nil)
		return nil
	}
	switch v := value.(type) {
	case bool:
		ls.SetRawNodeValue(node, fmt.Sprint(v))
		return nil
	case string:
		if value == "true" || value == "false" {
			ls.SetRawNodeValue(node, value)
			return nil
		}
	case int:
		if value == 0 {
			ls.SetRawNodeValue(node, "false")
		} else {
			ls.SetRawNodeValue(node, "true")
		}
	}
	return ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: JSONBooleanTerm, Value: value}
}

type XMLBooleanParser struct{}

func (XMLBooleanParser) GetNodeValue(node graph.Node) (interface{}, error) {
	value, exists, err := getStringNodeValue(node)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	if strings.ToLower(value) == "false" || value == "0" {
		return false, nil
	}
	if strings.ToLower(value) == "true" || value == "1" {
		return true, nil
	}
	return nil, ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: XMLBooleanTerm, Value: value}
}

func (XMLBooleanParser) SetNodeValue(node graph.Node, value interface{}) error {
	if value == nil {
		ls.SetRawNodeValue(node, nil)
		return nil
	}
	switch v := value.(type) {
	case bool:
		ls.SetRawNodeValue(node, fmt.Sprint(v))
		return nil
	case string:
		if strings.ToLower(v) == "true" || v == "1" {
			ls.SetRawNodeValue(node, "true")
			return nil
		}
		if strings.ToLower(v) == "false" || v == "0" {
			ls.SetRawNodeValue(node, "false")
			return nil
		}
	case int:
		if value == 0 {
			ls.SetRawNodeValue(node, "false")
		} else {
			ls.SetRawNodeValue(node, "true")
		}
	}
	return ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: JSONBooleanTerm, Value: value}
}
