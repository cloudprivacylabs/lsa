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
	"fmt"
	"strings"

	"github.com/cloudprivacylabs/lpg/v2"
	"github.com/cloudprivacylabs/opencypher"
)

// An instance of CompileContext is passed to all compile
// functions. It contains a statement cache to prevent recompiling
// same statemement multiple times
type CompileContext struct {
	statementCache map[any]any
}

// GetCompiledStatement returns the compiled statement if any
func (c *CompileContext) GetCompiledStatement(statement any) (any, bool) {
	if c.statementCache != nil {
		return nil, false
	}
	stmt, exists := c.statementCache[statement]
	return stmt, exists
}

// SetCompiledStatement sets a compiled statement
func (c *CompileContext) SetCompiledStatement(statement, compiled any) {
	if c.statementCache == nil {
		c.statementCache = make(map[any]any)
	}
	c.statementCache[statement] = compiled
}

func (c *CompileContext) CompileOpencypher(statement string) (opencypher.Evaluatable, error) {
	compiled, exists := c.GetCompiledStatement(statement)
	if exists {
		return compiled.(opencypher.Evaluatable), nil
	}
	compiled, err := opencypher.Parse(statement)
	if err != nil {
		return nil, err
	}
	c.SetCompiledStatement(statement, compiled)
	return compiled.(opencypher.Evaluatable), nil
}

// NodeCompiler interface represents term compilation algorithm when
// the term is a node.
//
// During schema compilation, if a node is found to be a semantic
// annotation node (i.e. not an attribute node), and if the term
// metadata for the node label implements NodeCompiler, this function
// is called to compile the node.
type NodeCompiler interface {
	// CompileNode gets a node and compiles the associated term on that
	// node. It should store the compiled state into node.Compiled with
	// the an opaque key
	CompileNode(*CompileContext, *Layer, *lpg.Node) error
}

// EdgeCompiler interface represents term compilation algorithm when
// the term is an edge
//
// During schema compilation, if the term metadata for the edge label
// implements EdgeCompiler, this method is called.
type EdgeCompiler interface {
	// CompileEdge gets an edge and compiles the associated term on that
	// edge. It should store the compiled state into edge.Compiled with
	// an opaque key
	CompileEdge(*CompileContext, *Layer, *lpg.Edge) error
}

// CompilablePropertyContainer contains properties and a compiled data map
type CompilablePropertyContainer interface {
	GetProperty(string) (interface{}, bool)
	SetProperty(string, interface{})
}

// TermCompiler interface represents term compilation algorithm. This
// is used to compile terms stored as node/edge properties. If the
// term metadata implements TermCompiler, this method is called, and
// the result is stored in the Compiled map of the node/edge with the
// term as the key.
type TermCompiler interface {
	// CompileTerm gets a node or edge, the term and its value, and
	// compiles it. It can store compilation data in the compiled data
	// map.
	CompileTerm(*CompileContext, CompilablePropertyContainer, string, *PropertyValue) error
}

type emptyCompiler struct{}

// CompileNode returns the value unmodified
func (emptyCompiler) CompileNode(*CompileContext, *Layer, *lpg.Node) error { return nil }
func (emptyCompiler) CompileEdge(*CompileContext, *Layer, *lpg.Edge) error { return nil }
func (emptyCompiler) CompileTerm(*CompileContext, CompilablePropertyContainer, string, *PropertyValue) error {
	return nil
}

// GetNodeCompiler return a compiler that will compile the term when
// the term is a node label
func GetNodeCompiler(term string) NodeCompiler {
	md := GetTermMetadata(term)
	if md == nil {
		return emptyCompiler{}
	}
	c, ok := md.(NodeCompiler)
	if ok {
		return c
	}
	return emptyCompiler{}
}

// GetEdgeCompiler return a compiler that will compile the term when
// the term is an edge label
func GetEdgeCompiler(term string) EdgeCompiler {
	md := GetTermMetadata(term)
	if md == nil {
		return emptyCompiler{}
	}
	c, ok := md.(EdgeCompiler)
	if ok {
		return c
	}
	return emptyCompiler{}
}

// GetTermCompiler return a compiler that will compile the term when
// the term is a node/edge property
func GetTermCompiler(term string) TermCompiler {
	md := GetTermMetadata(term)
	if md == nil {
		return emptyCompiler{}
	}
	c, ok := md.(TermCompiler)
	if ok {
		return c
	}
	return emptyCompiler{}
}

