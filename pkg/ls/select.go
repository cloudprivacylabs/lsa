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
	"encoding/json"

	"github.com/gobwas/glob"

	"github.com/bserdar/digraph"
)

// A NodePredicate determines if a node is selected or not.
type NodePredicate interface {
	EvaluateNode(digraph.Node) (bool, error)
}

// FalseNodePredicate returns false
type FalseNodePredicate struct{}

func (FalseNodePredicate) EvaluateNode(digraph.Node) (bool, error) { return false, nil }

func (FalseNodePredicate) MarshalJSON() ([]byte, error) {
	return json.Marshal(false)
}

// TrueNodePredicate returns true
type TrueNodePredicate struct{}

func (TrueNodePredicate) EvaluateNode(digraph.Node) (bool, error) { return true, nil }

func (TrueNodePredicate) MarshalJSON() ([]byte, error) {
	return json.Marshal(true)
}

type NodePredicates []NodePredicate

func (p *NodePredicates) UmarshalJSON(in []byte) error {
	var data []json.RawMessage
	if err := json.Unmarshal(in, &data); err != nil {
		return err
	}
	*p = make([]NodePredicate, 0, len(data))
	for _, x := range data {
		pred, err := UnmarshalNodePredicate(x)
		if err != nil {
			return err
		}
		*p = append(*p, pred)
	}
	return nil
}

// ANDNodePredicate combines multiple node predicates with a logical AND
type ANDNodePredicate struct {
	Options NodePredicates `json:"$and"`
}

// NewANDNodePredicate returns a predicate of the form AND: [options...]
func NewANDNodePredicate(options ...NodePredicate) ANDNodePredicate {
	return ANDNodePredicate{Options: options}
}

// EvaluateNode returns true only if all options of the predicate return true
func (p ANDNodePredicate) EvaluateNode(node digraph.Node) (bool, error) {
	for _, option := range p.Options {
		v, err := option.EvaluateNode(node)
		if err != nil {
			return false, err
		}
		if !v {
			return false, nil
		}
	}
	return true, nil
}

func unmarshalANDNodePredicate(in json.RawMessage) (NodePredicate, error) {
	var options NodePredicates
	if err := json.Unmarshal(in, &options); err != nil {
		return nil, err
	}
	return ANDNodePredicate{Options: options}, nil
}

// NodeIDPredicate selects nodes by ID
type NodeIDPredicate struct {
	ID string `json:"$id"`
}

// NewNodeIDPredicate returns a new node id predicate with the given ID
func NewNodeIDPredicate(id string) (NodeIDPredicate, error) {
	return NodeIDPredicate{ID: id}, nil
}

// EvaluateNode selects node based on the given id
func (p NodeIDPredicate) EvaluateNode(node digraph.Node) (bool, error) {
	s := node.Label()
	str, ok := s.(string)
	if !ok {
		return false, nil
	}
	return str == p.ID, nil
}

func unmarshalNodeIDPredicate(in json.RawMessage) (NodePredicate, error) {
	var id string
	if err := json.Unmarshal(in, &id); err != nil {
		return nil, err
	}
	return NodeIDPredicate{ID: id}, nil
}

// NodeIDGlobPredicate selects nodes by ID glob
type NodeIDGlobPredicate struct {
	IDGlob string `json:"$idglob"`

	compiled glob.Glob
}

// NewNodeIDGlobPredicate returns a new node id predicate with the given glob
func NewNodeIDGlobPredicate(idGlob string) (*NodeIDGlobPredicate, error) {
	ret := NodeIDGlobPredicate{IDGlob: idGlob}
	if err := ret.compile(); err != nil {
		return nil, err
	}
	return &ret, nil
}

func (p *NodeIDGlobPredicate) compile() error {
	var err error
	p.compiled, err = glob.Compile(p.IDGlob)
	return err
}

// EvaluateNode selects node based on the given glob
func (p *NodeIDGlobPredicate) EvaluateNode(node digraph.Node) (bool, error) {
	s := node.Label()
	str, ok := s.(string)
	if !ok {
		return false, nil
	}
	if p.compiled == nil {
		if err := p.compile(); err != nil {
			return false, err
		}
	}
	return p.compiled.Match(str), nil
}

func unmarshalNodeIDGlobPredicate(in json.RawMessage) (NodePredicate, error) {
	var id string
	if err := json.Unmarshal(in, &id); err != nil {
		return nil, err
	}
	return NewNodeIDGlobPredicate(id)
}

var predicateFactories = map[string]func(json.RawMessage) (NodePredicate, error){
	"$and":    unmarshalANDNodePredicate,
	"$id":     unmarshalNodeIDPredicate,
	"$idglob": unmarshalNodeIDGlobPredicate,
}

// UnmarshalNodePredicate unmarshals a node predicate from JSON
func UnmarshalNodePredicate(in []byte) (NodePredicate, error) {
	if string(in) == "false" {
		return FalseNodePredicate{}, nil
	}
	if string(in) == "true" {
		return TrueNodePredicate{}, nil
	}

	var input map[string]json.RawMessage
	if err := json.Unmarshal(in, &input); err != nil {
		return nil, err
	}
	if len(input) == 0 {
		return FalseNodePredicate{}, nil
	}
	if len(input) > 1 {
		return nil, ErrInvalidPredicate
	}
	for k, v := range input {
		factory, ok := predicateFactories[k]
		if !ok {
			return nil, ErrUnknownOperator{k}
		}
		value, err := factory(v)
		if err != nil {
			return nil, err
		}
		return value, nil
	}
	return nil, nil
}

// SelectNodes selects some nodes from the graph based on predicate
func SelectNodes(in *digraph.Graph, predicate NodePredicate) ([]digraph.Node, error) {
	ret := make([]digraph.Node, 0)
	for nodes := in.AllNodes(); nodes.HasNext(); {
		node := nodes.Next()
		include, err := predicate.EvaluateNode(node)
		if err != nil {
			return nil, err
		}
		if include {
			ret = append(ret, node)
		}
	}
	return ret, nil
}
