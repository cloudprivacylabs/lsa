package cmd

import (
	"bytes"
	"testing"

	"github.com/cloudprivacylabs/lsa/examples"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)
	return output, err
}

func executeCommandC(root *cobra.Command, args ...string) (c *cobra.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}

func TestPipeline(t *testing.T) {
	js := examples.EmbededFiles
	var inputFiles []string
	files, err := js.ReadDir("contact")
	if err != nil {
		t.Log(err)
	}
	const upDir string = "../../examples/contact/"
	for _, f := range files {
		switch {
		case f.Name() == "person.schema.json":
			inputFiles = append(inputFiles, upDir+f.Name())
		case f.Name() == "person-dpv.bundle.json":
			inputFiles = append(inputFiles, upDir+f.Name())
		}
	}
	rootCmd := rootCmd
	// fileDir, err := rootFS.ReadDir("../examples/contact")
	// if err != nil {
	// 	t.Log(err)
	// 	t.Fatal(err)
	// }
	// t.Log(fileDir)

	t.Log(inputFiles)
	args := []string{"pipeline", "--file", "../../examples/contact/pipeline.json", "../../examples/contact/person.schema.json"}
	//args = append(args, inputFiles...)
	rootCmd.SetArgs(args)
	t.Log(args)

	// tt := []struct {
	// 	args []string
	// 	err  error
	// 	out  string
	// }{
	// 	{
	// 		args: []string{"person.schema.json"},
	// 		err:  nil,
	// 		out:  "{}",
	// 	},
	// }
	// for _, tc := range tt {
	// 	out, err := execute(t, pipe, tc.args...)
	// 	t.Log(out)
	// 	t.Log(err)
	// }
	pipelineCmd := pipelineCmd
	t.Log(pipelineCmd)
	// for _, c := range rootCmd.Commands() {
	// 	if c.Use == "pipeline" {
	pipelineCmd.SetArgs(inputFiles)
	pipelineCmd.ValidArgs = inputFiles
	pipelineCmd.Args = func(c *cobra.Command, args []string) error {
		args = inputFiles
		return nil
	}
	pipelineCmd.ParseFlags(inputFiles)

	// pipelineCmd.SetArgs([]string{"pipeline.json"})
	rootCmd.AddCommand(pipelineCmd)
	// if err := pipelineCmd.Execute(); err != nil {
	// 	t.Log(err)
	// 	t.Fatal(err)
	// }
	out, err := executeCommand(pipelineCmd, inputFiles...)
	if err != nil {
		t.Log(err)
		t.Fatal()
	}
	t.Log(out)
	assert.Equal(t, out, "{}")

	// b := bytes.NewBufferString("")
	// pipelineCmd.SetOut(b)
	//pipelineCmd.Execute()
	// out, err := ioutil.ReadAll(b)
	// t.Log(string(out))
	// output, err := executeCommand(c, "../../examples/contact/person.schema.json",
	// 	"../../examples/contact/person-dpv.bundle.json",
	// 	"../../examples/contact/person-dpv.overlay.json",
	// 	"../../examples/contact/contact.schema.json",
	// 	"../../examples/contact/contact-dpv.overlay.json")
	// t.Log(output)

	// t.Log(err)
	// c.SetArgs([]string{"../examples/contact/person.schema.json"})
	// if err := c.Execute(); err != nil {
	// 	t.Log(err)
	// 	t.Fatal(err)
	// }
	// }
	// }
	t.Fail()
}
