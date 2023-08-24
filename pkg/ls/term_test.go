package ls

import (
	"encoding/json"
	"testing"
)

func TestJSONTerm(t *testing.T) {
	type metadata struct {
		JSONTermMetadata
	}

	type Msg struct {
		Value string `json:"val"`
	}

	md := metadata{}
	md.NewInstance = func() any { return &Msg{} }

	jt := RegisterJSONTerm(NewTerm("", "TestJSONTerm.term").SetMetadata(md))
	sch := `{
   "nodes": [
       {
          "n":0,
          "labels": [ "https://lschema.org/Schema"],
          "properties": {
             "https://lschema.org/nodeId": "id"
          },
          "edges": [
             {
                "to": 1,
                "label": "https://lschema.org/layer"
             }
          ]
       },
       {
           "n":1,
           "labels": [ "https://lschema.org/Attribute", "https://lschema.org/Value"],
           "properties": {
               "https://lschema.org/nodeId": "attr1",
               "TestJSONTerm.term": {
                   "val": "1"
               }
            }
       }
  ]
}`
	mr := NewJSONMarshaler(nil)
	gr := NewLayerGraph()
	err := mr.Unmarshal([]byte(sch), gr)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	layer := LayersFromGraph(gr)[0]
	err = CompileGraphNodeTerms(gr)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	nd := layer.GetSchemaRootNode()
	pv, _ := nd.GetProperty(jt.Name)
	if _, ok := pv.(PropertyValue).Value().(json.RawMessage); !ok {
		t.Errorf("Not json: %v", pv.(PropertyValue).Value())
	}
	x, _ := nd.GetProperty(jt.Name + "_compiled")
	v, ok := x.(*Msg)
	if !ok {
		t.Errorf("Wrong type: %v", x)
	}
	if v.Value != "1" {
		t.Errorf("Wrong data: %v", v)
	}
}
