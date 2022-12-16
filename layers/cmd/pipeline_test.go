package cmd

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/cloudprivacylabs/lsa/layers/cmd/pipeline"
	"github.com/joho/godotenv"
)

func TestPersonPipeline(t *testing.T) {
	steps, err := pipeline.ReadPipeline("testdata/ingest_person_pipeline.json")
	if err != nil {
		t.Error(err)
		return
	}
	oldTarget := ExportTarget
	var buf bytes.Buffer
	ExportTarget = &buf
	env, err := godotenv.Read(".env")
	if err != nil {
		t.Error(err)
		return
	}
	_, err = runPipeline(steps, env, "", []string{"testdata/person_sample.json"})
	ExportTarget = oldTarget
	if err != nil {
		t.Error(err)
	}
	var v interface{}
	d, err := ioutil.ReadFile("testdata/person_sample.json")
	if err != nil {
		panic(err)
	}
	var expected interface{}
	json.Unmarshal(d, &expected)
	json.Unmarshal(buf.Bytes(), &v)
	if !reflect.DeepEqual(v, expected) {
		t.Errorf("Got %v expected %v", v, expected)
		t.Logf("Expected: %s", string(d))
		t.Logf("Got: %s", buf.String())
	}
}
