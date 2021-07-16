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
	CompileNode(LayerNode) error
}

// EdgeCompiler interface represents term compilation algorithm when
// the term is an edge
//
// During schema compilation, if the term metadata for the edge label
// implements EdgeCompiler, this method is called.
type EdgeCompiler interface {
	// CompileEdge gets an edge and compiles the associated term on that
	// edge. It should store tje compiled state into edge.Compiled with
	// an opaque key
	CompileEdge(LayerEdge) error
}

// TermCompiler interface represents term compilation algorithm. This
// is used to compile terms stored as node/edge properties. If the
// term metadata implements TermCompiler, this method is called, and
// the result is stored in the Compiled map of the node/edge with the
// term as the key.
type TermCompiler interface {
	// CompileTerm gets a term and its value, and returns an object that
	// will be placed in the Compiled map of the node or the edge.
	CompileTerm(string, *PropertyValue) (interface{}, error)
}

type emptyCompiler struct{}

// CompileNode returns the value unmodified
func (emptyCompiler) CompileNode(LayerNode) error                             { return nil }
func (emptyCompiler) CompileEdge(LayerEdge) error                             { return nil }
func (emptyCompiler) CompileTerm(string, *PropertyValue) (interface{}, error) { return nil, nil }

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
