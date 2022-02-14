package types

import (
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/cloudprivacylabs/lsa/pkg/opencypher/graph"
)

type getSetTestCase struct {
	name           string
	srcTypes       []string
	srcValue       interface{}
	targetTypes    []string
	expectedValue  interface{}
	expectGetError bool
	expectSetError bool
}

func (tc getSetTestCase) run(t *testing.T) {
	g := graph.NewOCGraph()
	srcNode := g.NewNode(tc.srcTypes, nil)
	ls.SetRawNodeValue(srcNode, tc.srcValue)
	targetNode := g.NewNode(tc.targetTypes, nil)
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
	if tc.expectedValue != ls.GetRawNodeValue(targetNode) {
		t.Log(tc.name)
		t.Errorf("Expecting %v got %v in %+v", tc.expectedValue, ls.GetRawNodeValue(targetNode), tc)
	}
}

func runGetSetTests(t *testing.T, cases []getSetTestCase) {
	for _, cs := range cases {
		cs.run(t)
	}
}
