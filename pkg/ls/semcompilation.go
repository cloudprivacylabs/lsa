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
	"github.com/cloudprivacylabs/opencypher"
	"github.com/cloudprivacylabs/opencypher/graph"
)

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
	CompileNode(*Layer, graph.Node) error
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
	CompileEdge(*Layer, graph.Edge) error
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
	CompileTerm(CompilablePropertyContainer, string, *PropertyValue) error
}

type emptyCompiler struct{}

// CompileNode returns the value unmodified
func (emptyCompiler) CompileNode(*Layer, graph.Node) error { return nil }
func (emptyCompiler) CompileEdge(*Layer, graph.Edge) error { return nil }
func (emptyCompiler) CompileTerm(CompilablePropertyContainer, string, *PropertyValue) error {
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

func (CompileOCSemantics) CompileTerm(target CompilablePropertyContainer, term string, value *PropertyValue) error {
	if value == nil {
		return nil
	}
	expr := make([]opencypher.Evaluatable, 0)
	for _, str := range value.MustStringSlice() {
		e, err := opencypher.Parse(str)
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
