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
package ls

import (
	"errors"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
)

// ErrInvalidEncodingIndex is used to return error about an invalid or
// unrecognized character encoding
var ErrInvalidEncodingIndex = errors.New("Invalid encoding index")

// EncodingIndex determines the encoding name index
type EncodingIndex int

// Encoding indexes. If encoding name index is unknown, use
// UnknownEncodingIndex
const (
	UnknownEncodingIndex EncodingIndex = 0
	MIMEEncodingIndex    EncodingIndex = 1
	IANAEncodingIndex    EncodingIndex = 2
	MIBEncodingIndex     EncodingIndex = 3
)

// GetEncoding returns the encoding based on the index value. In index
// is UnknownEncodingIndex, it tries to resolve the name using IANA,
// MIME, and MIB indexes, in that order.
func (index EncodingIndex) Encoding(name string) (encoding.Encoding, error) {
	switch index {
	case UnknownEncodingIndex:
		e, _ := ianaindex.IANA.Encoding(name)
		if e != nil {
			return e, nil
		}
		e, _ = ianaindex.MIME.Encoding(name)
		if e != nil {
			return e, nil
		}
		return ianaindex.MIB.Encoding(name)
	case MIMEEncodingIndex:
		return ianaindex.MIME.Encoding(name)
	case IANAEncodingIndex:
		return ianaindex.IANA.Encoding(name)
	case MIBEncodingIndex:
		return ianaindex.MIB.Encoding(name)
	}
	return nil, ErrInvalidEncodingIndex
}
