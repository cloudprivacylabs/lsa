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

import ()

// EntitySchemaTerm is inserted by the schema compilation to mark
// entity roots. It records the schema ID containing the entity
// definition.
var EntitySchemaTerm = StringTerm{
	Term: NewTerm(LS, "entitySchema").
		SetType(StringType{}).
		SetComposition(ErrorComposition).
		SetTags(SchemaElementTag).
		Register(),
}

// EntityIDFieldsTerm is a string or []string that lists the attribute IDs
// for entity ID. It is defined at the root node of a layer. All
// attribute IDs must refer to value nodes.
var EntityIDFieldsTerm = StringSliceTerm{
	Term: NewTerm(LS, "entityIdFields").SetComposition(OverrideComposition).SetTags(SchemaElementTag).Register(),
}

// EntityIDTerm is a string or []string that gives the unique ID of
// an entity. This is a node property at the root node of an entity
var EntityIDTerm = StringSliceTerm{
	Term: NewTerm(LS, "entityId").SetComposition(OverrideComposition).Register(),
}
