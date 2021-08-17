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
	//"github.com/cloudprivacylabs/lsa/pkg/ls"
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
	check("123", ctx, ValueOf(123))
	check(`"123"`, ctx, ValueOf("123"))
	check("null", ctx, ValueOf(nil))
	check("true", ctx, ValueOf(true))
	check("false", ctx, ValueOf(false))

	ctx.Set("abc", "123")
	check("abc", ctx, ValueOf("123"))
	check(`abc.length`, ctx, ValueOf(3))
	ctx.Set("arr", []string{"a", "b", "c"})
	check(`arr.has("a")`, ctx, ValueOf(true))
	check(`arr.has("d")`, ctx, ValueOf(false))
	check(`!arr`, ctx, ValueOf(false))
	check(`!false`, ctx, ValueOf(true))
	check(`abc=="123"`, ctx, ValueOf(true))
	check(`abc!="123"`, ctx, ValueOf(false))

	check(`x=1`, ctx, ValueOf(1))
	// i, err := ctx.Get("x").AsInt()
	// if err != nil {
	// 	t.Error(err)
	// }
	// if i != 1 {
	// 	t.Errorf("ctx: %v", ctx)
	// }
}

// func TestExprTokenizer(t *testing.T) {
// 	check := func(input string, err error, expected ...interface{}) {
// 		result, e := tokenizeExpression(input)
// 		if e != err {
// 			t.Errorf("Expected error=%v but got %v for input %s", err, e, input)
// 			return
// 		}
// 		if len(result) != len(expected) {
// 			t.Errorf("Expected %v, got %v for input %s", expected, result, input)
// 			return
// 		}
// 		for i := range result {
// 			if result[i] != expected[i] {
// 				t.Errorf("Expected %v, got %v for token %d in %s", expected[i], result[i], i, input)
// 			}
// 		}
// 	}
// 	check(" $var ", nil, symbolToken("$var"))
// 	check("var . to ", nil, symbolToken("var"), delimiterToken("."), symbolToken("to"))
// 	check("var.to ", nil, symbolToken("var"), delimiterToken("."), symbolToken("to"))
// 	check("var.to", nil, symbolToken("var"), delimiterToken("."), symbolToken("to"))
// 	check("var.to[x] ", nil, symbolToken("var"), delimiterToken("."), symbolToken("to"), delimiterToken("["), symbolToken("x"), delimiterToken("]"))
// 	check(`"asasd"[x.`, nil, stringToken("asasd"), delimiterToken("["), symbolToken("x"), delimiterToken("."))
// 	check(`"abc\"\\"`, nil, stringToken(`abc"\`))
// }

// func TestExprParser(t *testing.T) {
// 	check := func(input string, err error, expected interface{}) {
// 		expr, e := ParseExpression(input)
// 		if e != err {
// 			t.Errorf("Expected error=%v but got %v for input %s", err, e, input)
// 		}
// 		if !reflect.DeepEqual(expected, expr) {
// 			t.Errorf("Expected output: %+v, got %+v", expected, expr)
// 		}
// 	}
// 	check("$var", nil, SymbolExpression("$var"))
// 	check("$var.a", nil, FieldAccess{Main: SymbolExpression("$var"), Field: "a"})
// 	check("a[x]", nil, IndexAccess{Main: SymbolExpression("a"), Index: SymbolExpression("x")})
// 	check("a.b[c.d]", nil, IndexAccess{Main: FieldAccess{Main: SymbolExpression("a"), Field: "b"}, Index: FieldAccess{Main: SymbolExpression("c"), Field: "d"}})
// 	check(`a.b["http://blah"]`, nil, IndexAccess{Main: FieldAccess{Main: SymbolExpression("a"), Field: "b"}, Index: StringExpression("http://blah")})
// }

// func TestExprEval(t *testing.T) {
// 	check := func(input string, ctx *Context, expected interface{}) {
// 		expr, err := ParseExpression(input)
// 		if err != nil {
// 			t.Errorf("Parse error: %s", err)
// 			return
// 		}
// 		if err := expr.Evaluate(ctx); err != nil {
// 			t.Errorf("Evaluation error: %s", err)
// 		}
// 		value, err := ctx.Pop()
// 		if err != nil {
// 			t.Errorf("No result")
// 			return
// 		}
// 		if !reflect.DeepEqual(value, expected) {
// 			t.Errorf("Expected %+v got %+v with %s", expected, value, input)
// 		}
// 	}

// 	check("$a", NewContext(map[string]interface{}{"a": "1"}), "1")
// 	node := ls.NewNode("id", "t1", "t2")
// 	check("$node.id", NewContext(map[string]interface{}{"node": node}), "id")
// 	node.GetProperties()["a"] = ls.StringPropertyValue("b")
// 	check(`$node.properties["a"]`, NewContext(map[string]interface{}{"node": node}), "b")
// }

// func TestPredicateEval(t *testing.T) {

// }
