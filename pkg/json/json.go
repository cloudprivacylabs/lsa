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

package json

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// OM represents the JSON object model that preserves order
type OM interface {
	Encode(io.Writer) error
}

// KeyValue is a JSON key-value pair
type KeyValue struct {
	Key   string
	Value OM
}

// Value is a JSON value. It is one of nil, string, bool, json.Number
type Value struct {
	Value interface{}
}

// Encode a value
func (e Value) Encode(w io.Writer) error {
	data, err := json.Marshal(e.Value)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// Encode a key-value pair
func (e KeyValue) Encode(w io.Writer) error {
	data, err := json.Marshal(e.Key)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	if _, err := w.Write([]byte{':'}); err != nil {
		return err
	}
	return e.Value.Encode(w)
}

// Object represents a JSON object
type Object struct {
	Values []KeyValue
}

// Encode a json object
func (e Object) Encode(w io.Writer) error {
	if _, err := w.Write([]byte{'{'}); err != nil {
		return err
	}
	for i, v := range e.Values {
		if i > 0 {
			if _, err := w.Write([]byte{','}); err != nil {
				return err
			}
		}
		if err := v.Encode(w); err != nil {
			return err
		}
	}
	if _, err := w.Write([]byte{'}'}); err != nil {
		return err
	}
	return nil
}

// Array represents a JSON array
type Array struct {
	Elements []OM
}

func (e Array) Encode(w io.Writer) error {
	if _, err := w.Write([]byte{'['}); err != nil {
		return err
	}
	for i, value := range e.Elements {
		if i > 0 {
			if _, err := w.Write([]byte{','}); err != nil {
				return err
			}
		}
		if err := value.Encode(w); err != nil {
			return err
		}
	}
	if _, err := w.Write([]byte{']'}); err != nil {
		return err
	}
	return nil
}

func Decode(decoder *json.Decoder) (OM, error) {
	decoder.UseNumber()
	return decode(decoder, ls.NewInterner())
}

func decode(decoder *json.Decoder, interner ls.StringInterner) (OM, error) {
	var ret OM

	tok, err := decoder.Token()
	if err == io.EOF {
		return ret, nil
	}
	if err != nil {
		return nil, err
	}
	if delim, ok := tok.(json.Delim); ok {
		switch delim {
		case '{':
			ret, err = decodeObject(decoder, interner)
		case '[':
			ret, err = decodeArray(decoder, interner)
		default:
			err = &json.SyntaxError{Offset: decoder.InputOffset()}
		}
	} else {
		ret, err = decodeValue(tok)
	}
	return ret, err
}

func decodeObject(decoder *json.Decoder, interner ls.StringInterner) (Object, error) {
	ret := Object{}
	for {
		tok, err := decoder.Token()
		if err == io.EOF {
			return ret, &json.SyntaxError{Offset: decoder.InputOffset()}
		}
		if err != nil {
			return ret, err
		}

		if delim, ok := tok.(json.Delim); ok {
			if delim == '}' {
				break
			}
			return ret, &json.SyntaxError{Offset: decoder.InputOffset()}
		}

		key, ok := tok.(string)
		if !ok {
			return ret, &json.SyntaxError{Offset: decoder.InputOffset()}
		}
		key = interner.Intern(key)

		value, err := decode(decoder, interner)
		if err != nil {
			return ret, err
		}
		ret.Values = append(ret.Values, KeyValue{Key: key, Value: value})
	}
	return ret, nil
}

func decodeElement(decoder *json.Decoder, interner ls.StringInterner) (OM, bool, error) {
	var ret OM

	tok, err := decoder.Token()
	if err == io.EOF {
		return ret, false, &json.SyntaxError{Offset: decoder.InputOffset()}
	}
	if err != nil {
		return nil, false, err
	}
	if delim, ok := tok.(json.Delim); ok {
		switch delim {
		case '{':
			ret, err = decodeObject(decoder, interner)
		case '[':
			ret, err = decodeArray(decoder, interner)
		case ']':
			return ret, true, nil
		default:
			err = &json.SyntaxError{Offset: decoder.InputOffset()}
		}
	} else {
		ret, err = decodeValue(tok)
	}
	return ret, false, err
}

func decodeArray(decoder *json.Decoder, interner ls.StringInterner) (Array, error) {
	ret := Array{}
	for {
		value, done, err := decodeElement(decoder, interner)
		if err != nil {
			return ret, err
		}
		if done {
			break
		}
		ret.Elements = append(ret.Elements, value)
	}
	return ret, nil
}

func decodeValue(tok json.Token) (Value, error) {
	ret := Value{}
	switch val := tok.(type) {
	case bool:
		ret.Value = val
	case float64:
		ret.Value = json.Number([]byte(fmt.Sprint(val)))
	case json.Number:
		ret.Value = val
	case string:
		ret.Value = val
	case nil:
		ret.Value = nil
	}
	return ret, nil
}
