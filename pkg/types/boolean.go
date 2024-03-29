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

	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

var JSONBooleanTerm = ls.NewTerm(JSON, "boolean", "json:boolean").SetComposition(ls.OverrideComposition).SetMetadata(struct {
	JSONBooleanParser
}{
	JSONBooleanParser{},
}).Register()

var XMLBooleanTerm = ls.NewTerm(XSD, "boolean", "xsd:boolean", "xs:boolean").SetComposition(ls.OverrideComposition).SetMetadata(struct {
	XMLBooleanParser
}{
	XMLBooleanParser{},
}).Register()

type JSONBooleanParser struct{}

func (JSONBooleanParser) GetNativeValue(value string, node *lpg.Node) (interface{}, error) {
	if value == "false" {
		return false, nil
	}
	if value == "true" {
		return true, nil
	}
	return nil, ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: JSONBooleanTerm.Name, Value: value}
}

func (JSONBooleanParser) FormatNativeValue(newValue, oldValue interface{}, node *lpg.Node) (string, error) {
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
	return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: JSONBooleanTerm.Name, Value: newValue}
}

type XMLBooleanParser struct{}

func (XMLBooleanParser) GetNativeValue(value string, node *lpg.Node) (interface{}, error) {
	if strings.ToLower(value) == "false" || value == "0" {
		return false, nil
	}
	if strings.ToLower(value) == "true" || value == "1" {
		return true, nil
	}
	return nil, ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: XMLBooleanTerm.Name, Value: value}
}

func (XMLBooleanParser) FormatNativeValue(newValue, oldValue interface{}, node *lpg.Node) (string, error) {
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
	return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: JSONBooleanTerm.Name, Value: newValue}
}
