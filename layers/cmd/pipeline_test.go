package cmd

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestPersonPipeline(t *testing.T) {
	steps, err := readPipeline("testdata/ingest_person_pipeline.json")
	if err != nil {
		t.Error(err)
		return
	}
	oldTarget := ExportTarget
	var buf bytes.Buffer
	ExportTarget = &buf
	_, err = runPipeline(steps, "", []string{"testdata/person_sample.json"}, false)
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
	}
}
