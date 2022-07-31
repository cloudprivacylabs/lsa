// Copyright 2021 Cloud Privacy Labs, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/cloudprivacylabs/lsa/layers/cmd/pipeline"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/types"
)

type ucumUnitService struct {
	serviceURL string
}

func (u ucumUnitService) Parse(in string) (types.Measure, error) {
	query := url.Values{}
	query.Set("value", in)
	rsp, err := http.Get(u.serviceURL + "/unit?" + query.Encode())
	if err != nil {
		return types.Measure{}, err
	}
	defer rsp.Body.Close()
	data, _ := ioutil.ReadAll(rsp.Body)
	if rsp.StatusCode != 200 {
		return types.Measure{}, fmt.Errorf("%d: %s", rsp.StatusCode, string(data))
	}
	var ret types.Measure
	if err := json.Unmarshal(data, &ret); err != nil {
		return types.Measure{}, err
	}
	return ret, nil
}

type ucumResponse struct {
	Status string      `json:"status"`
	ToVal  json.Number `json:"toVal"`
	Msg    []string    `json:"msg"`
	ToUnit struct {
		Name   string `json:"name_"`
		CSCode string `json:"csCode_"`
	} `json:"toUnit"`
}

func (u ucumUnitService) Convert(measure types.Measure, targetUnit string, domain string) (types.Measure, error) {
	query := url.Values{}
	query.Set("value", measure.Value)
	query.Set("unit", measure.Unit)
	query.Set("output", targetUnit)
	rsp, err := http.Get(u.serviceURL + "/convert?" + query.Encode())
	if err != nil {
		return types.Measure{}, err
	}
	defer rsp.Body.Close()
	data, _ := ioutil.ReadAll(rsp.Body)
	if rsp.StatusCode != 200 {
		return types.Measure{}, fmt.Errorf("%d: %s", rsp.StatusCode, string(data))
	}
	var result ucumResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return types.Measure{}, err
	}
	if result.Status != "succeeded" {
		return types.Measure{}, fmt.Errorf("%+v", result)
	}
	ret := types.Measure{
		Value: string(result.ToVal),
		Unit:  result.ToUnit.CSCode,
	}
	return ret, nil

}

type MeasureStep struct {
	BaseIngestParams
	SchemaNodeIDs []string `json:"schemaNodeIds" yaml:"schemaNodeIds"`

	initialized bool
	layer       *ls.Layer
}

func (MeasureStep) Help() {
	fmt.Println(`Process measures
Create/validate/update measure nodes in a graph

operation: measures
params:
  schemaNodeIds:
  - id1
  - id2

  # Specify the schema the input graph was ingested with`)
	fmt.Println(baseIngestParamsHelp)
}

func (ms *MeasureStep) Run(pipeline *pipeline.PipelineContext) error {
	if !ms.initialized {
		if ms.IsEmptySchema() {
			ms.layer, _ = pipeline.Properties["layer"].(*ls.Layer)
		} else {
			var err error
			ms.layer, err = LoadSchemaFromFileOrRepo(pipeline.Context, ms.CompiledSchema, ms.Repo, ms.Schema, ms.Type, ms.Bundle)
			if err != nil {
				return err
			}
		}
		if ms.layer == nil {
			return fmt.Errorf("No schema")
		}
		ms.initialized = true
	}
	builder := ls.NewGraphBuilder(pipeline.GetGraphRW(), ls.GraphBuilderOptions{
		EmbedSchemaNodes: true,
	})

	pipeline.Context.GetLogger().Debug(map[string]interface{}{"pipeline": "measures"})
	if len(ms.SchemaNodeIDs) == 0 {
		if err := types.BuildMeasureNodesForLayer(pipeline.Context, builder, ms.layer); err != nil {
			return err
		}
	} else {
		for _, id := range ms.SchemaNodeIDs {
			attr := ms.layer.GetAttributeByID(id)
			if attr == nil {
				return fmt.Errorf("Cannot find attribute with id: %s", id)
			}
			if err := types.BuildMeasureNodes(pipeline.Context, builder, attr); err != nil {
				return err
			}
		}
	}
	return pipeline.Next()
}

func init() {
	rootCmd.AddCommand(measuresCmd)
	measuresCmd.Flags().String("input", "json", "Input graph format (json, jsonld)")
	measuresCmd.Flags().String("output", "json", "Output format, json, jsonld, or dot")
	measuresCmd.Flags().StringSlice("schemaNodeId", nil, "Process measure processing for instances of these schema nodes only")
	addSchemaFlags(measuresCmd.Flags())

	pipeline.RegisterPipelineStep("measures", func() pipeline.Step { return &MeasureStep{} })
}

var measuresCmd = &cobra.Command{
	Use:   "measures",
	Short: "Process measurements in a graph",
	Long: `Process the measurements in a graph.

`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		step := &MeasureStep{}
		step.fromCmd(cmd)
		p := []pipeline.Step{
			NewReadGraphStep(cmd),
			step,
			NewWriteGraphStep(cmd),
		}
		_, err := runPipeline(p, "", args)
		return err
	},
}
