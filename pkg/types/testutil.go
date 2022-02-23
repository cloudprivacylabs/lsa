package types

import (
	"fmt"
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

type getSetTestCase struct {
	name             string
	srcTypes         []string
	srcValue         interface{}
	targetTypes      []string
	expectedValue    interface{}
	expectGetError   bool
	expectSetError   bool
	srcProperties    map[string]interface{}
	targetProperties map[string]interface{}
}

func (tc getSetTestCase) run(t *testing.T) {
	g := graph.NewOCGraph()
	srcNode := g.NewNode(tc.srcTypes, tc.srcProperties)
	ls.SetRawNodeValue(srcNode, fmt.Sprint(tc.srcValue))
	targetNode := g.NewNode(tc.targetTypes, tc.targetProperties)
	v, err := ls.GetNodeValue(srcNode)
	if err != nil {
		t.Log(tc.name)
		if tc.expectGetError {
			return
		}
		t.Errorf("Unexpected get error in %+v: %s", tc, err)
		return
	}
	if err == nil && tc.expectGetError {
		t.Log(tc.name)
		t.Errorf("Expecting get error, got none in %+v", tc)
		return
	}
	err = ls.SetNodeValue(targetNode, v)
	if err != nil {
		t.Log(tc.name)
		if tc.expectSetError {
			return
		}
		t.Errorf("Unexpected set error in %+v: %s", tc, err)
		return
	}
	if err == nil && tc.expectSetError {
		t.Errorf("Expecting set error, got none in %+v", tc)
		return
	}
	if s, _ := ls.GetRawNodeValue(targetNode); s != tc.expectedValue {
		t.Log(tc.name)
		t.Errorf("Expecting %v got %v in %+v", tc.expectedValue, s, tc)
	}
}

func runGetSetTests(t *testing.T, cases []getSetTestCase) {
	for _, cs := range cases {
		cs.run(t)
	}
}
