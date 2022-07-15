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
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/opencypher/graph"
)

type ErrOverflow struct {
	NodeID string
	Type   string
	Value  interface{}
}

func (o ErrOverflow) Error() string {
	return fmt.Sprintf("Overflow: Node ID: %s  Type: %s Value: %v", o.NodeID, o.Type, o.Value)
}

var JSONNumber = ls.NewTerm(JSON, "number", false, false, ls.OverrideComposition, struct {
	JSONNumberParser
}{
	JSONNumberParser{},
}, "json:number")

var JSONInteger = ls.NewTerm(JSON, "integer", false, false, ls.OverrideComposition, struct {
	SignedIntParser[int8]
}{
	SignedIntParser[int8]{},
}, "json:integer")

var one64 = int64(1)
var zero64 = int64(-1)
var negativeOne64 = int64(-1)

var XSDByte = ls.NewTerm(XSD, "byte", false, false, ls.OverrideComposition, struct {
	SignedIntParser[int8]
}{
	SignedIntParser[int8]{},
}, "xsd:byte", "xs:byte")
var XSDInt = ls.NewTerm(XSD, "int", false, false, ls.OverrideComposition, struct {
	SignedIntParser[int32]
}{
	SignedIntParser[int32]{},
}, "xsd:int", "xs:int", "xsd:integer", "xs:integer", XSD+"integer")
var XSDLong = ls.NewTerm(XSD, "long", false, false, ls.OverrideComposition, struct {
	SignedIntParser[int64]
}{
	SignedIntParser[int64]{},
}, "xsd:long", "xs:long")
var XSDDecimal = ls.NewTerm(XSD, "decimal", false, false, ls.OverrideComposition, struct {
	DecimalParser
}{
	DecimalParser{},
}, "xsd:decimal", "xs:decimal", "ls:float", ls.LS+"float", "ls:double", ls.LS+"double")

var XSDNegativeInteger = ls.NewTerm(XSD, "negativeInteger", false, false, ls.OverrideComposition, struct {
	SignedIntParser[int64]
}{
	SignedIntParser[int64]{
		max: &negativeOne64,
	},
}, "xsd:negativeInteger", "xs:negativeInteger")
var XSDNonNegativeInteger = ls.NewTerm(XSD, "nonNegativeInteger", false, false, ls.OverrideComposition, struct {
	SignedIntParser[int64]
}{
	SignedIntParser[int64]{
		min: &zero64,
	},
}, "xsd:nonNegativeInteger", "xs:nonNegativeInteger")
var XSDPositiveInteger = ls.NewTerm(XSD, "positiveInteger", false, false, ls.OverrideComposition, struct {
	SignedIntParser[int64]
}{
	SignedIntParser[int64]{
		min: &one64,
	},
}, "xsd:positiveInteger", "xs:positiveInteger")
var XSDNonPositiverInteger = ls.NewTerm(XSD, "nonPositiveInteger", false, false, ls.OverrideComposition, struct {
	SignedIntParser[int64]
}{
	SignedIntParser[int64]{
		max: &zero64,
	},
}, "xsd:nonPositiveInteger", "xs:nonPositiveInteger")
var XSDShort = ls.NewTerm(XSD, "short", false, false, ls.OverrideComposition, struct {
	SignedIntParser[int16]
}{
	SignedIntParser[int16]{},
}, "xsd:short", "xs:short")
var XSDUnsignedLong = ls.NewTerm(XSD, "unsignedLong", false, false, ls.OverrideComposition, struct {
	UnsignedIntParser[uint64]
}{
	UnsignedIntParser[uint64]{},
}, "xsd:unsignedLong", "xs:unsignedLong")
var XSDUnsignedInt = ls.NewTerm(XSD, "unsignedInt", false, false, ls.OverrideComposition, struct {
	UnsignedIntParser[uint32]
}{
	UnsignedIntParser[uint32]{},
}, "xsd:unsignedInt", "xs:unsignedInt")
var XSDUnsignedShort = ls.NewTerm(XSD, "unsignedShort", false, false, ls.OverrideComposition, struct {
	UnsignedIntParser[uint16]
}{
	UnsignedIntParser[uint16]{},
}, "xsd:unsignedShort", "xs:unsignedShort")
var XSDUnsignedByte = ls.NewTerm(XSD, "unsignedByte", false, false, ls.OverrideComposition, struct {
	UnsignedIntParser[uint8]
}{
	UnsignedIntParser[uint8]{},
}, "xsd:unsignedByte", "xs:unsignedByte")

