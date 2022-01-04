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

// LinkField links an ingested document node to the node it is linked
// to in the graph.The `documentRoot` is the root node of the document
// containing this link. The `referenceSchemaNode` is the compiled
// schema node that has the reference annotations. This function
// locates the object referenced by the referenceSchemaNode in
// `graph`, using the values collected from `documentRoot`. This is
// done by getting the ID values from the `ReferenceIDValue`
// annotation of the schema node, and find an object in `graph` that
// has the type referenced, with `ReferenceIDField` values equal to
// `ReferenceIDValue` values.
func LinkField(documentRoot Node, referenceSchemaNode Node, graph *digraph.Graph) error {
	referenceIDField := referenceSchemaNode.GetProperties()[ReferenceIDFieldTerm]
	if referenceIDField == nil {
		return nil
	}
	referenceIDValue := referenceSchemaNode.GetProperties()[ReferenceIDValueTerm]
	if referenceIDValue == nil {
		return nil
	}
	var idFields []string
	if referenceIDField.IsString() {
		idFields = []string{referenceIDField.AsString()}
	} else {
		idFields = referenceIDField.AsStringSlice()
	}
	var idValueFields []string
	if referenceIDValue.IsString() {
		idValueFields = []string{referenceIDValue.AsString()}
	} else {
		idValueFields = referenceIDValue.AsStringSlice()
	}

	IterateDescendants(documentRoot, func(node Node, _ []Node) bool {
		instances := InstanceOf(node)
	}, func(edge Edge, _ []Node) EdgeFuncResult {
		if edge.GetLabel() == HasTerm {
			return FollowEdgeResult
		}
		return SkipEdgeResult
	}, false)
}
