package layers

import (
	"encoding/json"
	"testing"

	"github.com/bserdar/digraph"
	"github.com/piprate/json-gold/ld"
)

func expandJsonld(t *testing.T, in string) []interface{} {
	proc := ld.NewJsonLdProcessor()
	var v interface{}
	if err := json.Unmarshal([]byte(in), &v); err != nil {
		t.Error(err)
		t.Fail()
	}
	ret, err := proc.Expand(v, nil)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	return ret
}

func TestUnmarshalJsonld(t *testing.T) {
	input := expandJsonld(t, `{
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
}`)

	g := &digraph.Graph{}
	node, err := UnmarshalExpandedAttribute(g, input[0].(map[string]interface{}))
	if err != nil {
		t.Error(err)
	}
	n := g.AllNodesWithLabel("attr1").All()[0]
	v, ok := n.Payload.(*ValueAttribute)
	if !ok {
		t.Errorf("Not a value")
	}
	if v.GetTypes()[0] != AttributeTypes.Value {
		t.Errorf("Not a value")
	}
	n = g.AllNodesWithLabel("attr3").All()[0]
	ref, ok := n.Payload.(*ReferenceAttribute)
	if !ok {
		t.Errorf("Not a reference")
	}
	if ref.GetReference() != "ref1" {
		t.Errorf("Wrong ref: %s", ref.GetReference())
	}
	edges := node.AllOutgoingEdgesWithLabel(TypeTerms.Attributes).All()
	if len(edges) != 3 {
		t.Errorf("Expected 3 got %d", len(edges))
	}
	found := false
	for _, x := range edges {
		if x.From() == node && x.To() == n {
			found = true
		}
	}
	if !found {
		t.Errorf("Not connected")
	}

	n2 := g.AllNodesWithLabel("attr2").All()[0]
	e := n2.AllOutgoingEdgesWithLabel("https://layeredschemas.org/attr/privacyClassification")
	if !e.HasNext() {
		t.Errorf("No privacy annotation")
	}
	aNode := e.Next().To()
	if aNode.Label() != "flg1" {
		t.Errorf("Wrong label: %v", aNode.Label())
	}
}
