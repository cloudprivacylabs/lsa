package cmd

import (
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func TestPipeline(t *testing.T) {
	if err := mockPipeline("testdata/pipeline.json", []string{"testdata/person_sample.json"}); err != nil {
		t.Error(err)
	}
	t.Fail()
}

func mockPipeline(file string, args []string) error {
	steps, err := readPipeline(file)
	if err != nil {
		return err
	}
	pipeline := &PipelineContext{
		Graph:       ls.NewDocumentGraph(),
		Context:     ls.DefaultContext(),
		InputFiles:  args,
		steps:       steps,
		currentStep: -1,
	}
	return pipeline.Next()
}
