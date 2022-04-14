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

package pp

import (
	"github.com/cloudprivacylabs/lsa/pkg/ls"
)

// MeasureTerm is used in node labels. It marks a node's type as a
// Measure.
//
// A measure node should have a value and a unit. There are several
// ways the values and units are specified.
//
// A node may specify value and unit in its node value.
//
// Node:
//  :Measure
//  value: value and unit
//
//
// A node may specify value and unit separately in its properties:
//
// Node:
//  :Measure
//  value: 123
//  measureUnit: <unit>
//
// A node may be point to other nodes containing value or unit.
//
//  Node:                                  Node:
//   :Measure                              value: <unit>
//   value: 123                            schemaNodeId: A
//   measureUnitNode: A
//
//  Node:                     Node:               Node:
//   :Measure                 value: <unit>       value: <value>
//   measureUnitNode: A       schemaNodeId: A     schemaNodeId: B
//   measureValueNode: B
//
//
var MeasureTerm = ls.NewTerm(ls.LS, "Measure", false, false, ls.OverrideComposition, measureAccessors)

// MeasureUnitTerm is a node property term giving measure unit
var MeasureUnitTerm = ls.NewTerm(ls.LS, "measureUnit", false, false, ls.OverrideComposition, nil)

// MeasureUnitNodeTerm gives the schema node ID of the unit node under
// the Measure node
var MeasureUnitNodeTerm = ls.NewTerm(ls.LS, "measureUnitNode", false, false, ls.OverrideComposition, nil)

// MeasureValueNodeTerm gives the schema node ID of the value node
// under the Measure node
var MeasureValueNodeTerm = ls.NewTerm(ls.Ls, "measureValueNode", false, false, ls.OverrideComposition, nil)
