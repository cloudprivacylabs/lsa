package ls

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/cloudprivacylabs/lpg"
)

// JSONMarshaler marshals/unmarshals a graph
type JSONMarshaler struct {
	lpg.JSON
}

func (JSONMarshaler) propertyUnmarshaler(key string, value json.RawMessage) (string, interface{}, error) {
	if key == NodeIDTerm {
		var s string
		if err := json.Unmarshal(value, &s); err != nil {
			return "", nil, err
		}
		return key, s, nil
	}
	// Must be string, or []string
	var str string
	var slice []string
	if err := json.Unmarshal(value, &str); err != nil {
		if err := json.Unmarshal(value, &slice); err != nil {
			return "", nil, fmt.Errorf("Value is not a string or []string")
		}
		return key, StringSlicePropertyValue(key, slice), nil
	}
	return key, StringPropertyValue(key, str), nil
}

func (JSONMarshaler) propertyMarshaler(key string, value interface{}) (string, json.RawMessage, error) {
	d, err := json.Marshal(value)
	return key, d, err
}

func NewJSONMarshaler(interner Interner) JSONMarshaler {
	ret := JSONMarshaler{
		JSON: lpg.JSON{
			Interner: interner,
		},
	}
	ret.PropertyUnmarshaler = ret.propertyUnmarshaler
	ret.PropertyMarshaler = ret.propertyMarshaler
	return ret
}

// Marshal marshals the graph as a JSON document
func (m JSONMarshaler) Marshal(g *lpg.Graph) ([]byte, error) {
	if m.PropertyMarshaler == nil {
		m.PropertyMarshaler = m.propertyMarshaler
	}
	buf := bytes.Buffer{}
	if err := m.Encode(g, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Unmarshal unmarshals a graph from JSON input
func (m JSONMarshaler) Unmarshal(in []byte, targetGraph *lpg.Graph) error {
	if m.PropertyUnmarshaler == nil {
		m.PropertyUnmarshaler = m.propertyUnmarshaler
	}
	dec := json.NewDecoder(bytes.NewReader(in))
	return m.Decode(targetGraph, dec)
}
