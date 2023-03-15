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

	"github.com/cloudprivacylabs/lpg"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type ErrOverflow struct {
	NodeID string
	Type   string
	Value  interface{}
}

func (o ErrOverflow) Error() string {
	return fmt.Sprintf("Overflow: Node ID: %s  Type: %s Value: %v", o.NodeID, o.Type, o.Value)
}

var JSONNumber = ls.NewTerm(JSON, "number", "json:number").SetComposition(ls.OverrideComposition).SetMetadata(struct {
	JSONNumberParser
}{
	JSONNumberParser{},
}).Register()

var JSONInteger = ls.NewTerm(JSON, "integer", "json:integer").SetComposition(ls.OverrideComposition).SetMetadata(struct {
	SignedIntParser[int8]
}{
	SignedIntParser[int8]{},
}).Register()

var one64 = int64(1)
var zero64 = int64(-1)
var negativeOne64 = int64(-1)

var XSDByte = ls.NewTerm(XSD, "byte", "xsd:byte", "xs:byte").SetComposition(ls.OverrideComposition).SetMetadata(struct {
	SignedIntParser[int8]
}{
	SignedIntParser[int8]{},
}).Register()
var XSDInt = ls.NewTerm(XSD, "int", "xsd:int", "xs:int", "xsd:integer", "xs:integer", XSD+"integer").SetComposition(ls.OverrideComposition).SetMetadata(struct {
	SignedIntParser[int32]
}{
	SignedIntParser[int32]{},
}).Register()
var XSDLong = ls.NewTerm(XSD, "long", "xsd:long", "xs:long", "int", "integer", ls.LS+"int", ls.LS+"integer", "ls:int", "ls:integer").SetComposition(ls.OverrideComposition).SetMetadata(struct {
	SignedIntParser[int64]
}{
	SignedIntParser[int64]{},
}).Register()
var XSDDecimal = ls.NewTerm(XSD, "decimal", "xsd:decimal", "xs:decimal", "ls:float", ls.LS+"float", "ls:double", ls.LS+"double", "double", "float").SetComposition(ls.OverrideComposition).SetMetadata(struct {
	DecimalParser
}{
	DecimalParser{},
}).Register()

var XSDNegativeInteger = ls.NewTerm(XSD, "negativeInteger", "xsd:negativeInteger", "xs:negativeInteger").SetComposition(ls.OverrideComposition).SetMetadata(struct {
	SignedIntParser[int64]
}{
	SignedIntParser[int64]{
		max: &negativeOne64,
	},
}).Register()
var XSDNonNegativeInteger = ls.NewTerm(XSD, "nonNegativeInteger", "xsd:nonNegativeInteger", "xs:nonNegativeInteger").SetComposition(ls.OverrideComposition).SetMetadata(struct {
	SignedIntParser[int64]
}{
	SignedIntParser[int64]{
		min: &zero64,
	},
}).Register()
var XSDPositiveInteger = ls.NewTerm(XSD, "positiveInteger", "xsd:positiveInteger", "xs:positiveInteger").SetComposition(ls.OverrideComposition).SetMetadata(struct {
	SignedIntParser[int64]
}{
	SignedIntParser[int64]{
		min: &one64,
	},
}).Register()
var XSDNonPositiverInteger = ls.NewTerm(XSD, "nonPositiveInteger", "xsd:nonPositiveInteger", "xs:nonPositiveInteger").SetComposition(ls.OverrideComposition).SetMetadata(struct {
	SignedIntParser[int64]
}{
	SignedIntParser[int64]{
		max: &zero64,
	},
}).Register()
var XSDShort = ls.NewTerm(XSD, "short", "xsd:short", "xs:short").SetComposition(ls.OverrideComposition).SetMetadata(struct {
	SignedIntParser[int16]
}{
	SignedIntParser[int16]{},
}).Register()
var XSDUnsignedLong = ls.NewTerm(XSD, "unsignedLong", "xsd:unsignedLong", "xs:unsignedLong").SetComposition(ls.OverrideComposition).SetMetadata(struct {
	UnsignedIntParser[uint64]
}{
	UnsignedIntParser[uint64]{},
}).Register()
var XSDUnsignedInt = ls.NewTerm(XSD, "unsignedInt", "xsd:unsignedInt", "xs:unsignedInt").SetComposition(ls.OverrideComposition).SetMetadata(struct {
	UnsignedIntParser[uint32]
}{
	UnsignedIntParser[uint32]{},
}).Register()
var XSDUnsignedShort = ls.NewTerm(XSD, "unsignedShort", "xsd:unsignedShort", "xs:unsignedShort").SetComposition(ls.OverrideComposition).SetMetadata(struct {
	UnsignedIntParser[uint16]
}{
	UnsignedIntParser[uint16]{},
}).Register()
var XSDUnsignedByte = ls.NewTerm(XSD, "unsignedByte", "xsd:unsignedByte", "xs:unsignedByte").SetComposition(ls.OverrideComposition).SetMetadata(struct {
	UnsignedIntParser[uint8]
}{
	UnsignedIntParser[uint8]{},
}).Register()

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

func (SignedIntParser[T]) GetNativeValue(value string, node *lpg.Node) (interface{}, error) {
	var ret T
	v, err := strconv.ParseInt(value, 10, 64)
	ret = T(v)
	if int64(ret) != v {
		return nil, ErrOverflow{NodeID: ls.GetNodeID(node), Type: "signed numeric value", Value: value}
	}
	return ret, err
}

func (parser SignedIntParser[T]) FormatNativeValue(newValue, oldValue interface{}, node *lpg.Node) (string, error) {
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

func (UnsignedIntParser[T]) GetNativeValue(value string, node *lpg.Node) (interface{}, error) {
	var ret T
	v, err := strconv.ParseUint(value, 10, 64)
	ret = T(v)
	if uint64(ret) != v {
		return nil, ErrOverflow{NodeID: ls.GetNodeID(node), Type: "unsigned numeric value", Value: value}
	}
	return ret, err
}

func (parser UnsignedIntParser[T]) FormatNativeValue(newValue, oldValue interface{}, node *lpg.Node) (string, error) {
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

func (DecimalParser) GetNativeValue(value string, node *lpg.Node) (interface{}, error) {
	v, err := strconv.ParseFloat(value, 64)
	return v, err
}

func (parser DecimalParser) FormatNativeValue(newValue, oldValue interface{}, node *lpg.Node) (string, error) {
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

func (JSONNumberParser) GetNativeValue(value string, node *lpg.Node) (interface{}, error) {
	return json.Number(value), nil
}

func (parser JSONNumberParser) FormatNativeValue(newValue, oldValue interface{}, node *lpg.Node) (string, error) {
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