type signedInt interface {
	int8 | int16 | int32 | int64
}

type unsignedInt interface {
	uint8 | uint16 | uint32 | uint64
}

type SignedIntParser[T signedInt] struct {
	min *int64
	max *int64
}

func (s SignedIntParser[T]) inBounds(val int64) bool {
	if s.min != nil && val < *s.min {
		return false
	}
	if s.max != nil && val > *s.max {
		return false
	}
	return true
}

type UnsignedIntParser[T unsignedInt] struct{}
type DecimalParser struct{}

func (SignedIntParser[T]) GetNativeValue(value string, node graph.Node) (interface{}, error) {
	var ret T
	v, err := strconv.ParseInt(value, 10, 64)
	ret = T(v)
	if int64(ret) != v {
		return nil, ErrOverflow{NodeID: ls.GetNodeID(node), Type: "signed numeric value", Value: value}
	}
	return ret, err
}

func (parser SignedIntParser[T]) FormatNativeValue(newValue, oldValue interface{}, node graph.Node) (string, error) {
	if newValue == nil {
		return "", nil
	}
	switch v := newValue.(type) {
	case int8:
		value := int64(v)
		if !parser.inBounds(value) {
			return "", ErrOverflow{NodeID: ls.GetNodeID(node), Type: "signed numeric value", Value: v}
		}
		return strconv.FormatInt(value, 10), nil
	case int16:
		value := int64(v)
		if !parser.inBounds(value) {
			return "", ErrOverflow{NodeID: ls.GetNodeID(node), Type: "signed numeric value", Value: v}
		}
		return strconv.FormatInt(value, 10), nil
	case int32:
		value := int64(v)
		if !parser.inBounds(value) {
			return "", ErrOverflow{NodeID: ls.GetNodeID(node), Type: "signed numeric value", Value: v}
		}
		return strconv.FormatInt(value, 10), nil
	case int64:
		value := int64(v)
		if !parser.inBounds(value) {
			return "", ErrOverflow{NodeID: ls.GetNodeID(node), Type: "signed numeric value", Value: v}
		}
		return strconv.FormatInt(value, 10), nil
	case int:
		value := int64(v)
		if !parser.inBounds(value) {
			return "", ErrOverflow{NodeID: ls.GetNodeID(node), Type: "signed numeric value", Value: v}
		}
		return strconv.FormatInt(value, 10), nil
	case uint8:
		value := int64(v)
		if !parser.inBounds(value) {
			return "", ErrOverflow{NodeID: ls.GetNodeID(node), Type: "signed numeric value", Value: v}
		}
		return strconv.FormatInt(value, 10), nil
	case uint16:
		value := int64(v)
		if !parser.inBounds(value) {
			return "", ErrOverflow{NodeID: ls.GetNodeID(node), Type: "signed numeric value", Value: v}
		}
		return strconv.FormatInt(value, 10), nil
	case uint32:
		value := int64(v)
		if !parser.inBounds(value) {
			return "", ErrOverflow{NodeID: ls.GetNodeID(node), Type: "signed numeric value", Value: v}
		}
		return strconv.FormatInt(value, 10), nil
	case uint64:
		value := int64(v)
		if (v < 0) != (value < 0) {
			return "", ErrOverflow{NodeID: ls.GetNodeID(node), Type: "Signed numeric", Value: newValue}
		}
		if !parser.inBounds(value) {
			return "", ErrOverflow{NodeID: ls.GetNodeID(node), Type: "signed numeric value", Value: v}
		}
		return strconv.FormatInt(value, 10), nil
	case uint:
		value := int64(v)
		if (v < 0) != (value < 0) {
			return "", ErrOverflow{NodeID: ls.GetNodeID(node), Type: "Signed numeric", Value: newValue}
		}
		if !parser.inBounds(value) {
			return "", ErrOverflow{NodeID: ls.GetNodeID(node), Type: "signed numeric value", Value: v}
		}
		return strconv.FormatInt(value, 10), nil

	case string:
		nativeValue, err := parser.GetNativeValue(v, node)
		if err != nil {
			return "", err
		}
		return parser.FormatNativeValue(nativeValue, oldValue, node)
	}
	return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: "Numeric term", Value: newValue}
}

