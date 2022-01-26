package types

import (
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

type getSetTestCase struct {
	srcTypes       []string
	srcValue       interface{}
	targetTypes    []string
	expectedValue  interface{}
	expectGetError bool
	expectSetError bool
}

func (tc getSetTestCase) run(t *testing.T) {
	srcNode := ls.NewNode("idsrc", tc.srcTypes...)
	srcNode.SetValue(tc.srcValue)
	targetNode := ls.NewNode("idtarget", tc.targetTypes...)
	v, err := ls.GetNodeValue(srcNode)
	if err != nil {
		if tc.expectGetError {
			return
		}
		t.Errorf("Unexpected get error in %+v: %s", tc, err)
		return
	}
	if err == nil && tc.expectGetError {
		t.Errorf("Expecting get error, got none in %+v", tc)
		return
	}
	err = ls.SetNodeValue(targetNode, v)
	if err != nil {
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
	if tc.expectedValue != targetNode.GetValue() {
		t.Errorf("Expecting %v got %v in %+v", tc.expectedValue, targetNode.GetValue(), tc)
	}
}

func runGetSetTests(t *testing.T, cases []getSetTestCase) {
	for _, cs := range cases {
		cs.run(t)
	}
}
