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

// NodeCompiler interface represents term compilation algorithm when the term is a node
type NodeCompiler interface {
	// CompileNode gets a node and compiles the associated term on that
	// node. It should store the compiled state into node.Compiled with
	// the an opaque key
	CompileNode(*SchemaNode) error
}

// EdgeCompiler interface represents term compilation algorithm when the term is an edge
type EdgeCompiler interface {
	// CompileEdge gets an edge and compiles the associated term on that
	// edge. It should store tje compiled state into edge.Compiled with
	// an opaque key
	CompileEdge(*SchemaEdge) error
}

// TermCompiler interface represents term compilation algorithm
type TermCompiler interface {
	// CompileTerm gets a term and its value, and returns an object that
	// will be placed in the Compiled map of the node or the edge.
	CompileTerm(string, interface{}) (interface{}, error)
}

type emptyCompiler struct{}

// CompileNode returns the value unmodified
func (emptyCompiler) CompileNode(*SchemaNode) error                        { return nil }
func (emptyCompiler) CompileEdge(*SchemaEdge) error                        { return nil }
func (emptyCompiler) CompileTerm(string, interface{}) (interface{}, error) { return nil, nil }

// GetNodeCompiler return a compiler that will compile the value
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

// GetEdgeCompiler return a compiler that will compile the value
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

// GetTermCompiler return a compiler that will compile the value
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
