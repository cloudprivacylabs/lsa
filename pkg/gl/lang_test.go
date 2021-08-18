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
package gl

import (
	"reflect"
	"testing"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

func TestExprParse(t *testing.T) {
	check := func(input string, ctx *Context, expected Value) {
		expr, err := ParseExpression(input)
		if err != nil {
			t.Error(err)
			return
		}
		value, err := expr.Evaluate(ctx)
		if err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(value, expected) {
			t.Errorf("input: %s expected: %v got: %v", input, expected, value)
		}
	}
	ctx := NewContext()
	// check("123", ctx, ValueOf(123))
	// check(`"123"`, ctx, ValueOf("123"))
	// check("null", ctx, ValueOf(nil))
	// check("true", ctx, ValueOf(true))
	// check("false", ctx, ValueOf(false))

	// ctx.Set("abc", "123")
	// check("abc", ctx, ValueOf("123"))
	// check(`abc.length`, ctx, ValueOf(3))
	// ctx.Set("arr", []string{"a", "b", "c"})
	// check(`arr.has("a")`, ctx, ValueOf(true))
	// check(`arr.has("d")`, ctx, ValueOf(false))
	// check(`!arr`, ctx, ValueOf(false))
	// check(`!false`, ctx, ValueOf(true))
	// check(`abc=="123"`, ctx, ValueOf(true))
	// check(`abc!="123"`, ctx, ValueOf(false))

	node1 := ls.NewNode("id1")
	node2 := ls.NewNode("id2")
	node1.Connect(node2, "edgeLabel")
	ctx.Set("node", ValueOf(node1))
	check("node.firstReachable(n->(n.id=='id2')).length", ctx, ValueOf(1))

	//	check(`x=1`, ctx, ValueOf(1))
}
