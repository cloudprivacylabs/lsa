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

// import (
// 	"github.com/bserdar/digraph"
// )

// /*

// Projection algorithm:

// First, an observation: Projection is straight-forward without
// arrays. A value of the source maps to a value of the target. When
// there are arrays involved, the scenario gets complicated, because we
// now have to have a way of deciding which array element we are
// operating on.

// Field mappings are done via 'projectTo' term. A 'projectTo' term
// contains a source, a target, and optional other attributes.

// Type mappings:

// Polymorphic fields cannot be mapped. Options of polymorphic field can
// be mapped.

// Composite field mappings are treated as object mappings. Unlisted mappings are invalid.

//   Value -> Value: The value of the source node is copied to the value of
//                   the target node
//   Value -> Array: If 'projectToIndex' is present, the field is projected to that
//                    array index. Otherwise, field values are accumulated as array
//                    elements in the target.

//   Object -> Object: All attributes under the object are mapped directly to the matching
//                     target fields.
//   Object -> Array: The array must be an object array. If 'projectToIndex' is present,
//                    the object is projected to that array index.  Otherwise, field values
//                    are accumulated as array elements in the target.

//   Array -> Value: Array must be a value array. Array is flattened with comma delimiter.
//                   Delimiter can be specified with `joinDelimiter'.
//   Array -> Array: Elements of the source are mapped to elements of the target.

// Algorithm sketch:

// Algorithm visits target schema nodes recursively.

// If target is object and source is object: recursively project all child nodes

// If target is value: source is either array or value array. Collect all source values and combine

// If target is array: source is value, object, or array.

//    If source is value or object, collect all source values and include as array
//    If source is array: For each element of source, create a target element and recursively project
// */

// var ProjectionTerms = struct {
// 	ProjectTo string
// }{
// 	ProjectTo: NewTerm(LS+"projection#to", false, false, OverrideComposition, nil),
// }

// type ProjectionSpec interface {
// 	GetFieldProjection(targetNodeID string) FieldProjectionSpec
// }

// type FieldProjectionSpec interface {
// }

// type projectionContext struct {
// 	seen           map[interface{}]struct{}
// 	projectionSpec ProjectionSpec
// }

// func (ctx *projectionContext) getProjection(targetNodeID string) FieldProjectionSpec {
// 	return ctx.projectionSpec.GetFieldProjection(targetNodeID)
// }

// func Project(targetGraph, sourceGraph *digraph.Graph, targetSchema *Layer, projection ProjectionSpec) error {
// 	return nil
// }

// func processTargetSchemaNodes(ctx projectionContext, node LayerNode) error {
// 	if _, ok := ctx.seen[node]; ok {
// 		return nil
// 	}
// 	ctx.seen[node] = struct{}{}

// 	projection := ctx.getProjection(node.GetID())
// 	switch {
// 	case node.HasType(AttributeTypes.Value):
// 		if projection == nil {
// 			if defaultValue, ok := node.GetProperties()[DefaultValueTerm]; ok {
// 				ctx.newValueNode(node.GetID(), defaultValue)
// 			}
// 		} else {
// 			ctx.newValueNode(node.GetID(), projection.getProjectedValue())
// 		}

// 	case node.HasType(AttributeTypes.Object):
// 	case node.HasType(AttributeTypes.Array):
// 	case node.HasType(AttributeTypes.Polymorphic):
// 	}
// 	return nil
// }
