package cmd

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/spf13/cobra"
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
	rootCmd := rootCmd
	// pipe := pipelineCmd
	// rootCmd.AddCommand(pipe)
	rootCmd.SetArgs([]string{"pipeline", "--file", "../../examples/contact/pipeline.json", "../../examples/contact/person.schema.json"})

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
	for _, c := range rootCmd.Commands() {
		if c.Use == "pipeline" {
			c.SetArgs([]string{"../../examples/contact/person.schema.json",
				"../../examples/contact/person-dpv.bundle.json",
				"../../examples/contact/person-dpv.overlay.json",
				"../../examples/contact/contact.schema.json",
				"../../examples/contact/contact-dpv.overlay.json",
			})
			t.Log(c)
			b := bytes.NewBufferString("")
			c.SetOut(b)
			c.Execute()
			out, err := ioutil.ReadAll(b)
			t.Log(string(out))
			// output, err := executeCommand(c, "../../examples/contact/person.schema.json",
			// 	"../../examples/contact/person-dpv.bundle.json",
			// 	"../../examples/contact/person-dpv.overlay.json",
			// 	"../../examples/contact/contact.schema.json",
			// 	"../../examples/contact/contact-dpv.overlay.json")
			// t.Log(output)

			t.Log(err)
			// c.SetArgs([]string{"../examples/contact/person.schema.json"})
			// if err := c.Execute(); err != nil {
			// 	t.Log(err)
			// 	t.Fatal(err)
			// }
		}
	}
	t.Fail()
}
