package cmd

import (
	"encoding/json"
	"testing"

	"github.com/cloudprivacylabs/lsa/layers/cmd/cmdutil"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func TestPipeline(t *testing.T) {
	fileArgs := []string{
		"../../examples/contact/person.schema.json",
		"../../examples/contact/person-dpv.bundle.json",
		"../../examples/contact/person-dpv.overlay.json",
		"../../examples/contact/contact.schema.json",
		"../../examples/contact/contact-dpv.overlay.json",
	}
	if err := mockPipeline("../../examples/contact/pipeline.json", fileArgs); err != nil {
		t.Error(err)
	}
	t.Fail()
}

func mockPipeline(file string, args []string) error {
	type stepMarshal struct {
		Operation string          `json:"operation" yaml:"operation"`
		Step      json.RawMessage `json:"params" yaml:"params"`
	}
	var stepMarshals []stepMarshal
	err := cmdutil.ReadJSONOrYAML(file, &stepMarshals)
	if err != nil {
		failErr(err)
	}
	initialGraph, _ := pipelineCmd.Flags().GetString("initialGraph")
	pipeline := &PipelineContext{
		Graph:      ls.NewDocumentGraph(),
		Context:    ls.DefaultContext(),
		InputFiles: make([]string, 0),
		steps:      []Step{},
	}
	if initialGraph != "" {
		pipeline.Graph, err = cmdutil.ReadJSONGraph([]string{initialGraph}, nil)
		if err != nil {
			failErr(err)
		}
	}
	if len(args) > 0 {
		pipeline.InputFiles = args
	}
	const upDir string = "../../examples/contact/"
	for _, stage := range stepMarshals {
		step := operations[stage.Operation]()
		if step != nil {
			if err := json.Unmarshal(stage.Step, step); err != nil {
				failErr(err)
			}
			switch v := step.(type) {
			case *JSONIngester:
				tmp := v.BaseIngestParams.Bundle
				v.BaseIngestParams.Bundle = upDir + tmp
			case *CSVIngester:
				tmp := v.BaseIngestParams.Bundle
				v.BaseIngestParams.Bundle = upDir + tmp
			case *XMLIngester:
				tmp := v.BaseIngestParams.Bundle
				v.BaseIngestParams.Bundle = upDir + tmp
			}
			pipeline.steps = append(pipeline.steps, step)
		}
	}
	pipeline.currentStep--
	if err := pipeline.Next(); err != nil {
		failErr(err)
	}
	return nil
}
