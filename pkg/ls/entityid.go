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

// EntityIDTerm marks a field of an entity as the entity unique
// ID. Fields contained within an entity will get IDs relative to
// this ID.
//
// This is a marker term. The contents are ignored. Existance of
// entityId term on a field marks it as entity id.
var EntityIDTerm = NewTerm(LS+"entityId", false, false, OverrideComposition, nil)

// GetEntityIDNodes returns all the nodes under entityRoot that are marked with EntityIDTerm
func GetEntityIDNodes(entityRoot Node) []Node {
	ret := make([]Node, 0)
	IterateDescendants(entityRoot, func(node Node, _ []Node) bool {
		if _, exists := node.GetProperties()[EntityIDTerm]; exists {
			ret = append(ret, node)
			return true
		}
		for _, schemaNode := range InstanceOf(node) {
			if _, exists := schemaNode.GetProperties()[EntityIDTerm]; exists {
				ret = append(ret, node)
				return true
			}
		}
		return true
	}, SkipSchemaNodes, false)
	return ret
}
