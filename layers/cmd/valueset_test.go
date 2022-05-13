package cmd

import (
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// Match looks at req.
func TestMatch(t *testing.T) {
	tests := []ls.ValuesetLookupRequest{
		{KeyValues: map[string]string{"": "some_value"}},
		{KeyValues: map[string]string{"k": "some_value"}},
		{KeyValues: map[string]string{"x": "some_value", "y": "another_value"}},
	}

	for _, tt := range tests {
		vsv := ValuesetValue{
			Result:       "some_value",
			ResultValues: tt.KeyValues,
		}
		vslResp, err := vsv.Match(tt)
		if err != nil || vslResp == nil {
			t.Errorf("Match failed %v", err)
		}
		t.Log(vslResp)
	}
}
