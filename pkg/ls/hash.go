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
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"strings"

	"github.com/cloudprivacylabs/lpg/v2"
)

var HashSHA256Term = StringSliceTerm{NewTerm(LS, "hash.sha256").SetComposition(OverrideComposition).SetMetadata(hashSemantics{}).SetTags(SchemaElementTag).Register()}
var HashSHA1Term = StringSliceTerm{NewTerm(LS, "hash.sha1").SetComposition(OverrideComposition).SetMetadata(hashSemantics{}).SetTags(SchemaElementTag).Register()}
var HashSHA512Term = StringSliceTerm{NewTerm(LS, "hash.sha512").SetComposition(OverrideComposition).SetMetadata(hashSemantics{}).SetTags(SchemaElementTag).Register()}
var HashTerm = StringSliceTerm{NewTerm(LS, "hash").SetComposition(OverrideComposition).SetMetadata(hashSemantics{}).SetTags(SchemaElementTag).Register()}

type hashSemantics struct{}

// ProcessNodePostDocIngest will search for nodes that are instances of
// the schema node ids given in the docnode, get a hash of those, and
// populate this node value with that hash
func (hashSemantics) ProcessNodePostDocIngest(schemaRootNode, schemaNode *lpg.Node, term PropertyValue, docNode *lpg.Node) error {
	termName := term.Sem().Name
	ix := strings.IndexRune(termName[len(LS):], '.')
	hashFunc := "sha256"
	if ix != -1 {
		hashFunc = termName[ix+1:]
	}
	entityRoot := GetEntityRoot(docNode)
	nodes := term.AsStringSlice()
	schemaNodes := make(map[string]*lpg.Node)
	IterateDescendants(schemaRootNode, func(node *lpg.Node) bool {
		id := GetNodeID(node)
		for _, n := range nodes {
			if id == n {
				schemaNodes[id] = node
				break
			}
		}
		return true
	}, SkipDocumentNodes, false)
	refs := make([]AttributeReference, len(nodes))
	for i := range nodes {
		sch := schemaNodes[nodes[i]]
		if sch != nil {
			ref, exists := GetAttributeReferenceBySchemaNode(schemaRootNode, sch, entityRoot)
			if exists {
				refs[i] = ref
			}
		}
	}
	collected := make(map[string][]string)
	IterateDescendants(entityRoot, func(node *lpg.Node) bool {
		instance := SchemaNodeIDTerm.PropertyValue(node)
		for _, ref := range refs {
			if ref.Node != nil && SchemaNodeIDTerm.PropertyValue(ref.Node) == instance {
				if ref.IsProperty() {
					p, ok := ref.AsPropertyValue()
					if ok {
						s, _ := p.Value().(string)
						collected[instance] = append(collected[instance], s)
					}
				} else {
					v, ok := GetRawNodeValue(node)
					if ok {
						collected[instance] = append(collected[instance], v)
					}
					break
				}
			}
		}
		return true
	}, FollowEdgesInEntity, false)
	var h hash.Hash
	switch hashFunc {
	case "sha256", "":
		h = sha256.New()
	case "sha1":
		h = sha1.New()
	case "sha512":
		h = sha512.New()
	default:
		return fmt.Errorf("Unknown hash function: %s", hashFunc)
	}
	for i, x := range refs {
		var key string
		if x.IsProperty() {
			key = SchemaNodeIDTerm.PropertyValue(x.Node)
		} else {
			key = nodes[i]
		}
		for _, y := range collected[key] {
			h.Write([]byte(y))
		}
	}
	value := fmt.Sprintf("%x", h.Sum(nil))
	SetRawNodeValue(docNode, value)
	SetEntityIDVectorElementFromNode(docNode, value)
	return nil
}
