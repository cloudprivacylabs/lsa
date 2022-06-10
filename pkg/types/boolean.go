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
	"github.com/cloudprivacylabs/opencypher/graph"
)

var JSONBooleanTerm = ls.NewTerm(JSON, "boolean", false, false, ls.OverrideComposition, struct {
	JSONBooleanParser
}{
	JSONBooleanParser{},
}, "json:boolean")

var XMLBooleanTerm = ls.NewTerm(XSD, "boolean", false, false, ls.OverrideComposition, struct {
	XMLBooleanParser
}{
	XMLBooleanParser{},
}, "xsd:boolean", "xs:boolean")

type JSONBooleanParser struct{}

func (JSONBooleanParser) GetNativeValue(value string, node graph.Node) (interface{}, error) {
	if value == "false" {
		return false, nil
	}
	if value == "true" {
		return true, nil
	}
	return nil, ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: JSONBooleanTerm, Value: value}
}

func (JSONBooleanParser) FormatNativeValue(newValue, oldValue interface{}, node graph.Node) (string, error) {
	if newValue == nil {
		return "", nil
	}
	switch v := newValue.(type) {
	case bool:
		return fmt.Sprint(v), nil
	case string:
		if v == "true" || v == "false" {
			return v, nil
		}
	case int:
		if v == 0 {
			return "false", nil
		}
		return "true", nil
	}
	return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: JSONBooleanTerm, Value: newValue}
}

type XMLBooleanParser struct{}

func (XMLBooleanParser) GetNativeValue(value string, node graph.Node) (interface{}, error) {
	if strings.ToLower(value) == "false" || value == "0" {
		return false, nil
	}
	if strings.ToLower(value) == "true" || value == "1" {
		return true, nil
	}
	return nil, ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: XMLBooleanTerm, Value: value}
}

func (XMLBooleanParser) FormatNativeValue(newValue, oldValue interface{}, node graph.Node) (string, error) {
	if newValue == nil {
		return "", nil
	}
	switch v := newValue.(type) {
	case bool:
		return fmt.Sprint(v), nil
	case string:
		if strings.ToLower(v) == "true" || v == "1" {
			return "true", nil
		}
		if strings.ToLower(v) == "false" || v == "0" {
			return "false", nil
		}
	case int:
		if v == 0 {
			return "false", nil
		}
		return "true", nil
	}
	return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: JSONBooleanTerm, Value: newValue}
}
