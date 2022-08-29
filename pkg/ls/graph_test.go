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

package ls

import (
	"testing"

	"github.com/cloudprivacylabs/lpg"
)

func TestWalkNodesInEntity(t *testing.T) {
	g := lpg.NewGraph()
	n1 := g.NewNode([]string{"root", DocumentNodeTerm}, nil)
	n2 := g.NewNode([]string{"eroot", DocumentNodeTerm}, map[string]interface{}{
		EntitySchemaTerm: StringPropertyValue(EntitySchemaTerm, "schA"),
	})
	n3 := g.NewNode([]string{"3", DocumentNodeTerm}, nil)
	n4 := g.NewNode([]string{"4", DocumentNodeTerm}, nil)
	n5 := g.NewNode([]string{"5", DocumentNodeTerm}, nil)
	n6 := g.NewNode([]string{"root2", DocumentNodeTerm}, map[string]interface{}{
		EntitySchemaTerm: StringPropertyValue(EntitySchemaTerm, "schB"),
	})
	n7 := g.NewNode([]string{"7"}, nil)

	// 1 --> 2(schA) --> 3 --> 4
	//         |         |
	//         |         + --> 5
	//         |
	//         +------> 7 -->6 (schB)
	g.NewEdge(n1, n2, "", nil)
	g.NewEdge(n2, n3, "", nil)
	g.NewEdge(n3, n4, "", nil)
	g.NewEdge(n3, n5, "", nil)
	g.NewEdge(n2, n7, "", nil)
	g.NewEdge(n7, n6, "", nil)

	result := make([]*lpg.Node, 0)
	accumulate := func(n *lpg.Node) bool {
		result = append(result, n)
		return true
	}
	WalkNodesInEntity(n3, accumulate)
	for _, n := range []*lpg.Node{n3, n4, n5, n7, n2} {
		found := 0
		for _, x := range result {
			if x == n {
				found++
			}
		}
		if found != 1 {
			t.Errorf("Node %v found %d times", n, found)
		}
	}
}
