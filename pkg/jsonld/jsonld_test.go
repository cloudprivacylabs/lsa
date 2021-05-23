package jsonld

import (
	"encoding/json"
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/layers"
)

func TestUnmarshalJsonld(t *testing.T) {
	var input interface{}
	err := json.Unmarshal([]byte(`{
"@context": "../../schemas/ls.jsonld",
"@type":"Schema",
"attributes": [
 {
  "@id": "attr1",
  "@type": "Value"
 },
 {
  "@id":  "attr2",
  "@type": "Value",
  "privacyClassification": [
    {
      "@value": "flg1"
    }
  ]
 },
 {
  "@id": "attr3",
  "@type": "Reference",
  "reference": "ref1"
 }
]
}`), &input)
	if err != nil {
		t.Error(err)
		return
	}

	layer, err := UnmarshalLayer(input)
	if err != nil {
		t.Error(err)
	}
	n := layer.Graph.AllNodesWithLabel("attr1").All()[0]
	v, ok := n.Payload.(*layers.SchemaNode)
	if !ok {
		t.Errorf("Not a value")
	}
	if !v.HasType(layers.AttributeTypes.Value) {
		t.Errorf("Not a value")
	}
	n = layer.Graph.AllNodesWithLabel("attr3").All()[0]
	ref := n.Payload.(*layers.SchemaNode)
	if ref.Properties[layers.TypeTerms.Reference] != layers.IRI("ref1") {
		t.Errorf("Wrong ref: %v", ref.Properties)
	}
	edges := layer.RootNode.AllOutgoingEdgesWithLabel(layers.TypeTerms.Attributes).All()
	if len(edges) != 3 {
		t.Errorf("Expected 3 got %d", len(edges))
	}

	n2 := layer.Graph.AllNodesWithLabel("attr2").All()[0]
	if n2.Payload.(*layers.SchemaNode).Properties["https://layeredschemas.org/attr/privacyClassification"] != "flg1" {
		t.Errorf("Wrong label: %v", n2.Payload)
	}
}

func TestMarshalJsonld(t *testing.T) {
	var input interface{}
	err := json.Unmarshal([]byte(`{
"@context": "../../schemas/ls.jsonld",
"@type":"Schema",
"attributes": [
 {
  "@id": "attr1",
  "@type": "Value"
 },
 {
  "@id":  "attr2",
  "@type": "Value",
  "privacyClassification": [
    {
      "@value": "flg1"
    }
  ]
 },
 {
  "@id": "attr3",
  "@type": "Reference",
  "reference": "ref1"
 }
]
}`), &input)
	if err != nil {
		t.Error(err)
		return
	}

	layer, err := UnmarshalLayer(input)
	if err != nil {
		t.Error(err)
	}
	out := MarshalLayer(layer)
	x, _ := json.MarshalIndent(out, "", "")
	t.Log(string(x))
	expected := `[
{
"@id": "_:b0",
"@type": [
"https://layeredschemas.org/Attribute",
"https://layeredschemas.org/Schema",
"https://layeredschemas.org/Object"
],
"https://layeredschemas.org/Object#attributes": [
{
"@id": "attr1",
"@type": [
"https://layeredschemas.org/Attribute",
"https://layeredschemas.org/Value"
]
},
{
"@id": "attr2",
"@type": [
"https://layeredschemas.org/Attribute",
"https://layeredschemas.org/Value"
],
"https://layeredschemas.org/attr/privacyClassification": [
{
"@value": "flg1"
}
]
},
{
"@id": "attr3",
"@type": [
"https://layeredschemas.org/Attribute",
"https://layeredschemas.org/Reference"
],
"https://layeredschemas.org/Reference#reference": [
{
"@id": "ref1"
}
]
}
]
}
]`

	if string(x) != expected {
		t.Errorf("Got %s Expected: %s", string(x), expected)
	}
}
