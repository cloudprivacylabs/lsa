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
	check := func(input string, scope *Scope, expected Value) {
		expr, err := Parse(input)
		if err != nil {
			t.Errorf("%s: %s", input, err)
			return
		}
		value, err := expr.Evaluate(scope)
		if err != nil {
			t.Errorf("%s: %s", input, err)
		} else if !reflect.DeepEqual(value, expected) {
			t.Errorf("input: %s expected: %v got: %v", input, expected, value)
		}
	}
	scope := NewScope()
	check("123", scope, ValueOf(123))
	check(`"123"`, scope, ValueOf("123"))
	check("null", scope, ValueOf(nil))
	check("true", scope, ValueOf(true))
	check("false", scope, ValueOf(false))

	scope.Set("abc", "123")
	check("abc", scope, ValueOf("123"))
	check(`abc.length`, scope, ValueOf(3))
	// Linefeed is valid whitespace
	check(`abc.
  length`, scope, ValueOf(3))
	scope.Set("arr", []string{"a", "b", "c"})
	check(`arr.has("a")`, scope, ValueOf(true))
	check(`arr.has("d")`, scope, ValueOf(false))
	check(`!arr`, scope, ValueOf(false))
	check(`!false`, scope, ValueOf(true))
	check(`abc=="123"`, scope, ValueOf(true))
	check(`abc!="123"`, scope, ValueOf(false))
	check(`newvar:=abc=="123"`, scope, ValueOf(true))
	if !reflect.DeepEqual(scope.Get("newvar"), ValueOf(true)) {
		t.Errorf("Assignment error")
	}
	check(`x:=1`, scope, ValueOf(1))

	node1 := ls.NewNode("id1")
	node2 := ls.NewNode("id2")
	ls.Connect(node1, node2, "edgeLabel")
	scope.Set("node", ValueOf(node1))
	check("node.firstReachable(n->n.id=='id2').length", scope, ValueOf(1))
	check("node.firstReachable(n->n.id=='id2'&&n.id=='id2').length", scope, ValueOf(1))
	check("node.firstReachable(n->{n.id=='id2';}).length", scope, ValueOf(1))
	check("123;abc.length;", scope, ValueOf(3))
	check("a:=1;  { a=2; } a;", scope, ValueOf(2))
	check("a:=1;  { a:=2; } a;", scope, ValueOf(1))
	check("x:=1;  { y:=2; } x;", scope, ValueOf(1))
}