// CompiledProperties is a lazy-initialized map
type CompiledProperties struct {
	m map[interface{}]interface{}
}

func (p *CompiledProperties) GetCompiledProperty(key interface{}) (interface{}, bool) {
	if p.m == nil {
		return nil, false
	}
	property, exists := p.m[key]
	return property, exists
}

func (p *CompiledProperties) SetCompiledProperty(key, value interface{}) {
	if p.m == nil {
		p.m = make(map[interface{}]interface{})
	}
	p.m[key] = value
}

func (p *CompiledProperties) CopyCompiledToMap(target map[interface{}]interface{}) {
	if p.m == nil {
		return
	}
	for k, v := range p.m {
		target[k] = v
	}
}

func (p *CompiledProperties) CopyTo(target *CompiledProperties) {
	if p.m == nil {
		return
	}
	if target.m == nil {
		target.m = make(map[interface{}]interface{})
	}
	for k, v := range p.m {
		target.m[k] = v
	}
}

// CompileOCSemantics is a compilation implementation for terms
// containing a slice of opencypher expressions. It compiles one or
// more expressions of the term, and places them im $compiled_term
// property.
type CompileOCSemantics struct{}

func (CompileOCSemantics) CompileTerm(ctx *CompileContext, target CompilablePropertyContainer, term string, value *PropertyValue) error {
	if value == nil {
		return nil
	}
	expr := make([]opencypher.Evaluatable, 0)
	for _, str := range value.MustStringSlice() {
		e, err := ctx.CompileOpencypher(str)
		if err != nil {
			return err
		}
		expr = append(expr, e)
	}
	target.SetProperty("$compiled_"+term, expr)
	return nil
}

func (CompileOCSemantics) Compiled(target CompilablePropertyContainer, term string) []opencypher.Evaluatable {
	x, _ := target.GetProperty("$compiled_" + term)
	if x == nil {
		return nil
	}
	ret, _ := x.([]opencypher.Evaluatable)
	return ret
}

// Evaluate all compiled expressions and return nonempty resultsets.
func (c CompileOCSemantics) Evaluate(target CompilablePropertyContainer, term string, evalCtx *opencypher.EvalContext) ([]opencypher.ResultSet, error) {
	ret := make([]opencypher.ResultSet, 0)
	for _, expr := range c.Compiled(target, term) {
		v, err := expr.Evaluate(evalCtx)
		if err != nil {
			return nil, err
		}
		if v.Get() == nil {
			continue
		}
		rs, ok := v.Get().(opencypher.ResultSet)
		if !ok {
			continue
		}
		if len(rs.Rows) == 0 {
			continue
		}
		ret = append(ret, rs)
	}
	return ret, nil
}

// DialectValueSemantics is a compilation implementation for terms
// containing an expression, or value. The property is expected to be:
//
//	dialect: value
//
// For example:
//
//	opencypher:  expr
//
// is an expression
//
//	literal: value
//
// is a literal
//
// Anything else is error
type DialectValueSemantics struct{}

type DialectValue struct {
	Dialect string
	Value   any
}

func (DialectValueSemantics) CompileTerm(ctx *CompileContext, target CompilablePropertyContainer, term string, value *PropertyValue) error {
	if value == nil {
		return nil
	}
	result := make([]DialectValue, 0)
	for _, str := range value.MustStringSlice() {
		parts := strings.SplitN(str, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("Invalid dialect:value expression: %s", str)
		}
		parts[0] = strings.TrimSpace(parts[0])
		switch parts[0] {
		case "opencypher":
			e, err := ctx.CompileOpencypher(parts[1])
			if err != nil {
				return err
			}
			result = append(result, DialectValue{
				Dialect: "opencypher",
				Value:   e,
			})
		case "literal":
			result = append(result, DialectValue{
				Dialect: "literal",
				Value:   parts[1],
			})
		default:
			return fmt.Errorf("Unknown dialect in expression: %s", parts[0])
		}
	}
	target.SetProperty("$compiled_"+term, result)
	return nil
}

func (DialectValueSemantics) Compiled(target CompilablePropertyContainer, term string) []DialectValue {
	x, _ := target.GetProperty("$compiled_" + term)
	if x == nil {
		return nil
	}
	ret, _ := x.([]DialectValue)
	return ret
}