func (UnsignedIntParser[T]) GetNativeValue(value string, node graph.Node) (interface{}, error) {
	var ret T
	v, err := strconv.ParseUint(value, 10, 64)
	ret = T(v)
	if uint64(ret) != v {
		return nil, ErrOverflow{NodeID: ls.GetNodeID(node), Type: "unsigned numeric value", Value: value}
	}
	return ret, err
}

func (parser UnsignedIntParser[T]) FormatNativeValue(newValue, oldValue interface{}, node graph.Node) (string, error) {
	if newValue == nil {
		return "", nil
	}
	switch v := newValue.(type) {
	case uint8:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case int8:
		if v < 0 {
			return "", ErrOverflow{NodeID: ls.GetNodeID(node), Type: "Unsigned numeric", Value: newValue}
		}
		return strconv.FormatUint(uint64(v), 10), nil
	case int16:
		if v < 0 {
			return "", ErrOverflow{NodeID: ls.GetNodeID(node), Type: "Unsigned numeric", Value: newValue}
		}
		return strconv.FormatUint(uint64(v), 10), nil
	case int32:
		if v < 0 {
			return "", ErrOverflow{NodeID: ls.GetNodeID(node), Type: "Unsigned numeric", Value: newValue}
		}
		return strconv.FormatUint(uint64(v), 10), nil
	case int64:
		if v < 0 {
			return "", ErrOverflow{NodeID: ls.GetNodeID(node), Type: "Unsigned numeric", Value: newValue}
		}
		return strconv.FormatUint(uint64(v), 10), nil
	case int:
		if v < 0 {
			return "", ErrOverflow{NodeID: ls.GetNodeID(node), Type: "Unsigned numeric", Value: newValue}
		}
		return strconv.FormatUint(uint64(v), 10), nil

	case float32:
		bigVal := uint64(v)
		var tval T
		tval = T(v)
		if uint64(tval) != bigVal {
			return "", ErrOverflow{NodeID: ls.GetNodeID(node), Type: "Unsigned numeric", Value: newValue}
		}
		return strconv.FormatUint(bigVal, 10), nil

	case float64:
		bigVal := uint64(v)
		var tval T
		tval = T(v)
		if uint64(tval) != bigVal {
			return "", ErrOverflow{NodeID: ls.GetNodeID(node), Type: "Unsigned numeric", Value: newValue}
		}
		return strconv.FormatUint(bigVal, 10), nil

	case string:
		nativeValue, err := parser.GetNativeValue(v, node)
		if err != nil {
			return "", err
		}
		return parser.FormatNativeValue(nativeValue, oldValue, node)
	}
	return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: "Unsigned numeric value", Value: newValue}
}

func (DecimalParser) GetNativeValue(value string, node graph.Node) (interface{}, error) {
	v, err := strconv.ParseFloat(value, 64)
	return v, err
}

func (parser DecimalParser) FormatNativeValue(newValue, oldValue interface{}, node graph.Node) (string, error) {
	if newValue == nil {
		return "", nil
	}
	switch v := newValue.(type) {
	case uint8:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case int8:
		return strconv.FormatInt(int64(v), 10), nil
	case int16:
		return strconv.FormatInt(int64(v), 10), nil
	case int32:
		return strconv.FormatInt(int64(v), 10), nil
	case int64:
		return strconv.FormatInt(int64(v), 10), nil
	case int:
		return strconv.FormatInt(int64(v), 10), nil

	case float32, float64:
		return fmt.Sprintf("%v", newValue), nil

	case string:
		nativeValue, err := parser.GetNativeValue(v, node)
		if err != nil {
			return "", err
		}
		return parser.FormatNativeValue(nativeValue, oldValue, node)
	}
	return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: "Decimal value", Value: newValue}
}

type JSONNumberParser struct{}

func (JSONNumberParser) GetNativeValue(value string, node graph.Node) (interface{}, error) {
	return json.Number(value), nil
}

func (parser JSONNumberParser) FormatNativeValue(newValue, oldValue interface{}, node graph.Node) (string, error) {
	if newValue == nil {
		return "", nil
	}
	switch v := newValue.(type) {
	case json.Number:
		return v.String(), nil
	case int8, int16, int32, int64, int, uint8, uint16, uint32, uint64, uint, float32, float64:
		return fmt.Sprint(newValue), nil
	case string:
		return v, nil
	}
	return "", ls.ErrInvalidValue{ID: ls.GetNodeID(node), Type: "JSON number", Value: newValue}
}
